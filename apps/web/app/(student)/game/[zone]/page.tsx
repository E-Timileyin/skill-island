"use client";

import dynamic from "next/dynamic";

const PhaserGame = dynamic(() => import("@/game/PhaserGame"), { ssr: false });

import React from "react";

export default function GameZonePage(props: any) {
  const params = React.use(props.params);
  return (
    <main className="flex min-h-screen flex-col items-center justify-center">
      <h1 className="mb-4 text-2xl font-bold">Zone: {params.zone}</h1>
      <PhaserGame />
    </main>
  );
}
