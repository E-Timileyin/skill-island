"use client";

import React from "react";

export default function SkillIslandHeader() {
  return (
    <div className="flex flex-col items-center mt-8 mb-6 select-none">
      {/* Wooden sign */}
      <div className="relative">
        <div className="bg-gradient-to-b from-yellow-400 to-orange-400 rounded-t-3xl rounded-b-xl shadow-lg px-8 py-4 border-4 border-yellow-700 text-center">
          <span className="text-5xl md:text-6xl font-extrabold text-yellow-100 drop-shadow-lg tracking-wide" style={{textShadow: "0 4px 8px #a15c1a, 0 1px 0 #fff"}}>
            SKILL ISLAND
          </span>
        </div>
        {/* Blue sub-banner */}
        <div className="absolute left-1/2 -translate-x-1/2 top-full mt-[-1.2rem] w-[22rem] md:w-[28rem] bg-blue-600 rounded-b-2xl border-4 border-blue-900 flex items-center justify-center py-3 shadow-md">
          <span className="text-white text-2xl font-bold mx-4">Explore</span>
          <span className="text-white text-2xl font-bold mx-2">•</span>
          <span className="text-white text-2xl font-bold mx-4">Learn</span>
          <span className="text-white text-2xl font-bold mx-2">•</span>
          <span className="text-white text-2xl font-bold mx-4">Team</span>
        </div>
      </div>
    </div>
  );
}
