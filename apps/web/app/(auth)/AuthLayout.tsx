"use client";

import Image from "next/image";
import logo from "@/public/assets/logo.png";
import authImage from "@/public/auth/authImage.jpeg";

export default function AuthLayout({ children }: { children: React.ReactNode }) {
  return (
    <main className="min-h-screen grid grid-cols-1 md:grid-cols-2">
      {/* Left: Form */}
      <div className="flex items-center justify-center h-full">
        <div className="w-full h-full flex items-center justify-center">
          {children}
        </div>
      </div>
      {/* Right: Image */}
      <div className="relative hidden md:block rounded-bl-4xl rounded-tl-4xl overflow-hidden">
        <Image
          src={authImage}
          alt="login-image"
          fill
          className="object-cover w-full h-full"
          priority
          quality={100}
        />
        {/* <div className="absolute inset-0 bg-black/30" /> */}
      </div>
    </main>
  )
}
