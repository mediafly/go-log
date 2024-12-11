package log

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"reflect"

	"github.com/pkg/errors"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

var ErrorLogKey = "error"

type ContextHandler struct {
	slog.Handler
}

func (h ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	r.AddAttrs(h.addTraceFromContext(ctx)...)
	return h.Handler.Handle(ctx, r)
}

func (h ContextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return ContextHandler{h.Handler.WithAttrs(attrs)}
}

func (h ContextHandler) WithGroup(name string) slog.Handler {
	return ContextHandler{h.Handler.WithGroup(name)}
}

func (h ContextHandler) addTraceFromContext(ctx context.Context) (as []slog.Attr) {
	span, ok := tracer.SpanFromContext(ctx)
	if ok {
		traceID := span.Context().TraceID()
		ddgroup := slog.Group("dd", slog.Uint64("trace_id", traceID))
		as = append(as, ddgroup)
	}
	return
}

type stackTracer interface {
	StackTrace() errors.StackTrace
}

type errorField struct {
	Kind    string `json:"kind"`
	Stack   string `json:"stack"`
	Message string `json:"message"`
}

func ErrorField(err error) errorField {
	var stack string
	if serr, ok := err.(stackTracer); ok {
		st := serr.StackTrace()
		stack = fmt.Sprintf("%+v", st)
		if len(stack) > 0 && stack[0] == '\n' {
			stack = stack[1:]
		}
	}
	return errorField{
		Kind:    reflect.TypeOf(err).String(),
		Stack:   stack,
		Message: err.Error(),
	}
}

func replaceAttr(_ []string, a slog.Attr) slog.Attr {
	switch a.Value.Kind() {
	case slog.KindAny:
		switch v := a.Value.Any().(type) {
		case error:
			a.Value = fmtErr(v)
		default:
			if a.Key == slog.TimeKey {
				a.Key = "@timestamp"
				return a
			}
			if a.Key == slog.MessageKey {
				a.Key = "log"
				return a
			}

			return a
		}
	}

	return a
}

// fmtErr returns a slog.Value with keys `message` and `kind`. If the error
// implements interface { StackTrace() errors.StackTrace }, the `stack` is populated
func fmtErr(err error) slog.Value {
	errorField := ErrorField(err)

	var groupValues []slog.Attr

	groupValues = append(groupValues,
		slog.Any("kind", errorField.Kind),
	)

	groupValues = append(groupValues,
		slog.Any("message", errorField.Message),
	)

	if errorField.Stack != "" {
		groupValues = append(groupValues,
			slog.Any("stack", errorField.Stack),
		)
	}

	return slog.GroupValue(groupValues...)
}

func SetupLog(logLevel slog.Level) {

	handlerOptions := &slog.HandlerOptions{
		Level:       logLevel, // Set the desired log level,
		ReplaceAttr: replaceAttr,
	}

	jsonHandler := slog.NewJSONHandler(os.Stdout, handlerOptions)
	ctxHandler := ContextHandler{jsonHandler}

	ddvars := slog.Group("dd",
		slog.String("env", os.Getenv("DD_ENV")),
		slog.String("service", os.Getenv("DD_SERVICE")),
		slog.String("version", os.Getenv("DD_VERSION")),
		slog.String("source", "golang"),
	)

	logger := slog.New(ctxHandler).With(ddvars)
	slog.SetDefault(logger)
}
