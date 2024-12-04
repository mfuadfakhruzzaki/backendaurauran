// utils/logger.go
package utils

import (
	"os"

	"github.com/mfuadfakhruzzaki/backendaurauran/config"
	"github.com/sirupsen/logrus"
)

var Logger *logrus.Logger

// InitLogger menginisialisasi logger berdasarkan konfigurasi
func InitLogger() {
    Logger = logrus.New()

    // Set log level
    switch config.AppConfig.Logger.Level {
    case "debug":
        Logger.SetLevel(logrus.DebugLevel)
    case "info":
        Logger.SetLevel(logrus.InfoLevel)
    case "warn":
        Logger.SetLevel(logrus.WarnLevel)
    case "error":
        Logger.SetLevel(logrus.ErrorLevel)
    default:
        Logger.SetLevel(logrus.InfoLevel)
    }

    // Set output ke stdout
    Logger.SetOutput(os.Stdout)

    // Set format log
    Logger.SetFormatter(&logrus.JSONFormatter{})
}
