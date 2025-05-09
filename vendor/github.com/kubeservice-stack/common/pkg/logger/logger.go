/*
Copyright 2022 The KubeService-Stack Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package logger

import (
	"fmt"
	"os"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/mattn/go-isatty"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	HTTPModule  = "http"
	CrashModule = "crash"
)

var (
	AccessLog = GetLogger(HTTPModule, "Access")
	CrashLog  = GetLogger(CrashModule, "Crash")
)

func SimpleTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
}

func SimpleLevelEncoder(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(LevelString(l))
}

func SimpleAccessLevelEncoder(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	if isTerminal {
		enc.AppendString(LevelString(l))
	}
}

func LevelString(l zapcore.Level) string {
	if !isTerminal {
		return l.CapitalString()
	}
	return l.CapitalString()
}

// linux 系统 stdout输出
func IsTerminal(f *os.File) bool {
	if runtime.GOOS == "windows" {
		return false
	}
	fd := f.Fd()
	return isatty.IsTerminal(fd) || isatty.IsCygwinTerminal(fd)
}

type Logger struct {
	module string
	role   string
	logger *zap.Logger
}

func (l *Logger) getInitializedOrDefaultLogger() *zap.Logger {
	if l.logger != nil {
		return l.logger
	}
	var item interface{}
	switch l.module {
	case HTTPModule:
		item = accessLogger.Load()
	case CrashModule:
		item = crashLogger.Load()
	default:
		item = mediaLogger.Load()
	}
	if item == nil {
		return defaultLogger
	}
	l.logger = item.(*zap.Logger)
	return l.logger
}

// TODO: append 浅拷贝问题
func (l *Logger) Debug(msg string, fields ...zap.Field) {
	l.getInitializedOrDefaultLogger().Debug(l.formatMsg(), append(fields, String("msg", msg))...)
}

func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.getInitializedOrDefaultLogger().Info(l.formatMsg(), append(fields, String("msg", msg))...)
}

func (l *Logger) Warn(msg string, fields ...zap.Field) {
	l.getInitializedOrDefaultLogger().Warn(l.formatMsg(), append(fields, String("msg", msg))...)
}

func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.getInitializedOrDefaultLogger().Error(l.formatMsg(), append(fields, String("msg", msg))...)
}

func (l *Logger) formatMsg() string {
	//if !isTerminal && l.module == HTTPModule {
	//	return msg
	//}
	moduleName := fmt.Sprintf("[%*s]", atomic.LoadUint32(&maxModuleNameLen), l.module)
	if l.role == "" {
		return fmt.Sprintf("%s: ",
			moduleName)
	}
	return fmt.Sprintf("%s [%s]: ",
		moduleName, l.role)
}

func String(key string, val string) zap.Field {
	return zap.Field{Key: key, Type: zapcore.StringType, String: val}
}

func Error(err error) zap.Field {
	return zap.NamedError("error", err)
}

func Uint16(key string, val uint16) zap.Field {
	return zap.Field{Key: key, Type: zapcore.Uint16Type, Integer: int64(val)}
}

func Uint32(key string, val uint32) zap.Field {
	return zap.Field{Key: key, Type: zapcore.Uint32Type, Integer: int64(val)}
}

func Stack() zap.Field {
	return zap.Stack("stack")
}

func Reflect(key string, val interface{}) zap.Field {
	return zap.Reflect(key, val)
}

func Any(key string, value interface{}) zap.Field {
	return zap.Any(key, value)
}

func Int32(key string, val int32) zap.Field {
	return zap.Field{Key: key, Type: zapcore.Int32Type, Integer: int64(val)}
}

func Int64(key string, val int64) zap.Field {
	return zap.Field{Key: key, Type: zapcore.Int64Type, Integer: val}
}
