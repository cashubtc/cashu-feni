package cashuLog

import (
	"encoding/json"
	"github.com/fatih/structs"
	log "github.com/sirupsen/logrus"
	"go.elastic.co/ecslogrus"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
)

func InitializeLogger(logLevel string) {
	level, err := log.ParseLevel(logLevel)
	if err != nil {
		level = log.TraceLevel
	}
	rotateFileHook, err := NewRotateFileHook(RotateFileConfig{
		Filename:   "log/out.log",
		MaxSize:    1, // megabytes
		MaxBackups: 3,
		MaxAge:     28,
		Compress:   true,
		Level:      level,
		Formatter:  &ecslogrus.Formatter{},
	})
	log.StandardLogger().ReportCaller = true
	if err != nil {
		panic(err)
	}
	log.AddHook(rotateFileHook)
}

// RotateFileHook is file rotation hook for logrus
type RotateFileHook struct {
	Config    RotateFileConfig
	logWriter io.Writer
}

// RotateFileConfig configuration for logfile rotation.
type RotateFileConfig struct {
	Filename   string
	MaxSize    int
	MaxBackups int
	MaxAge     int
	Compress   bool
	Level      log.Level
	Formatter  log.Formatter
}

func NewRotateFileHook(config RotateFileConfig) (log.Hook, error) {
	hook := RotateFileHook{
		Config: config,
	}
	hook.logWriter = &lumberjack.Logger{
		Filename:   config.Filename,
		MaxSize:    config.MaxSize,
		MaxBackups: config.MaxBackups,
		MaxAge:     config.MaxAge,
		Compress:   config.Compress,
	}

	return &hook, nil
}

type Loggable interface {
	Log() map[string]interface{}
}

func WithLoggable(l Loggable, more ...interface{}) log.Fields {
	var logMap log.Fields
	if l != nil {
		logMap = l.Log()
	}
	f := log.Fields{}
	for _, m := range more {
		switch m.(type) {
		case error:
			f["error.message"] = m.(error)
		case Loggable:
			for s, i := range m.(Loggable).Log() {
				f[s] = i
			}
		}
	}
	for s, i := range logMap {
		f[s] = i
	}
	return f
}

func (hook *RotateFileHook) Levels() []log.Level {
	return log.AllLevels[:hook.Config.Level+1]
}

func (hook *RotateFileHook) Fire(entry *log.Entry) (err error) {
	b, err := hook.Config.Formatter.Format(entry)
	if err != nil {
		return err
	}
	_, err = hook.logWriter.Write(b)
	if err != nil {
		return
	}
	return nil
}

func ToMap(i interface{}) map[string]interface{} {
	return structs.Map(i)
}
func ToJson(i interface{}) string {
	b, err := json.Marshal(i)
	if err != nil {
		return err.Error()
	}
	return string(b)
}
