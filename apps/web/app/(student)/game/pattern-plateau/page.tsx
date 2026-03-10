"use client";

import React, { useState } from 'react';
import { motion, AnimatePresence } from 'motion/react';
import { Star, X as XIcon, Circle as CircleIcon, Trophy, Play, Home } from 'lucide-react';
import Image from 'next/image';
import Link from 'next/link';

type Player = 'X' | 'O' | null;
type Winner = Player | 'Draw';

const WINNING_COMBINATIONS = [
  [0, 1, 2], [3, 4, 5], [6, 7, 8], // Rows
  [0, 3, 6], [1, 4, 7], [2, 5, 8], // Columns
  [0, 4, 8], [2, 4, 6]             // Diagonals
];

export default function PatternPlateauPage() {
  const [board, setBoard] = useState<Player[]>(Array(9).fill(null));
  const [isXNext, setIsXNext] = useState(true);
  const [winner, setWinner] = useState<Winner>(null);
  const [winningLine, setWinningLine] = useState<number[] | null>(null);
  const [scoreX, setScoreX] = useState(0);
  const [scoreO, setScoreO] = useState(0);
  const [roundCompleted, setRoundCompleted] = useState(false);

  const checkWinner = (currentBoard: Player[]) => {
    for (const combo of WINNING_COMBINATIONS) {
      const [a, b, c] = combo;
      if (currentBoard[a] && currentBoard[a] === currentBoard[b] && currentBoard[a] === currentBoard[c]) {
        return { winner: currentBoard[a], combo };
      }
    }
    // Check for draw
    if (!currentBoard.includes(null)) {
      return { winner: 'Draw', combo: null };
    }
    return null;
  };

  const handleClick = (index: number) => {
    if (board[index] || winner || roundCompleted) return;

    const newBoard = [...board];
    newBoard[index] = isXNext ? 'X' : 'O';
    setBoard(newBoard);

    const result = checkWinner(newBoard);
    if (result) {
      if (result.winner === 'Draw') {
        setWinner('Draw');
        setRoundCompleted(true);
      } else {
        setWinner(result.winner as Winner);
        setWinningLine(result.combo);
        setRoundCompleted(true);
        if (result.winner === 'X') setScoreX(s => s + 1);
        else setScoreO(s => s + 1);
      }
    } else {
      setIsXNext(!isXNext);
    }
  };

  const resetGame = () => {
    setBoard(Array(9).fill(null));
    setIsXNext(true);
    setWinner(null);
    setWinningLine(null);
    setRoundCompleted(false);
  };

  const getWinningLineStyle = () => {
    if (!winningLine) return {};
    const [a, b, c] = winningLine;
    
    // Horizontal
    if (a === 0 && c === 2) return { top: '16.6%', left: '-5%', right: '-5%', height: '24px', borderRadius: '12px' };
    if (a === 3 && c === 5) return { top: '50%', left: '-5%', right: '-5%', height: '24px', transform: 'translateY(-50%)', borderRadius: '12px' };
    if (a === 6 && c === 8) return { top: '83.3%', left: '-5%', right: '-5%', height: '24px', borderRadius: '12px' };
    
    // Vertical
    if (a === 0 && c === 6) return { left: '16.6%', top: '-5%', bottom: '-5%', width: '24px', borderRadius: '12px' };
    if (a === 1 && c === 7) return { left: '50%', top: '-5%', bottom: '-5%', width: '24px', transform: 'translateX(-50%)', borderRadius: '12px' };
    if (a === 2 && c === 8) return { left: '83.3%', top: '-5%', bottom: '-5%', width: '24px', borderRadius: '12px' };
    
    // Diagonal
    if (a === 0 && c === 8) return { top: '50%', left: '50%', width: '130%', height: '24px', transform: 'translate(-50%, -50%) rotate(45deg)', borderRadius: '12px' };
    if (a === 2 && c === 6) return { top: '50%', left: '50%', width: '130%', height: '24px', transform: 'translate(-50%, -50%) rotate(-45deg)', borderRadius: '12px' };
    
    return {};
  };

  return (
    <div className="relative min-h-screen flex flex-col items-center justify-center p-4 overflow-hidden select-none">
      {/* Background Image full screen */}
      <div className="absolute inset-0 z-0 bg-[#A6E5FF]">
        <Image
          src="/assets/images/bg-pattern-plateau.jpg"
          alt="Pattern Plateau Background"
          fill
          className="object-cover"
          priority
        />
        <div className="absolute inset-0 bg-blue-100/30" />
      </div>

      {/* Confetti Effect (Simple) */}
      {roundCompleted && (
        <div className="fixed inset-0 pointer-events-none z-50 overflow-hidden">
          {[...Array(40)].map((_, i) => (
            <motion.div
              key={i}
              initial={{ 
                top: '-10%', 
                left: `${Math.random() * 100}%`,
                rotate: 0,
                scale: Math.random() * 0.8 + 0.5
              }}
              animate={{ 
                top: '110%', 
                rotate: Math.random() * 720 - 360,
                left: `${Math.random() * 100}%`
              }}
              transition={{ 
                duration: Math.random() * 3 + 2, 
                repeat: winner ? Infinity : 0,
                ease: "linear",
                delay: Math.random() * 0.5
              }}
              className={`absolute w-3 h-3 md:w-4 md:h-4 rounded-sm ${
                ['bg-[#FF5C5C]', 'bg-[#FFDE59]', 'bg-[#5CE1E6]', 'bg-[#7ED957]', 'bg-[#CB6CE6]'][i % 5]
              }`}
            />
          ))}
        </div>
      )}

      {/* Main Container */}
      <div className="relative z-10 flex flex-col items-center w-full max-w-lg mt-4">
        
        {/* Top Progress Bar */}
        <div className="relative w-full max-w-sm h-10 bg-[#0288D1]/30 rounded-full border-4 border-white mb-8 flex items-center shadow-[0_4px_10px_rgba(0,0,0,0.15)] overflow-visible">
          <div className="h-full w-full rounded-full overflow-hidden absolute inset-0">
            <motion.div 
              className="h-full bg-gradient-to-r from-[#4FC3F7] to-[#FFDE59] rounded-full"
              initial={{ width: '0%' }}
              animate={{ width: `${Math.min(((scoreX + scoreO) % 10) * 10 || 10, 100)}%` }}
              transition={{ duration: 0.5 }}
            />
          </div>
          {/* Star Cap */}
          <div className="absolute right-0 top-1/2 -translate-y-1/2 translate-x-1/2 bg-[#FFDE59] rounded-full p-1.5 border-[3px] border-white shadow-md z-10">
            <Star className="text-white fill-white" size={24} />
          </div>
        </div>

        {/* Board Enclosure */}
        <div className="relative flex flex-col items-center">
          
          {/* Top Curve Banner */}
          {!roundCompleted && (
            <div className="absolute -top-10 z-30 flex items-center justify-center">
              <div className="relative bg-[#A1E35E] border-4 border-white shadow-[0_6px_0_rgba(100,180,50,1)] rounded-3xl px-12 py-3 flex items-center justify-center -skew-x-2 transform transition-transform">
                <div className="absolute -left-6 top-1/2 -translate-y-1/2 w-16 h-16 rounded-full border-4 border-white bg-[#D6F0FF] overflow-hidden shadow-xl z-20 flex items-center justify-center scale-110">
                  <img src="https://api.dicebear.com/7.x/avataaars/svg?seed=Felix&backgroundColor=b6e3f4" alt="Player Avatar" className="w-full h-full object-cover" />
                  <div className="absolute bottom-0 right-0 bg-[#42A5F5] rounded-full p-[2px] border-2 border-white shadow-sm">
                    {isXNext ? <XIcon size={12} className="text-white stroke-[3]"/> : <CircleIcon size={12} className="text-white stroke-[3]"/>}
                  </div>
                </div>
                <span className="text-white text-3xl font-bolder font-extrabold drop-shadow-[0_2px_1px_rgba(0,0,0,0.3)] tracking-wide ml-4 translate-y-[1px]">
                  {isXNext ? "X's Turn!" : "O's Turn!"}
                </span>
                <Star className="absolute right-2 top-2 text-[#FFDE59] fill-[#FFDE59] scale-75 opacity-80" size={16} />
                <Star className="absolute right-8 bottom-1 text-[#FFDE59] fill-[#FFDE59] scale-50 opacity-90" size={16} />
              </div>
            </div>
          )}

          {/* The Game Grid Canvas */}
          <div className="relative bg-[#FFF9ED] backdrop-blur-md p-6 sm:p-8 rounded-[2rem] border-[6px] border-[#FFE9B8] shadow-[0_25px_50px_rgba(0,0,0,0.15)] mt-4 z-20">
            <div className="grid grid-cols-3 gap-2 sm:gap-4 relative z-10">
              {/* Grid Lines (CSS based instead of borders on every box) */}
              <div className="absolute top-1/3 left-0 right-0 h-1 sm:h-2 bg-[#F1DEB4] -translate-y-1/2 rounded-full opacity-60 pointer-events-none" />
              <div className="absolute top-2/3 left-0 right-0 h-1 sm:h-2 bg-[#F1DEB4] -translate-y-1/2 rounded-full opacity-60 pointer-events-none" />
              <div className="absolute left-1/3 top-0 bottom-0 w-1 sm:w-2 bg-[#F1DEB4] -translate-x-1/2 rounded-full opacity-60 pointer-events-none" />
              <div className="absolute left-2/3 top-0 bottom-0 w-1 sm:w-2 bg-[#F1DEB4] -translate-x-1/2 rounded-full opacity-60 pointer-events-none" />

              {board.map((cell, i) => (
                <motion.button
                  key={i}
                  whileHover={!cell && !roundCompleted ? { scale: 1.05 } : {}}
                  whileTap={!cell && !roundCompleted ? { scale: 0.95 } : {}}
                  onClick={() => handleClick(i)}
                  className="w-20 h-20 sm:w-28 sm:h-28 flex items-center justify-center relative focus:outline-none z-10 bg-transparent rounded-2xl"
                >
                  <AnimatePresence>
                    {cell === 'X' && (
                      <motion.div
                        initial={{ scale: 0, rotate: -45 }}
                        animate={{ scale: 1, rotate: 0 }}
                        className="text-[#42A5F5]"
                      >
                        <XIcon size={80} strokeWidth={3} className="drop-shadow-[0_4px_2px_rgba(66,165,245,0.4)]" />
                      </motion.div>
                    )}
                    {cell === 'O' && (
                      <motion.div
                        initial={{ scale: 0 }}
                        animate={{ scale: 1 }}
                        className="text-[#FFA726]"
                      >
                        <CircleIcon size={70} strokeWidth={4} className="drop-shadow-[0_4px_2px_rgba(255,167,38,0.4)]" />
                      </motion.div>
                    )}
                  </AnimatePresence>
                </motion.button>
              ))}
            </div>

            {/* Glowing Winning Line Overlay */}
            {winner && winner !== 'Draw' && (
              <motion.div 
                initial={{ opacity: 0, scale: 0 }}
                animate={{ opacity: 1, scale: 1 }}
                transition={{ type: "spring", stiffness: 100, damping: 10 }}
                className="absolute z-30 shadow-[0_0_20px_rgba(255,222,89,0.8)]"
                style={{
                  ...getWinningLineStyle(),
                  background: 'linear-gradient(90deg, #FFDE59 0%, #FF9900 100%)',
                }}
              >
                {/* Sparkles on the line */}
                <Star className="absolute top-1/2 left-1/4 -translate-y-1/2 text-white fill-white scale-50 opacity-80" size={24} />
                <Star className="absolute top-1/2 left-3/4 -translate-y-1/2 text-white fill-white scale-75 opacity-90" size={24} />
              </motion.div>
            )}
          </div>

          {/* Winner Banner (Overlaps the bottom of the board) */}
          <AnimatePresence>
            {roundCompleted && (
              <motion.div 
                initial={{ opacity: 0, scale: 0.5, y: -20 }}
                animate={{ opacity: 1, scale: 1, y: 0 }}
                className="absolute -bottom-6 z-40"
              >
                <div className="bg-[#4CAF50] border-[3px] border-white shadow-[0_8px_0_rgba(50,150,50,1),0_15px_20px_rgba(0,0,0,0.3)] rounded-full px-8 py-2 flex items-center justify-center transform -skew-x-2 gap-3">
                  <Trophy className="text-[#FFDE59] fill-[#FFDE59]" size={24} />
                  <span className="text-white text-3xl font-extrabold drop-shadow-[0_2px_1px_rgba(0,0,0,0.4)] tracking-wider">
                    {winner === 'Draw' ? "It's a Draw!" : `${winner} Wins!`}
                  </span>
                  <Trophy className="text-[#FFDE59] fill-[#FFDE59]" size={24} />
                </div>
              </motion.div>
            )}
          </AnimatePresence>
        </div>

        {/* Footer Actions / Players */}
        <div className="mt-16 w-full flex justify-center z-30">
          {!roundCompleted ? (
            <div className="flex gap-4 sm:gap-8">
              {/* Player 1 Pill */}
              <div className={`flex items-center bg-[#42A5F5] rounded-full pl-1.5 pr-6 sm:pr-8 py-2 border-[3px] shadow-[0_6px_0_#1E88E5,0_10px_20px_rgba(0,0,0,0.2)] gap-3 transition-all duration-300 ${isXNext ? 'border-[#FFDE59] scale-110 shadow-[0_6px_0_#1E88E5,0_0_20px_rgba(255,222,89,0.5)]' : 'border-white'}`}>
                <div className="bg-white rounded-full p-1.5 shadow-inner flex items-center justify-center">
                  <XIcon size={24} className="text-[#42A5F5] stroke-[4]" />
                </div>
                <span className="text-white font-extrabold text-xl md:text-2xl drop-shadow-[0_1px_1px_rgba(0,0,0,0.3)]">Player 1</span>
              </div>

              {/* Player 2 Pill */}
              <div className={`flex items-center bg-[#FFA726] rounded-full pl-1.5 pr-6 sm:pr-8 py-2 border-[3px] shadow-[0_6px_0_#F57C00,0_10px_20px_rgba(0,0,0,0.2)] gap-3 transition-all duration-300 ${!isXNext ? 'border-[#FFDE59] scale-110 shadow-[0_6px_0_#F57C00,0_0_20px_rgba(255,222,89,0.5)]' : 'border-white'}`}>
                <div className="bg-white rounded-full p-1.5 shadow-inner flex items-center justify-center">
                  <CircleIcon size={24} className="text-[#FFA726] stroke-[4]" />
                </div>
                <span className="text-white font-extrabold text-xl md:text-2xl drop-shadow-[0_1px_1px_rgba(0,0,0,0.3)]">Player 2</span>
              </div>
            </div>
          ) : (
            <div className="flex gap-4 sm:gap-6">
              {/* Play Again Button */}
              <motion.button
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                whileHover={{ scale: 1.05 }}
                whileTap={{ scale: 0.95 }}
                onClick={resetGame}
                className="flex items-center bg-[#FFDE59] rounded-full pl-3 pr-6 sm:pr-8 py-2.5 border-[3px] border-white shadow-[0_6px_0_#D4B227,0_10px_20px_rgba(0,0,0,0.2)] gap-2 hover:bg-[#FFE87A] transition-colors"
              >
                <Play className="text-[#8FB339] fill-[#8FB339]" size={24} />
                <span className="text-[#7A9E27] font-extrabold text-xl md:text-2xl">Play Again</span>
              </motion.button>

              {/* Home Button */}
              <Link href="/island" passHref>
                <motion.a
                  initial={{ opacity: 0, y: 20 }}
                  animate={{ opacity: 1, y: 0 }}
                  whileHover={{ scale: 1.05 }}
                  whileTap={{ scale: 0.95 }}
                  className="flex items-center bg-[#42A5F5] rounded-full pl-3 pr-6 sm:pr-8 py-2.5 border-[3px] border-white shadow-[0_6px_0_#1E88E5,0_10px_20px_rgba(0,0,0,0.2)] gap-2 hover:bg-[#64B5F6] transition-colors cursor-pointer"
                >
                  <Home className="text-white" size={24} />
                  <span className="text-white font-extrabold text-xl md:text-2xl">Home</span>
                </motion.a>
              </Link>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
