export interface Block {
  id: string;
  x: number;
  y: number;
  shape: string;
  colour: string;
}

export function lerpBlock(current: { x: number; y: number }, target: { x: number; y: number }, factor: number): { x: number; y: number } {
  return {
    x: current.x + (target.x - current.x) * factor,
    y: current.y + (target.y - current.y) * factor
  };
}

export function getBlockColour(colour: string): number {
  switch (colour) {
    case 'red': return 0xe74c3c;
    case 'blue': return 0x3498db;
    case 'green': return 0x2ecc71;
    case 'yellow': return 0xf1c40f;
    case 'purple': return 0x9b59b6;
    default: return 0x95a5a6;
  }
}

export function getBlockDimensions(shape: string): { w: number; h: number } {
  // Base dimensions defined heavily by the server logic bounds and game mechanics.
  // We mirror server blocks (all represented as rectangles)
  switch (shape) {
    case 'square': return { w: 40, h: 40 };
    case 'wide_rect': return { w: 80, h: 40 };
    case 'tall_rect': return { w: 40, h: 80 };
    case 'large_square': return { w: 80, h: 80 };
    default: return { w: 80, h: 40 };
  }
}

export function calculateCoM(blocks: Block[]): number {
  if (blocks.length === 0) return 0.5;

  let totalWeight = 0;
  let weightedCenterX = 0;

  for (const block of blocks) {
    const dims = getBlockDimensions(block.shape);
    const weight = (dims.w * dims.h) / (40 * 40); // Base unit area is 1600
    totalWeight += weight;
    weightedCenterX += block.x * weight;
  }

  return weightedCenterX / totalWeight;
}
