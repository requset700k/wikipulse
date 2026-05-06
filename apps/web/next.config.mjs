// BACKEND_URL: Next.js 서버(SSR/API Route)에서만 사용하는 서버 사이드 환경변수.
// K8s deployment.yaml에서 런타임 주입. 로컬 개발 시 기본값 localhost:8080.
const BACKEND_URL = process.env.BACKEND_URL || 'http://localhost:8080';

/** @type {import('next').NextConfig} */
const nextConfig = {
  // standalone: Docker 이미지 최적화 모드.
  // npm run build 후 .next/standalone에 필요한 파일만 추출 → 이미지 크기 대폭 감소.
  output: 'standalone',
  async rewrites() {
    return [
      {
        // 브라우저의 /api/* 요청을 Next.js 서버가 백엔드로 프록시.
        source: '/api/:path*',
        destination: `${BACKEND_URL}/api/:path*`,
      },
    ];
  },
};

export default nextConfig;
