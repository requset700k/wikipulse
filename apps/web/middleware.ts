// 인증 게이트 미들웨어 — Edge Runtime에서 실행됨 (서버 컴포넌트 렌더링 전).
// 로그인 안 된 사용자를 /login으로 리다이렉트.
// 현재는 쿠키 존재 여부만 체크 (토큰 검증은 추후 추가 예정).

import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';

// 인증 없이 접근 가능한 경로
const PUBLIC_PATHS = ['/login', '/callback'];

// Keycloak 미연동 환경에서 인증 게이트 비활성화. K8s에서 AUTH_ENABLED=true 주입 시 활성화.
const authEnabled = process.env.AUTH_ENABLED === 'true';

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;

  // 로그인/콜백 페이지는 인증 없이 접근 가능
  if (PUBLIC_PATHS.some((p) => pathname === p || pathname.startsWith(p + '/'))) {
    return NextResponse.next();
  }

  if (!authEnabled) return NextResponse.next();

  // Keycloak 로그인 후 백엔드가 설정하는 HTTP-only 쿠키로 인증 확인
  const token = request.cookies.get('access_token');
  if (!token) {
    const loginUrl = new URL('/login', request.url);
    loginUrl.searchParams.set('from', pathname); // 로그인 후 원래 경로로 복귀
    return NextResponse.redirect(loginUrl);
  }

  return NextResponse.next();
}

export const config = {
  // _next 정적 파일과 /api/* (백엔드 프록시)는 미들웨어 대상 제외
  matcher: ['/((?!_next/static|_next/image|favicon.ico|api).*)'],
};
