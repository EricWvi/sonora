package log

import (
	"context"
	"fmt"
	"log/slog"
	"os"
)

const RequestIDCtxKey = "RequestId"
const RequestIDKey = "requestId"

var (
	WorkerLogId = "9c89f949-ea0a-4e2b-b4f2-518dacd64ba8"
	WorkerCtx   = context.WithValue(context.Background(), RequestIDCtxKey, WorkerLogId)

	mediaLogId = "d002789b-07b0-4ee1-ae56-66596c956562"
	MediaCtx   = context.WithValue(context.Background(), RequestIDCtxKey, mediaLogId)

	idemLogId = "55483de4-93c3-4ef3-bb24-99a3ce7b6e56"
	IdemCtx   = context.WithValue(context.Background(), RequestIDCtxKey, idemLogId)
)

var slogger *slog.Logger

func InitLogger(level slog.Level) {
	slogger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))
}

func Fatal(c context.Context, msg string, args ...any) {
	slogger.With(RequestIDKey, c.Value(RequestIDCtxKey)).Error(msg, args...)
	os.Exit(1)
}

func Fatalf(c context.Context, format string, args ...any) {
	slogger.With(RequestIDKey, c.Value(RequestIDCtxKey)).Error(fmt.Sprintf(format, args...))
	os.Exit(1)
}

func Error(c context.Context, msg string, args ...any) {
	slogger.With(RequestIDKey, c.Value(RequestIDCtxKey)).Error(msg, args...)
}

func Errorf(c context.Context, format string, args ...any) {
	slogger.With(RequestIDKey, c.Value(RequestIDCtxKey)).Error(fmt.Sprintf(format, args...))
}

func Warn(c context.Context, msg string, args ...any) {
	slogger.With(RequestIDKey, c.Value(RequestIDCtxKey)).Warn(msg, args...)
}

func Warnf(c context.Context, format string, args ...any) {
	slogger.With(RequestIDKey, c.Value(RequestIDCtxKey)).Warn(fmt.Sprintf(format, args...))
}

func Info(c context.Context, msg string, args ...any) {
	slogger.With(RequestIDKey, c.Value(RequestIDCtxKey)).Info(msg, args...)
}

func Infof(c context.Context, format string, args ...any) {
	slogger.With(RequestIDKey, c.Value(RequestIDCtxKey)).Info(fmt.Sprintf(format, args...))
}

func Debug(c context.Context, msg string, args ...any) {
	slogger.With(RequestIDKey, c.Value(RequestIDCtxKey)).Debug(msg, args...)
}

func Debugf(c context.Context, format string, args ...any) {
	slogger.With(RequestIDKey, c.Value(RequestIDCtxKey)).Debug(fmt.Sprintf(format, args...))
}
