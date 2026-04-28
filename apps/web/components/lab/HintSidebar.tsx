// AI 학습 도우미 사이드바.
// 수강생이 막혔을 때 Gemini(양성호 담당 AI BFF)에 힌트를 요청.
// 힌트 단계: 1(막연) → 2(구체적) → 3(직접적). 정답 직접 제시는 하지 않음(소크라테스 방식).
// 분당 6회 Rate Limit — 백엔드 Redis INCR+Expire 패턴으로 관리.

'use client';

import { useState } from 'react';
import { api } from '@/lib/api';

interface Props {
  sessionId: string;
  stepId: number;
  terminalHistory?: string; // 현재까지 입력한 터미널 내용 — AI가 문맥 파악에 사용
  onClose: () => void;
}

interface HintResult {
  hint_text: string;
  related_docs: string[];
  hints_remaining: number;
}

export function HintSidebar({ sessionId, stepId, terminalHistory, onClose }: Props) {
  const [level, setLevel] = useState(1);
  const [result, setResult] = useState<HintResult | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  async function requestHint() {
    setLoading(true);
    setError('');
    try {
      const res = await api.sessions.hint(sessionId, stepId, level, terminalHistory);
      setResult(res);
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : '힌트 요청 실패';
      setError(msg.includes('rate limit') ? '분당 힌트 요청 횟수를 초과했습니다.' : msg);
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="flex flex-col h-full bg-slate-900 border-l border-slate-700">
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-3 border-b border-slate-800">
        <div className="flex items-center gap-2">
          <span className="text-amber-400 text-sm">💡</span>
          <span className="text-white text-sm font-medium">AI 학습 도우미</span>
        </div>
        <button onClick={onClose} className="text-slate-500 hover:text-white text-sm">
          ✕
        </button>
      </div>

      <div className="flex-1 overflow-y-auto p-4 space-y-4">
        {/* Hint level selector */}
        <div>
          <p className="text-slate-400 text-xs mb-2">힌트 단계 선택</p>
          <div className="flex gap-2">
            {[1, 2, 3].map((l) => (
              <button
                key={l}
                onClick={() => {
                  setLevel(l);
                  setResult(null);
                }}
                className={`flex-1 py-1.5 rounded-lg text-sm font-medium transition-colors ${
                  level === l
                    ? 'bg-amber-500/20 text-amber-400 border border-amber-500/40'
                    : 'bg-slate-800 text-slate-400 hover:text-white border border-slate-700'
                }`}
              >
                {l === 1 ? '막연한' : l === 2 ? '구체적' : '직접적'}
              </button>
            ))}
          </div>
          <p className="text-slate-600 text-xs mt-1">
            {level === 1 && '방향만 알려드려요'}
            {level === 2 && '좀 더 구체적인 힌트예요'}
            {level === 3 && '거의 정답에 가까운 힌트예요'}
          </p>
        </div>

        {/* Result */}
        {result && (
          <div className="bg-slate-800/60 border border-slate-700 rounded-xl p-4">
            <p className="text-slate-200 text-sm leading-relaxed">{result.hint_text}</p>
            {result.related_docs.length > 0 && (
              <div className="mt-3 pt-3 border-t border-slate-700">
                <p className="text-slate-500 text-xs mb-1">관련 문서</p>
                {result.related_docs.map((doc) => (
                  <a
                    key={doc}
                    href={doc}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-brand-400 text-xs hover:underline block"
                  >
                    {doc}
                  </a>
                ))}
              </div>
            )}
            <p className="text-slate-600 text-xs mt-3">남은 횟수: {result.hints_remaining}회/분</p>
          </div>
        )}

        {error && (
          <p className="text-red-400 text-sm bg-red-500/10 border border-red-500/20 rounded-lg px-3 py-2">
            {error}
          </p>
        )}
      </div>

      {/* Footer */}
      <div className="px-4 py-3 border-t border-slate-800">
        <button
          onClick={requestHint}
          disabled={loading}
          className="w-full bg-amber-500 hover:bg-amber-600 disabled:opacity-50 text-white text-sm font-medium py-2 rounded-lg transition-colors"
        >
          {loading ? '힌트 가져오는 중...' : '힌트 요청'}
        </button>
        <p className="text-slate-600 text-xs text-center mt-2">
          소크라테스식 힌트 · 정답 직접 제시 안 함
        </p>
      </div>
    </div>
  );
}
