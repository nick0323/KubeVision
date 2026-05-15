package api

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
)

type WebSocketManager struct {
	maxConns     int32
	activeCount  atomic.Int32
	shuttingDown atomic.Bool
	shutdownCh   chan struct{}
	wg           sync.WaitGroup
	stopOnce     sync.Once
}

func NewWebSocketManager(maxConnections int) *WebSocketManager {
	m := &WebSocketManager{
		shutdownCh: make(chan struct{}),
	}
	if maxConnections > 0 {
		m.maxConns = int32(maxConnections)
	} else {
		m.maxConns = 1 << 30
	}
	return m
}

func (m *WebSocketManager) Acquire() error {
	if m == nil {
		return nil
	}
	if m.shuttingDown.Load() {
		return fmt.Errorf("server is shutting down")
	}
	current := m.activeCount.Add(1)
	if current > m.maxConns {
		m.activeCount.Add(-1)
		return fmt.Errorf("connection limit exceeded")
	}
	m.wg.Add(1)
	return nil
}

func (m *WebSocketManager) Release() {
	if m == nil {
		return
	}
	m.activeCount.Add(-1)
	m.wg.Done()
}

func (m *WebSocketManager) ShutdownCtx(ctx context.Context) context.Context {
	if m == nil || ctx == nil {
		return ctx
	}
	shutdownCtx, cancel := context.WithCancel(ctx)
	go func() {
		select {
		case <-m.shutdownCh:
			cancel()
		case <-shutdownCtx.Done():
		}
	}()
	return shutdownCtx
}

func (m *WebSocketManager) Shutdown(ctx context.Context) error {
	if m == nil {
		return nil
	}
	m.shuttingDown.Store(true)
	m.stopOnce.Do(func() {
		close(m.shutdownCh)
	})

	done := make(chan struct{})
	go func() {
		m.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (m *WebSocketManager) ActiveCount() int32 {
	if m == nil {
		return 0
	}
	return m.activeCount.Load()
}

func (m *WebSocketManager) IsShuttingDown() bool {
	if m == nil {
		return false
	}
	return m.shuttingDown.Load()
}
