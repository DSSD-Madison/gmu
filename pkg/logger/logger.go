package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"sync"
)

type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)

	DebugContext(ctx context.Context, msg string, args ...any)
	InfoContext(ctx context.Context, msg string, args ...any)
	WarnContext(ctx context.Context, msg string, args ...any)
	ErrorContext(ctx context.Context, msg string, args ...any)

	With(args ...any) Logger
}

type slogWrapper struct {
	*slog.Logger
}

func (w *slogWrapper) With(args ...any) Logger {
	return &slogWrapper{Logger: w.Logger.With(args...)}
}

type HandlerOptions struct {
	Mode        string
	AddSource   bool
	Level       slog.Leveler
	ReplaceAttr func(groups []string, a slog.Attr) slog.Attr
}

const (
	reset = "\033[0m"

	black        = 30
	red          = 31
	green        = 32
	yellow       = 33
	blue         = 34
	magenta      = 35
	cyan         = 36
	lightGray    = 37
	darkGray     = 90
	lightRed     = 91
	lightGreen   = 92
	lightYellow  = 93
	lightBlue    = 94
	lightMagenta = 95
	lightCyan    = 96
	white        = 97
)

func colorize(colorCode int, v string) string {
	return fmt.Sprintf("\033[%sm%s%s", strconv.Itoa(colorCode), v, reset)
}

type Handler struct {
	handler     slog.Handler
	bytes       *bytes.Buffer
	mutex       *sync.Mutex
	prettyPrint bool
}

func (h *Handler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &Handler{
		handler:     h.handler.WithAttrs(attrs),
		bytes:       h.bytes,
		mutex:       h.mutex,
		prettyPrint: h.prettyPrint,
	}
}

func (h *Handler) WithGroup(name string) slog.Handler {
	return &Handler{handler: h.handler.WithGroup(name), bytes: h.bytes, mutex: h.mutex}
}

const (
	timeFormat = "[15:04:05.000]"
)

func (h *Handler) Handle(ctx context.Context, r slog.Record) error {
	level := r.Level.String() + ":"

	if h.prettyPrint {
		switch r.Level {
		case slog.LevelDebug:
			level = colorize(darkGray, level)
		case slog.LevelInfo:
			level = colorize(cyan, level)
		case slog.LevelWarn:
			level = colorize(lightYellow, level)
		case slog.LevelError:
			level = colorize(lightRed, level)
		}
	}

	attrs, err := h.computeAttrs(ctx, r)
	if err != nil {
		return err
	}

	jsonPayload, err := json.Marshal(attrs)
	if err != nil {
		return fmt.Errorf("error when marshalling attrs: %w", err)
	}

	if h.prettyPrint {
		fmt.Println(
			colorize(white, r.Time.Format(timeFormat)),
			level,
			colorize(white, r.Message),
			colorize(lightGray, string(jsonPayload)),
		)
	} else {
		fmt.Println(
			r.Time.Format(timeFormat),
			level,
			r.Message,
			string(jsonPayload),
		)
	}

	return nil
}

// TODO: Potentially replace the bytes buffer with a sync pool
func (h *Handler) computeAttrs(
	ctx context.Context,
	r slog.Record,
) (map[string]any, error) {
	h.mutex.Lock()
	defer func() {
		h.bytes.Reset()
		h.mutex.Unlock()
	}()
	if err := h.handler.Handle(ctx, r); err != nil {
		return nil, fmt.Errorf("error when calling inner handler's Handle: %w", err)
	}

	var attrs map[string]any
	err := json.Unmarshal(h.bytes.Bytes(), &attrs)
	if err != nil {
		return nil, fmt.Errorf("error when unmarshalling inner handler's Handle result: %w", err)
	}
	return attrs, nil
}

func suppressDefaults(
	next func([]string, slog.Attr) slog.Attr,
) func([]string, slog.Attr) slog.Attr {
	return func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.TimeKey ||
			a.Key == slog.LevelKey ||
			a.Key == slog.MessageKey {
			return slog.Attr{}
		}
		if next == nil {
			return a
		}
		return next(groups, a)
	}
}

func newCustomHandler(opts *HandlerOptions) *Handler {
	if opts == nil {
		opts = &HandlerOptions{}
	}
	var prettyPrint bool

	switch opts.Mode {
	case "dev":
		prettyPrint = true
	case "prod":
		prettyPrint = false
	default:
		prettyPrint = false
	}

	b := &bytes.Buffer{}
	return &Handler{
		bytes: b,
		handler: slog.NewJSONHandler(b, &slog.HandlerOptions{
			Level:       opts.Level,
			AddSource:   opts.AddSource,
			ReplaceAttr: suppressDefaults(opts.ReplaceAttr),
		}),
		mutex:       &sync.Mutex{},
		prettyPrint: prettyPrint,
	}
}

func New(opts *HandlerOptions) Logger {
	customHandler := newCustomHandler(opts)
	slogger := slog.New(customHandler)

	return &slogWrapper{Logger: slogger}
}
