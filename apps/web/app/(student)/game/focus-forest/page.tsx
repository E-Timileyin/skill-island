"use client";

import React, { useState, useEffect, useRef, useCallback } from "react";
import { motion, AnimatePresence } from "motion/react";
import { Star, Home, RotateCcw, Clock } from "lucide-react";
import Image from "next/image";
import Link from "next/link";
import { initSession, submitSession } from "@/lib/api";

const BUTTERFLY_TYPES = [
  { id: "butterfly_blue", emoji: "🦋", color: "#60a5fa" },
  { id: "butterfly_orange", emoji: "🦋", color: "#fb923c" },
  { id: "butterfly_pink", emoji: "🦋", color: "#f472b6" },
  { id: "butterfly_green", emoji: "🦋", color: "#4ade80" },
];

const BEE_TYPE = { id: "bee", emoji: "🐝" };

interface Target {
  id: string;
  type: string;
  emoji: string;
  color?: string;
  fromX: number;
  toX: number;
  y: number;
  amplitude: number;
  duration: number;
  isBee: boolean;
  spawnTime: number;
}

const SESSION_DURATION_MS = 60000;

export default function FocusForestPage() {
  const [phase, setPhase] = useState<"loading" | "playing" | "result">("loading");
  const [targets, setTargets] = useState<Target[]>([]);
  const [score, setScore] = useState(0);
  const [butterfliesCaught, setButterfliesCaught] = useState(0);
  const [beesTapped, setBeesTapped] = useState(0);
  const [timeRemaining, setTimeRemaining] = useState(SESSION_DURATION_MS);
  const [sessionToken, setSessionToken] = useState("");
  const [actionLog, setActionLog] = useState<any[]>([]);
  const [result, setResult] = useState<any>(null);
  const [error, setError] = useState("");
  const [saving, setSaving] = useState(false);
  
  const gameAreaRef = useRef<HTMLDivElement>(null);
  const targetIdRef = useRef(0);
  
  // HUD Calculation
  const timeSeconds = Math.ceil(timeRemaining / 1000);
  const timePercent = (timeRemaining / SESSION_DURATION_MS) * 100;
  
  const calcStars = () => {
    if (result?.stars_earned != null) return result.stars_earned;
    if (butterfliesCaught >= 25) return 3;
    if (butterfliesCaught >= 15) return 2;
    if (butterfliesCaught >= 5) return 1;
    return 0;
  };

  // Init
  useEffect(() => {
    const init = async () => {
      try {
        const res = await initSession("focus_forest");
        setSessionToken(res.session_token);
      } catch (err) {
        console.warn("Offline play - initSession failed:", err);
      }
      setPhase("playing");
    };
    init();
  }, []);

  const spawnTarget = useCallback(() => {
    const isBee = Math.random() < 0.2; // 20% Bees
    const type: { id: string; emoji: string; color?: string } = isBee 
      ? BEE_TYPE 
      : BUTTERFLY_TYPES[Math.floor(Math.random() * BUTTERFLY_TYPES.length)];
    
    const fromLeft = Math.random() > 0.5;
    const fromX = fromLeft ? -15 : 115;
    const toX = fromLeft ? 115 : -15;
    const y = 15 + Math.random() * 70; // 15-85% height
    
    const newTarget: Target = {
      id: `t_${targetIdRef.current++}`,
      type: type.id,
      emoji: type.emoji,
      color: type.color,
      fromX,
      toX,
      y,
      amplitude: 5 + Math.random() * 10,
      duration: 3 + Math.random() * 3,
      isBee,
      spawnTime: Date.now(),
    };

    setTargets((prev) => [...prev, newTarget]);
    
    // Cleanup target from state after its duration
    setTimeout(() => {
      setTargets((prev) => prev.filter((t) => t.id !== newTarget.id));
    }, newTarget.duration * 1000 + 500);
  }, []);

  // Timers
  useEffect(() => {
    if (phase !== "playing") return;

    // Game End Timer
    const endTimer = setTimeout(() => {
      endGame();
    }, SESSION_DURATION_MS);

    // Visual Timer update
    const interval = setInterval(() => {
      setTimeRemaining((prev) => Math.max(0, prev - 100));
    }, 100);

    // Spawn loop
    const spawnTimer = setInterval(() => {
      spawnTarget();
    }, 1100);

    return () => {
      clearTimeout(endTimer);
      clearInterval(interval);
      clearInterval(spawnTimer);
    };
  }, [phase, spawnTarget]);

  const catchTarget = (t: Target, xPercent: number, yPercent: number) => {
    if (phase !== "playing") return;
    
    // Log action
    setActionLog((prev) => [
      ...prev,
      {
        type: "tap",
        target_id: t.id,
        target_type: t.type,
        tap_x: xPercent,
        tap_y: yPercent,
        client_timestamp: SESSION_DURATION_MS - timeRemaining, // Elapsed time
      }
    ]);

    setTargets((prev) => prev.filter((item) => item.id !== t.id));

    if (t.isBee) {
      setBeesTapped((v) => v + 1);
      setScore((v) => Math.max(0, v - 1));
    } else {
      setButterfliesCaught((v) => v + 1);
      setScore((v) => v + 1);
    }
  };

  const handleMiss = (e: React.MouseEvent) => {
    if (phase !== "playing") return;
    const rect = gameAreaRef.current?.getBoundingClientRect();
    if (!rect) return;
    const x = ((e.clientX - rect.left) / rect.width) * 100;
    const y = ((e.clientY - rect.top) / rect.height) * 100;

    setActionLog((prev) => [
      ...prev,
      {
        type: "tap",
        target_id: null,
        tap_x: x,
        tap_y: y,
        client_timestamp: SESSION_DURATION_MS - timeRemaining, // Elapsed time
      }
    ]);
  };

  const endGame = async () => {
    setPhase("result");
    setSaving(true);
    try {
      const res = await submitSession({
        session_token: sessionToken,
        actions: actionLog,
      } as any);
      setResult(res);
    } catch {
      setError("Progress saved locally.");
    } finally {
      setSaving(false);
    }
  };

  // Result UI
  if (phase === "result") {
    const finalStars = calcStars();
    return (
      <div className="relative min-h-screen flex items-center justify-center p-4">
        <div className="absolute inset-0 z-0">
          <Image src="/assets/images/bg-focus-forest.jpg" alt="Focus Forest" fill className="object-cover" />
          <div className="absolute inset-0 bg-black/60" />
        </div>
        
        <motion.div 
          initial={{ scale: 0.9, opacity: 0 }}
          animate={{ scale: 1, opacity: 1 }}
          className="relative z-10 bg-white/95 rounded-[2.5rem] p-8 shadow-2xl max-w-md w-full flex flex-col items-center gap-6 border-b-8 border-gray-200"
        >
          <h2 className="text-4xl font-black text-gray-800 tracking-tight">Forest Clean!</h2>
          
          <div className="flex gap-2">
            {[0, 1, 2].map((i) => (
              <Star key={i} size={50} className={i < finalStars ? "text-yellow-400 fill-yellow-400" : "text-gray-200 fill-gray-200"} />
            ))}
          </div>

          <div className="grid grid-cols-2 gap-4 w-full">
            <div className="bg-emerald-50 rounded-2xl p-4 text-center">
              <p className="text-xs text-emerald-600 font-bold uppercase">Butterflies</p>
              <p className="text-3xl font-black text-emerald-700">{butterfliesCaught}</p>
            </div>
            <div className="bg-amber-50 rounded-2xl p-4 text-center">
              <p className="text-xs text-amber-600 font-bold uppercase">Bees</p>
              <p className="text-3xl font-black text-amber-700">{beesTapped}</p>
            </div>
          </div>

          <div className="text-center">
            <p className="text-sm font-bold text-gray-500">XP GAINED</p>
            <p className="text-4xl font-black text-blue-600">+{result?.xp_earned ?? 0}</p>
          </div>

          <div className="flex gap-4 w-full">
            <button onClick={() => window.location.reload()} className="flex-1 bg-emerald-500 hover:bg-emerald-600 active:translate-y-1 text-white font-bold py-4 rounded-3xl shadow-[0_6px_0_#059669]">
              <RotateCcw size={24} className="inline mr-2" /> Again
            </button>
            <Link href="/island" className="flex-1">
              <button className="w-full bg-blue-500 hover:bg-blue-600 active:translate-y-1 text-white font-bold py-4 rounded-3xl shadow-[0_6px_0_#2563eb]">
                <Home size={24} className="inline mr-2" /> Island
              </button>
            </Link>
          </div>
          {saving && <p className="text-xs text-gray-400 animate-pulse">Syncing session...</p>}
        </motion.div>
      </div>
    );
  }

  return (
    <div className="relative min-h-screen flex flex-col overflow-hidden select-none bg-emerald-900">
      {/* Background */}
      <div className="absolute inset-0 z-0">
        <Image src="/assets/images/bg-focus-forest.jpg" alt="Focus Forest" fill className="object-cover" priority />
      </div>

      {/* HUD Header */}
      <div className="relative z-30 p-6 flex justify-between items-start w-full max-w-7xl mx-auto">
        <div className="flex flex-col gap-2">
          {/* Timer Pill */}
          <div className="bg-white/90 backdrop-blur rounded-full px-5 py-2 shadow-xl border-2 border-white/50 flex items-center gap-3">
            <Clock size={24} className={timeSeconds <= 10 ? "text-red-500 animate-pulse" : "text-emerald-600"} />
            <span className={`text-2xl font-black ${timeSeconds <= 10 ? "text-red-500" : "text-gray-700"}`}>{timeSeconds}s</span>
          </div>
          {/* Progress Bar */}
          <div className="w-48 h-3 bg-black/20 rounded-full overflow-hidden border border-white/30">
            <motion.div 
              className={`h-full ${timeSeconds <= 10 ? 'bg-red-500' : 'bg-emerald-400'}`}
              animate={{ width: `${timePercent}%` }}
              transition={{ duration: 0.1 }}
            />
          </div>
        </div>

        {/* Score Pill */}
        <div className="flex flex-col items-center gap-2">
          <div className="bg-white/90 backdrop-blur rounded-full px-8 py-3 shadow-xl border-2 border-white/50 flex items-center gap-3">
             <span className="text-4xl">🦋</span>
             <span className="text-3xl font-black text-emerald-800">{butterfliesCaught}</span>
          </div>
          <div className="flex gap-1">
             {[0, 1, 2].map(i => (
               <Star key={i} size={28} className={`${i < calcStars() ? 'text-yellow-400 fill-yellow-400' : 'text-white/30 fill-white/10'} drop-shadow-lg`} />
             ))}
          </div>
        </div>
      </div>

      {/* Game Stage */}
      <div 
        ref={gameAreaRef}
        className="relative flex-1 z-10 cursor-crosshair"
        onClick={handleMiss}
      >
        <AnimatePresence>
          {targets.map((t) => (
            <TargetComponent 
              key={t.id} 
              target={t} 
              onCatch={catchTarget} 
            />
          ))}
        </AnimatePresence>
      </div>

      {/* Loading Overlay */}
      {phase === "loading" && (
        <div className="absolute inset-0 z-50 flex items-center justify-center bg-black/40 backdrop-blur-md">
          <motion.div 
            animate={{ scale: [1, 1.1, 1] }}
            transition={{ repeat: Infinity, duration: 1.5 }}
            className="bg-white rounded-[3rem] px-12 py-8 shadow-2xl flex flex-col items-center gap-4"
          >
             <span className="text-7xl">🦋</span>
             <p className="text-2xl font-black text-emerald-800">Heading to the Woods...</p>
          </motion.div>
        </div>
      )}
    </div>
  );
}

function TargetComponent({ target, onCatch }: { target: Target; onCatch: (t: Target, x: number, y: number) => void }) {
  const [isCaught, setIsCaught] = useState(false);
  
  const handleTap = (e: React.MouseEvent) => {
    e.stopPropagation();
    if (isCaught) return;
    setIsCaught(true);
    
    // Calculate click percent relative to viewport (since we use percentages for positioning)
    const x = (e.clientX / window.innerWidth) * 100;
    const y = (e.clientY / window.innerHeight) * 100;
    
    onCatch(target, x, y);
  };

  return (
    <motion.div
      initial={{ left: `${target.fromX}%`, top: `${target.y}%`, scale: 0 }}
      animate={{ 
        left: `${target.toX}%`,
        scale: 1,
      }}
      exit={{ scale: 0, opacity: 0 }}
      transition={{ 
        left: { duration: target.duration, ease: "linear" },
        scale: { duration: 0.3 }
      }}
      className="absolute cursor-pointer z-20 pointer-events-auto"
      style={{ transform: "translate(-50%, -50%)" }}
      onClick={handleTap}
    >
       {/* Vertical Waving Motion */}
       <motion.div
         animate={{ 
           y: [-target.amplitude, target.amplitude, -target.amplitude],
           rotate: [target.toX > target.fromX ? 0 : 0, target.toX > target.fromX ? 20 : -20]
         }}
         transition={{ 
           y: { duration: 1.8 + Math.random(), repeat: Infinity, ease: "easeInOut" },
           rotate: { duration: 0.5, repeat: Infinity, repeatType: "reverse" }
         }}
         className="relative"
       >
         <span 
           className="text-5xl md:text-6xl drop-shadow-[0_4px_8px_rgba(0,0,0,0.3)] block"
           style={{ filter: target.isBee ? 'none' : `drop-shadow(0 0 10px ${target.color}80)` }}
         >
           {target.emoji}
         </span>
         
         {!target.isBee && (
           <motion.div 
             animate={{ scale: [1, 1.5, 1], opacity: [0, 0.5, 0] }}
             transition={{ repeat: Infinity, duration: 1 }}
             className="absolute inset-0 rounded-full bg-white/30 blur-xl -z-10"
           />
         )}
       </motion.div>
    </motion.div>
  );
}
