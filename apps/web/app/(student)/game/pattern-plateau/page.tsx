"use client";

import React, { useState } from 'react';
import { motion, AnimatePresence } from 'motion/react';
import { Star, X as XIcon, Circle as CircleIcon, Trophy } from 'lucide-react';
import Image from 'next/image';

type Player = 'X' | 'O' | null;

const WINNING_COMBINATIONS = [
  [0, 1, 2], [3, 4, 5], [6, 7, 8], // Rows
  [0, 3, 6], [1, 4, 7], [2, 5, 8], // Columns
  [0, 4, 8], [2, 4, 6]             // Diagonals
];

export default function PatternPlateauPage() {
  const [board, setBoard] = useState<Player[]>(Array(9).fill(null));
  const [isXNext, setIsXNext] = useState(true);
  const [winner, setWinner] = useState<Player>(null);
  const [winningLine, setWinningLine] = useState<number[] | null>(null);
  const [scoreX, setScoreX] = useState(0);
  const [scoreO, setScoreO] = useState(0);

  const checkWinner = (currentBoard: Player[]) => {
    for (const combo of WINNING_COMBINATIONS) {
      const [a, b, c] = combo;
      if (currentBoard[a] && currentBoard[a] === currentBoard[b] && currentBoard[a] === currentBoard[c]) {
        return { winner: currentBoard[a], combo };
      }
    }
    return null;
  };

  const handleClick = (index: number) => {
    if (board[index] || winner) return;

    const newBoard = [...board];
    newBoard[index] = isXNext ? 'X' : 'O';
    setBoard(newBoard);

    const result = checkWinner(newBoard);
    if (result) {
      setWinner(result.winner);
      setWinningLine(result.combo);
      if (result.winner === 'X') setScoreX(s => s + 1);
      else setScoreO(s => s + 1);
    } else {
      setIsXNext(!isXNext);
    }
  };

  const resetGame = () => {
    setBoard(Array(9).fill(null));
    setIsXNext(true);
    setWinner(null);
    setWinningLine(null);
  };

  const getWinningLineStyle = () => {
    if (!winningLine) return {};
    const [a, b, c] = winningLine;
    
    // Horizontal
    if (a === 0 && c === 2) return { top: '16.6%', left: '5%', right: '5%', height: '8px' };
    if (a === 3 && c === 5) return { top: '50%', left: '5%', right: '5%', height: '8px', transform: 'translateY(-50%)' };
    if (a === 6 && c === 8) return { top: '83.3%', left: '5%', right: '5%', height: '8px' };
    
    // Vertical
    if (a === 0 && c === 6) return { left: '16.6%', top: '5%', bottom: '5%', width: '8px' };
    if (a === 1 && c === 7) return { left: '50%', top: '5%', bottom: '5%', width: '8px', transform: 'translateX(-50%)' };
    if (a === 2 && c === 8) return { left: '83.3%', top: '5%', bottom: '5%', width: '8px' };
    
    // Diagonal
    if (a === 0 && c === 8) return { top: '50%', left: '50%', width: '120%', height: '8px', transform: 'translate(-50%, -50%) rotate(45deg)' };
    if (a === 2 && c === 6) return { top: '50%', left: '50%', width: '120%', height: '8px', transform: 'translate(-50%, -50%) rotate(-45deg)' };
    
    return {};
  };

  return (
    <div className="relative min-h-screen flex flex-col items-center justify-center overflow-hidden p-4">
      {/* Background Image */}
      <div className="absolute inset-0 z-0">
        <Image
          src="/assets/images/bg-pattern-plateau.jpg"
          alt="Pattern Plateau Background"
          fill
          className="object-cover scale-105"
          priority
        />
        <div className="absolute inset-0 bg-blue-900/10 backdrop-blur-[2px]" />
      </div>
      
      {/* Static Confetti/Sparkles */}
      <div className="absolute inset-0 pointer-events-none z-0">
        {[...Array(15)].map((_, i) => (
          <div 
            key={i}
            className={`absolute w-3 h-3 rounded-sm opacity-40 rotate-45 ${
              ['bg-yellow-300', 'bg-green-300', 'bg-blue-300', 'bg-red-300'][i % 4]
            }`}
            style={{
              top: `${Math.random() * 100}%`,
              left: `${Math.random() * 100}%`,
            }}
          />
        ))}
      </div>

      {/* Main Content (Z-index to be above background) */}
      <div className="relative z-10 flex flex-col items-center">
        {/* Header Progress Bar */}
        <div className="relative w-80 h-10 bg-[#0288D1]/30 rounded-full border-4 border-white mb-6 flex items-center px-1 shadow-lg overflow-hidden backdrop-blur-sm">
          <motion.div 
            className="h-full bg-gradient-to-r from-[#4FC3F7] to-[#FFB300] rounded-full"
            initial={{ width: '0%' }}
            animate={{ width: `${((scoreX + scoreO) % 10) * 10 || 50}%` }}
          />
          <div className="absolute right-0 top-1/2 -translate-y-1/2 translate-x-1/4 bg-yellow-400 rounded-full p-1 border-2 border-white shadow-md">
            <Star className="text-white fill-white" size={20} />
          </div>
        </div>

        {/* Turn Banner */}
        <div className="relative mb-6">
          <div className="absolute -left-10 -top-4 w-20 h-20 rounded-full border-4 border-white bg-white overflow-hidden shadow-xl z-20">
            <img 
              src="https://api.dicebear.com/7.x/avataaars/svg?seed=Felix&backgroundColor=b6e3f4" 
              alt="Player Avatar" 
              className="w-full h-full object-cover"
            />
            <div className="absolute bottom-1 right-1 bg-blue-500 rounded-full p-1 border-2 border-white shadow-sm">
              <XIcon size={14} className="text-white stroke-[4]" />
            </div>
          </div>
          
          <div className="bg-gradient-to-b from-[#8BC34A] to-[#4CAF50] px-16 py-4 rounded-[2rem] border-4 border-white shadow-2xl flex items-center gap-4">
            <span className="text-white text-3xl sm:text-4xl font-bold tracking-wide drop-shadow-[0_2px_2px_rgba(0,0,0,0.3)]">
              {winner ? (winner === 'X' ? 'Player 1 Wins!' : 'Player 2 Wins!') : `${isXNext ? "Player 1's" : "Player 2's"} Turn!`}
            </span>
            <Star className="text-yellow-300 fill-yellow-300" size={24} />
          </div>
        </div>

        {/* Game Board */}
        <div className="relative bg-white/90 backdrop-blur-md p-5 rounded-[2.5rem] border-[10px] border-[#64B5F6] shadow-[0_20px_50px_rgba(0,0,0,0.4)]">
          <div className="grid grid-cols-3 gap-4">
            {board.map((cell, i) => (
              <motion.button
                key={i}
                whileHover={{ scale: 1.02 }}
                whileTap={{ scale: 0.98 }}
                onClick={() => handleClick(i)}
                className="w-24 h-24 sm:w-28 sm:h-28 bg-[#FFFDE7] rounded-3xl border-4 border-[#BBDEFB] flex items-center justify-center relative overflow-hidden shadow-[inset_0_-4px_0_rgba(0,0,0,0.1)] focus:outline-none"
              >
                <AnimatePresence>
                  {cell === 'X' && (
                    <motion.div
                      initial={{ scale: 0, rotate: -45 }}
                      animate={{ scale: 1, rotate: 0 }}
                      className="text-[#42A5F5]"
                    >
                      <XIcon size={64} strokeWidth={3} />
                    </motion.div>
                  )}
                  {cell === 'O' && (
                    <motion.div
                      initial={{ scale: 0 }}
                      animate={{ scale: 1 }}
                      className="text-[#FFA726]"
                    >
                      <CircleIcon size={56} strokeWidth={4} />
                    </motion.div>
                  )}
                </AnimatePresence>
              </motion.button>
            ))}
          </div>

          {/* Winning Line Overlay */}
          {winner && (
            <motion.div 
              initial={{ opacity: 0, scale: 0 }}
              animate={{ opacity: 1, scale: 1 }}
              className="absolute bg-green-500 rounded-full z-10 shadow-lg"
              style={getWinningLineStyle()}
            />
          )}
        </div>

        {/* Footer Controls */}
        <div className="mt-10 flex flex-col sm:flex-row gap-6 sm:gap-10">
          <div className={`flex items-center bg-[#42A5F5] rounded-full pl-2 pr-8 py-3 border-4 ${isXNext && !winner ? 'border-yellow-400 scale-110 shadow-[0_0_20px_rgba(250,204,21,0.5)]' : 'border-white'} shadow-[0_8px_0_#1E88E5,0_15px_25px_rgba(0,0,0,0.2)] gap-4 transition-all`}>
            <div className="bg-white rounded-full p-2 shadow-inner">
              <XIcon size={28} className="text-[#42A5F5] stroke-[4]" />
            </div>
            <div className="flex flex-col">
              <span className="text-white font-bold text-xl leading-none">Player 1</span>
              <span className="text-white/80 font-bold text-sm">Score: {scoreX}</span>
            </div>
          </div>

          <div className={`flex items-center bg-[#FFA726] rounded-full pl-2 pr-8 py-3 border-4 ${!isXNext && !winner ? 'border-yellow-400 scale-110 shadow-[0_0_20px_rgba(250,204,21,0.5)]' : 'border-white'} shadow-[0_8px_0_#F57C00,0_15px_25px_rgba(0,0,0,0.2)] gap-4 transition-all`}>
            <div className="bg-white rounded-full p-2 shadow-inner">
              <CircleIcon size={28} className="text-[#FFA726] stroke-[4]" />
            </div>
            <div className="flex flex-col">
              <span className="text-white font-bold text-xl leading-none">Player 2</span>
              <span className="text-white/80 font-bold text-sm">Score: {scoreO}</span>
            </div>
          </div>
        </div>
      </div>

      {/* Reset Button (Floating) */}
      {winner && (
        <motion.button
          initial={{ y: 50, opacity: 0 }}
          animate={{ y: 0, opacity: 1 }}
          onClick={resetGame}
          className="mt-8 relative z-20 bg-yellow-400 hover:bg-yellow-500 text-white px-8 py-3 rounded-full font-bold text-xl border-4 border-white shadow-xl flex items-center gap-2 transform active:scale-95 transition-all"
        >
          <Trophy size={24} />
          Play Again!
        </motion.button>
      )}

      {/* Confetti Effect (Simple) */}
      {winner && (
        <div className="fixed inset-0 pointer-events-none z-50">
          {[...Array(30)].map((_, i) => (
            <motion.div
              key={i}
              initial={{ 
                top: -20, 
                left: `${Math.random() * 100}%`,
                rotate: 0,
                scale: Math.random() * 0.5 + 0.5
              }}
              animate={{ 
                top: '100%', 
                rotate: 360,
                left: `${Math.random() * 100}%`
              }}
              transition={{ 
                duration: Math.random() * 2 + 2, 
                repeat: Infinity,
                ease: "linear",
                delay: Math.random() * 0.5
              }}
              className={`absolute w-3 h-3 rounded-sm ${
                ['bg-red-500', 'bg-blue-500', 'bg-yellow-400', 'bg-green-500', 'bg-purple-500'][i % 5]
              }`}
            />
          ))}
        </div>
      )}
    </div>
  );
}
