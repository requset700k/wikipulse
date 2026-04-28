// xterm.js 기반 브라우저 터미널 컴포넌트.
// 백엔드 WebSocket(/api/v1/sessions/:id/ws)에 직접 연결해 VM PTY 스트림을 주고받는다.
// Next.js rewrite는 HTTP만 프록시하므로 WS는 NEXT_PUBLIC_WS_URL로 직접 연결.
//
// xterm 초기화 시 주의사항:
//   - open() 전에 requestAnimationFrame 두 번 대기 → DOM에 실제 크기가 잡힌 후 렌더러 초기화
//   - FitAddon: 컨테이너 크기에 맞게 터미널 cols/rows 자동 조정 (ResizeObserver 연동)
//   - IME(한글 등 조합 입력) 처리: compositionend에서 직접 전송, 중복 onData는 억제

'use client';

import { useEffect, useRef } from 'react';

// xterm 타입 선언이 복잡하고 동적 import로 로드하므로 any 처리
// eslint-disable-next-line @typescript-eslint/no-explicit-any
type XTermInstance = any;

interface Props {
  sessionId: string;
  onDisconnect?: () => void;
}

export function Terminal({ sessionId, onDisconnect }: Props) {
  const containerRef = useRef<HTMLDivElement>(null);
  const termRef = useRef<XTermInstance | null>(null);
  const wsRef = useRef<WebSocket | null>(null);
  // onDisconnect를 ref로 관리 — 함수 참조가 바뀌어도 effect 재실행 방지
  const onDisconnectRef = useRef(onDisconnect);
  useEffect(() => {
    onDisconnectRef.current = onDisconnect;
  }, [onDisconnect]);

  useEffect(() => {
    let isMounted = true;

    async function init() {
      if (!containerRef.current) return;

      const { Terminal: XTerm } = await import('@xterm/xterm');
      const { FitAddon } = await import('@xterm/addon-fit');

      if (!isMounted || !containerRef.current) return;

      const term: XTermInstance = new XTerm({
        cursorBlink: true,
        fontSize: 14,
        fontFamily: '"JetBrains Mono", "Fira Code", monospace',
        theme: {
          background: '#0d1117',
          foreground: '#e6edf3',
          cursor: '#58a6ff',
          selectionBackground: '#264f78',
          black: '#0d1117',
          red: '#ff7b72',
          green: '#3fb950',
          yellow: '#d29922',
          blue: '#58a6ff',
          magenta: '#bc8cff',
          cyan: '#39c5cf',
          white: '#b1bac4',
        },
      });

      const fitAddon = new FitAddon();
      term.loadAddon(fitAddon);

      // 컨테이너에 실제 크기가 잡힌 다음 open() — 그래야 xterm 내부 renderer가
      // 바로 dimensions를 읽어도 undefined가 나지 않음.
      await new Promise<void>((resolve) =>
        requestAnimationFrame(() => requestAnimationFrame(() => resolve())),
      );

      if (!isMounted || !containerRef.current) return;

      term.open(containerRef.current);
      fitAddon.fit(); // open() 직후 바로 fit — 내부 타이머 전에 dimensions 확정
      termRef.current = term;

      // ResizeObserver for responsive resize
      const ro = new ResizeObserver(() => {
        if (containerRef.current) fitAddon.fit();
      });
      ro.observe(containerRef.current);

      // WebSocket — direct to backend (Next.js rewrites don't proxy WS)
      const backendWS = process.env.NEXT_PUBLIC_WS_URL ?? 'ws://localhost:8080';
      const token = process.env.NODE_ENV === 'development' ? 'dev-token' : '';
      const wsURL = `${backendWS}/api/v1/sessions/${sessionId}/ws${token ? `?token=${token}` : ''}`;

      const ws = new WebSocket(wsURL);
      wsRef.current = ws;
      ws.binaryType = 'arraybuffer';

      ws.onopen = () => {
        if (isMounted) term.writeln('\r\n\x1b[32m● Connected\x1b[0m\r\n');
      };

      ws.onmessage = (e: MessageEvent) => {
        if (!isMounted) return;
        const data: Uint8Array | string =
          e.data instanceof ArrayBuffer ? new Uint8Array(e.data) : (e.data as string);
        try {
          term.write(data);
        } catch {
          /* renderer not ready yet */
        }
      };

      ws.onclose = () => {
        if (isMounted) {
          term.writeln('\r\n\x1b[31m● Connection closed\x1b[0m');
          onDisconnectRef.current?.();
        }
      };

      ws.onerror = () => {
        if (isMounted) term.writeln('\r\n\x1b[31m● WebSocket error\x1b[0m');
      };

      // IME(한글·일본어·중국어) 입력 처리
      //
      // 문제: "안녕하세요" 입력 시 음절 전환마다
      //   compositionend("안") → compositionstart("ㄴ") → onData("안") 순서로 발생.
      //   compositionstart가 onData보다 먼저 와서 isComposing=true가 되면
      //   "안"이 차단되고 마지막 "요"만 전송됨.
      //
      // 해결: compositionend에서 e.data를 직접 전송하고,
      //   이후 중복으로 오는 onData는 suppressNextData 플래그로 차단.
      let isComposing = false;
      let suppressNextData = false;

      const textarea =
        containerRef.current?.querySelector<HTMLTextAreaElement>('.xterm-helper-textarea');

      if (textarea) {
        textarea.addEventListener('compositionstart', () => {
          isComposing = true;
        });
        textarea.addEventListener('compositionend', (e: CompositionEvent) => {
          isComposing = false;
          if (ws.readyState === WebSocket.OPEN && e.data) {
            ws.send(e.data);
            suppressNextData = true; // 동일 문자 onData 중복 전송 방지
          }
        });
      }

      term.onData((data: string) => {
        if (ws.readyState !== WebSocket.OPEN) return;
        if (isComposing) return;
        if (suppressNextData) {
          suppressNextData = false;
          return;
        }
        ws.send(data);
      });

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
  }, [sessionId]); // onDisconnect는 ref로 관리 — 변경 시 effect 재실행 안 함

  return (
    <div ref={containerRef} className="w-full h-full bg-[#0d1117]" style={{ padding: '8px' }} />
  );
}
