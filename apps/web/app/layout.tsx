import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "Skill Island",
  description: "Gamified educational platform for SEND teenagers",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body className="font-nunito antialiased">{children}</body>
    </html>
  );
}
