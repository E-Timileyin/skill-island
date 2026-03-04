"use client";

import { useState } from "react";

const AVATARS = [
  { id: 1, emoji: "🦊", bg: "bg-orange-400" },
  { id: 2, emoji: "🐙", bg: "bg-purple-400" },
  { id: 3, emoji: "🦋", bg: "bg-blue-400" },
  { id: 4, emoji: "🌿", bg: "bg-green-400" },
  { id: 5, emoji: "🌟", bg: "bg-yellow-400" },
  { id: 6, emoji: "🐳", bg: "bg-cyan-400" },
];

interface ProfileSetupProps {
  onSubmit: (data: {
    nickname: string;
    avatar_id: number;
    play_mode: string;
  }) => void;
  isLoading: boolean;
  error?: string;
}

const NICKNAME_REGEX = /^[a-zA-Z0-9 ]+$/;

export default function ProfileSetup({
  onSubmit,
  isLoading,
  error,
}: ProfileSetupProps) {
  const [nickname, setNickname] = useState("");
  const [avatarId, setAvatarId] = useState(1);
  const [playMode, setPlayMode] = useState<"solo" | "team">("solo");
  const [validationError, setValidationError] = useState<string | null>(null);

  const selectedAvatar = AVATARS.find((a) => a.id === avatarId) ?? AVATARS[0];

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setValidationError(null);

    const trimmed = nickname.trim();
    if (!trimmed) {
      setValidationError("Nickname is required");
      return;
    }
    if (trimmed.length > 20) {
      setValidationError("Nickname must be 20 characters or fewer");
      return;
    }
    if (!NICKNAME_REGEX.test(trimmed)) {
      setValidationError("Nickname can only contain letters, numbers, and spaces");
      return;
    }

    onSubmit({ nickname: trimmed, avatar_id: avatarId, play_mode: playMode });
  }

  const displayError = validationError || error;

  return (
    <form
      onSubmit={handleSubmit}
      className="w-full max-w-md space-y-6 rounded-2xl bg-white p-8 shadow-lg"
    >
      <h1 className="text-2xl font-bold text-center text-gray-800">
        Create Your Profile
      </h1>

      {displayError && (
        <p className="text-sm text-red-600 text-center rounded-lg bg-red-50 p-2">
          {displayError}
        </p>
      )}

      {/* Avatar preview */}
      <div className="flex flex-col items-center gap-2">
        <div
          className={`flex h-20 w-20 items-center justify-center rounded-full text-4xl ${selectedAvatar.bg}`}
        >
          {selectedAvatar.emoji}
        </div>
        <span className="text-sm text-gray-500">Your avatar</span>
      </div>

      {/* Avatar grid */}
      <div>
        <label className="block text-sm font-semibold text-gray-700 mb-2">
          Choose an avatar
        </label>
        <div className="grid grid-cols-3 gap-3">
          {AVATARS.map((avatar) => (
            <button
              key={avatar.id}
              type="button"
              onClick={() => setAvatarId(avatar.id)}
              className={`flex h-16 w-full items-center justify-center rounded-xl text-2xl transition-all ${
                avatar.bg
              } ${
                avatarId === avatar.id
                  ? "ring-4 ring-blue-500 ring-offset-2 scale-110"
                  : "opacity-70 hover:opacity-100 hover:scale-105"
              }`}
            >
              {avatar.emoji}
            </button>
          ))}
        </div>
      </div>

      {/* Nickname input */}
      <div>
        <label
          htmlFor="nickname"
          className="block text-sm font-semibold text-gray-700 mb-1"
        >
          Nickname
        </label>
        <input
          id="nickname"
          type="text"
          maxLength={20}
          value={nickname}
          onChange={(e) => setNickname(e.target.value)}
          placeholder="Enter your nickname"
          className="block w-full rounded-lg border border-gray-300 px-3 py-2 focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
        />
      </div>

      {/* Play mode toggle */}
      <div>
        <label className="block text-sm font-semibold text-gray-700 mb-2">
          Play mode
        </label>
        <div className="flex gap-2">
          <button
            type="button"
            onClick={() => setPlayMode("solo")}
            className={`flex-1 rounded-lg py-2 text-sm font-semibold transition-colors ${
              playMode === "solo"
                ? "bg-blue-600 text-white"
                : "bg-gray-100 text-gray-600 hover:bg-gray-200"
            }`}
          >
            Solo
          </button>
          <button
            type="button"
            onClick={() => setPlayMode("team")}
            className={`flex-1 rounded-lg py-2 text-sm font-semibold transition-colors ${
              playMode === "team"
                ? "bg-blue-600 text-white"
                : "bg-gray-100 text-gray-600 hover:bg-gray-200"
            }`}
          >
            Team
          </button>
        </div>
      </div>

      {/* Submit */}
      <button
        type="submit"
        disabled={isLoading}
        className="w-full rounded-lg bg-green-500 py-3 text-lg font-bold text-white transition-colors hover:bg-green-600 disabled:opacity-50"
      >
        {isLoading ? "Setting up…" : "Let\u2019s Go!"}
      </button>
    </form>
  );
}
