"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { register } from "@/lib/api";

export default function RegisterPage() {
  const router = useRouter();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [role, setRole] = useState<"student" | "parent" | "educator">(
    "student"
  );
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

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
    <main className="flex min-h-screen items-center justify-center bg-gray-50">
      <form
        onSubmit={handleSubmit}
        className="w-full max-w-sm space-y-4 rounded-lg bg-white p-8 shadow"
      >
        <h1 className="text-2xl font-bold text-center">Create Account</h1>

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

        <label className="block">
          <span className="text-sm font-semibold">Password</span>
          <input
            type="password"
            required
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            className="mt-1 block w-full rounded border px-3 py-2"
          />
        </label>

        <label className="block">
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
          className="w-full rounded bg-blue-600 py-2 text-white font-semibold hover:bg-blue-700 disabled:opacity-50"
        >
          {loading ? "Creating account…" : "Register"}
        </button>

        <p className="text-sm text-center text-gray-600">
          Already have an account?{" "}
          <a href="/login" className="text-blue-600 hover:underline">
            Sign in
          </a>
        </p>
      </form>
    </main>
  );
}
