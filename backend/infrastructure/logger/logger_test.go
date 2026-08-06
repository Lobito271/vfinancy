package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
)

func TestNew_JSONFormat(t *testing.T) {
	var buf bytes.Buffer
	l := NewWith(&buf, "info", "json", "stdout")

	l.Info("hello", "key", "value")

	var rec map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &rec); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, buf.String())
	}
	if rec["msg"] != "hello" {
		t.Errorf("expected msg=hello, got %v", rec["msg"])
	}
	if rec["key"] != "value" {
		t.Errorf("expected key=value, got %v", rec["key"])
	}
	if rec["level"] != "INFO" {
		t.Errorf("expected level=INFO, got %v", rec["level"])
	}
}

func TestNew_TextFormat(t *testing.T) {
	var buf bytes.Buffer
	l := NewWith(&buf, "debug", "text", "stdout")

	l.Debug("ping")

	out := buf.String()
	if !strings.Contains(out, "ping") {
		t.Errorf("expected output to contain 'ping', got %q", out)
	}
	if !strings.Contains(out, "level=DEBUG") {
		t.Errorf("expected level=DEBUG in output, got %q", out)
	}
}

func TestNew_LevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	l := NewWith(&buf, "warn", "json", "stdout")

	l.Info("ignored")
	l.Warn("kept")

	out := buf.String()
	if strings.Contains(out, "ignored") {
		t.Errorf("expected Info to be filtered at warn level, got %s", out)
	}
	if !strings.Contains(out, "kept") {
		t.Errorf("expected Warn to be logged, got %s", out)
	}
}

func TestParseLevel(t *testing.T) {
	cases := []struct {
		level string
		logAt slog.Level
		want  string
	}{
		{"debug", slog.LevelDebug, "DEBUG"},
		{"warn", slog.LevelWarn, "WARN"},
		{"error", slog.LevelError, "ERROR"},
		{"info", slog.LevelInfo, "INFO"},
		{"", slog.LevelInfo, "INFO"},
		{"other", slog.LevelInfo, "INFO"},
	}
	for _, c := range cases {
		var buf bytes.Buffer
		l := NewWith(&buf, c.level, "json", "stdout")
		l.Log(context.Background(), c.logAt, "x")
		var rec map[string]any
		if err := json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &rec); err != nil {
			t.Errorf("level=%q: not valid json: %v", c.level, err)
			continue
		}
		if rec["level"] != c.want {
			t.Errorf("level=%q want %q, got %v", c.level, c.want, rec["level"])
		}
	}
}
