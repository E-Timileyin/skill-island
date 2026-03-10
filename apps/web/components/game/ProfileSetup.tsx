"use client";

import React, { useState } from "react";
import { AVATARS } from "@/lib/avatars";
import Image from "next/image";
import { CheckCircle2, Box, Users } from "lucide-react";
import { createProfile } from "@/lib/api"; // Added back just in case, though it seems unused in this file now

export type ProfileSetupProps = {
  onSubmit: (data: { nickname: string; avatar_id: number; play_mode: string }) => Promise<void>;
  isLoading: boolean;
  error?: string;
};

export default function ProfileSetup({ onSubmit, isLoading, error }: ProfileSetupProps) {
  const [selectedAvatar, setSelectedAvatar] = useState(1); // default to first avatar id
  const [nickname, setNickname] = useState("");
  const [playMode, setPlayMode] = useState("");

  const handleSubmit = async () => {
    await onSubmit({
      nickname,
      avatar_id: selectedAvatar,
      play_mode: playMode,
    });
  };
 
  // Find the selected avatar object
  const selectedAvatarObj = AVATARS.find(a => a.id === selectedAvatar) || AVATARS[0];

  return (
    <>
    <div className="relative w-full min-h-screen flex flex-col items-center justify-center">
      <div
        className="w-full flex flex-col items-center pointer-events-none z-30
          absolute top-0 left-1/2 -translate-x-1/2
          md:static md:translate-x-0 md:mt-0 md:mb-8"
        style={{ maxWidth: 700 }}
      >
        <Image
          src="/assets/logo/profile-setup.png"
          alt="Welcome to Skill Island! Choose Your Avatar!"
          width={700}
          height={220}
          className="object-contain drop-shadow-lg select-none absolute  md:-top-[5.7rem]"
          priority
        />
      </div>     
      {/* Main Full-Screen Background Image */}
      <div className="fixed inset-0 z-0 w-screen h-screen">
        <Image
          src="/assets/images/bg-profile-setup.webp"
          alt="Mushroom Village Background"
          fill
          className="object-cover"
          priority
        />
      </div>

      {/* Logo fixed at the very top */}
  

      <div className="w-full max-w-3xl pt-2 relative z-10 flex flex-col items-center">
        {/* PROFILE COMPONENT */}
        <div className="bg-amber-50 rounded-3xl p-6 shadow-xl border-4 border-white w-full relative mt-40 z-10">
          <div className="absolute inset-2 border-2 border-dashed border-amber-200 rounded-2xl pointer-events-none"></div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-8 relative z-10">
          {/* Left Side */}
          <div className="flex flex-col gap-6">
            <div className="grid grid-cols-4 gap-4">
              {AVATARS.map((avatar) => {
                const isSelected = selectedAvatar === avatar.id;
                return (
                  <button
                    key={avatar.id}
                    onClick={() => setSelectedAvatar(avatar.id)}
                    className={`relative w-16 h-16 rounded-full flex items-center justify-center transition-transform overflow-hidden ${
                      isSelected
                        ? "ring-4 ring-green-400 scale-110 shadow-lg"
                        : "hover:scale-105 border-2 border-amber-200 bg-white"
                    }`}
                  >
                    <Image
                      src={avatar.src}
                      alt={avatar.label}
                      width={64}
                      height={64}
                      className="object-cover"
                    />
                    {isSelected && (
                      <div className="absolute -bottom-1 -right-1 bg-white rounded-full">
                        <CheckCircle2 className="w-5 h-5 text-green-500 shadow-sm" />
                      </div>
                    )}
                  </button>
                );
              })}
            </div>

            <div className="flex items-center gap-4">
              <label className="text-amber-800 font-bold text-lg">
                Nickname:
              </label>
              <input
                type="text"
                value={nickname}
                onChange={(e) => setNickname(e.target.value)}
                className="flex-1 bg-white border-2 border-amber-200 rounded-xl px-4 py-2 font-bold text-gray-700 focus:outline-none focus:ring-2 focus:ring-amber-400"
              />
            </div>
          </div>

          {/* Right Side */}
          <div className="flex flex-col gap-6 items-center">
            <div className="bg-sky-100 border-4 border-amber-600 rounded-2xl p-4 w-full flex flex-col items-center relative overflow-hidden">
              <div className="absolute inset-0 z-0">
                <Image
                  src="/assets/images/bg-profile-setup.webp"
                  alt="Avatar Background"
                  fill
                  className="object-cover"
                  priority
                />
              </div>
              <div className="absolute -top-4 bg-amber-600 text-white px-4 py-1 rounded-full font-bold text-sm z-10">
                Preview
              </div>
              <div className={`w-32 h-32 rounded-full border-4 border-white overflow-hidden mb-4 mt-2 shadow-inner bg-white z-10`}>
                <Image
                  src={selectedAvatarObj.src}
                  alt="Selected Avatar"
                  width={128}
                  height={128}
                  className="object-cover"
                />
              </div>
              <div className="bg-amber-600 text-white px-6 py-2 rounded-lg font-bold w-full text-center z-10">
                {nickname || "Player"}
              </div>
            </div>

            <div className="w-full">
              <div className="text-amber-800 font-bold text-lg mb-2 text-center">
                Play Mode:
              </div>
              <div className="flex gap-4 justify-center">
                <button
                  onClick={() => setPlayMode("solo")}
                  className={`flex items-center gap-2 px-6 py-3 rounded-xl font-bold transition-all ${
                    playMode === "solo"
                      ? "bg-sky-400 text-white shadow-[0_4px_0_#0284c7] translate-y-0"
                      : "bg-amber-100 text-amber-800 shadow-[0_4px_0_#fcd34d] hover:translate-y-1 hover:shadow-[0_2px_0_#fcd34d]"
                  }`}
                >
                  <Box className="w-6 h-6" /> Solo
                  {playMode === "solo" && (
                    <CheckCircle2 className="w-5 h-5 text-green-300 ml-1" />
                  )}
                </button>
                <button
                  onClick={() => setPlayMode("team")}
                  className={`flex items-center gap-2 px-6 py-3 rounded-xl font-bold transition-all ${
                    playMode === "team"
                      ? "bg-orange-400 text-white shadow-[0_4px_0_#c2410c] translate-y-0"
                      : "bg-amber-100 text-amber-800 shadow-[0_4px_0_#fcd34d] hover:translate-y-1 hover:shadow-[0_2px_0_#fcd34d]"
                  }`}
                >
                  <Users className="w-6 h-6" /> Team
                  {playMode === "team" && (
                    <CheckCircle2 className="w-5 h-5 text-green-300 ml-1" />
                  )}
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <button
        onClick={handleSubmit}
        className="mt-8 bg-gradient-to-b from-orange-400 to-orange-600 text-white font-extrabold text-2xl px-12 py-4 rounded-full shadow-[0_8px_0_#c2410c] border-4 border-white hover:translate-y-2 hover:shadow-[0_0px_0_#c2410c] transition-all z-20"
        disabled={isLoading}
      >
        {isLoading ? "Loading…" : "Let’s Go!"}
      </button>
      {error && <div className="mt-4 text-red-600 font-bold">{error}</div>}
    </div>
  </div>
</>
  );
}
