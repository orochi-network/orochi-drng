package logger

import (
	"encoding/hex"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var once sync.Once
var sugar *zap.SugaredLogger

func init() {
	once.Do(func() {
		config := zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		logger, err := config.Build()
		if err == nil {
			sugar = logger.Sugar()
			sugar.Debug("Logger online")
			defer func() {
				logger.Sync()
				sugar.Sync()
			}()
			return
		}
		panic(err)
	})

}

//GetSugarLogger get singleton sugar logger
func GetSugarLogger() *zap.SugaredLogger {
	return sugar
}

//HexDump for debug purpose
func HexDump(title string, data []byte) {
	sugar := GetSugarLogger()
	content := hex.Dump(data)
	sugar.Debugf("%s\n%s", title, content[:len(content)-1])
}
