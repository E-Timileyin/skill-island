
import LoginPage from '@/app/(auth)/login/page'
import Image from 'next/image';

// Import logo
import logo from '@/public/assets/logo/logo.png';

export default function Home() {
  return (
    <main className="relative min-h-screen w-full flex flex-col items-center justify-center overflow-hidden">
      {/* Background Image */}
      <div className="absolute inset-0 z-0">
        <Image
          src="/assets/images/bg-login.jpg"
          alt="Castle Background"
          fill
          className="object-cover"
          priority
        />
      </div>
      {/* Logo at the top */}
      <div className="relative z-20 flex flex-col items-center mt-12 mb-8">
          <div className="absolute top-0 left-0 w-full flex justify-center z-10">
            <Image
              src={logo}
              alt="Skill Island Logo"
              width={400}
              height={120}
              className="object-contain drop-shadow-lg"
              priority
            />
          </div>
      </div>
      <div className="relative z-10 flex flex-col items-center justify-center w-full">h
        {/* ...existing code... */}
      </div>

      {/* Get Started button at the bottom, styled like login */}
      <button
        className="fixed bottom-8 left-1/2 -translate-x-1/2 w-[320px] py-4 rounded-2xl transition-all duration-200 transform hover:scale-105 active:scale-95 shadow-xl border-4 border-white overflow-hidden z-20 mb-20"
        style={{
          background:
            "linear-gradient(180deg, #fbbf24 0%, #f97316 50%, #ea580c 100%)",
        }}
      >
        <div className="absolute top-1 left-4 right-4 h-1/3 bg-white/40 rounded-full blur-[2px]" />
        <span className="relative z-10 text-2xl font-black text-white drop-shadow-md tracking-wider">
          Get Started
        </span>
        <div className="absolute inset-0 rounded-xl shadow-[inset_0_-4px_8px_rgba(0,0,0,0.3)]" />
      </button>
    </main>
  );
}
