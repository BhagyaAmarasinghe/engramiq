import type { Metadata } from "next";
import { Providers } from "./providers";
import "./globals.css";

export const metadata: Metadata = {
  title: "EngramIQ - Advanced Site Memory",
  description: "Solar asset management with intelligent document processing and natural language queries",
  keywords: "solar asset management, AI, document processing, site memory",
  authors: [{ name: "EngramIQ" }],
  viewport: "width=device-width, initial-scale=1",
  themeColor: "#17c480",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body className="min-h-screen bg-primary-dark-blue font-figtree">
        <Providers>
          {children}
        </Providers>
      </body>
    </html>
  );
}