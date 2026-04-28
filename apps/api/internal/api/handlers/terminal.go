// 터미널 WebSocket 핸들러 — 수강생/강사 터미널 연결 관리.
// 개발 모드(Stub): 로컬 bash PTY 실행 + PTY 레지스트리에 등록 (강사 관전 지원).
// 프로덕션: VM SSH에 WebSocket 프록시 연결.
package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/creack/pty"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/kt-techuplabs/cledyu/backend/internal/infra/vm"
	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"
)

// WebSocket 업그레이더. CheckOrigin은 Kong Gateway 뒤에서 Origin 검증이 이중으로 되므로 여기선 생략.
var upgrader = websocket.Upgrader{
	HandshakeTimeout: 10 * time.Second,
	CheckOrigin:      func(r *http.Request) bool { return true },
}

// ── Student terminal ──────────────────────────────────────────────────────

// TerminalWS는 수강생의 xterm.js와 VM을 잇는 WebSocket 핸들러.
// 로컬 개발(Stub provider)이면 로컬 bash PTY를 시작하고,
// 프로덕션이면 VM의 SSH에 프록시 연결한다.
func (h *Handler) TerminalWS(c *gin.Context) {
	sessionID := c.Param("id")
	userID := c.GetString("user_id")

	ip, port, provider, err := h.sessions.VMInfo(c.Request.Context(), sessionID, userID)
	if err != nil {
		h.err(c, http.StatusBadRequest, err.Error())
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.log.Error("ws upgrade", zap.Error(err))
		return
	}
	defer conn.Close()
	// WebSocket 연결은 장시간 유지되므로 데드라인 제거
	conn.SetReadDeadline(time.Time{})
	conn.SetWriteDeadline(time.Time{})

	if provider == vm.ProviderStub || port == 0 {
		h.runLocalShell(conn, sessionID) // 개발 모드: 로컬 bash
	} else {
		proxySSH(conn, ip, port, h.log) // 프로덕션: VM SSH 프록시
	}
}

func (h *Handler) runLocalShell(conn *websocket.Conn, sessionID string) {
	cmd := exec.Command("/bin/bash", "--norc", "--noprofile")
	cmd.Env = []string{
		"TERM=xterm-256color",
		"HOME=" + os.Getenv("HOME"),
		"PATH=" + os.Getenv("PATH"),
		"LANG=en_US.UTF-8",
		"LC_ALL=en_US.UTF-8",
	}
	ptmx, err := pty.Start(cmd)
	if err != nil {
		h.log.Error("pty start", zap.Error(err))
		conn.WriteMessage(websocket.TextMessage, []byte("failed to start shell\r\n")) //nolint:errcheck
		return
	}

	entry := h.terminals.Register(sessionID, ptmx)
	h.log.Info("terminal registered", zap.String("session", sessionID))
	defer func() {
		if cmd.Process != nil {
			cmd.Process.Kill() //nolint:errcheck
		}
		ptmx.Close()
		h.terminals.Remove(sessionID)
		h.log.Info("terminal unregistered", zap.String("session", sessionID))
	}()

	// PTY output → student WebSocket + broadcast to instructor subscribers
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := ptmx.Read(buf)
			if n > 0 {
				conn.WriteMessage(websocket.BinaryMessage, buf[:n]) //nolint:errcheck
				entry.Broadcast(buf[:n])
			}
			if err != nil {
				return
			}
		}
	}()

	// Student WebSocket input → PTY
	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			return
		}
		ptmx.Write(data) //nolint:errcheck
	}
}

// ── Instructor terminals ──────────────────────────────────────────────────

func (h *Handler) InstructorTerminalWS(c *gin.Context) {
	sessionID := c.Param("id")

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()
	conn.SetReadDeadline(time.Time{})
	conn.SetWriteDeadline(time.Time{})

	entry, ok := h.terminals.Get(sessionID)
	if !ok {
		conn.WriteMessage(websocket.TextMessage, []byte("session not active\r\n")) //nolint:errcheck
		return
	}

	subID := newID()
	ch := entry.Subscribe(subID)
	defer entry.Unsubscribe(subID)

	conn.WriteMessage(websocket.TextMessage, []byte("\r\n\x1b[33m[강사 관전 모드 — 읽기 전용]\x1b[0m\r\n")) //nolint:errcheck

	for data := range ch {
		if err := conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
			return
		}
	}
}

func (h *Handler) InstructorInjectCommand(c *gin.Context) {
	sessionID := c.Param("id")
	var req struct {
		Command string `json:"command" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.err(c, http.StatusBadRequest, err.Error())
		return
	}

	entry, ok := h.terminals.Get(sessionID)
	if !ok {
		h.err(c, http.StatusNotFound, "session terminal not active")
		return
	}

	// \r 로 현재 입력 중인 줄을 취소한 뒤 명령어 실행.
	// PTY stdin에 ANSI 코드를 쓰면 bash가 명령어로 해석하므로 배너는 넣지 않음.
	// 강사 확인은 instructor 대시보드의 토스트로 처리.
	entry.Ptmx.WriteString("\r" + req.Command + "\n") //nolint:errcheck
	c.JSON(http.StatusOK, gin.H{"message": "command injected"})
}

// ── SSH proxy (production) ────────────────────────────────────────────────

func proxySSH(conn *websocket.Conn, ip string, port int, log *zap.Logger) {
	cfg := &ssh.ClientConfig{
		User:            "ubuntu",
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec
		Timeout:         10 * time.Second,
	}
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", ip, port), cfg)
	if err != nil {
		log.Error("ssh dial", zap.Error(err))
		conn.WriteMessage(websocket.TextMessage, []byte("ssh connection failed\r\n")) //nolint:errcheck
		return
	}
	defer client.Close()

	sshSess, err := client.NewSession()
	if err != nil {
		return
	}
	defer sshSess.Close()

	sshSess.RequestPty("xterm-256color", 40, 120, ssh.TerminalModes{}) //nolint:errcheck
	stdin, _ := sshSess.StdinPipe()
	stdout, _ := sshSess.StdoutPipe()
	sshSess.Shell() //nolint:errcheck

	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := stdout.Read(buf)
			if n > 0 {
				conn.WriteMessage(websocket.BinaryMessage, buf[:n]) //nolint:errcheck
			}
			if err != nil {
				return
			}
		}
	}()

	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			return
		}
		io.WriteString(stdin, string(data)) //nolint:errcheck
	}
}

func newID() string {
	b := make([]byte, 8)
	os.ReadFile("/dev/urandom") //nolint:errcheck
	return fmt.Sprintf("%x", b)
}
