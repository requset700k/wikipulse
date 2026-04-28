'use client';

// 앱 전역 Context Provider 모음. app/layout.tsx에서 children을 감쌈.
// TanStack Query: 서버 상태(API 응답) 캐싱 및 동기화 담당.

import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { useState } from 'react';

export function Providers({ children }: { children: React.ReactNode }) {
  // QueryClient를 컴포넌트 내부 state로 생성 → SSR 시 요청 간 공유 방지
  const [queryClient] = useState(
    () =>
      new QueryClient({
        defaultOptions: {
          queries: {
            staleTime: 30_000, // 30초 동안 캐시된 데이터를 fresh로 간주 (불필요한 재요청 방지)
            retry: 1, // 실패 시 1회 재시도
          },
        },
      }),
  );

  return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>;
}
