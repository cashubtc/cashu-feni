package main

import (
	"github.com/gohumble/cashu-feni/api"
	_ "github.com/gohumble/cashu-feni/docs"
	log "github.com/sirupsen/logrus"
	"go.elastic.co/ecslogrus"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
)

// @title Cashu (Feni) golang mint
// @version 0.0.1
// @description Ecash wallet and mint with Bitcoin Lightning support.
// @contact.url https://8333.space:3338
func main() {
	initializeLogger()
	log.Info("starting (feni) cashu mint server")
	m := api.New()
	m.StartServer()
}

func initializeLogger() {
	level, err := log.ParseLevel(api.Config.LogLevel)
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
