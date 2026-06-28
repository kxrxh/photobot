package logger

import (
	"context"
	"io"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/trace"
)

// swappableWriter is a thread-safe io.Writer that allows swapping the underlying writer at runtime.
type swappableWriter struct {
	mu     sync.RWMutex
	writer io.Writer
}

func (s *swappableWriter) Write(p []byte) (n int, err error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.writer.Write(p)
}

func (s *swappableWriter) SetWriter(w io.Writer) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.writer = w
}

// WrapGlobalWriter wraps the current output writer (e.g. OTLP JSON tap).
func WrapGlobalWriter(wrapper func(inner io.Writer) io.Writer) {
	globalWriter.mu.Lock()
	defer globalWriter.mu.Unlock()
	globalWriter.writer = wrapper(globalWriter.writer)
}

var (
	globalWriter = &swappableWriter{writer: os.Stdout}
	Logger       = zerolog.New(globalWriter).With().Timestamp().Logger()
)

func init() {
	log.Logger = Logger
	InitLogger(false)
}

// InitLogger sets global log level and writer: debug uses console; otherwise JSON to stdout.
func InitLogger(debug bool) {
	zerolog.TimeFieldFormat = time.RFC3339

	if debug {
		consoleWriter := zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
		globalWriter.SetWriter(consoleWriter)
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		globalWriter.SetWriter(os.Stdout)
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}

// WithTrace returns log enriched with trace_id and span_id when ctx carries a span.
func WithTrace(ctx context.Context, log zerolog.Logger) zerolog.Logger {
	sc := trace.SpanContextFromContext(ctx)
	if !sc.IsValid() {
		return log
	}
	return log.With().
		Str("trace_id", sc.TraceID().String()).
		Str("span_id", sc.SpanID().String()).
		Logger()
}
