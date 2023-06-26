package logger

import "github.com/wailsapp/wails/v2/pkg/logger"

var log logger.Logger

func InitLogger() {
	dl := logger.NewDefaultLogger()
	log = dl
}

func Debug(info string) {
	log.Debug(info)
}

func Error(info string) {
	log.Error(info)
}
