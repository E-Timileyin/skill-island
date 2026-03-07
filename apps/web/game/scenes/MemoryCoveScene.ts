import Phaser from 'phaser';
import { buildButtonID, getShapeColour, getEncouragementText } from '../logic/memoryCove';
import EventBus from '../events/EventBus';

const ELEMENT_DISPLAY_MS = 800;
const ELEMENT_GAP_MS = 300;
const FEEDBACK_PAUSE_MS = 600;
const MAX_ROUNDS = 10;

const SHAPES = ['circle', 'square', 'triangle', 'star'];
const COLOURS = ['red', 'blue', 'green', 'yellow'];

export default class MemoryCoveScene extends Phaser.Scene {
  state = 'IDLE';
  round = 1;
  sequence: { Shape: string, Colour: string }[] = [];
  actionLog: any[] = [];
  sessionToken = '';
  seed = 0;
  inputStartTs = 0;
  elementIndex = 0;

  constructor() {
    super('MemoryCoveScene');
  }

  preload() {
    // PreloadScene handles assets
  }

  async create() {
    this.state = 'IDLE';
    this.round = 1;
    this.actionLog = [];
    this.elementIndex = 0;
    // Get seed + session_token from API
    const res = await fetch('/api/sessions/init', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ game_type: 'memory_cove', mode: 'solo' })
    });
    const { session_token, seed } = await res.json();
    this.sessionToken = session_token;
    this.seed = seed;
    this.sequence = this.generateSequence(seed, this.getSequenceLength(this.round));
    this.showIdle();
  }

  showIdle() {
    this.state = 'IDLE';
    this.add.text(400, 100, 'Watch carefully!', { fontSize: '32px', color: '#fff' }).setOrigin(0.5);
    this.time.delayedCall(1000, () => this.showSequence());
  }

  showSequence() {
    this.state = 'SHOWING_SEQUENCE';
    let x = 400, y = 250;
    let i = 0;
    const showNext = () => {
      if (i >= this.sequence.length) {
        this.time.delayedCall(ELEMENT_GAP_MS, () => this.waitForInput());
        return;
      }
      const elem = this.sequence[i];
      const colour = getShapeColour(elem.Colour);
      const rect = this.add.rectangle(x, y, 120, 120, colour, 1).setOrigin(0.5);
      this.add.text(x, y, elem.Shape, { fontSize: '28px', color: '#000' }).setOrigin(0.5);
      this.time.delayedCall(ELEMENT_DISPLAY_MS, () => {
        rect.destroy();
        i++;
        this.time.delayedCall(ELEMENT_GAP_MS, showNext);
      });
    };
    showNext();
  }

  waitForInput() {
    this.state = 'WAITING_FOR_INPUT';
    this.inputStartTs = Date.now();
    this.elementIndex = 0;
    this.renderButtons();
  }

  renderButtons() {
    const y = 450;
    const btns = [
      { shape: 'circle', colour: 'red' },
      { shape: 'square', colour: 'blue' },
      { shape: 'triangle', colour: 'green' },
      { shape: 'star', colour: 'yellow' }
    ];
    btns.forEach((btn, idx) => {
      const x = 200 + idx * 200;
      const colour = getShapeColour(btn.colour);
      const rect = this.add.rectangle(x, y, 100, 100, colour, 1).setOrigin(0.5);
      const label = this.add.text(x, y, btn.shape, { fontSize: '24px', color: '#000' }).setOrigin(0.5);
      rect.setInteractive();
      rect.on('pointerdown', () => this.handleButtonPress(btn.shape, btn.colour, rect));
      // Highlight expected button
      if (btn.shape === this.sequence[this.elementIndex].Shape && btn.colour === this.sequence[this.elementIndex].Colour) {
        rect.setStrokeStyle(4, 0xffff00);
      }
    });
  }

  handleButtonPress(shape: string, colour: string, rect: Phaser.GameObjects.Rectangle) {
    if (this.state !== 'WAITING_FOR_INPUT') return;
    const expected = this.sequence[this.elementIndex];
    const correct = shape === expected.Shape && colour === expected.Colour;
    const action = {
      Type: 'press',
      ButtonID: buildButtonID(shape, colour),
      ElementIndex: this.elementIndex,
      ClientTimestamp: Date.now() - this.inputStartTs
    };
    this.actionLog.push(action);
    if (correct) {
      const emitter = this.add.particles(rect.x, rect.y, 'sparkle', { speed: 100, lifespan: 400, tint: 0x00ff00, maxParticles: 10 });
    } else {
      this.tweens.add({ targets: rect, x: rect.x + 10, yoyo: true, repeat: 3, duration: 60, onComplete: () => { rect.x -= 10; } });
    }
    this.elementIndex++;
    if (this.elementIndex >= this.sequence.length) {
      this.time.delayedCall(FEEDBACK_PAUSE_MS, () => this.roundComplete());
    }
  }

  roundComplete() {
    this.state = 'ROUND_COMPLETE';
    const accuracy = this.actionLog.filter(a => a.ButtonID === buildButtonID(this.sequence[a.ElementIndex].Shape, this.sequence[a.ElementIndex].Colour)).length / this.sequence.length;
    const text = getEncouragementText(accuracy);
    this.add.text(400, 350, text, { fontSize: '28px', color: '#fff' }).setOrigin(0.5);
    this.round++;
    if (this.round > MAX_ROUNDS) {
      this.gameOver();
    } else {
      this.sequence = this.generateSequence(this.seed, this.getSequenceLength(this.round));
      this.actionLog = [];
      this.elementIndex = 0;
      this.time.delayedCall(FEEDBACK_PAUSE_MS, () => this.showIdle());
    }
  }

  gameOver() {
    this.state = 'GAME_OVER';
    this.input.enabled = false;
    EventBus.emit('game:session-end', {
      sessionToken: this.sessionToken,
      actions: this.actionLog
    });
    this.add.text(400, 350, 'Session complete!', { fontSize: '32px', color: '#fff' }).setOrigin(0.5);
  }

  generateSequence(seed: number, length: number) {
    // Deterministic sequence generation (client display only)
    const rand = new Phaser.Math.RandomDataGenerator([seed.toString()]);
    const seq = [];
    for (let i = 0; i < length; i++) {
      seq.push({
        Shape: SHAPES[rand.integerInRange(0, SHAPES.length - 1)],
        Colour: COLOURS[rand.integerInRange(0, COLOURS.length - 1)]
      });
    }
    return seq;
  }

  getSequenceLength(round: number) {
    if (round >= 11) return 8;
    if (round >= 9) return 7;
    if (round >= 7) return 6;
    if (round >= 5) return 5;
    if (round >= 3) return 4;
    return 3;
  }
}
