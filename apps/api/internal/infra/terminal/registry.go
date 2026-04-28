// Package terminal manages active PTY sessions so instructors can subscribe
// to a student's terminal output without a separate SSH connection.
package terminal

import (
	"os"
	"sync"
)

// Global is the process-wide registry of active local-shell sessions.
var Global = &Registry{sessions: make(map[string]*Entry)}

type Registry struct {
	mu       sync.RWMutex
	sessions map[string]*Entry
}

// Entry represents one active PTY session and its broadcast subscribers.
type Entry struct {
	Ptmx *os.File
	mu   sync.Mutex
	subs map[string]chan []byte
}

func (r *Registry) Register(sessionID string, ptmx *os.File) *Entry {
	r.mu.Lock()
	defer r.mu.Unlock()
	e := &Entry{Ptmx: ptmx, subs: make(map[string]chan []byte)}
	r.sessions[sessionID] = e
	return e
}

func (r *Registry) Get(sessionID string) (*Entry, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	e, ok := r.sessions[sessionID]
	return e, ok
}

func (r *Registry) Remove(sessionID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if e, ok := r.sessions[sessionID]; ok {
		e.mu.Lock()
		for _, ch := range e.subs {
			close(ch)
		}
		e.mu.Unlock()
	}
	delete(r.sessions, sessionID)
}

// Subscribe returns a channel that receives copies of all PTY output.
func (e *Entry) Subscribe(subID string) chan []byte {
	e.mu.Lock()
	defer e.mu.Unlock()
	ch := make(chan []byte, 256)
	e.subs[subID] = ch
	return ch
}

func (e *Entry) Unsubscribe(subID string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if ch, ok := e.subs[subID]; ok {
		close(ch)
		delete(e.subs, subID)
	}
}

// Broadcast copies data to all subscriber channels (instructor viewers).
func (e *Entry) Broadcast(data []byte) {
	e.mu.Lock()
	defer e.mu.Unlock()
	for _, ch := range e.subs {
		cp := make([]byte, len(data))
		copy(cp, data)
		select {
		case ch <- cp:
		default: // slow subscriber — drop rather than block
		}
	}
}
