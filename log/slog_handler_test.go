package log

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"os"
	"strings"
	"testing"
)

func Test_SetupLogWithWriter_BytesBufferContainsLogs(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	testWriter := io.MultiWriter(os.Stderr, buf)
	SetupLogWithWriter(slog.LevelInfo, testWriter)

	testMessage := "this is our test message"
	slog.InfoContext(context.Background(), testMessage, slog.String("key", "value"))

	// Check if the log output contains the expected message
	if buf.Len() == 0 {
		t.Error("Expected log output, but got none.")
	}
	if buf.String() == "" {
		t.Error("Expected log output, but got empty string.")
	}
	if !strings.Contains(buf.String(), testMessage) {
		t.Errorf("Expected log output to contain %v, but got: %v", testMessage, buf.String())
	}
}
