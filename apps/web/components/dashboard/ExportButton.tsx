"use client";

import { useState } from "react";

export default function ExportButton() {
  const [showToast, setShowToast] = useState(false);

  const handleClick = () => {
    setShowToast(true);
    setTimeout(() => setShowToast(false), 3000);
  };

  return (
    <div className="relative inline-block">
      <button
        onClick={handleClick}
        className="rounded-xl bg-indigo-100 px-4 py-2 text-sm font-semibold text-indigo-700 transition-colors hover:bg-indigo-200"
      >
        📥 Download Report
      </button>

      {showToast && (
        <div className="absolute right-0 top-12 z-10 rounded-lg bg-gray-800 px-4 py-2 text-sm text-white shadow-lg">
          Report generation coming soon
        </div>
      )}
    </div>
  );
}
