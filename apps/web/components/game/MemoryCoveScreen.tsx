import React from "react";
// import { WoodenSign } from "../WoodenSign";

export function MemoryCoveScreen() {
  return (
    <div className="flex flex-col items-center w-full max-w-3xl">
      {/* <WoodenSign text="Memory Cove" /> */}
      <h2 className="text-white font-extrabold text-2xl mb-8 drop-shadow-[0_2px_2px_rgba(0,0,0,0.5)] text-center">
        Repeat the Sequence to Earn Stars!
      </h2>

      {/* Sequence Display */}
      <div className="flex gap-4 mb-12">
        <div className="w-16 h-16 rounded-full bg-red-500 border-4 border-white shadow-lg flex items-center justify-center text-white font-bold text-3xl">
          1
        </div>
        <div className="w-16 h-16 rounded-full bg-yellow-400 border-4 border-white shadow-lg flex items-center justify-center text-white font-bold text-3xl">
          2
        </div>
        <div className="w-16 h-16 rounded-full bg-green-500 border-4 border-white shadow-lg flex items-center justify-center text-white font-bold text-3xl">
          3
        </div>
        <div className="w-16 h-16 rounded-full bg-blue-500 border-4 border-white shadow-lg flex items-center justify-center text-white font-bold text-3xl">
          4
        </div>
      </div>

      {/* Interactive Buttons */}
      <div className="bg-amber-700 border-b-8 border-amber-900 rounded-2xl p-6 shadow-2xl relative">
        {/* Wood texture details */}
        <div className="absolute inset-2 border-2 border-amber-800/30 rounded-xl pointer-events-none"></div>
        <div className="absolute top-4 left-4 w-3 h-3 rounded-full bg-amber-900 opacity-50"></div>
        <div className="absolute top-4 right-4 w-3 h-3 rounded-full bg-amber-900 opacity-50"></div>
        <div className="absolute bottom-4 left-4 w-3 h-3 rounded-full bg-amber-900 opacity-50"></div>
        <div className="absolute bottom-4 right-4 w-3 h-3 rounded-full bg-amber-900 opacity-50"></div>

        <div className="flex gap-6 relative z-10">
          <button className="w-24 h-24 rounded-full bg-gradient-to-b from-red-400 to-red-600 border-4 border-gray-300 shadow-[0_8px_0_#991b1b] hover:translate-y-2 hover:shadow-[0_0px_0_#991b1b] transition-all active:scale-95"></button>
          <button className="w-24 h-24 rounded-full bg-gradient-to-b from-yellow-300 to-yellow-500 border-4 border-gray-300 shadow-[0_8px_0_#b45309] hover:translate-y-2 hover:shadow-[0_0px_0_#b45309] transition-all active:scale-95"></button>
          <button className="w-24 h-24 rounded-full bg-gradient-to-b from-green-400 to-green-600 border-4 border-gray-300 shadow-[0_8px_0_#166534] hover:translate-y-2 hover:shadow-[0_0px_0_#166534] transition-all active:scale-95"></button>
          <button className="w-24 h-24 rounded-full bg-gradient-to-b from-blue-400 to-blue-600 border-4 border-gray-300 shadow-[0_8px_0_#1e3a8a] hover:translate-y-2 hover:shadow-[0_0px_0_#1e3a8a] transition-all active:scale-95"></button>
        </div>
      </div>
    </div>
  );
}
