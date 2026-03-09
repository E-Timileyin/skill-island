"use client";

import React, { useState, useEffect, useCallback, useRef } from "react";
import { motion, AnimatePresence } from "motion/react";
import { Star, Home, RotateCcw, Eye, HandMetal } from "lucide-react";
import Image from "next/image";
import Link from "next/link";
import { initSession, submitSession } from "@/lib/api";

/* ── Colour Palette ─── */
const COLORS = [
  { id: "red", bg: "from-rose-400 to-rose-600", border: "#be123c", glow: "rgba(244,63,94,0.6)", hex: "#f43f5e", lit: "from-rose-300 to-rose-400" },
  { id: "yellow", bg: "from-amber-300 to-amber-500", border: "#b45309", glow: "rgba(245,158,11,0.6)", hex: "#f59e0b", lit: "from-amber-200 to-amber-300" },
  { id: "green", bg: "from-emerald-400 to-emerald-600", border: "#166534", glow: "rgba(16,185,129,0.6)", hex: "#10b981", lit: "from-emerald-300 to-emerald-400" },
  { id: "blue", bg: "from-sky-400 to-sky-600", border: "#1e3a8a", glow: "rgba(59,130,246,0.6)", hex: "#3b82f6", lit: "from-sky-300 to-sky-400" },
];

const MAX_ROUNDS = 10;
const SEQUENCE_DISPLAY_MS = 600;
const SEQUENCE_GAP_MS = 300;

type GamePhase = "idle" | "showing" | "input" | "feedback" | "result";

function seededRandom(seed: number) {
  let s = seed;
  return () => {
    s = (s * 16807) % 2147483647;
    return (s - 1) / 2147483646;
  };
}

function generateSequence(seed: number, length: number): number[] {
  const rand = seededRandom(seed);
  return Array.from({ length }, () => Math.floor(rand() * COLORS.length));
}

function calcStars(accuracy: number): number {
  if (accuracy >= 0.9) return 3;
  if (accuracy >= 0.7) return 2;
  if (accuracy >= 0.4) return 1;
  return 0;
}

export default function MemoryCovePage() {
  const [phase, setPhase] = useState<GamePhase>("idle");
  const [round, setRound] = useState(1);
  const [sequence, setSequence] = useState<number[]>([]);
  const [playerIndex, setPlayerIndex] = useState(0);
  const [litButton, setLitButton] = useState<number | null>(null);
  const [correctCount, setCorrectCount] = useState(0);
  const [totalAttempts, setTotalAttempts] = useState(0);
  const [stars, setStars] = useState(0);
  const [shakeButton, setShakeButton] = useState<number | null>(null);
  const [sessionToken, setSessionToken] = useState("");
  const [seed, setSeed] = useState(0);
  const [actionLog, setActionLog] = useState<any[]>([]);
  const [startTime, setStartTime] = useState(0);
  const [result, setResult] = useState<any>(null);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState("");
  const timeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  // Sequence length increases with rounds
  const getSeqLength = (r: number) => {
    if (r >= 9) return 7;
    if (r >= 7) return 6;
    if (r >= 5) return 5;
    if (r >= 3) return 4;
    return 3;
  };

  // Initialize session on mount
  useEffect(() => {
    initSession("memory_cove")
      .then((res) => {
        setSessionToken(res.session_token);
        setSeed(res.seed);
        const seq = generateSequence(res.seed, 3);
        setSequence(seq);
        setStartTime(Date.now());
      })
      .catch(() => {
        // Fallback: play offline with random seed
        const fallbackSeed = Math.floor(Math.random() * 99999);
        setSeed(fallbackSeed);
        const seq = generateSequence(fallbackSeed, 3);
        setSequence(seq);
        setStartTime(Date.now());
      });

    return () => {
      if (timeoutRef.current) clearTimeout(timeoutRef.current);
    };
  }, []);

  // Start showing sequence automatically once it's generated
  useEffect(() => {
    if (sequence.length > 0 && phase === "idle") {
      timeoutRef.current = setTimeout(() => {
        showSequence();
      }, 1200);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [sequence, phase]);

  const showSequence = useCallback(() => {
    setPhase("showing");
    setPlayerIndex(0);
    let i = 0;

    const showNext = () => {
      if (i >= sequence.length) {
        setLitButton(null);
        timeoutRef.current = setTimeout(() => {
          setPhase("input");
        }, 400);
        return;
      }
      setLitButton(sequence[i]);
      i++;
      timeoutRef.current = setTimeout(() => {
        setLitButton(null);
        timeoutRef.current = setTimeout(showNext, SEQUENCE_GAP_MS);
      }, SEQUENCE_DISPLAY_MS);
    };

    timeoutRef.current = setTimeout(showNext, 500);
  }, [sequence]);

  const handleColorPress = (colorIndex: number) => {
    if (phase !== "input") return;

    const expected = sequence[playerIndex];
    const correct = colorIndex === expected;

    // Log action
    const action = {
      Type: "press",
      ButtonID: `${COLORS[colorIndex].id}`,
      ElementIndex: playerIndex,
      ClientTimestamp: Date.now() - startTime,
    };
    setActionLog((prev) => [...prev, action]);

    // Flash the button
    setLitButton(colorIndex);
    setTimeout(() => setLitButton(null), 200);

    if (correct) {
      setCorrectCount((c) => c + 1);
    } else {
      setShakeButton(colorIndex);
      setTimeout(() => setShakeButton(null), 400);
    }
    setTotalAttempts((t) => t + 1);

    const nextIndex = playerIndex + 1;
    setPlayerIndex(nextIndex);

    if (nextIndex >= sequence.length) {
      // Round complete
      const roundAccuracy = (correctCount + (correct ? 1 : 0)) / sequence.length;
      const roundStars = calcStars(roundAccuracy);
      setStars(Math.max(stars, roundStars));
      setPhase("feedback");

      if (round >= MAX_ROUNDS) {
        // Game over — submit
        timeoutRef.current = setTimeout(() => {
          submitResults([...actionLog, action]);
        }, 1500);
      } else {
        // Next round
        timeoutRef.current = setTimeout(() => {
          const nextRound = round + 1;
          setRound(nextRound);
          const newSeq = generateSequence(seed + nextRound, getSeqLength(nextRound));
          setSequence(newSeq);
          setPlayerIndex(0);
          setPhase("idle");
        }, 1500);
      }
    }
  };

  const submitResults = async (actions: any[]) => {
    setSubmitting(true);
    try {
      const res = await submitSession({
        session_token: sessionToken,
        actions,
        game_type: "memory_cove",
        mode: "solo",
        duration_ms: Date.now() - startTime,
      });
      setResult(res);
      setPhase("result");
    } catch {
      setError("Could not save your session. You can still play again!");
      setPhase("result");
    } finally {
      setSubmitting(false);
    }
  };

  const phaseText = () => {
    switch (phase) {
      case "idle": return "Get Ready...";
      case "showing": return "👀 Watch Carefully!";
      case "input": return "🎯 Your Turn!";
      case "feedback": return round >= MAX_ROUNDS ? "🎉 Amazing!" : "✨ Great Job!";
      case "result": return "🏆 Session Complete!";
      default: return "";
    }
  };

  const accuracy = totalAttempts > 0 ? correctCount / totalAttempts : 0;

  // Result Screen
  if (phase === "result") {
    const finalStars = result?.stars_earned ?? calcStars(accuracy);
    const xpEarned = result?.xp_earned ?? 0;

    return (
      <div className="relative min-h-screen flex flex-col items-center justify-center overflow-hidden select-none">
        <div className="absolute inset-0 z-0">
          <Image src="/assets/images/bg-memory-cove.jpg" alt="Memory Cove" fill className="object-cover" priority />
          <div className="absolute inset-0 bg-gradient-to-t from-black/60 to-transparent" />
        </div>

        <motion.div
          initial={{ scale: 0.8, opacity: 0 }}
          animate={{ scale: 1, opacity: 1 }}
          className="relative z-10 bg-white/90 backdrop-blur-lg rounded-[2.5rem] p-10 shadow-[0_30px_60px_rgba(0,0,0,0.4)] border-4 border-white/50 flex flex-col items-center gap-6 max-w-md w-full mx-4"
        >
          <h2 className="text-4xl font-extrabold text-gray-800">Session Complete!</h2>

          <div className="flex gap-2">
            {[0, 1, 2].map((i) => (
              <motion.div
                key={i}
                initial={{ scale: 0, rotate: -30 }}
                animate={{ scale: 1, rotate: 0 }}
                transition={{ delay: 0.3 + i * 0.2, type: "spring", stiffness: 300 }}
              >
                <Star
                  size={48}
                  className={i < finalStars ? "text-yellow-400 fill-yellow-400 drop-shadow-lg" : "text-gray-300 fill-gray-300"}
                />
              </motion.div>
            ))}
          </div>

          <div className="grid grid-cols-2 gap-4 w-full text-center">
            <div className="bg-emerald-50 rounded-2xl p-4">
              <p className="text-sm text-emerald-600 font-bold">Accuracy</p>
              <p className="text-3xl font-extrabold text-emerald-700">{Math.round(accuracy * 100)}%</p>
            </div>
            <div className="bg-sky-50 rounded-2xl p-4">
              <p className="text-sm text-sky-600 font-bold">XP Earned</p>
              <p className="text-3xl font-extrabold text-sky-700">+{xpEarned}</p>
            </div>
            <div className="bg-amber-50 rounded-2xl p-4">
              <p className="text-sm text-amber-600 font-bold">Rounds</p>
              <p className="text-3xl font-extrabold text-amber-700">{round}/{MAX_ROUNDS}</p>
            </div>
            <div className="bg-purple-50 rounded-2xl p-4">
              <p className="text-sm text-purple-600 font-bold">Score</p>
              <p className="text-3xl font-extrabold text-purple-700">{result?.score ?? correctCount}</p>
            </div>
          </div>

          {error && <p className="text-red-500 text-sm">{error}</p>}

          <div className="flex gap-4 w-full">
            <button
              onClick={() => window.location.reload()}
              className="flex-1 flex items-center justify-center gap-2 bg-amber-400 hover:bg-amber-500 text-white font-extrabold text-lg py-3 rounded-full shadow-[0_4px_0_#b45309] active:shadow-none active:translate-y-1 transition-all"
            >
              <RotateCcw size={20} />
              Play Again
            </button>
            <Link href="/island" className="flex-1">
              <button className="w-full flex items-center justify-center gap-2 bg-sky-500 hover:bg-sky-600 text-white font-extrabold text-lg py-3 rounded-full shadow-[0_4px_0_#0369a1] active:shadow-none active:translate-y-1 transition-all">
                <Home size={20} />
                Island
              </button>
            </Link>
          </div>
        </motion.div>
      </div>
    );
  }

  return (
    <div className="relative min-h-screen flex flex-col items-center justify-center overflow-hidden select-none p-4">
      {/* Background */}
      <div className="absolute inset-0 z-0">
        <Image src="/assets/images/bg-memory-cove.jpg" alt="Memory Cove" fill className="object-cover" priority />
        <div className="absolute inset-0 bg-black/20" />
      </div>

      {submitting && (
        <div className="fixed inset-0 bg-black/60 z-50 flex items-center justify-center">
          <div className="bg-white rounded-2xl px-8 py-6 shadow-2xl flex items-center gap-4">
            <div className="w-8 h-8 border-4 border-sky-500 border-t-transparent rounded-full animate-spin" />
            <span className="text-xl font-bold text-gray-700">Saving session...</span>
          </div>
        </div>
      )}

      {/* Game UI */}
      <div className="relative z-10 flex flex-col items-center w-full max-w-md gap-6">

        {/* Top Bar - Round & Stars */}
        <div className="w-full flex items-center justify-between">
          <div className="bg-white/90 backdrop-blur-sm rounded-full px-5 py-2 shadow-lg border-2 border-white/50">
            <span className="font-extrabold text-gray-700 text-lg">Round {round}/{MAX_ROUNDS}</span>
          </div>
          <div className="flex gap-1">
            {[0, 1, 2].map((i) => (
              <Star
                key={i}
                size={28}
                className={`${i < stars ? "text-yellow-400 fill-yellow-400" : "text-white/50 fill-white/20"} drop-shadow-md transition-all`}
              />
            ))}
          </div>
        </div>

        {/* Phase Banner */}
        <motion.div
          key={phaseText()}
          initial={{ scale: 0.8, opacity: 0 }}
          animate={{ scale: 1, opacity: 1 }}
          className="bg-gradient-to-b from-[#8BC34A] to-[#4CAF50] px-10 py-3 rounded-[2rem] border-4 border-white shadow-[0_6px_0_rgba(50,150,50,1),0_12px_20px_rgba(0,0,0,0.3)] flex items-center gap-3"
        >
          {phase === "showing" && <Eye size={28} className="text-white" />}
          {phase === "input" && <HandMetal size={28} className="text-white" />}
          <span className="text-white text-2xl sm:text-3xl font-extrabold drop-shadow-[0_2px_2px_rgba(0,0,0,0.3)] tracking-wide">
            {phaseText()}
          </span>
        </motion.div>

        {/* Sequence Display Dots */}
        <div className="flex gap-2 min-h-[28px]">
          {phase === "input" && sequence.map((_, i) => (
            <motion.div
              key={i}
              initial={{ scale: 0 }}
              animate={{ scale: 1 }}
              transition={{ delay: i * 0.05 }}
              className={`w-5 h-5 rounded-full border-2 border-white shadow-md transition-all ${
                i < playerIndex
                  ? "bg-emerald-400 scale-110"
                  : i === playerIndex
                    ? "bg-white animate-pulse"
                    : "bg-white/30"
              }`}
            />
          ))}
        </div>

        {/* Color Buttons Grid */}
        <div className="relative bg-amber-700 border-b-8 border-amber-900 rounded-[2rem] p-6 sm:p-8 shadow-[0_20px_40px_rgba(0,0,0,0.4)]">
          {/* Wood screws */}
          <div className="absolute top-4 left-4 w-3 h-3 rounded-full bg-amber-900/50 shadow-inner" />
          <div className="absolute top-4 right-4 w-3 h-3 rounded-full bg-amber-900/50 shadow-inner" />
          <div className="absolute bottom-4 left-4 w-3 h-3 rounded-full bg-amber-900/50 shadow-inner" />
          <div className="absolute bottom-4 right-4 w-3 h-3 rounded-full bg-amber-900/50 shadow-inner" />
          <div className="absolute inset-3 border-2 border-amber-800/20 rounded-2xl pointer-events-none" />

          <div className="grid grid-cols-2 gap-4 sm:gap-6 relative z-10">
            {COLORS.map((color, idx) => {
              const isLit = litButton === idx;
              const isShaking = shakeButton === idx;
              const interactable = phase === "input";

              return (
                <motion.button
                  key={color.id}
                  onClick={() => handleColorPress(idx)}
                  disabled={!interactable}
                  animate={isShaking ? { x: [0, -8, 8, -8, 8, 0] } : {}}
                  transition={isShaking ? { duration: 0.4 } : {}}
                  whileHover={interactable ? { scale: 1.05 } : {}}
                  whileTap={interactable ? { scale: 0.95, y: 4 } : {}}
                  className={`
                    w-28 h-28 sm:w-32 sm:h-32 rounded-full 
                    bg-gradient-to-b ${isLit ? color.lit : color.bg}
                    border-4 border-gray-200 
                    shadow-[0_8px_0_${color.border}]
                    transition-all duration-150
                    ${interactable ? "cursor-pointer hover:brightness-110" : "cursor-default"}
                    ${isLit ? `shadow-[0_0_30px_${color.glow},0_8px_0_${color.border}] brightness-125 scale-110` : ""}
                    focus:outline-none
                  `}
                  style={{
                    boxShadow: isLit
                      ? `0 0 40px ${color.glow}, 0 4px 0 ${color.border}`
                      : `0 8px 0 ${color.border}, 0 12px 20px rgba(0,0,0,0.2)`,
                  }}
                />
              );
            })}
          </div>
        </div>

        {/* Accuracy Bar */}
        <div className="w-full bg-white/20 rounded-full h-3 backdrop-blur-sm overflow-hidden">
          <motion.div
            className="h-full bg-gradient-to-r from-emerald-400 to-sky-400 rounded-full"
            animate={{ width: `${totalAttempts > 0 ? accuracy * 100 : 0}%` }}
            transition={{ duration: 0.3 }}
          />
        </div>
        <p className="text-white/80 text-sm font-bold">
          Accuracy: {totalAttempts > 0 ? Math.round(accuracy * 100) : 0}%
        </p>
      </div>
    </div>
  );
}
