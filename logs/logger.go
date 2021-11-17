package logs

import (
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	"daobackend/config"
)

var logger *logrus.Logger

func initLogger() {
	conf := config.GetConfig()

	logger = logrus.New()
	if conf.Dev {
		logger.SetLevel(logrus.DebugLevel)
	} else {
		logger.SetLevel(logrus.InfoLevel)
	}

	formatter := &logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05.000",
		FullTimestamp:   true,
	}
	logger.SetReportCaller(true)
	logger.SetFormatter(formatter)
	pathMap := lfshook.PathMap{
		logrus.InfoLevel:  "./logs/info.log",
		logrus.WarnLevel:  "./logs/warn.log",
		logrus.ErrorLevel: "./logs/error.log",
		logrus.FatalLevel: "./logs/error.log",
		logrus.PanicLevel: "./logs/error.log",
	}
	logger.Hooks.Add(lfshook.NewHook(
		pathMap,
		formatter,
	))
	logger.WriterLevel(logrus.InfoLevel)
}
func GetLogger() *logrus.Logger {
	if logger == nil {
		initLogger()
	}
	return logger
}
