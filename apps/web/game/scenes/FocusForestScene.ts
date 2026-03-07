import * as Phaser from 'phaser';
import eventBus from '../events/EventBus';
import { initSession, getSessionManifest } from '@/lib/api';
import {
  getSpeedForDifficulty,
  getSpawnIntervalForDifficulty,
  getTargetColour
} from '../logic/spawnPatterns';

interface FocusForestAction {
  type: string;
  tap_x: number;
  tap_y: number;
  client_timestamp: number;
}

const DESPAWN_AFTER_MS = 3000;
const SESSION_DURATION_MS = 60000;

export default class FocusForestScene extends Phaser.Scene {
  private sessionToken: string | null = null;
  private manifest: any[] = [];
  private actionLog: FocusForestAction[] = [];
  
  private startTime: number = 0;
  private timeRemainingMs: number = SESSION_DURATION_MS;
  private difficultyLevel: number = 1;
  private butterfliesHit: number = 0;
  
  private spawnEvent: Phaser.Time.TimerEvent | null = null;
  private tickEvent: Phaser.Time.TimerEvent | null = null;
  private activeTargets: Map<string, Phaser.GameObjects.Image> = new Map();

  constructor() {
    super('FocusForestScene');
  }

  async create() {
    this.cameras.main.setBackgroundColor('#2A3B4C'); // Foresty background
    
    try {
      // 1. Init session
      const { session_token, seed, difficulty_level } = await initSession('focus_forest');
      this.sessionToken = session_token;
      this.difficultyLevel = difficulty_level || 1;

      // 2. Fetch manifest
      this.manifest = await getSessionManifest(this.sessionToken);

      // 3. Setup game loop
      this.actionLog = [];
      this.timeRemainingMs = SESSION_DURATION_MS;
      this.startTime = this.time.now;

      const spawnInterval = getSpawnIntervalForDifficulty(this.difficultyLevel);

      // Spawn loop
      this.spawnEvent = this.time.addEvent({
        delay: spawnInterval,
        callback: this.spawnNextTarget,
        callbackScope: this,
        loop: true,
      });

      // Session UI tick loop
      this.tickEvent = this.time.addEvent({
        delay: 1000,
        callback: () => {
          this.timeRemainingMs -= 1000;
          eventBus.emit('game:ui-update', {
            timeRemainingMs: Math.max(0, this.timeRemainingMs),
            butterfliesHit: this.butterfliesHit
          });
          
          if (this.timeRemainingMs <= 0) {
            this.endSession();
          }
        },
        callbackScope: this,
        loop: true
      });

      // Tap handling
      this.input.on('pointerdown', this.handleTap, this);

    } catch (err) {
      console.error('Failed to init Focus Forest session', err);
      // Depending on app state, might want to emit error
    }
  }

  private spawnNextTarget() {
    if (this.manifest.length === 0) return;
    
    // Dequeue next target
    const target = this.manifest.shift();
    
    // Fallback if dimensions aren't reliable yet
    const width = this.scale.width || 800;
    const height = this.scale.height || 600;
    
    const startX = target.position_x * width;
    const startY = target.position_y * height;
    
    const color = getTargetColour(target.target_type);
    
    // Create circle using Graphics then generate texture (or use generic image)
    const graphics = this.add.graphics();
    graphics.fillStyle(color, 1);
    graphics.fillCircle(20, 20, 20); // 40x40
    graphics.generateTexture(target.target_id, 40, 40);
    graphics.destroy();

    const img = this.add.image(startX, startY, target.target_id);
    img.setInteractive();
    // Save target info onto the sprite for interaction parsing
    img.setData('target_id', target.target_id);
    img.setData('target_type', target.target_type);
    
    this.activeTargets.set(target.target_id, img);

    // Movement tween
    const speedPxPerS = getSpeedForDifficulty(this.difficultyLevel);
    const distanceToTravel = width + 100 - startX; // Move past right edge
    const duration = (distanceToTravel / speedPxPerS) * 1000;

    this.tweens.add({
      targets: img,
      x: width + 100,
      duration: duration,
      ease: 'Linear'
    });

    // Despawn timeout
    this.time.delayedCall(DESPAWN_AFTER_MS, () => {
      this.destroyTarget(target.target_id);
    });
  }

  private destroyTarget(id: string) {
    const img = this.activeTargets.get(id);
    if (img) {
      img.destroy();
      this.activeTargets.delete(id);
      this.textures.remove(id); // Cleanup memory
    }
  }

  private handleTap(pointer: Phaser.Input.Pointer) {
    if (this.timeRemainingMs <= 0) return;

    const tap_x = pointer.x / this.scale.width;
    const tap_y = pointer.y / this.scale.height;

    this.actionLog.push({
      type: 'tap',
      tap_x,
      tap_y,
      client_timestamp: this.time.now
    });

    // Check hit intersections
    // Note: since targets are constantly moving, simple distance check vs visual position
    let hitFound = false;
    for (const [id, img] of this.activeTargets.entries()) {
      const dist = Phaser.Math.Distance.Between(pointer.x, pointer.y, img.x, img.y);
      if (dist < 40) { // 40px hit radius
        const targetType = img.getData('target_type');
        hitFound = true;
        
        if (targetType.startsWith('butterfly')) {
          this.butterfliesHit++;
          // Particle burst (simulated by scale/fade tween for now)
          this.tweens.add({
            targets: img,
            scale: 1.5,
            alpha: 0,
            duration: 300,
            onComplete: () => this.destroyTarget(id)
          });
        } else if (targetType === 'bee') {
          // Shake tween
          this.tweens.add({
            targets: img,
            x: img.x + 10,
            yoyo: true,
            repeat: 3,
            duration: 50,
            onComplete: () => {
              this.tweens.add({
                targets: img,
                alpha: 0,
                duration: 200,
                onComplete: () => this.destroyTarget(id)
              });
            }
          });
        }
        break; // Only hit one target per tap
      }
    }

    if (!hitFound) {
      // Miss ripple feedback
      const ripple = this.add.circle(pointer.x, pointer.y, 5, 0xffffff, 0.5);
      this.tweens.add({
        targets: ripple,
        radius: 30,
        alpha: 0,
        duration: 400,
        onComplete: () => ripple.destroy()
      });
    }
  }

  private endSession() {
    if (this.spawnEvent) this.spawnEvent.remove();
    if (this.tickEvent) this.tickEvent.remove();

    for (const id of this.activeTargets.keys()) {
      this.destroyTarget(id);
    }

    eventBus.emit('game:session-end', {
      game_type: 'focus_forest',
      session_token: this.sessionToken,
      actions: this.actionLog,
      duration_ms: SESSION_DURATION_MS,
      butterfliesHit: this.butterfliesHit
    });
  }
}
