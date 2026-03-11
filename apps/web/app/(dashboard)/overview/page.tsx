"use client";

import React, { useEffect, useState } from 'react';
import { Download, Star, CheckCircle2, Flag, PieChart, Search, Users, BarChart3, ClipboardList, X, Clock, Zap, Target, ChevronDown } from 'lucide-react';
import { motion, AnimatePresence } from 'motion/react';
import { useAuthStore } from "@/hooks/useAuthStore";
import { setAuthCallbacks } from "@/lib/api";
import { getAnalyticsOverview, type WeeklySummary } from "@/lib/api";
import { useRouter } from "next/navigation";

const SkillCard = ({ 
  title, 
  stats, 
  status, 
  icon, 
  bgColor, 
  borderColor, 
  delay = 0,
  onClick
}: { 
  title: string; 
  stats: string; 
  status: string; 
  icon: React.ReactNode; 
  bgColor: string;
  borderColor: string;
  delay?: number;
  onClick?: () => void;
}) => (
  <motion.div 
    initial={{ opacity: 0, scale: 0.9 }}
    animate={{ opacity: 1, scale: 1 }}
    transition={{ delay, duration: 0.4 }}
    onClick={onClick}
    className={`relative ${bgColor} border-4 ${borderColor} rounded-[2rem] p-4 shadow-sm flex flex-col items-center text-center cursor-pointer hover:shadow-md transition-all active:scale-95 group`}
  >
    <div className={`absolute -top-5 left-1/2 -translate-x-1/2 h-10 px-6 ${borderColor.replace('border-', 'bg-')} rounded-full border-4 border-white flex items-center justify-center shadow-md z-20`}>
       <h3 className="text-xl font-black text-slate-700 tracking-tight whitespace-nowrap">{title}</h3>
    </div>
    
    <div className="flex items-center gap-2 w-full mt-6">
      <div className="w-28 h-28 flex items-center justify-center shrink-0">
        {icon}
      </div>
      
      <div className="flex-1 space-y-2">
        <div className="bg-white rounded-full py-2 px-3 shadow-inner border border-slate-100 flex items-center justify-center">
          <span className="text-slate-700 font-black text-xs whitespace-nowrap">{stats}</span>
        </div>
        <div className="bg-white rounded-full py-2 px-3 shadow-inner border border-slate-100 flex items-center justify-center gap-1">
          <span className="text-slate-700 font-black text-xs whitespace-nowrap">{status}</span>
        </div>
      </div>
    </div>
  </motion.div>
);

export default function DashboardOverviewPage() {
  const [showAttentionDetails, setShowAttentionDetails] = useState(false);
  const { user, accessToken, fetchUser, logout } = useAuthStore();
  const [summary, setSummary] = useState<WeeklySummary | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const router = useRouter();

  // Initialize auth callbacks
  useEffect(() => {
    setAuthCallbacks(
      () => useAuthStore.getState().accessToken,
      () => useAuthStore.getState().refreshTokens()
    );
  }, []);

  useEffect(() => {
    // No access token - redirect to login
    if (!accessToken) {
      router.replace("/login");
      return;
    }

    // Fetch user if we have token but no user
    if (accessToken && !user) {
      fetchUser().catch(() => router.replace("/login"));
      return;
    }

    if (!user) return;

    const params = new URLSearchParams(window.location.search);
    const profileId = params.get("profile_id");

    if (!profileId) {
      setIsLoading(false);
      return;
    }

    getAnalyticsOverview(profileId)
      .then(setSummary)
      .catch((err) => {
        setError(err?.message || "Failed to load analytics data");
      })
      .finally(() => setIsLoading(false));
  }, [user, accessToken, router, fetchUser]);

  const handleLogout = async () => {
    logout();
    router.push("/login");
  };

  const dashboardLabel = user?.role === "educator" ? "Educator Dashboard" : "Parent Dashboard";
  const childNickname = summary?.student_nickname || "Your child";

  function formatPercent(value: number | null): string {
    if (value === null) return "—";
    return `${Math.round(value * 100)}%`;
  }

  if (!accessToken || !user || isLoading) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-[#FFF9F0]">
        <p className="text-lg font-bold text-slate-600">Loading dashboard…</p>
      </div>
    );
  }

  const noData = !summary || summary.message === "No data yet";

  return (
    <div className="min-h-screen bg-[#FFF9F0] font-sans text-slate-900 overflow-x-hidden">
      {/* Background Beach Elements */}
      <div className="fixed inset-0 pointer-events-none overflow-hidden">
        {/* Bottom Waves */}
        <div className="absolute bottom-0 w-full h-48">
          <svg className="absolute bottom-0 w-full h-full" viewBox="0 0 1440 320" preserveAspectRatio="none">
            <path fill="#A5D8FF" fillOpacity="0.4" d="M0,160L48,176C96,192,192,224,288,224C384,224,480,192,576,165.3C672,139,768,117,864,138.7C960,160,1056,224,1152,234.7C1248,245,1344,203,1392,181.3L1440,160L1440,320L1392,320C1344,320,1248,320,1152,320C1056,320,960,320,864,320C768,320,672,320,576,320C480,320,384,320,288,320C192,320,96,320,48,320L0,320Z"></path>
            <path fill="#74C0FC" fillOpacity="0.6" d="M0,224L48,213.3C96,203,192,181,288,186.7C384,192,480,224,576,218.7C672,213,768,171,864,149.3C960,128,1056,128,1152,149.3C1248,171,1344,213,1392,234.7L1440,256L1440,320L1392,320C1344,320,1248,320,1152,320C1056,320,960,320,864,320C768,320,672,320,576,320C480,320,384,320,288,320C192,320,96,320,48,320L0,320Z"></path>
            <path fill="#339AF0" fillOpacity="0.8" d="M0,288L48,272C96,256,192,224,288,224C384,224,480,256,576,266.7C672,277,768,267,864,245.3C960,224,1056,192,1152,186.7C1248,181,1344,203,1392,213.3L1440,224L1440,320L1392,320C1344,320,1248,320,1152,320C1056,320,960,320,864,320C768,320,672,320,576,320C480,320,384,320,288,320C192,320,96,320,48,320L0,320Z"></path>
          </svg>
        </div>
      </div>

      <AnimatePresence>
        {showAttentionDetails && (
          <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
            <motion.div 
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              onClick={() => setShowAttentionDetails(false)}
              className="absolute inset-0 bg-slate-900/40 backdrop-blur-sm"
            />
            <motion.div 
              initial={{ opacity: 0, scale: 0.9, y: 20 }}
              animate={{ opacity: 1, scale: 1, y: 0 }}
              exit={{ opacity: 0, scale: 0.9, y: 20 }}
              className="relative w-full max-w-lg bg-white rounded-[3rem] shadow-2xl overflow-hidden border-8 border-[#FFD43B]"
            >
              <div className="bg-[#FFF3BF] p-8 flex justify-between items-center border-b-4 border-[#FFD43B]">
                <div className="flex items-center gap-4">
                  <div className="bg-white p-3 rounded-2xl shadow-sm">
                    <Search className="w-8 h-8 text-amber-600" />
                  </div>
                  <h2 className="text-3xl font-black text-slate-800">Attention Details</h2>
                </div>
                <button onClick={() => setShowAttentionDetails(false)} className="p-2 hover:bg-black/5 rounded-full">
                  <X className="w-8 h-8 text-slate-600" />
                </button>
              </div>

              <div className="p-10 space-y-10">
                <div className="grid grid-cols-2 gap-6">
                  <div className="bg-amber-50 p-6 rounded-[2rem] border-4 border-amber-100">
                    <div className="flex items-center gap-2 text-amber-600 mb-2">
                      <Clock className="w-5 h-5" />
                      <span className="text-xs font-black uppercase tracking-widest">Avg. Focus</span>
                    </div>
                    <div className="text-3xl font-black text-slate-800">{noData ? "—" : "45m"}</div>
                  </div>
                  <div className="bg-emerald-50 p-6 rounded-[2rem] border-4 border-emerald-100">
                    <div className="flex items-center gap-2 text-emerald-600 mb-2">
                      <Zap className="w-5 h-5" />
                      <span className="text-xs font-black uppercase tracking-widest">Peak Time</span>
                    </div>
                    <div className="text-3xl font-black text-slate-800">{noData ? "—" : "10:30 AM"}</div>
                  </div>
                </div>

                <div className="bg-slate-50 p-8 rounded-[2rem] border-4 border-slate-100">
                  <h4 className="text-sm font-black text-slate-400 uppercase tracking-[0.2em] mb-6">Focus Breakdown</h4>
                  {noData ? (
                    <p className="text-slate-500 font-medium">No detailed focus breakdown is available yet.</p>
                  ) : (
                    <div className="space-y-6">
                      {[
                        { label: 'Pattern Matching', value: 85, color: 'bg-amber-400' },
                        { label: 'Visual Tracking', value: 62, color: 'bg-blue-400' },
                        { label: 'Auditory Focus', value: 45, color: 'bg-rose-400' },
                      ].map((item) => (
                        <div key={item.label}>
                          <div className="flex justify-between text-sm font-black text-slate-700 mb-2">
                            <span>{item.label}</span>
                            <span>{item.value}%</span>
                          </div>
                          <div className="h-4 w-full bg-white rounded-full overflow-hidden border-2 border-slate-100 shadow-inner">
                            <motion.div 
                              initial={{ width: 0 }}
                              animate={{ width: `${item.value}%` }}
                              transition={{ duration: 1, delay: 0.2 }}
                              className={`h-full ${item.color}`}
                            />
                          </div>
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              </div>

              <div className="p-8 bg-slate-50 border-t-4 border-slate-100 flex justify-end">
                <button 
                  onClick={() => setShowAttentionDetails(false)}
                  className="px-10 py-3 bg-slate-800 text-white font-black rounded-2xl hover:bg-slate-700 transition-colors shadow-lg"
                >
                  CLOSE
                </button>
              </div>
            </motion.div>
          </div>
        )}
      </AnimatePresence>

      {/* Header */}
      <header className="relative pt-12 pb-8 flex justify-between items-center z-10 px-6 sm:px-12 flex-col sm:flex-row gap-6">
        <div className="relative flex items-center">
          {/* Logo Banner */}
          <div className="bg-[#FFF3BF] border-4 border-[#FFD43B] rounded-2xl px-10 py-4 shadow-xl flex items-center gap-4 transform -rotate-1 relative">
            <div className="absolute -top-12 -left-8 w-24 h-24 hidden sm:block">
               <svg viewBox="0 0 100 100" className="w-full h-full drop-shadow-xl">
                 {/* Palm Tree */}
                 <path d="M40 85 L35 95 L45 95 Z" fill="#8B4513" />
                 <path d="M40 85 C20 85 10 65 40 25 C70 65 60 85 40 85" fill="#2D5A27" />
                 {/* Treasure Chest */}
                 <rect x="55" y="75" width="25" height="15" fill="#5D4037" rx="2" />
                 <path d="M55 75 Q67.5 65 80 75" fill="#8D6E63" />
                 <circle cx="67.5" cy="82.5" r="2" fill="#FFD700" />
               </svg>
            </div>
            <h1 className="text-4xl sm:text-5xl font-black text-[#1C7ED6] tracking-tighter drop-shadow-md italic">Skill Island</h1>
          </div>
        </div>

        <div className="flex items-center gap-4 flex-col sm:flex-row">
          <div className="bg-white border-4 border-[#74C0FC] rounded-full px-6 sm:px-10 py-3 shadow-lg flex items-center gap-4 relative">
            <h2 className="text-xl sm:text-2xl font-black text-[#1C7ED6] italic">{dashboardLabel}</h2>
            <div className="absolute -right-16 -top-8 w-20 h-20 hidden sm:block">
               <svg viewBox="0 0 100 100" className="w-full h-full drop-shadow-sm">
                 {/* Seagull */}
                 <path d="M20 40 Q40 20 60 40 Q80 20 100 40" stroke="#1C7ED6" strokeWidth="4" fill="none" strokeLinecap="round" />
                 <circle cx="60" cy="35" r="2" fill="#1C7ED6" />
               </svg>
            </div>
          </div>
          <button 
            onClick={handleLogout}
            className="rounded-full bg-white/50 hover:bg-white/80 border-4 border-[#74C0FC] px-6 py-3 font-black text-[#1C7ED6] shadow-sm transition-all active:scale-95"
          >
            Logout
          </button>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-6xl mx-auto px-8 pb-24 relative z-10">
        <div className="text-center mb-8">
          <div className="flex items-center justify-center gap-6 mb-4 mt-8">
            <div className="h-[2px] w-20 sm:w-32 bg-slate-200 border-t-2 border-dotted" />
            <h2 className="text-xl font-black text-slate-600 uppercase tracking-[0.2em]">{childNickname}'s Weekly Summary</h2>
            <div className="h-[2px] w-20 sm:w-32 bg-slate-200 border-t-2 border-dotted" />
          </div>
          {noData && (
            <p className="text-slate-500 font-bold max-w-lg mx-auto mt-4 bg-white/50 p-4 rounded-xl">
              No analytics data available yet. Data will appear after {childNickname} plays some game sessions.
            </p>
          )}
          {error && (
            <p className="text-rose-500 font-bold max-w-lg mx-auto mt-4 bg-rose-50 p-4 rounded-xl border-2 border-rose-200">
              {error}
            </p>
          )}
        </div>

        {/* Row 1 */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-10 mb-12">
          <SkillCard 
            title="Attention"
            stats={`Focus Score: ${noData ? "—" : formatPercent(summary?.attention_score ?? null)}`}
            status={noData ? "Play to unlock" : "Great Concentration! ⭐"}
            bgColor="bg-[#FFF9DB]"
            borderColor="border-[#FFD43B]"
            delay={0.1}
            onClick={() => { if (!noData) setShowAttentionDetails(true) }}
            icon={
              <div className="relative w-full h-full flex items-center justify-center">
                <svg viewBox="0 0 100 100" className="w-24 h-24 drop-shadow-sm">
                  {/* Owl */}
                  <circle cx="50" cy="50" r="40" fill="#E9ECEF" stroke="#8B4513" strokeWidth="2" />
                  <circle cx="35" cy="40" r="10" fill="white" stroke="#8B4513" strokeWidth="1" />
                  <circle cx="65" cy="40" r="10" fill="white" stroke="#8B4513" strokeWidth="1" />
                  <circle cx="35" cy="40" r="4" fill="#8B4513" />
                  <circle cx="65" cy="40" r="4" fill="#8B4513" />
                  <path d="M45 60 Q50 70 55 60" fill="#FFD43B" stroke="#8B4513" strokeWidth="1" />
                  {/* Clock Overlay */}
                  <circle cx="80" cy="80" r="15" fill="white" stroke="#FFD43B" strokeWidth="2" />
                  <path d="M80 70 L80 80 L85 80" stroke="#8B4513" strokeWidth="2" strokeLinecap="round" />
                </svg>
              </div>
            }
          />
          <SkillCard 
            title="Memory"
            stats={`Memory Score: ${noData ? "—" : formatPercent(summary?.memory_score ?? null)}`}
            status={noData ? "Play to unlock" : "Good Recall Skills! ✅"}
            bgColor="bg-[#EBFBEE]"
            borderColor="border-[#69DB7C]"
            delay={0.2}
            icon={
              <div className="relative w-full h-full flex items-center justify-center">
                <svg viewBox="0 0 100 100" className="w-24 h-24 drop-shadow-sm">
                  {/* Elephant */}
                  <circle cx="50" cy="50" r="40" fill="#D0EBFF" stroke="#1C7ED6" strokeWidth="2" />
                  <circle cx="35" cy="40" r="4" fill="#1C7ED6" />
                  <path d="M20 50 Q0 50 10 70" stroke="#1C7ED6" strokeWidth="4" fill="none" />
                  {/* Puzzle Overlay */}
                  <rect x="65" y="65" width="20" height="20" fill="#69DB7C" rx="4" />
                  <circle cx="75" cy="65" r="4" fill="#69DB7C" />
                </svg>
              </div>
            }
          />
          <SkillCard 
            title="Social Engagement"
            stats={`Sessions this week: ${noData ? "0" : (summary?.sessions_this_week ?? 0)}`}
            status={noData ? "Play to unlock" : "Very Cooperative! ❤️"}
            bgColor="bg-[#FFF5F5]"
            borderColor="border-[#FF8787]"
            delay={0.3}
            icon={
              <div className="relative w-full h-full flex items-center justify-center">
                <svg viewBox="0 0 100 100" className="w-24 h-24 drop-shadow-sm">
                  {/* Two Kids */}
                  <circle cx="35" cy="50" r="25" fill="#FFF3BF" stroke="#FF8787" strokeWidth="2" />
                  <circle cx="65" cy="50" r="25" fill="#FFD8A8" stroke="#FF8787" strokeWidth="2" />
                  <circle cx="30" cy="45" r="2" fill="#FF8787" />
                  <circle cx="40" cy="45" r="2" fill="#FF8787" />
                  <circle cx="60" cy="45" r="2" fill="#FF8787" />
                  <circle cx="70" cy="45" r="2" fill="#FF8787" />
                  <path d="M30 55 Q35 60 40 55" stroke="#FF8787" strokeWidth="1" fill="none" />
                  <path d="M60 55 Q65 60 70 55" stroke="#FF8787" strokeWidth="1" fill="none" />
                </svg>
              </div>
            }
          />
        </div>

        {/* Row 2 */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-10 max-w-4xl mx-auto mb-16">
          <SkillCard 
            title="Progress"
            stats={`Total XP: ${noData ? "0" : (summary?.total_xp ?? 0)}`}
            status={noData ? "Play to unlock" : "Keep it Up! 🚩"}
            bgColor="bg-[#E7F5FF]"
            borderColor="border-[#74C0FC]"
            delay={0.4}
            icon={
              <div className="relative w-full h-full flex items-center justify-center">
                <svg viewBox="0 0 100 100" className="w-24 h-24 drop-shadow-sm">
                  {/* Chart & Flag */}
                  <rect x="20" y="60" width="10" height="20" fill="#74C0FC" />
                  <rect x="35" y="40" width="10" height="40" fill="#74C0FC" />
                  <rect x="50" y="20" width="10" height="60" fill="#74C0FC" />
                  <path d="M70 20 L70 80" stroke="#FF6B6B" strokeWidth="4" strokeLinecap="round" />
                  <path d="M70 20 L90 30 L70 40" fill="#FF6B6B" />
                </svg>
              </div>
            }
          />
          <SkillCard 
            title="Reports"
            stats={`Total Stars: ${noData ? "0" : (summary?.total_stars ?? 0)}`}
            status="Detailed Insights! 📊"
            bgColor="bg-[#F3F0FF]"
            borderColor="border-[#B197FC]"
            delay={0.5}
            icon={
              <div className="relative w-full h-full flex items-center justify-center">
                <svg viewBox="0 0 100 100" className="w-24 h-24 drop-shadow-sm">
                  {/* Clipboard & Pie Chart */}
                  <rect x="25" y="20" width="50" height="60" fill="white" stroke="#B197FC" strokeWidth="2" rx="4" />
                  <rect x="40" y="15" width="20" height="10" fill="#B197FC" rx="2" />
                  <circle cx="50" cy="50" r="15" fill="#D0BFFF" />
                  <path d="M50 50 L50 35 A15 15 0 0 1 65 50 Z" fill="#B197FC" />
                  <circle cx="75" cy="75" r="12" fill="white" stroke="#B197FC" strokeWidth="2" />
                  <path d="M70 70 L80 80" stroke="#B197FC" strokeWidth="3" strokeLinecap="round" />
                </svg>
              </div>
            }
          />
        </div>

        {/* Download Button */}
        <div className="flex justify-center">
          <motion.button 
            whileHover={{ scale: 1.05, y: -5 }}
            whileTap={{ scale: 0.95 }}
            className="bg-[#FFF9DB] border-8 border-[#E9ECEF] rounded-[4rem] px-8 sm:px-20 py-6 sm:py-8 shadow-2xl flex items-center gap-4 sm:gap-8 group hover:border-[#FFD43B] transition-all relative overflow-hidden"
          >
            <div className="absolute inset-0 bg-white/20 opacity-0 group-hover:opacity-100 transition-opacity" />
            <div className="bg-[#8B4513] p-4 rounded-2xl shadow-xl group-hover:scale-110 transition-transform z-10 hidden sm:block">
              <Download className="w-12 h-12 text-white" />
            </div>
            <span className="text-3xl sm:text-5xl font-black text-slate-800 tracking-tighter z-10 italic">Download Report</span>
          </motion.button>
        </div>
      </main>

      {/* Footer Dotted Line */}
      <div className="max-w-6xl mx-auto px-8 pb-8 relative z-10">
        <div className="border-t-4 border-dotted border-slate-200" />
      </div>
    </div>
  );
}
