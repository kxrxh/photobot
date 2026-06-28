package observability

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"time"

	otellog "go.opentelemetry.io/otel/log"
	otelglobal "go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/trace"
)

const otelLogScope = "csort.ru/analysis-service"

func wrapWriterWithOTLPJSONTap(next io.Writer) io.Writer {
	lg := otelglobal.Logger(otelLogScope)
	return &otlpJSONTap{next: next, lg: lg}
}

type otlpJSONTap struct {
	next io.Writer
	lg   otellog.Logger
}

func (w *otlpJSONTap) Write(p []byte) (n int, err error) {
	n, err = w.next.Write(p)
	line := bytes.TrimRight(p, "\n\r")
	if len(line) > 0 {
		w.emitLine(line)
	}
	return n, err
}

func (w *otlpJSONTap) emitLine(line []byte) {
	var m map[string]interface{}
	if err := json.Unmarshal(line, &m); err != nil {
		var r otellog.Record
		r.SetTimestamp(time.Now())
		r.SetSeverity(otellog.SeverityInfo)
		r.SetSeverityText("info")
		r.SetBody(otellog.StringValue(string(line)))
		w.lg.Emit(context.Background(), r)
		return
	}

	var r otellog.Record
	if ts := parseZerologTime(m); !ts.IsZero() {
		r.SetTimestamp(ts)
	} else {
		r.SetTimestamp(time.Now())
	}

	sev, text := mapZerologLevelToOtel(m)
	r.SetSeverity(sev)
	r.SetSeverityText(text)
	r.SetBody(otellog.StringValue(bodyFromZerologMap(m)))
	addZerologMapAsAttributes(&r, m)

	ctx := contextFromTraceFields(m)
	w.lg.Emit(ctx, r)
}

func parseZerologTime(m map[string]interface{}) time.Time {
	raw, ok := m["time"].(string)
	if !ok || raw == "" {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return time.Time{}
	}
	return t
}

func mapZerologLevelToOtel(m map[string]interface{}) (otellog.Severity, string) {
	lv, _ := m["level"].(string)
	switch lv {
	case "trace":
		return otellog.SeverityTrace, lv
	case "debug":
		return otellog.SeverityDebug, lv
	case "info":
		return otellog.SeverityInfo, lv
	case "warn", "warning":
		return otellog.SeverityWarn, lv
	case "error":
		return otellog.SeverityError, lv
	case "fatal", "panic":
		return otellog.SeverityFatal, lv
	default:
		return otellog.SeverityInfo, lv
	}
}

func bodyFromZerologMap(m map[string]interface{}) string {
	msg, _ := m["message"].(string)
	if msg != "" {
		return msg
	}

	method := stringField(m, "http.method", "method")
	urlFull := stringField(m, "url.full", "url")
	path := stringField(m, "url.path", "path")
	status := intFromMap(m, "http.status_code", "status")
	latencyMs := intFromMap(m, "http.latency_ms")
	latencyStr := stringField(m, "http.latency", "latency")

	if method != "" && (urlFull != "" || path != "") && status > 0 {
		target := urlFull
		if target == "" {
			target = path
		}
		if latencyMs > 0 {
			return fmt.Sprintf("HTTP %s %s -> %d in %dms", method, target, status, latencyMs)
		}
		if latencyStr != "" {
			return fmt.Sprintf("HTTP %s %s -> %d (%s)", method, target, status, latencyStr)
		}
		return fmt.Sprintf("HTTP %s %s -> %d", method, target, status)
	}

	return ""
}

func stringField(m map[string]interface{}, keys ...string) string {
	for _, k := range keys {
		if s, ok := m[k].(string); ok && s != "" {
			return s
		}
	}
	return ""
}

func intFromMap(m map[string]interface{}, keys ...string) int {
	for _, k := range keys {
		switch v := m[k].(type) {
		case float64:
			return int(v)
		case int:
			return v
		case int64:
			return int(v)
		case string:
			if n, err := strconv.Atoi(v); err == nil {
				return n
			}
		}
	}
	return 0
}

func addZerologMapAsAttributes(r *otellog.Record, m map[string]interface{}) {
	const maxAttrs = 48
	n := 0
	for k, v := range m {
		if n >= maxAttrs {
			break
		}
		switch k {
		case "level", "message", "time":
			continue
		}
		if kv, ok := anyToLogKV(k, v); ok {
			r.AddAttributes(kv)
			n++
		}
	}
}

func anyToLogKV(k string, v interface{}) (otellog.KeyValue, bool) {
	switch x := v.(type) {
	case string:
		if x == "" {
			return otellog.KeyValue{}, false
		}
		return otellog.String(k, x), true
	case bool:
		return otellog.Bool(k, x), true
	case float64:
		if x == float64(int64(x)) {
			return otellog.Int64(k, int64(x)), true
		}
		return otellog.Float64(k, x), true
	case int:
		return otellog.Int64(k, int64(x)), true
	case int64:
		return otellog.Int64(k, x), true
	case json.Number:
		if i, err := x.Int64(); err == nil {
			return otellog.Int64(k, i), true
		}
		if f, err := x.Float64(); err == nil {
			return otellog.Float64(k, f), true
		}
	case map[string]interface{}, []interface{}:
		b, err := json.Marshal(x)
		if err != nil {
			return otellog.KeyValue{}, false
		}
		return otellog.String(k, string(b)), true
	default:
		s := fmt.Sprint(x)
		if s == "" {
			return otellog.KeyValue{}, false
		}
		return otellog.String(k, s), true
	}
	return otellog.KeyValue{}, false
}

func contextFromTraceFields(m map[string]interface{}) context.Context {
	ctx := context.Background()
	tidStr, _ := m["trace_id"].(string)
	sidStr, _ := m["span_id"].(string)
	if len(tidStr) != 32 || len(sidStr) != 16 {
		return ctx
	}
	tid, err := trace.TraceIDFromHex(tidStr)
	if err != nil {
		return ctx
	}
	sid, err := trace.SpanIDFromHex(sidStr)
	if err != nil {
		return ctx
	}
	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    tid,
		SpanID:     sid,
		TraceFlags: trace.FlagsSampled,
	})
	return trace.ContextWithSpanContext(ctx, sc)
}
