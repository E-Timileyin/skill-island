"use client";
import { useState } from "react";
import { useRouter } from "next/navigation";
import { register } from "@/lib/api";
import AuthLayout from "../AuthLayout";
import Image from "next/image";
import logo from '@/public/assets/logo.png'


export default function RegisterPage() {
  const router = useRouter();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [role, setRole] = useState<"student" | "parent" | "educator">(
    "student"
  );
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [isStartHovered, setIsStartHovered ] = useState(false)

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    setLoading(true);

    try {
      const user = await register(email, password, role);
      if (user.role === "student") {
        router.push("/island");
      } else {
        router.push("/overview");
      }
    } catch (err: unknown) {
      const apiErr = err as { message?: string };
      setError(apiErr.message || "Registration failed");
    } finally {
      setLoading(false);
    }
  }

  return (
    <AuthLayout>
      <form
        onSubmit={handleSubmit}
        className="w-full md:w-[40rem] lg:w-[48rem] mx-auto space-y-4 rounded-4xl flex flex-col items-center justify-center"
      >
        <div className="flex justify-center mb-4">
          <Image src={logo} alt="Logo" className="h-[11rem] md:h-[12rem] w-auto" />
        </div>
        {error && (
          <p className="text-sm text-red-600 text-center">{error}</p>
        )}
        <label className="block">
          <span className="text-sm font-semibold">Email</span>
          <input
            type="email"
            required
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            className="mt-1 block w-full rounded border px-3 py-2"
          />
        </label>
        <label className="block pb-2">
          <span className="text-sm font-semibold">Password</span>
          <input
            type="password"
            required
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            className="mt-1 block w-full rounded border px-3 py-2"
          />
        </label>
        <label className="block pb-2">
          <span className="text-sm font-semibold">Role</span>
          <select
            value={role}
            onChange={(e) =>
              setRole(e.target.value as "student" | "parent" | "educator")
            }
            className="mt-1 block w-full rounded border px-3 py-2"
          >
            <option value="student">Student</option>
            <option value="parent">Parent</option>
            <option value="educator">Educator</option>
          </select>
        </label>
         <button
          type="submit"
          disabled={loading}
          onMouseEnter={() => setIsStartHovered(true)}
          onMouseLeave={() => setIsStartHovered(false)}
          className={`
            group relative px-14 py-2 rounded-full 
            transition-all duration-200 transform hover:scale-105 active:scale-95
            shadow-[0_10px_20px_rgba(234,88,12,0.4)]/10
            border-4 border-white
          `}
          style={{
            background:
              "linear-gradient(180deg, #fbbf24 0%, #f97316 50%, #ea580c 100%)",
          }}
        >
          {/* Glossy Highlight Top */}
          <div className="absolute top-2 left-4 right-4 h-1/3 bg-white/30 rounded-full blur-[2px]" />
          {/* Button Text */}
          <span className="relative z-10 text-xl font-black text-white drop-shadow-md tracking-wide">
            {loading ? "Creating account…" : "Register"}
          </span>
          {/* Inner Shadow/Glow for depth */}
          <div className="absolute inset-0 rounded-full shadow-[inset_0_-4px_6px_rgba(0,0,0,0.2)]" />
        </button>
        <p className="text-sm text-center text-gray-600">
          Already have an account?{" "}
          <a href="/login" className="text-blue-600 hover:underline">
            Sign in
          </a>
        </p>
      </form>
    </AuthLayout>
  );
}
