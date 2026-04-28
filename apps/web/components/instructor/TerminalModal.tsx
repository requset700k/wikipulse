// 강사 관전 모드 모달 — 수강생의 터미널 출력을 실시간으로 읽기 전용으로 표시.
// 백엔드 /api/v1/instructor/sessions/:id/ws에 연결 → PTY Broadcast 스트림 수신.
// disableStdin: true로 강사는 입력 불가 (관전 전용). 명령 주입은 별도 inject API 사용.
// Terminal.tsx와 xterm 초기화 로직이 동일하지만 역할이 달라 분리.

'use client';

import { useEffect, useRef, useState } from 'react';

interface Props {
  sessionId: string;
  studentName: string;
  onClose: () => void;
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
type XTermInstance = any;

type ConnState = 'connecting' | 'connected' | 'closed' | 'error';

export function TerminalModal({ sessionId, studentName, onClose }: Props) {
  const containerRef = useRef<HTMLDivElement>(null);
  const termRef = useRef<XTermInstance | null>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const [connState, setConnState] = useState<ConnState>('connecting');

  useEffect(() => {
    let isMounted = true;

    async function init() {
      if (!containerRef.current) return;

      const { Terminal: XTerm } = await import('@xterm/xterm');
      const { FitAddon } = await import('@xterm/addon-fit');

      if (!isMounted || !containerRef.current) return;

      const term: XTermInstance = new XTerm({
        cursorBlink: false,
        fontSize: 13,
        fontFamily: '"JetBrains Mono", "Fira Code", monospace',
        disableStdin: true,
        theme: {
          background: '#0d1117',
          foreground: '#e6edf3',
          cursor: 'transparent',
        },
      });

      const fitAddon = new FitAddon();
      term.loadAddon(fitAddon);

      await new Promise<void>((resolve) =>
        requestAnimationFrame(() => requestAnimationFrame(() => resolve())),
      );

      if (!isMounted || !containerRef.current) return;

      term.open(containerRef.current);
      fitAddon.fit();
      termRef.current = term;

      // open() 후 한 프레임 더 대기 — 렌더러가 첫 draw를 마쳐야 write()가 동작
      await new Promise<void>((resolve) => requestAnimationFrame(() => resolve()));
      if (!isMounted || !containerRef.current) {
        term.dispose();
        return;
      }

      const ro = new ResizeObserver(() => {
        if (containerRef.current) fitAddon.fit();
      });
      ro.observe(containerRef.current);

      const backendWS = process.env.NEXT_PUBLIC_WS_URL ?? 'ws://localhost:8080';
      const token = process.env.NODE_ENV === 'development' ? 'dev-token' : '';
      const wsURL = `${backendWS}/api/v1/instructor/sessions/${sessionId}/ws${token ? `?token=${token}` : ''}`;

      const ws = new WebSocket(wsURL);
      wsRef.current = ws;
      ws.binaryType = 'arraybuffer';

      ws.onopen = () => {
        if (isMounted) setConnState('connected');
      };

      ws.onmessage = (e: MessageEvent) => {
        if (!isMounted) return;
        const data: Uint8Array | string =
          e.data instanceof ArrayBuffer ? new Uint8Array(e.data) : (e.data as string);
        try {
          term.write(data);
        } catch (e) {
          console.warn('term.write failed', e);
        }
      };

      ws.onclose = () => {
        if (isMounted) {
          setConnState('closed');
          term.writeln('\r\n\x1b[31m● 연결 종료\x1b[0m');
          term.writeln('\x1b[33m수강생 터미널이 열려 있는지 확인하세요.\x1b[0m');
        }
      };

      ws.onerror = () => {
        if (isMounted) setConnState('error');
      };

      return () => ro.disconnect();
    }

    const cleanup = init();

    return () => {
      isMounted = false;
      wsRef.current?.close();
      termRef.current?.dispose();
      termRef.current = null;
      cleanup.then((fn) => fn?.());
    };
  }, [sessionId]);

  const stateConfig: Record<ConnState, { label: string; color: string }> = {
    connecting: { label: '연결 중...', color: 'text-amber-400' },
    connected: { label: '연결됨', color: 'text-emerald-400' },
    closed: { label: '연결 끊김', color: 'text-red-400' },
    error: { label: '연결 오류', color: 'text-red-400' },
  };

  const state = stateConfig[connState];

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 backdrop-blur-sm"
      onClick={(e) => e.target === e.currentTarget && onClose()}
    >
      <div
        className="w-[820px] max-w-[95vw] bg-slate-900 border border-slate-700 rounded-2xl shadow-2xl flex flex-col overflow-hidden"
        style={{ height: '500px' }}
      >
        {/* Header */}
        <div className="flex items-center justify-between px-4 py-3 border-b border-slate-800 flex-shrink-0">
          <div className="flex items-center gap-3">
            <span className="text-xs bg-amber-500/20 text-amber-400 border border-amber-500/30 px-2 py-0.5 rounded-md">
              관전 모드
            </span>
            <span className="text-slate-300 text-sm font-medium">{studentName}</span>
            <span className={`text-xs ${state.color}`}>● {state.label}</span>
          </div>
          <button
            onClick={onClose}
            className="text-slate-500 hover:text-white text-sm transition-colors"
          >
            닫기 ✕
          </button>
        </div>

        {/* Terminal container — 명시적 flex-1 + overflow-hidden */}
        <div className="flex-1 overflow-hidden bg-[#0d1117]">
          <div ref={containerRef} className="w-full h-full" style={{ padding: '8px' }} />
        </div>
      </div>
    </div>
  );
}
