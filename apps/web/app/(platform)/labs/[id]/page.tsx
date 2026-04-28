// Lab 실습 화면 — 좌: xterm.js 터미널 / 우: 단계 목록 + AI 힌트.
// KodeKloud의 2분할 레이아웃을 계승.
//
// 화면 진입 흐름:
//   1. Lab 정보 로드 (GET /api/v1/labs/:id)
//   2. 세션 생성 (POST /api/v1/sessions) → VM 프로비저닝 시작
//   3. 2초마다 폴링 (GET /api/v1/sessions/:id) → status === "ready" 대기
//   4. 터미널 WebSocket 연결 (/api/v1/sessions/:id/ws)

'use client';

import { useEffect, useRef, useState, useCallback } from 'react';
import { useRouter } from 'next/navigation';
import { Terminal } from '@/components/terminal/Terminal';
import { StepList } from '@/components/lab/StepList';
import { HintSidebar } from '@/components/lab/HintSidebar';
import { api } from '@/lib/api';
import type { Lab, Session, StepProgress } from '@/lib/types';

// Lab DSL에서 단계 타이틀을 가져올 때까지 보여주는 임시 목 데이터 (Week 4 이후 제거)
const MOCK_STEPS = [
  { id: 1, title: 'VM에 접속해서 현재 디렉토리 확인' },
  { id: 2, title: '파일 시스템 구조 탐색' },
  { id: 3, title: '프로세스 목록 조회' },
  { id: 4, title: '네트워크 설정 확인' },
];

/** 페이지 로딩 단계 — UI 메시지 분기에 사용 */
type Phase = 'loading-lab' | 'starting' | 'provisioning' | 'ready' | 'error';
/** 단계 검증 상태 — 버튼 비활성화 및 피드백 메시지에 사용 */
type ValidationState = 'idle' | 'pending' | 'pass' | 'fail';

export default function LabSessionPage({ params }: { params: { id: string } }) {
  const router = useRouter();
  const [lab, setLab] = useState<Lab | null>(null);
  const [session, setSession] = useState<Session | null>(null);
  const [phase, setPhase] = useState<Phase>('loading-lab');
  const [errorMsg, setErrorMsg] = useState('');
  const [elapsed, setElapsed] = useState(0);

  const [steps, setSteps] = useState<StepProgress[]>([]);
  const [currentStep, setCurrentStep] = useState(1);
  const [validation, setValidation] = useState<ValidationState>('idle');
  const [showHint, setShowHint] = useState(false);

  // Load lab
  useEffect(() => {
    api.labs
      .get(params.id)
      .then(setLab)
      .catch(() => {
        setPhase('error');
        setErrorMsg('Lab을 찾을 수 없습니다.');
      });
  }, [params.id]);

  // 세션 생성 — React StrictMode에서 useEffect가 두 번 실행되는 것을 ref로 방지.
  // 중복 실행 시 서버에 VM이 두 번 프로비저닝되는 문제가 생기므로 필수.
  const creatingRef = useRef(false);
  useEffect(() => {
    if (!lab || creatingRef.current) return;
    creatingRef.current = true;
    setPhase('starting');
    api.sessions
      .create(lab.id)
      .then((s) => {
        setSession(s);
        setPhase('provisioning');
      })
      .catch((e) => {
        setPhase('error');
        setErrorMsg(e.message ?? '세션 생성 실패');
      });
  }, [lab]);

  // Poll until ready
  useEffect(() => {
    if (!session || phase !== 'provisioning') return;
    const timer = setInterval(() => setElapsed((e) => e + 1), 1000);
    const poller = setInterval(async () => {
      const updated = await api.sessions.get(session.id).catch(() => null);
      if (!updated) return;
      setSession(updated);
      if (updated.status === 'ready') {
        setPhase('ready');
        clearInterval(poller);
        clearInterval(timer);
        loadSteps(updated.id);
      } else if (updated.status === 'failed') {
        setPhase('error');
        setErrorMsg('VM 프로비저닝 실패');
        clearInterval(poller);
        clearInterval(timer);
      }
    }, 2000);
    return () => {
      clearInterval(timer);
      clearInterval(poller);
    };
  }, [session, phase]);

  async function loadSteps(sessionId: string) {
    const res = await api.sessions.steps(sessionId).catch(() => null);
    if (res) setSteps(res.items ?? []);
  }

  // steps 폴링 — validation 진행 중일 때만 (idle이면 중단)
  useEffect(() => {
    if (phase !== 'ready' || !session || validation !== 'pending') return;
    const poller = setInterval(() => loadSteps(session.id), 3000);
    return () => clearInterval(poller);
  }, [phase, session, validation]);

  // Trigger validation
  async function handleValidate() {
    if (!session || validation === 'pending') return;
    setValidation('pending');
    try {
      await api.sessions.validate(session.id, currentStep);
      // Poll for result
      const check = setInterval(async () => {
        await loadSteps(session.id);
        const step = steps.find((s) => s.step_id === currentStep);
        if (step?.status === 'passed') {
          setValidation('pass');
          setCurrentStep((c) => c + 1);
          clearInterval(check);
          setTimeout(() => setValidation('idle'), 2000);
        } else if (step?.status === 'failed') {
          setValidation('fail');
          clearInterval(check);
          setTimeout(() => setValidation('idle'), 2000);
        }
      }, 1000);
      setTimeout(() => {
        clearInterval(check);
        setValidation('idle');
      }, 10000);
    } catch {
      setValidation('idle');
    }
  }

  const handleTerminate = useCallback(async () => {
    if (session) await api.sessions.delete(session.id).catch(() => {});
    router.push('/labs');
  }, [session, router]);

  // ── Provisioning / error screen ──────────────────────────────────────────
  if (phase !== 'ready') {
    return (
      <div className="flex items-center justify-center h-[calc(100vh-3.5rem)]">
        <div className="text-center max-w-sm">
          {phase === 'error' ? (
            <>
              <div className="text-4xl mb-4">⚠️</div>
              <p className="text-red-400 font-medium mb-4">
                {errorMsg || '알 수 없는 오류가 발생했습니다.'}
              </p>
              <div className="flex flex-col gap-2 items-center">
                <button
                  onClick={() => window.location.reload()}
                  className="text-sm bg-brand-500 hover:bg-brand-600 text-white px-4 py-2 rounded-lg transition-colors"
                >
                  페이지 새로고침
                </button>
                <button
                  onClick={() => router.push('/labs')}
                  className="text-slate-500 hover:text-white text-sm transition-colors"
                >
                  ← Labs로 돌아가기
                </button>
              </div>
            </>
          ) : (
            <>
              <div className="inline-block w-10 h-10 border-2 border-brand-500 border-t-transparent rounded-full animate-spin mb-6" />
              <p className="text-white font-semibold mb-1">
                {phase === 'loading-lab' && 'Lab 정보 로딩 중...'}
                {phase === 'starting' && '세션 생성 중...'}
                {phase === 'provisioning' && 'VM 프로비저닝 중...'}
              </p>
              {phase === 'provisioning' && (
                <>
                  <p className="text-slate-400 text-sm mb-1">
                    {session?.vm_provider === 'ec2' ? 'AWS EC2' : '온프렘 KubeVirt'} VM 부팅 중
                  </p>
                  <p className="text-slate-600 text-xs font-mono">{elapsed}초 경과</p>
                </>
              )}
            </>
          )}
        </div>
      </div>
    );
  }

  // ── 2-panel layout ───────────────────────────────────────────────────────
  return (
    <div className="flex h-[calc(100vh-3.5rem)] overflow-hidden -mx-4 sm:-mx-6 -my-8">
      {/* Left: Terminal */}
      <div className="flex-1 flex flex-col min-w-0 border-r border-slate-800">
        <div className="flex items-center justify-between px-4 py-2 bg-slate-900 border-b border-slate-800 flex-shrink-0">
          <div className="flex items-center gap-2">
            <span className="w-2.5 h-2.5 rounded-full bg-emerald-400 animate-pulse" />
            <span className="text-slate-300 text-sm font-mono">
              {session?.vm_provider === 'ec2' ? 'ec2' : 'vm'} — bash
            </span>
          </div>
          <button
            onClick={handleTerminate}
            className="text-slate-500 hover:text-red-400 text-xs transition-colors"
          >
            세션 종료
          </button>
        </div>
        <div className="flex-1 min-h-0">
          <Terminal
            sessionId={session!.id}
            onDisconnect={() => {
              setErrorMsg('터미널 연결이 끊겼습니다. 백엔드 서버를 확인하세요.');
              setPhase('error');
            }}
          />
        </div>
      </div>

      {/* Right: Lab portal + optional hint sidebar */}
      <div className="flex flex-shrink-0">
        {/* Lab portal */}
        <div className="w-72 flex flex-col bg-slate-900 border-r border-slate-800 overflow-y-auto scrollbar-thin">
          <div className="px-4 py-4 border-b border-slate-800">
            <h2 className="text-white font-semibold text-sm">{lab?.title}</h2>
            <p className="text-slate-400 text-xs mt-0.5">
              {lab?.step_count}단계 · {lab?.duration_min}분
            </p>
          </div>

          <div className="px-4 py-4 flex-1">
            <p className="text-slate-500 text-xs font-medium uppercase tracking-wider mb-3">
              진행 상황
            </p>
            <StepList steps={MOCK_STEPS} progress={steps} currentStep={currentStep} />
          </div>

          {/* Validation feedback */}
          {validation !== 'idle' && (
            <div
              className={`mx-4 mb-3 px-3 py-2 rounded-lg text-sm text-center ${
                validation === 'pending'
                  ? 'bg-slate-700 text-slate-300'
                  : validation === 'pass'
                    ? 'bg-emerald-500/20 text-emerald-400 border border-emerald-500/30'
                    : 'bg-red-500/20 text-red-400 border border-red-500/30'
              }`}
            >
              {validation === 'pending' && '⏳ 검증 중...'}
              {validation === 'pass' && '✓ 통과! 다음 단계로 이동합니다'}
              {validation === 'fail' && '✗ 아직 조건을 충족하지 못했어요'}
            </div>
          )}

          <div className="px-4 py-4 border-t border-slate-800 space-y-2">
            <button
              onClick={handleValidate}
              disabled={validation === 'pending'}
              className="w-full bg-brand-500 hover:bg-brand-600 disabled:opacity-50 text-white text-sm font-medium py-2 rounded-lg transition-colors"
            >
              단계 완료 확인
            </button>
            <button
              onClick={() => setShowHint((v) => !v)}
              className="w-full bg-slate-800 hover:bg-slate-700 text-slate-300 text-sm py-2 rounded-lg transition-colors"
            >
              {showHint ? '힌트 닫기' : '💡 AI 힌트 요청'}
            </button>
          </div>
        </div>

        {/* Hint sidebar */}
        {showHint && session && (
          <div className="w-72 flex-shrink-0">
            <HintSidebar
              sessionId={session.id}
              stepId={currentStep}
              onClose={() => setShowHint(false)}
            />
          </div>
        )}
      </div>
    </div>
  );
}
