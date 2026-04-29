// 앱 루트 레이아웃 — 모든 페이지에 공통 적용. Inter 폰트, 전역 CSS 설정.
import type { Metadata } from 'next';
import { Inter } from 'next/font/google';
import './globals.css';

const inter = Inter({ subsets: ['latin'] });

export const metadata: Metadata = {
  title: 'Cledyu',
  description: '클라우드 엔지니어링 인터랙티브 실습 플랫폼',
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="ko">
      <body className={inter.className}>{children}</body>
    </html>
  );
}
