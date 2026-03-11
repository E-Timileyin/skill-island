"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { login } from "@/lib/api";
import logo from "@/public/assets/logo.png";
import Image from "next/image";

export default function LoginPage() {
  const router = useRouter();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [isStartHovered, setIsStartHovered] = useState(false);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    setLoading(true);

    try {
      await login(email, password);
      // Now fetch user info using /api/auth/me
      const user = await import("@/lib/api").then(api => api.getMe());
      console.log("User after login:", user);
      if (user.role === "student") {
        router.push("/island");
        console.log("routing based on user role:", user.role);
      } else {
        router.push("/overview");
        console.log("routing to the ");
      }
    } catch (err: unknown) {
      const apiErr = err as { message?: string };

      if (apiErr.message){
        setError(apiErr.message || "Login failed"); 
        console.error("Unknown error during login:", err);
      } else{
        setError(apiErr.message || "Login failed. Please check your credentials or try again.");
      }
    }
  }

  return (
    <div className="relative min-h-screen flex items-center justify-center overflow-hidden">
      {/* Background Image */}
      <div className="absolute inset-0 z-0">
        <Image
          src="/assets/images/bg-login-1.jpg"
          alt="Background"
          fill
          className="object-cover"
          priority
        />
        {/* <div className="absolute inset-0 bg-black/20 backdrop-blur-[2px]" /> */}
      </div>

      <form
        onSubmit={handleSubmit}
        className="relative z-10 w-full max-w-lg mx-auto space-y-6 flex flex-col items-center justify-center bg-white/50 backdrop-blur-md p-10 rounded-[2.5rem] shadow-2xl border-4 border-white/50"
      >
        <div className="flex justify-center mb-2">
          <Image src={logo} alt="Logo" className="h-[9rem] md:h-[10rem] w-auto drop-shadow-lg" />
        </div>

        <div className="w-full space-y-4">
          {error && (
            <div className="bg-red-100 border-2 border-red-200 text-red-700 px-4 py-2 rounded-xl text-center font-bold animate-shake">
              {error}
            </div>
          )}

          <div>
            <label className="block text-sm font-bold text-blue-900 ml-1 mb-1">Email</label>
            <input
              type="email"
              required
              placeholder="Enter your email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className="block w-full rounded-2xl border-2 border-blue-100 bg-white/90 px-4 py-3 text-gray-700 focus:border-blue-400 focus:outline-none focus:ring-4 focus:ring-blue-100 transition-all font-medium"
            />
          </div>

          <div>
            <label className="block text-sm font-bold text-blue-900 ml-1 mb-1">Password</label>
            <input
              type="password"
              required
              placeholder="Enter your password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="block w-full rounded-2xl border-2 border-blue-100 bg-white/90 px-4 py-3 text-gray-700 focus:border-blue-400 focus:outline-none focus:ring-4 focus:ring-blue-100 transition-all font-medium"
            />
          </div>
        </div>

        <button
          type="submit"
          disabled={loading}
          onMouseEnter={() => setIsStartHovered(true)}
          onMouseLeave={() => setIsStartHovered(false)}
          className={`
            group relative w-full py-4 rounded-2xl
            transition-all duration-200 transform hover:scale-105 active:scale-95
            shadow-xl border-4 border-white overflow-hidden
          `}
          style={{
            background:
              "linear-gradient(180deg, #fbbf24 0%, #f97316 50%, #ea580c 100%)",
          }}
        >
          {/* Glossy Highlight Top */}
          <div className="absolute top-1 left-4 right-4 h-1/3 bg-white/40 rounded-full blur-[2px]" />
          
          <span className="relative z-10 text-2xl font-black text-white drop-shadow-md tracking-wider">
            {loading ? "Signing in…" : "Sign In"}
          </span>
          
          <div className="absolute inset-0 rounded-xl shadow-[inset_0_-4px_8px_rgba(0,0,0,0.3)]" />
        </button>

        <p className="text-base font-bold text-center text-blue-900">
          Don&apos;t have an account?{" "}
          <a href="/register" className="text-orange-600 hover:text-orange-700 underline underline-offset-4 decoration-2 transition-colors">
            Register for free
          </a>
        </p>
      </form>
    </div>
  );
}
