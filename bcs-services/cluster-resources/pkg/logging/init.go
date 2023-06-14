/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package logging

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/ctxkey"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/config"
)

var loggerInitOnce sync.Once

// 如果要进一步性能，可以使用SugaredLogger
var logger *zap.Logger

var levelMap = map[string]zapcore.Level{
	"debug": zapcore.DebugLevel,
	"info":  zapcore.InfoLevel,
	"warn":  zapcore.WarnLevel,
	"error": zapcore.ErrorLevel,
	"panic": zapcore.PanicLevel,
	"fatal": zapcore.FatalLevel,
}

// InitLogger xxx
func InitLogger(logConf *config.LogConf) {
	loggerInitOnce.Do(func() {
		// 使用 zap 记录日志，格式为 json
		logger = newZapJSONLogger(logConf)
	})
}

// getEncoder 修改时间并设置日志级别为大写，例如 日志级别: DEBUG/INFO, 时间格式: 2022-01-04 10:33:08
func getEncoder() zapcore.Encoder {
	return zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		MessageKey:    "msg",
		LevelKey:      "level",
		EncodeLevel:   zapcore.CapitalLevelEncoder,
		TimeKey:       "ts",
		StacktraceKey: "stacktrace",
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format("2006-01-02 15:04:05"))
		},
		CallerKey:    "caller",
		EncodeCaller: zapcore.ShortCallerEncoder,
		EncodeDuration: func(d time.Duration, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendInt64(int64(d) / 1000000)
		},
	})
}

func newZapJSONLogger(conf *config.LogConf) *zap.Logger {
	writer, err := getWriter(conf)
	if err != nil {
		panic(err)
	}
	w := &zapcore.BufferedWriteSyncer{
		WS:            zapcore.AddSync(writer),
		FlushInterval: time.Duration(conf.FlushInterval) * time.Second,
	}

	// 设置日志级别
	l, ok := levelMap[conf.Level]
	if !ok {
		l = zap.InfoLevel
	}

	core := zapcore.NewCore(getEncoder(), w, l)
	// AddCallerSkip 由于对 logger 进行封装，设置 caller 记录位置往上一层
	// AddStacktrace 设置 Error 及以上级别允许打印堆栈信息
	return zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1), zap.AddStacktrace(zap.ErrorLevel))
}

// GetLogger xxx
// : 是否分为不同的类型，比如请求第三方、API等，根据不同的配置，设置不同的日志
func GetLogger() *zap.Logger {
	// 未执行日志组件初始化时，日志输出到 stderr
	if logger == nil {
		stderrLogger, _ := zap.NewProductionConfig().Build(zap.AddCallerSkip(1))
		return stderrLogger
	}
	return logger
}

// Info 同 Warn，Error 等为封装在 logging 模块下的快捷方法，
// 使用默认 logger，避免使用时手动 GetLogger，可按需添加 Panic 等
// 参考用法：
// import (
//
//	log ".../pkg/logging"
//
// )
//
//	func main() {
//			log.Info(ctx, "log content: %s", content)
//	}
func Info(ctx context.Context, msg string, vars ...interface{}) {
	GetLogger().Info(wrapLogMsg(ctx, fmt.Sprintf(msg, vars...)))
}

// Warn ....
func Warn(ctx context.Context, msg string, vars ...interface{}) {
	GetLogger().Warn(wrapLogMsg(ctx, fmt.Sprintf(msg, vars...)))
}

// Error xxx
func Error(ctx context.Context, msg string, vars ...interface{}) {
	GetLogger().Error(wrapLogMsg(ctx, fmt.Sprintf(msg, vars...)))
}

// wrapLogMsg 向日志中补充 requestID，username 信息
func wrapLogMsg(ctx context.Context, msg string) string {
	requestID, username := ctx.Value(ctxkey.RequestIDKey), ctx.Value(ctxkey.UsernameKey)
	return fmt.Sprintf("requestID: %v, username: %v, ", requestID, username) + msg
}
