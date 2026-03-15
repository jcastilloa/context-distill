package server

import (
	"bytes"
	"fmt"
	"io"
	"sync"
)

type protocolAwareWriter struct {
	target io.Writer
	mode   *stdioProtocolMode
	mu     sync.Mutex
}

func newProtocolAwareWriter(target io.Writer, mode *stdioProtocolMode) *protocolAwareWriter {
	return &protocolAwareWriter{
		target: target,
		mode:   mode,
	}
}

func (w *protocolAwareWriter) Write(payload []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.mode.Get() != stdioProtocolContentLength {
		return w.target.Write(payload)
	}

	body := bytes.TrimSpace(payload)
	if len(body) == 0 {
		return len(payload), nil
	}

	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(body))
	if _, err := io.WriteString(w.target, header); err != nil {
		return 0, err
	}
	if _, err := w.target.Write(body); err != nil {
		return 0, err
	}

	return len(payload), nil
}
