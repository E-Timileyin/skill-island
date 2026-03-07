import * as Phaser from 'phaser';
import eventBus from '../events/EventBus';
import { getBlockColour, getBlockDimensions, calculateCoM } from '../logic/teamTower';

const LERP_FACTOR = 0.25;
const HEARTBEAT_INTERVAL_MS = 5000;
const BLOCK_HEIGHT_PX = 40;

interface TeamTowerAction {
  type: string;
  position_x: number;
  client_timestamp: number;
}

interface BlockData {
  id: string;
  x: number;
  y: number;
  colour: string;
  shape: string;
}

interface TowerState {
  blocks: BlockData[];
  stable: boolean;
  current_height: number;
  target_height: number;
  active_player: string;
  turn_number: number;
  next_block_shape: string;
  group_xp: number;
}

export default class TeamTowerScene extends Phaser.Scene {
  private playerRole: "player_1" | "player_2" | null = null;
  private myTurn: boolean = false;
  private currentState: TowerState | null = null;
  private blockObjects: Map<string, Phaser.GameObjects.Rectangle> = new Map();
  private ws: WebSocket | null = null;
  private heartbeatInterval: any = null;
  private actionLog: TeamTowerAction[] = [];
  private sessionStartTime: number = 0;
  
  private dropGuide!: Phaser.GameObjects.Rectangle;
  private previewBlock: Phaser.GameObjects.Rectangle | null = null;
  private isReconnecting = false;
  private reconnectAttempts = 0;

  constructor() {
    super('TeamTowerScene');
  }

  create() {
    this.cameras.main.setBackgroundColor('#87CEEB'); // Sky blue for tower
    
    // Draw base platform
    const width = this.scale.width;
    const height = this.scale.height;
    this.add.rectangle(width / 2, height - 10, width, 20, 0x2c3e50);

    // Drop guide
    this.dropGuide = this.add.rectangle(width / 2, height / 2, 80, height, 0xbdc3c7, 0.2);
    this.dropGuide.setVisible(false);
    
    this.connectWebSocket();
    this.sessionStartTime = Date.now();

    this.input.on('pointermove', this.handlePointerMove, this);
    this.input.on('pointerdown', this.handlePointerDown, this);
  }

  update() {
    // Lerp blocks
    if (this.currentState && this.currentState.stable) {
      for (const block of this.currentState.blocks) {
        const obj = this.blockObjects.get(block.id);
        if (obj) {
          const dims = getBlockDimensions(block.shape);
          const targetX = block.x * this.scale.width;
          const targetY = this.scale.height - (block.y * BLOCK_HEIGHT_PX) - 20 - (dims.h / 2); // 20 is base platform

          obj.x += (targetX - obj.x) * LERP_FACTOR;
          obj.y += (targetY - obj.y) * LERP_FACTOR;
        }
      }
    }
  }

  private connectWebSocket() {
    const wsHost = process.env.NEXT_PUBLIC_WS_URL || "ws://localhost:8080";
    this.ws = new WebSocket(`${wsHost}/ws/game`);
    
    this.ws.onopen = () => {
      console.log('WS connected');
      this.isReconnecting = false;
      this.reconnectAttempts = 0;
      this.ws?.send(JSON.stringify({ type: "join_room", game_type: "team_tower" }));
      eventBus.emit('game:ui-update', { phase: "waiting" });
      
      this.heartbeatInterval = setInterval(() => {
        if (this.ws?.readyState === WebSocket.OPEN) {
          this.ws.send(JSON.stringify({ type: "heartbeat_ping", timestamp: Date.now() }));
        }
      }, HEARTBEAT_INTERVAL_MS);
    };

    this.ws.onmessage = (event) => this.handleServerMessage(event);
    this.ws.onclose = () => this.handleDisconnect();
    this.ws.onerror = (err) => console.error("WS Error:", err);
  }

  private handleServerMessage(event: MessageEvent) {
    try {
      const msg = JSON.parse(event.data);
      switch (msg.type) {
        case "waiting_for_partner":
          eventBus.emit('game:ui-update', { phase: "waiting" });
          break;
        case "room_ready":
          this.playerRole = msg.player_role;
          eventBus.emit('game:ui-update', { phase: "ready", playerRole: this.playerRole, opponentAvatar: msg.opponent_avatar });
          setTimeout(() => {
            eventBus.emit('game:ui-update', { phase: "playing" });
          }, 1500);
          break;
        case "state_update":
          // If we were reconnecting, tell UI we are back
          if (this.isReconnecting) {
             this.isReconnecting = false;
             eventBus.emit('game:ui-update', { phase: "partner_reconnected" });
          }
          let stateRaw = msg.game_state;
          if (typeof stateRaw === 'string') {
            stateRaw = JSON.parse(stateRaw);
          }
          this.reconcileState(stateRaw as TowerState);
          break;
        case "action_rejected":
          if (msg.reason === "not_your_turn") {
            // Screen shake or UI reaction
            this.cameras.main.shake(150, 0.005);
          } else if (msg.reason === "out_of_bounds") {
            this.cameras.main.shake(150, 0.01);
          }
          break;
        case "player_disconnected":
          eventBus.emit('game:ui-update', { phase: "partner_disconnected" });
          break;
        case "idle_warning":
          eventBus.emit('game:ui-update', { phase: "idle_warning", secondsRemaining: msg.seconds_remaining });
          break;
        case "session_end":
          if (this.heartbeatInterval) clearInterval(this.heartbeatInterval);
          this.ws?.close();
          // Provide final UI broadcast to be safe
          let endStateRaw = msg.final_state;
          if (typeof endStateRaw === 'string') {
             endStateRaw = JSON.parse(endStateRaw);
          }
          if (endStateRaw) this.reconcileState(endStateRaw as TowerState);

          eventBus.emit('game:session-end', {
            game_type: "team_tower",
            outcome: msg.outcome,
            stars: msg.stars,
            group_xp: msg.group_xp,
            room_session_id: msg.room_session_id,
            duration_ms: Date.now() - this.sessionStartTime
          });
          break;
      }
    } catch (e) {
      console.error("Failed to parse WS message", e);
    }
  }

  private handleDisconnect() {
    if (this.heartbeatInterval) clearInterval(this.heartbeatInterval);
    
    // Attempt reconnect if we were playing or paused
    if (!this.currentState || this.currentState.stable === false) {
      // Game over or hasn't started
      return;
    }

    eventBus.emit('game:ui-update', { phase: "reconnecting" });
    this.isReconnecting = true;
    
    setTimeout(() => {
      this.reconnectAttempts++;
      if (this.reconnectAttempts <= 2) {
        this.connectWebSocket();
      } else {
        eventBus.emit('game:session-end', {
          game_type: "team_tower",
          outcome: "incomplete",
          stars: 1,
          duration_ms: Date.now() - this.sessionStartTime
        });
      }
    }, 2000);
  }

  private reconcileState(newState: TowerState) {
    if (!newState) return;
    
    const wasStable = this.currentState ? this.currentState.stable : true;
    
    for (const block of newState.blocks) {
      if (!this.blockObjects.has(block.id)) {
        const color = getBlockColour(block.colour);
        const dims = getBlockDimensions(block.shape);
        const startX = block.x * this.scale.width;
        // Start high up for spawn animation
        const startY = -100;

        const rect = this.add.rectangle(startX, startY, dims.w, dims.h, color);
        rect.setStrokeStyle(2, 0x000000);
        this.blockObjects.set(block.id, rect);

        // Spawn tween
        this.tweens.add({
          targets: rect,
          scaleX: { from: 0.1, to: 1.0 },
          scaleY: { from: 0.1, to: 1.0 },
          duration: 150,
          ease: 'Back.easeOut'
        });
      }
    }

    // Check falls
    if (newState.stable === false && wasStable === true) {
      this.cameras.main.shake(300, 0.01);
      
      this.blockObjects.forEach((rect) => {
        this.tweens.add({
          targets: rect,
          angle: Phaser.Math.Between(-45, 45),
          duration: 600,
          ease: 'Power1'
        });
        this.tweens.add({
          targets: rect,
          y: rect.y + 200,
          alpha: 0,
          duration: 800,
          ease: 'Cubic.easeIn'
        });
      });
    }

    // Check Wins
    if (newState.current_height >= newState.target_height && wasStable) {
       this.blockObjects.forEach(rect => {
         // Tiny jump reaction
         this.tweens.add({
           targets: rect,
           y: rect.y - 20,
           yoyo: true,
           duration: 300,
           ease: 'Sine.easeInOut'
         });
       });

       // Particle emitter from top of tower (rough center top calculation)
       if (newState.blocks.length > 0) {
          const topBlock = newState.blocks[newState.blocks.length - 1];
          const tx = topBlock.x * this.scale.width;
          const ty = this.scale.height - (topBlock.y * BLOCK_HEIGHT_PX) - 20;

          const emitter = this.add.particles(tx, ty, 'sparkle', {
            speed: { min: -200, max: 200 },
            angle: { min: 0, max: 360 },
            scale: { start: 1, end: 0 },
            lifespan: 1000,
            quantity: 20,
            blendMode: 'ADD'
          });
          emitter.explode(40);
       }
    }
    
    this.currentState = newState;
    this.myTurn = (newState.active_player === this.playerRole);

    // Sync UI
    eventBus.emit('game:ui-update', {
      phase: "playing",
      groupXP: newState.group_xp,
      activePlayer: newState.active_player,
      myRole: this.playerRole,
      turnNumber: newState.turn_number,
      nextBlockShape: newState.next_block_shape
    });

    // Update ghost preview block based on next_block_shape
    if (this.previewBlock) {
      this.previewBlock.destroy();
      this.previewBlock = null;
    }
    if (this.myTurn && newState.stable) {
      const dims = getBlockDimensions(newState.next_block_shape);
      this.previewBlock = this.add.rectangle(this.input.x, -50, dims.w, dims.h, 0xffffff, 0.5);
      this.previewBlock.setStrokeStyle(2, 0xffffff, 1);
      this.dropGuide.setVisible(true);
      this.dropGuide.width = dims.w;
    } else {
      this.dropGuide.setVisible(false);
    }
  }

  private handlePointerMove(pointer: Phaser.Input.Pointer) {
    if (!this.myTurn || !this.currentState || !this.currentState.stable) return;

    // Clamp between 0.05 and 0.95 horizontally
    const minX = 0.05 * this.scale.width;
    const maxX = 0.95 * this.scale.width;
    const clampedX = Phaser.Math.Clamp(pointer.x, minX, maxX);

    if (this.previewBlock) {
       this.previewBlock.x = clampedX;
       // Target Y sits slightly above where it will land roughly
       const topY = this.scale.height - (this.currentState.current_height * BLOCK_HEIGHT_PX) - 60;
       this.previewBlock.y = topY;
    }

    this.dropGuide.x = clampedX;
  }

  private handlePointerDown(pointer: Phaser.Input.Pointer) {
     if (!this.myTurn || !this.currentState || !this.currentState.stable) return;
     if (!this.ws || this.ws.readyState !== WebSocket.OPEN) return;

     const minX = 0.05 * this.scale.width;
     const maxX = 0.95 * this.scale.width;
     const clampedX = Phaser.Math.Clamp(pointer.x, minX, maxX);
     
     const positionX = clampedX / this.scale.width;

     const action = {
       type: "place_block",
       position_x: positionX,
       client_timestamp: Date.now()
     };

     this.ws.send(JSON.stringify(action));

     // Ghost drop trail
     const trail = this.add.line(0, 0, clampedX, 0, clampedX, this.scale.height, 0xffffff, 0.4);
     trail.setOrigin(0, 0);
     this.tweens.add({
       targets: trail,
       alpha: 0,
       duration: 500,
       onComplete: () => trail.destroy()
     });
     
     // Tentatively hide previews until turn swaps back to me
     this.myTurn = false;
     if (this.previewBlock) {
       this.previewBlock.setVisible(false);
     }
     this.dropGuide.setVisible(false);
  }

  shutdown() {
    if (this.heartbeatInterval) clearInterval(this.heartbeatInterval);
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }
}
