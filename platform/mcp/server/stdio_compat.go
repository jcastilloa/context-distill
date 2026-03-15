package server

import (
	"bufio"
	"context"
	"errors"
	"io"
	"os"
	"os/signal"
	"syscall"

	mcpserver "github.com/mark3labs/mcp-go/server"
)

func serveStdioCompat(server *mcpserver.MCPServer) error {
	mode := newStdioProtocolMode()
	adaptedOutput := newProtocolAwareWriter(os.Stdout, mode)
	adaptedInputReader, adaptedInputWriter := io.Pipe()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGTERM, syscall.SIGINT)
	defer signal.Stop(signalChannel)

	go func() {
		select {
		case <-signalChannel:
			cancel()
		case <-ctx.Done():
		}
	}()

	go normalizeStdioInput(ctx, os.Stdin, adaptedInputWriter, mode)

	stdioServer := mcpserver.NewStdioServer(server)
	err := stdioServer.Listen(ctx, adaptedInputReader, adaptedOutput)
	if errors.Is(err, context.Canceled) {
		return nil
	}
	return err
}

func normalizeStdioInput(
	ctx context.Context,
	stdin io.Reader,
	normalizedWriter *io.PipeWriter,
	mode *stdioProtocolMode,
) {
	defer normalizedWriter.Close()

	reader := bufio.NewReader(stdin)
	for {
		select {
		case <-ctx.Done():
			_ = normalizedWriter.CloseWithError(ctx.Err())
			return
		default:
		}

		message, detectedProtocol, err := readStdioMessage(reader)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return
			}
			_ = normalizedWriter.CloseWithError(err)
			return
		}
		if len(message) == 0 {
			continue
		}

		mode.SetIfUnknown(detectedProtocol)

		if _, err := normalizedWriter.Write(message); err != nil {
			return
		}
		if _, err := normalizedWriter.Write([]byte{'\n'}); err != nil {
			return
		}
	}
}
