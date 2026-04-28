// 백엔드 API 클라이언트.
// 모든 HTTP 요청은 Next.js rewrite를 통해 /api/* → BACKEND_URL/api/*로 프록시됨.
// WebSocket은 rewrite 대상이 아니므로 Terminal 컴포넌트에서 NEXT_PUBLIC_WS_URL로 직접 연결.

import type { Lab, Session, StepProgress, User, Badge, LeaderboardEntry } from './types';

interface Paginated<T> {
  items: T[];
  total: number;
}

// 개발 모드에서는 Keycloak 대신 dev-token으로 백엔드 stub JWT 미들웨어를 통과
const DEV_HEADERS: Record<string, string> =
  process.env.NODE_ENV === 'development' ? { Authorization: 'Bearer dev-token' } : {};

/** 모든 API 요청의 공통 래퍼. 에러 응답은 백엔드의 { error: string } 포맷으로 throw. */
async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(path, {
    ...options,
    credentials: 'include', // Keycloak 쿠키(access_token) 자동 포함
    headers: {
      'Content-Type': 'application/json',
      ...DEV_HEADERS,
      ...options?.headers,
    },
  });

  if (!res.ok) {
    const body = await res.json().catch(() => ({ error: res.statusText }));
    throw new Error(body.error ?? 'Request failed');
  }

  if (res.status === 204) return undefined as T; // DELETE 등 응답 바디 없는 경우
  return res.json() as T;
}

export const api = {
  labs: {
    list: () => request<Paginated<Lab>>('/api/v1/labs'),
    get: (id: string) => request<Lab>(`/api/v1/labs/${id}`),
  },

  sessions: {
    create: (labId: string) =>
      request<Session>('/api/v1/sessions', {
        method: 'POST',
        body: JSON.stringify({ lab_id: labId }),
      }),
    get: (id: string) => request<Session>(`/api/v1/sessions/${id}`),
    delete: (id: string) => request<void>(`/api/v1/sessions/${id}`, { method: 'DELETE' }),
    steps: (id: string) => request<Paginated<StepProgress>>(`/api/v1/sessions/${id}/steps`),
    validate: (id: string, stepId: number) =>
      request<{ status: string; message: string }>(`/api/v1/sessions/${id}/validate`, {
        method: 'POST',
        body: JSON.stringify({ step_id: stepId }),
      }),
    hint: (id: string, stepId: number, level: number, history?: string) =>
      request<{ hint_text: string; related_docs: string[]; hints_remaining: number }>(
        `/api/v1/sessions/${id}/hint`,
        {
          method: 'POST',
          body: JSON.stringify({ step_id: stepId, hint_level: level, terminal_history: history }),
        },
      ),
  },

  auth: {
    me: () => request<User>('/api/v1/me'),
    logout: () => request<void>('/api/v1/auth/logout', { method: 'POST' }),
  },

  leaderboard: {
    get: () => request<Paginated<LeaderboardEntry>>('/api/v1/leaderboard'),
  },

  me: {
    badges: () => request<Paginated<Badge>>('/api/v1/me/badges'),
  },

  instructor: {
    sessions: () => request<Paginated<Session>>('/api/v1/instructor/sessions'),
    inject: (sessionId: string, command: string) =>
      request<void>(`/api/v1/instructor/sessions/${sessionId}/inject`, {
        method: 'POST',
        body: JSON.stringify({ command }),
      }),
  },
};
