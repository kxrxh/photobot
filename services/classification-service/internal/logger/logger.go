package logger

import (
	"context"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/trace"
)

const serviceName = "classification-service"

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

func WrapGlobalWriter(wrapper func(inner io.Writer) io.Writer) {
	globalWriter.mu.Lock()
	defer globalWriter.mu.Unlock()
	globalWriter.writer = wrapper(globalWriter.writer)
}

var (
	globalWriter = &swappableWriter{writer: os.Stdout}
	Logger       zerolog.Logger
)

func init() {
	InitLogger()
}

func InitLogger() {
	zerolog.TimeFieldFormat = time.RFC3339

	isDebug := strings.ToLower(os.Getenv("DEBUG")) == "true"

	ctx := zerolog.New(globalWriter).With().Timestamp().Str("service", serviceName)
	if isDebug {
		ctx = ctx.Caller()
	}
	Logger = ctx.Logger()
	log.Logger = Logger

	if isDebug {
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

func GetLogger(component string) zerolog.Logger {
	return Logger.With().Str("component", component).Logger()
}

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
