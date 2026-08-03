package logger

import (
	"errors"
	"genshin-quiz/internal/enum"
	"syscall"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

//nolint:gochecknoglobals // 全局 logger 详见 Init()/L 的使用说明
var L *zap.Logger = zap.NewNop()

func Init(env string) error {
	var config zap.Config

	if env == "production" {
		config = zap.NewProductionConfig()
		config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
		// 生产环境不使用彩色输出
		config.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	} else {
		config = zap.NewDevelopmentConfig()
		config.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
		// 开发环境使用彩色输出
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// 设置更友好的时间格式和颜色编码
	if env != string(enum.PROD) {
		config.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("15:04:05.000")
		// 使用短路径显示 caller 信息
		config.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	}

	logger, err := config.Build(zap.AddStacktrace(zapcore.ErrorLevel))
	if err != nil {
		return err
	}

	L = logger
	return nil
}

// Sync 在程序退出前调用，确保 buffer 里的日志被刷新.
func Sync() {
	err := L.Sync()
	if err != nil &&
		!errors.Is(err, syscall.EINVAL) &&
		!errors.Is(err, syscall.EBADF) &&
		!errors.Is(err, syscall.ENOTTY) {
		// 这里选择打印而不是 panic，因为程序即将退出
		L.Error("failed to sync logger", zap.Error(err))
	}
}
