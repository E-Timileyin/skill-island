"use client";

import dynamic from "next/dynamic";

const PhaserGame = dynamic(() => import("@/game/PhaserGame"), { ssr: false });

export default function GameZonePage({
  params,
}: {
  params: { zone: string };
}) {
  return (
    <main className="flex min-h-screen flex-col items-center justify-center">
      <h1 className="mb-4 text-2xl font-bold">Zone: {params.zone}</h1>
      <PhaserGame />
    </main>
  );
}
