package log

import (
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// inspired by https://www.datadoghq.com/blog/go-logging/#implement-a-standard-logging-interface
type Event struct {
	id      int
	message string
}

type stdLogger struct {
	*logrus.Entry
}

func newLogger() *stdLogger {
	log := &stdLogger{logrus.WithFields(getStats())}
	log.Logger.SetFormatter(&logrus.JSONFormatter{})
	return log
}

func getStats() logrus.Fields {
	return logrus.Fields{
		"startTime": time.Now(),
	}
}

var logger = newLogger()

func SetOperatorId(operatorId uuid.UUID) {
	logger.Data["operator"] = operatorId
}

func Print(args ...interface{}) {
	logger.Print(args)
}

func Printf(format string, args ...interface{}) {
	logger.Printf(format, args...)
}

func Warn(args ...interface{}) {
	logger.Warn(args)
}

func Warnf(format string, args ...interface{}) {
	logger.Warnf(format, args...)
}

func Error(args ...interface{}) {
	logger.Error(args)
}

func Errorf(format string, args ...interface{}) {
	logger.Errorf(format, args...)
}

func Fatal(args ...interface{}) {
	logger.Fatal(args)
}

func Fatalf(format string, args ...interface{}) {
	logger.Fatalf(format, args...)
}

func Panic(args ...interface{}) {
	logger.Panic(args)
}

func Panicf(format string, args ...interface{}) {
	logger.Panicf(format, args...)
}
