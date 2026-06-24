// Package logger định nghĩa interface ILogger để usecase ghi log mà không
// phụ thuộc vào thư viện log cụ thể. Bản triển khai mặc định dùng stdlib log.
package logger

import (
	stdlog "log"
	"os"
)

type ILogger interface {
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Debugf(format string, args ...interface{})
}

type stdLogger struct {
	l *stdlog.Logger
}

// New trả về ILogger ghi ra stderr với tiền tố mức log.
func New() ILogger {
	return &stdLogger{l: stdlog.New(os.Stderr, "", stdlog.LstdFlags)}
}

func (s *stdLogger) Infof(format string, args ...interface{}) { s.l.Printf("[INFO] "+format, args...) }
func (s *stdLogger) Warnf(format string, args ...interface{}) { s.l.Printf("[WARN] "+format, args...) }
func (s *stdLogger) Errorf(format string, args ...interface{}) {
	s.l.Printf("[ERROR] "+format, args...)
}
func (s *stdLogger) Debugf(format string, args ...interface{}) {
	s.l.Printf("[DEBUG] "+format, args...)
}
