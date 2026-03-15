package server

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"
)

func TestReadStdioMessageLineDelimited(t *testing.T) {
	r := bufio.NewReader(strings.NewReader("{\"jsonrpc\":\"2.0\",\"id\":1}\n"))

	msg, protocol, err := readStdioMessage(r)
	if err != nil {
		t.Fatalf("readStdioMessage() error = %v", err)
	}

	if protocol != stdioProtocolLineDelimited {
		t.Fatalf("unexpected protocol: %v", protocol)
	}

	if got, want := string(msg), `{"jsonrpc":"2.0","id":1}`; got != want {
		t.Fatalf("unexpected msg: got %q want %q", got, want)
	}
}

func TestReadStdioMessageContentLength(t *testing.T) {
	body := `{"jsonrpc":"2.0","id":1,"method":"initialize"}`
	payload := fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(body), body)
	r := bufio.NewReader(strings.NewReader(payload))

	msg, protocol, err := readStdioMessage(r)
	if err != nil {
		t.Fatalf("readStdioMessage() error = %v", err)
	}

	if protocol != stdioProtocolContentLength {
		t.Fatalf("unexpected protocol: %v", protocol)
	}

	if got := string(msg); got != body {
		t.Fatalf("unexpected msg: got %q want %q", got, body)
	}
}

func TestReadStdioMessageContentLengthWithExtraHeaders(t *testing.T) {
	body := `{"jsonrpc":"2.0","id":2,"method":"ping"}`
	payload := fmt.Sprintf("Content-Length: %d\r\nContent-Type: application/json\r\n\r\n%s", len(body), body)
	r := bufio.NewReader(strings.NewReader(payload))

	msg, protocol, err := readStdioMessage(r)
	if err != nil {
		t.Fatalf("readStdioMessage() error = %v", err)
	}

	if protocol != stdioProtocolContentLength {
		t.Fatalf("unexpected protocol: %v", protocol)
	}

	if got := string(msg); got != body {
		t.Fatalf("unexpected msg: got %q want %q", got, body)
	}
}

func TestReadStdioMessageInvalidContentLength(t *testing.T) {
	r := bufio.NewReader(strings.NewReader("Content-Length: abc\r\n\r\n{}"))

	_, _, err := readStdioMessage(r)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestProtocolAwareWriterLineDelimited(t *testing.T) {
	var out bytes.Buffer
	mode := newStdioProtocolMode()
	mode.Set(stdioProtocolLineDelimited)
	w := newProtocolAwareWriter(&out, mode)

	in := []byte("{\"jsonrpc\":\"2.0\",\"id\":1}\n")
	n, err := w.Write(in)
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	if n != len(in) {
		t.Fatalf("unexpected written size: got %d want %d", n, len(in))
	}

	if got := out.String(); got != string(in) {
		t.Fatalf("unexpected output: got %q want %q", got, string(in))
	}
}

func TestProtocolAwareWriterContentLength(t *testing.T) {
	var out bytes.Buffer
	mode := newStdioProtocolMode()
	mode.Set(stdioProtocolContentLength)
	w := newProtocolAwareWriter(&out, mode)

	body := `{"jsonrpc":"2.0","id":1,"result":{}}`
	in := []byte(body + "\n")
	n, err := w.Write(in)
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	if n != len(in) {
		t.Fatalf("unexpected written size: got %d want %d", n, len(in))
	}

	if got, want := out.String(), fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(body), body); got != want {
		t.Fatalf("unexpected output: got %q want %q", got, want)
	}
}

func TestReadStdioMessageEOFWithNoData(t *testing.T) {
	r := bufio.NewReader(strings.NewReader(""))
	_, _, err := readStdioMessage(r)
	if err == nil {
		t.Fatal("expected EOF")
	}

	if err != io.EOF {
		t.Fatalf("unexpected error: %v", err)
	}
}
