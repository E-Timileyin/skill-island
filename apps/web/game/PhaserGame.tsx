"use client";

import { useEffect, useRef } from "react";
import eventBus from "@/game/events/EventBus";

export default function PhaserGame() {
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    console.log("Phaser ready");

    const handleSessionEnd = (...args: unknown[]) => {
      console.log("game:session-end received", args);
    };

    eventBus.on("game:session-end", handleSessionEnd);

    return () => {
      eventBus.off("game:session-end", handleSessionEnd);
    };
  }, []);

  return (
    <div
      ref={containerRef}
      id="phaser-container"
      className="h-[600px] w-[800px] rounded border bg-gray-900"
    />
  );
}
