package server

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

func readStdioMessage(reader *bufio.Reader) ([]byte, stdioProtocol, error) {
	for {
		nextByte, err := reader.Peek(1)
		if err != nil {
			return nil, stdioProtocolUnknown, err
		}

		if isWhitespace(nextByte[0]) {
			if _, err := reader.ReadByte(); err != nil {
				return nil, stdioProtocolUnknown, err
			}
			continue
		}

		if nextByte[0] == '{' || nextByte[0] == '[' {
			return readLineDelimitedMessage(reader)
		}

		return readFramedOrFallbackMessage(reader)
	}
}

func readLineDelimitedMessage(reader *bufio.Reader) ([]byte, stdioProtocol, error) {
	line, err := reader.ReadBytes('\n')
	if errors.Is(err, io.EOF) && len(line) == 0 {
		return nil, stdioProtocolLineDelimited, io.EOF
	}
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, stdioProtocolLineDelimited, err
	}

	message := bytes.TrimSpace(line)
	if len(message) == 0 {
		if errors.Is(err, io.EOF) {
			return nil, stdioProtocolLineDelimited, io.EOF
		}
		return nil, stdioProtocolLineDelimited, nil
	}

	return message, stdioProtocolLineDelimited, nil
}

func readFramedOrFallbackMessage(reader *bufio.Reader) ([]byte, stdioProtocol, error) {
	headerLine, err := reader.ReadString('\n')
	if errors.Is(err, io.EOF) && strings.TrimSpace(headerLine) == "" {
		return nil, stdioProtocolUnknown, io.EOF
	}
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, stdioProtocolUnknown, err
	}

	trimmedHeader := strings.TrimSpace(headerLine)
	if trimmedHeader == "" {
		if errors.Is(err, io.EOF) {
			return nil, stdioProtocolUnknown, io.EOF
		}
		return nil, stdioProtocolUnknown, nil
	}

	contentLength, hasHeader, parseErr := parseContentLengthHeader(trimmedHeader)
	if parseErr != nil {
		return nil, stdioProtocolContentLength, parseErr
	}
	if !hasHeader {
		return []byte(trimmedHeader), stdioProtocolLineDelimited, nil
	}

	if err := consumeFrameHeaders(reader); err != nil {
		return nil, stdioProtocolContentLength, err
	}

	body, err := readFrameBody(reader, contentLength)
	if err != nil {
		return nil, stdioProtocolContentLength, err
	}

	return bytes.TrimSpace(body), stdioProtocolContentLength, nil
}

func parseContentLengthHeader(header string) (int, bool, error) {
	const prefix = "content-length:"
	lowerHeader := strings.ToLower(header)
	if !strings.HasPrefix(lowerHeader, prefix) {
		return 0, false, nil
	}

	lengthText := strings.TrimSpace(header[len(prefix):])
	length, err := strconv.Atoi(lengthText)
	if err != nil || length < 0 {
		return 0, true, fmt.Errorf("invalid Content-Length value %q", lengthText)
	}

	return length, true, nil
}

func consumeFrameHeaders(reader *bufio.Reader) error {
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		if strings.TrimSpace(line) == "" {
			return nil
		}
	}
}

func readFrameBody(reader *bufio.Reader, length int) ([]byte, error) {
	if length == 0 {
		return []byte{}, nil
	}

	body := make([]byte, length)
	if _, err := io.ReadFull(reader, body); err != nil {
		return nil, err
	}

	return body, nil
}

func isWhitespace(value byte) bool {
	switch value {
	case ' ', '\t', '\r', '\n':
		return true
	default:
		return false
	}
}
