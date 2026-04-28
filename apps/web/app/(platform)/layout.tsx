// 인증이 필요한 플랫폼 영역의 공통 레이아웃.
// /labs, /labs/[id], /instructor 등 로그인 후 접근하는 모든 페이지에 적용.
// Navbar + 최대 너비 컨테이너를 공통으로 제공.
import { Navbar } from '@/components/ui/Navbar';

export default function PlatformLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="min-h-screen bg-slate-950">
      <Navbar />
      <main className="max-w-7xl mx-auto px-4 sm:px-6 py-8">{children}</main>
    </div>
  );
}
