export function getSpeedForDifficulty(level: number): number {
  switch (level) {
    case 4: return 160;
    case 3: return 130;
    case 2: return 100;
    case 1:
    default: return 80;
  }
}

export function getSpawnIntervalForDifficulty(level: number): number {
  switch (level) {
    case 4: return 700;
    case 3: return 800;
    case 2: return 1000;
    case 1:
    default: return 1200;
  }
}

export function getTargetColour(targetType: string): number {
  switch (targetType) {
    case "butterfly_blue": return 0x4A90E2;
    case "butterfly_orange": return 0xF5A623;
    case "butterfly_red": return 0xD0021B;
    case "bee": return 0xF8E71C;
    default: return 0xFFFFFF;
  }
}
