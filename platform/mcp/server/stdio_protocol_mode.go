package server

import "sync"

type stdioProtocol int

const (
	stdioProtocolUnknown stdioProtocol = iota
	stdioProtocolLineDelimited
	stdioProtocolContentLength
)

type stdioProtocolMode struct {
	mu       sync.RWMutex
	protocol stdioProtocol
}

func newStdioProtocolMode() *stdioProtocolMode {
	return &stdioProtocolMode{protocol: stdioProtocolUnknown}
}

func (m *stdioProtocolMode) Get() stdioProtocol {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.protocol
}

func (m *stdioProtocolMode) Set(protocol stdioProtocol) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.protocol = protocol
}

func (m *stdioProtocolMode) SetIfUnknown(protocol stdioProtocol) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.protocol == stdioProtocolUnknown {
		m.protocol = protocol
	}
}
