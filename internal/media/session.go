package media

import "sync/atomic"

type SessionStats struct {
	viewers int64
}

func NewSessionStats() *SessionStats {
	return &SessionStats{}
}

func (s *SessionStats) AddViewer() {
	atomic.AddInt64(&s.viewers, 1)
}

func (s *SessionStats) RemoveViewer() {
	atomic.AddInt64(&s.viewers, -1)
}

func (s *SessionStats) ViewerCount() int64 {
	return atomic.LoadInt64(&s.viewers)
}
