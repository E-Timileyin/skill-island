// Memory Cove game logic helpers

export function buildButtonID(shape: string, colour: string): string {
  return `${shape}-${colour}`;
}

export function getShapeColour(colour: string): number {
  switch (colour) {
    case 'red': return 0xff4444;
    case 'blue': return 0x4488ff;
    case 'green': return 0x44ff44;
    case 'yellow': return 0xffee44;
    default: return 0xffffff;
  }
}

export function getEncouragementText(accuracy: number): string {
  if (accuracy >= 0.9) return 'Great!';
  if (accuracy >= 0.7) return 'Nice try!';
  return 'Keep going!';
}
