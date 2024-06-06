package util

import (
	"os"
	"sync"

	"go.uber.org/zap"
)

var L *zap.Logger
var S *zap.SugaredLogger
var setupLock sync.Mutex

func SetupLog() {
	setupLock.Lock()
	defer setupLock.Unlock()
	switch os.Getenv("QRYSTAL_LOGGING_CONFIG") {
	case "development":
		L, _ = zap.NewDevelopment()
	default:
		L, _ = zap.NewProduction()
	}
	S = L.Sugar()
	zap.ReplaceGlobals(L)
}
