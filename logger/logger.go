package logger

import (
	"bytes"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"

	nested "github.com/antonfisher/nested-logrus-formatter"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
)

var defaultLogger logrusImpl
var defaultLoggerOnce sync.Once

// Data is
type Data struct {
	ClientIP string `` // c.ClientIP() set from controller before entering service
	Session  string `` // ksuid.New().String() or gonanoid.Generate(x, 12) set from first caller
	UserID   string `` // from client
	Type     string `` // MOB/BOF/MSQ/SYS/SCH --> mobile / backoffice / message queuing / system / scheduller
}

// ILogger is
type ILogger interface {
	Debug(data interface{}, description string, args ...interface{})
	Info(data interface{}, description string, args ...interface{})
	Warn(data interface{}, description string, args ...interface{})
	Error(data interface{}, description string, args ...interface{})
	Fatal(data interface{}, description string, args ...interface{})
	Panic(data interface{}, description string, args ...interface{})
	WithFile(appsName, filename string, maxAge int)
}

// LogrusImpl is
type logrusImpl struct {
	theLogger *logrus.Logger
	useFile   bool
}

// GetLog is
func GetLog() ILogger {
	defaultLoggerOnce.Do(func() {
		defaultLogger = logrusImpl{theLogger: logrus.New()}
		defaultLogger.useFile = false
		defaultLogger.theLogger.SetFormatter(&nested.Formatter{
			NoColors:        false,
			HideKeys:        true,
			TimestampFormat: "0102 150405.000",
			FieldsOrder:     []string{"func"},
		})
	})
	return &defaultLogger
}

// WithFile is command to state the log will printing to files
// the rolling log file will put in logs/ directory
//
// filename is just a name of log file without any extension
//
// maxAge is age (in days) of the logs file before it gets purged from the file system
func (l *logrusImpl) WithFile(appsName, filename string, maxAge int) {

	if !l.useFile {

		if maxAge <= 0 {
			panic("maxAge should > 0")
		}

		path := filename + ".log"
		writer, _ := rotatelogs.New(
			"./logs/"+path+".%Y%m%d",
			rotatelogs.WithLinkName(path),
			rotatelogs.WithMaxAge(time.Duration(maxAge*24)*time.Hour),
			rotatelogs.WithRotationTime(time.Duration(1*24)*time.Hour),
		)

		defaultLogger.theLogger.AddHook(lfshook.NewHook(
			lfshook.WriterMap{
				logrus.InfoLevel:  writer,
				logrus.WarnLevel:  writer,
				logrus.ErrorLevel: writer,
				logrus.DebugLevel: writer,
			},
			defaultLogger.theLogger.Formatter,
		))

		l.useFile = true
	}
}

func (l *logrusImpl) getLogEntry(data map[string]string) *logrus.Entry {
	pc, _, _, _ := runtime.Caller(2)
	funcName := runtime.FuncForPC(pc).Name()

	var buffer bytes.Buffer

	buffer.WriteString("|FN:")

	x := strings.LastIndex(funcName, "/")

	buffer.WriteString(funcName[x+1:])

	if data == nil {
		return l.theLogger.WithField("info", buffer.String())
	}

	if clientIP, ok := data["clientIP"]; ok {
		buffer.WriteString("|IP:")
		buffer.WriteString(clientIP)
	}

	if session, ok := data["session"]; ok {
		buffer.WriteString("|SS:")
		buffer.WriteString(session)
	}

	if userID, ok := data["userId"]; ok {
		buffer.WriteString("|US:")
		buffer.WriteString(userID)
	}

	if types, ok := data["types"]; ok {
		buffer.WriteString("|TY:")
		buffer.WriteString(types)
	}

	return l.theLogger.WithField("info", buffer.String())
}

// Debug is
func (l *logrusImpl) Debug(data interface{}, description string, args ...interface{}) {
	l.getLogEntry(nil).Debugf(description, args...)
}

// Info is
func (l *logrusImpl) Info(data interface{}, description string, args ...interface{}) {
	l.getLogEntry(nil).Infof(description, args...)
}

// Warn is
func (l *logrusImpl) Warn(data interface{}, description string, args ...interface{}) {
	l.getLogEntry(nil).Warnf(description, args...)
}

// Error is
func (l *logrusImpl) Error(data interface{}, description string, args ...interface{}) {
	l.getLogEntry(nil).Errorf(description, args...)
}

// Fatal is
func (l *logrusImpl) Fatal(data interface{}, description string, args ...interface{}) {
	l.getLogEntry(nil).Fatalf(description, args...)
}

// Panic is
func (l *logrusImpl) Panic(data interface{}, description string, args ...interface{}) {
	l.getLogEntry(nil).Panicf(description, args...)
}
