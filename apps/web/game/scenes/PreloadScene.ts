import * as Phaser from 'phaser';

export default class PreloadScene extends Phaser.Scene {
  constructor() {
    super('PreloadScene');
  }

  preload() {
    // Generate placeholder rectangle assets for shapes
    const graphics = this.add.graphics();
    
    graphics.fillStyle(0xff4444, 1);
    graphics.fillRect(0, 0, 32, 32);
    graphics.generateTexture('circle_red', 32, 32);
    
    graphics.clear();
    graphics.fillStyle(0x4488ff, 1);
    graphics.fillRect(0, 0, 32, 32);
    graphics.generateTexture('square_blue', 32, 32);

    graphics.clear();
    graphics.fillStyle(0x44ff44, 1);
    graphics.fillRect(0, 0, 32, 32);
    graphics.generateTexture('triangle_green', 32, 32);

    graphics.clear();
    graphics.fillStyle(0xffee44, 1);
    graphics.fillRect(0, 0, 32, 32);
    graphics.generateTexture('star_yellow', 32, 32);

    // Sparkle particle placeholder
    graphics.clear();
    graphics.fillStyle(0xffffff, 1);
    graphics.fillRect(0, 0, 8, 8);
    graphics.generateTexture('sparkle', 8, 8);
    
    graphics.destroy();
  }

  create() {
    this.scene.start('MemoryCoveScene');
  }
}
