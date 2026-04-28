// 강사 대시보드 — 현재 활성 수강생 세션 목록을 카드 그리드로 표시.
// 기능: Lab별 필터, 수동 새로고침, 터미널 관전(TerminalModal), 명령 주입.
// instructor 또는 admin 역할만 접근 가능 (백엔드 RequireRole 미들웨어 + 프론트 라우팅).

'use client';

import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { api } from '@/lib/api';
import { TerminalModal } from '@/components/instructor/TerminalModal';

interface InstructorSession {
  id: string;
  user_name: string;
  user_email: string;
  lab_id: string;
  status: string;
  current_step: number;
  steps_done: number;
  started_at: string;
  has_terminal: boolean;
}

const STATUS_COLOR: Record<string, string> = {
  provisioning: 'text-amber-400',
  ready: 'text-emerald-400',
  active: 'text-emerald-400',
  completed: 'text-slate-500',
  failed: 'text-red-400',
};

const LAB_NAMES: Record<string, string> = {
  'lab-001': 'Linux Basics',
  'lab-002': 'Ansible Fundamentals',
  'lab-003': 'Terraform Introduction',
  'lab-004': 'Kubernetes Introduction',
};

type InjectState = { cmd: string; ok: boolean };

export default function InstructorPage() {
  const [watching, setWatching] = useState<InstructorSession | null>(null);
  const [selectedLab, setSelectedLab] = useState<string>('all');
  const [injectMap, setInjectMap] = useState<Record<string, InjectState>>({});

  function showInject(sessionId: string, cmd: string, ok: boolean) {
    setInjectMap((prev) => ({ ...prev, [sessionId]: { cmd, ok } }));
    setTimeout(() => {
      setInjectMap((prev) => {
        const next = { ...prev };
        delete next[sessionId];
        return next;
      });
    }, 3000);
  }

  const { data, isLoading, refetch, isFetching } = useQuery({
    queryKey: ['instructor-sessions'],
    queryFn: () =>
      api.instructor.sessions() as unknown as Promise<{
        items: InstructorSession[];
        total: number;
      }>,
  });

  const sessions = data?.items ?? [];

  // 활성 Lab 목록 (세션에 있는 것만)
  const activeLabs = Array.from(new Set(sessions.map((s) => s.lab_id)));

  const filtered =
    selectedLab === 'all' ? sessions : sessions.filter((s) => s.lab_id === selectedLab);

  return (
    <div>
      {/* Header */}
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white">강사 대시보드</h1>
          <p className="text-slate-400 text-sm mt-1">수강생 Lab 세션 현황</p>
        </div>
        <button
          onClick={() => refetch()}
          disabled={isFetching}
          className="flex items-center gap-2 px-4 py-2 bg-slate-800 hover:bg-slate-700 disabled:opacity-50 text-slate-300 text-sm rounded-lg transition-colors"
        >
          {isFetching ? (
            <span className="inline-block w-3.5 h-3.5 border border-slate-400 border-t-transparent rounded-full animate-spin" />
          ) : (
            <span>↻</span>
          )}
          새로고침
        </button>
      </div>

      {/* Lab 필터 */}
      {!isLoading && activeLabs.length > 1 && (
        <div className="flex gap-2 mb-6 flex-wrap">
          <button
            onClick={() => setSelectedLab('all')}
            className={`px-4 py-1.5 rounded-full text-sm font-medium transition-colors ${
              selectedLab === 'all'
                ? 'bg-brand-500 text-white'
                : 'bg-slate-800 text-slate-400 hover:text-white'
            }`}
          >
            전체 ({sessions.length})
          </button>
          {activeLabs.map((labId) => {
            const count = sessions.filter((s) => s.lab_id === labId).length;
            return (
              <button
                key={labId}
                onClick={() => setSelectedLab(labId)}
                className={`px-4 py-1.5 rounded-full text-sm font-medium transition-colors ${
                  selectedLab === labId
                    ? 'bg-brand-500 text-white'
                    : 'bg-slate-800 text-slate-400 hover:text-white'
                }`}
              >
                {LAB_NAMES[labId] ?? labId} ({count})
              </button>
            );
          })}
        </div>
      )}

      {/* Grid */}
      {isLoading && (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
          {[...Array(3)].map((_, i) => (
            <div
              key={i}
              className="bg-slate-800/50 border border-slate-700 rounded-xl p-5 animate-pulse h-44"
            />
          ))}
        </div>
      )}

      {!isLoading && filtered.length === 0 && (
        <div className="text-center py-20 text-slate-500">
          <p className="text-4xl mb-4">📋</p>
          <p>
            {selectedLab === 'all'
              ? '현재 활성 세션이 없습니다.'
              : '이 Lab의 활성 세션이 없습니다.'}
          </p>
        </div>
      )}

      {filtered.length > 0 && (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
          {filtered.map((s) => (
            <SessionCard
              key={s.id}
              session={s}
              injectState={injectMap[s.id] ?? null}
              onWatch={() => setWatching(s)}
              onInject={showInject}
            />
          ))}
        </div>
      )}

      {watching && (
        <TerminalModal
          sessionId={watching.id}
          studentName={watching.user_name}
          onClose={() => setWatching(null)}
        />
      )}
    </div>
  );
}

function SessionCard({
  session,
  injectState,
  onWatch,
  onInject,
}: {
  session: InstructorSession;
  injectState: InjectState | null;
  onWatch: () => void;
  onInject: (sessionId: string, cmd: string, ok: boolean) => void;
}) {
  async function inject() {
    const cmd = window.prompt('주입할 명령어:');
    if (!cmd) return;
    try {
      await api.instructor.inject(session.id, cmd);
      onInject(session.id, cmd, true);
    } catch (e) {
      onInject(session.id, e instanceof Error ? e.message : '주입 실패', false);
    }
  }

  const elapsed = Math.floor((Date.now() - new Date(session.started_at).getTime()) / 60000);

  return (
    <div className="bg-slate-800/50 border border-slate-700 rounded-xl p-5 hover:border-slate-600 transition-colors">
      {/* Student info */}
      <div className="flex items-start justify-between mb-3">
        <div>
          <p className="text-white font-medium text-sm">{session.user_name}</p>
          <p className="text-slate-500 text-xs">{session.user_email}</p>
        </div>
        <span className={`text-xs font-medium ${STATUS_COLOR[session.status] ?? 'text-slate-400'}`}>
          {session.status}
        </span>
      </div>

      <div className="flex items-center justify-between text-xs text-slate-500 mb-3">
        <span>
          Lab: <span className="text-slate-400">{session.lab_id}</span>
        </span>
        <span>{elapsed}분 경과</span>
      </div>

      {/* Progress */}
      <div className="mb-4">
        <div className="flex justify-between text-xs text-slate-500 mb-1">
          <span>단계 {session.current_step}</span>
          <span>{session.steps_done}단계 완료</span>
        </div>
        <div className="h-1.5 bg-slate-700 rounded-full overflow-hidden">
          <div
            className="h-full bg-brand-500 rounded-full transition-all"
            style={{ width: `${Math.min((session.steps_done / 8) * 100, 100)}%` }}
          />
        </div>
      </div>

      {/* 명령 주입 결과 인라인 표시 */}
      {injectState && (
        <div
          className={`mb-3 px-3 py-1.5 rounded-lg text-xs flex items-center gap-1.5 ${
            injectState.ok
              ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20'
              : 'bg-red-500/10 text-red-400 border border-red-500/20'
          }`}
        >
          <span>{injectState.ok ? '✓' : '✗'}</span>
          <span className="font-mono truncate">{injectState.cmd}</span>
          <span className="ml-auto flex-shrink-0">{injectState.ok ? '주입됨' : '실패'}</span>
        </div>
      )}

      {/* Actions */}
      <div className="flex gap-2">
        <button
          onClick={onWatch}
          disabled={!session.has_terminal}
          title={!session.has_terminal ? '수강생 터미널이 아직 연결되지 않았습니다' : ''}
          className="flex-1 text-xs bg-slate-700 hover:bg-brand-500/20 hover:text-brand-400 hover:border-brand-500/30 text-slate-300 border border-slate-600 py-1.5 rounded-lg transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
        >
          {session.has_terminal ? '터미널 관전' : '터미널 미연결'}
        </button>
        <button
          onClick={inject}
          className="flex-1 text-xs bg-slate-700 hover:bg-slate-600 text-slate-300 border border-slate-600 py-1.5 rounded-lg transition-colors"
        >
          명령 주입
        </button>
      </div>
    </div>
  );
}
