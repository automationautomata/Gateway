package logging

import (
	"context"
	"log/slog"
)

type SlogAdapter struct {
	inner *slog.Logger
}

func NewSlogAdapter(l *slog.Logger) *SlogAdapter {
	return &SlogAdapter{inner: l}
}

func (l *SlogAdapter) Component(name string) *SlogAdapter {
	return NewSlogAdapter(l.inner.With("component", name))
}

func (l *SlogAdapter) Debug(ctx context.Context, msg string, fields map[string]any) {
	l.inner.LogAttrs(ctx, slog.LevelDebug, msg, mapToAttrs(fields)...)
}

func (l *SlogAdapter) Info(ctx context.Context, msg string, fields map[string]any) {
	l.inner.LogAttrs(ctx, slog.LevelInfo, msg, mapToAttrs(fields)...)
}

func (l *SlogAdapter) Warn(ctx context.Context, msg string, fields map[string]any) {
	l.inner.LogAttrs(ctx, slog.LevelWarn, msg, mapToAttrs(fields)...)
}

func (l *SlogAdapter) Error(ctx context.Context, msg string, fields map[string]any) {
	l.inner.LogAttrs(ctx, slog.LevelError, msg, mapToAttrs(fields)...)
}

func mapToAttrs(fields map[string]any) []slog.Attr {
	attrs := make([]slog.Attr, 0, len(fields))
	for k, v := range fields {
		attrs = append(attrs, slog.Any(k, v))
	}
	return attrs
}
