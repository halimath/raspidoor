package logging

import (
	"fmt"
	"io"
	"log/syslog"
	"os"
	"time"
)

type Logger interface {
	Info(string, ...interface{})
	Warn(string, ...interface{})
	Error(string, ...interface{})
	Err(error)
	Close() error
}

func Stdout() Logger {
	return &writerLogger{w: os.Stdout}
}

type writerLogger struct {
	w io.Writer
}

var _ Logger = &writerLogger{}

func (l *writerLogger) log(level string, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(l.w, "%s [%s] %s\n", time.Now().Format(time.RFC3339), level, msg)
}

func (l *writerLogger) Info(format string, args ...interface{})  { l.log("INFO", format, args...) }
func (l *writerLogger) Warn(format string, args ...interface{})  { l.log("WARN", format, args...) }
func (l *writerLogger) Error(format string, args ...interface{}) { l.log("ERROR", format, args...) }
func (l *writerLogger) Err(err error)                            { l.log("ERROR", "an error occured: %s", err) }

func (l *writerLogger) Close() error {
	if c, ok := l.w.(io.Closer); ok {
		return c.Close()
	}

	return nil
}

func Syslog() (Logger, error) {
	w, err := syslog.New(syslog.LOG_INFO, "raspidoor")
	if err != nil {
		return nil, err
	}

	return &syslogLogger{w}, err
}

type syslogLogger struct {
	w *syslog.Writer
}

var _ Logger = &syslogLogger{}

func (l *syslogLogger) Info(format string, args ...interface{}) {
	l.w.Info(fmt.Sprintf(format, args...))
}
func (l *syslogLogger) Warn(format string, args ...interface{}) {
	l.w.Warning(fmt.Sprintf(format, args...))
}
func (l *syslogLogger) Error(format string, args ...interface{}) {
	l.w.Err(fmt.Sprintf(format, args...))
}
func (l *syslogLogger) Err(err error) {
	l.w.Err(fmt.Sprintf("an error occured: %s", err))
}

func (l *syslogLogger) Close() error {
	return l.w.Close()
}
