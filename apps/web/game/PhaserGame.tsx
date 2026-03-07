"use client";

import { useEffect, useRef } from "react";

interface PhaserGameProps {
  scene?: string; // e.g. "FocusForestScene" or "MemoryCoveScene"
}

export default function PhaserGame({ scene = "MemoryCoveScene" }: PhaserGameProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const gameRef = useRef<any>(null);

  useEffect(() => {
    if (!containerRef.current || gameRef.current) return;

    // All imports are dynamic so Phaser never runs during SSR
    Promise.all([
      import("phaser"),
      import("@/game/scenes/PreloadScene"),
      import("@/game/scenes/MemoryCoveScene"),
      import("@/game/scenes/FocusForestScene"),
      import("@/game/scenes/TeamTowerScene"),
    ]).then(([Phaser, { default: PreloadScene }, { default: MemoryCoveScene }, { default: FocusForestScene }, { default: TeamTowerScene }]) => {
      const config: Phaser.Types.Core.GameConfig = {
        type: Phaser.AUTO,
        parent: containerRef.current!,
        width: 800,
        height: 600,
        backgroundColor: "#2A3B4C",
        scene: [PreloadScene, MemoryCoveScene, FocusForestScene, TeamTowerScene],
      };

      const game = new Phaser.Game(config);
      gameRef.current = game;

      game.events.on("ready", () => {
        game.scene.start(scene);
      });
    });

    return () => {
      if (gameRef.current) {
        gameRef.current.destroy(true);
        gameRef.current = null;
      }
    };
  }, [scene]);

  return (
    <div
      ref={containerRef}
      id="phaser-container"
      className="overflow-hidden rounded-xl border-4 border-gray-800 shadow-2xl"
    />
  );
}
