/*
 * Author:slive
 * DATE:2020/7/24
 */
package logger

import (
	"fmt"
	logx "github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/writer"
	"gsfly/config"
	"io/ioutil"
	"log"
	"os"
	"path"
	"runtime"
	"strings"
)

const (
	LOG_DEBUG = logx.DebugLevel
	LOG_INFO  = logx.InfoLevel
	LOG_WARN  = logx.WarnLevel
	LOG_ERROR = logx.ErrorLevel
)

var logLevel logx.Level

func init() {
	logdir := config.Global_Conf.LogConf.LogDir
	// 创建日志文件夹
	_ = mkdirLog(logdir)

	filePath := path.Join(logdir, config.Global_Conf.LogConf.LogFile)
	log.Println("filePath:", filePath)
	var logfile *os.File = nil
	if checkFileExist(filePath) {
		// 文件存在，打开
		logfile, _ = os.OpenFile(filePath, os.O_APPEND|os.O_RDWR, 0644)
	} else {
		// 文件不存在，创建
		logfile, _ = os.OpenFile(filePath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	}
	logLevel = logx.DebugLevel
	logx.SetLevel(logLevel)

	logx.SetFormatter(&logx.TextFormatter{
		DisableTimestamp: false,
		FullTimestamp:    true,
		TimestampFormat:  "2006-01-02 15:04:05.999999999",
		DisableSorting:   false,
	})

	logx.SetOutput(ioutil.Discard) // Send all logs to nowhere by default

	logx.AddHook(&writer.Hook{ // Send logs with level higher than warning to stderr
		Writer: os.Stderr,
		LogLevels: []logx.Level{
			logx.PanicLevel,
			logx.FatalLevel,
			logx.ErrorLevel,
			logx.WarnLevel,
		},
	})

	logx.AddHook(&writer.Hook{ // Send logs with level higher than warning to stderr
		Writer: logfile,
		LogLevels: []logx.Level{
			logx.PanicLevel,
			logx.FatalLevel,
			logx.ErrorLevel,
			logx.WarnLevel,
			logx.InfoLevel,
			logx.DebugLevel,
		},
	})

	logx.AddHook(&writer.Hook{ // Send info and debug logs to stdout
		Writer: os.Stdout,
		LogLevels: []logx.Level{
			logx.InfoLevel,
			logx.DebugLevel,
		},
	})
}

func checkFileExist(filename string) bool {
	exist := true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		exist = false
	}
	return exist
}

func mkdirLog(dir string) (e error) {
	_, er := os.Stat(dir)
	defer func() {
		err := recover()
		if err != nil {
			log.Println("mkdirLog error:", err)
		}
	}()
	b := er == nil || os.IsExist(er)
	if !b {
		if err := os.MkdirAll(dir, 0775); err != nil {
			if os.IsPermission(err) {
				er = err
			}
		}
	}
	return er
}

func logwrite(level logx.Level, logs []interface{}) {
	if logs != nil {
		finalLog := convertFinalLog(logs)
		// finalOutLog := time.Now().Format("2006-01-02 15:04:05.999999999") + " " + finalLog
		switch level {
		case LOG_DEBUG:
			logx.Debug(finalLog)
			// loggOut.Debug(finalOutLog)
			break
		case LOG_INFO:
			logx.Info(finalLog)
			// loggOut.Info(finalOutLog)
			break
		case LOG_WARN:
			logx.Warn(finalLog)
			// loggOut.Warn(finalOutLog)
			break
		case LOG_ERROR:
			logx.Debug(finalLog)
			// loggOut.Debug(finalOutLog)
			break
		default:
			logx.Println(finalLog)
			// loggOut.Println(finalOutLog)
		}
	}
}

func convertFinalLog(logs []interface{}) string {
	lfmt := ""
	_, file, line, ok := runtime.Caller(3)
	if ok {
		sps := strings.Split(file, "/")
		if sps != nil {
			file = sps[len(sps)-1]
		}
		lfmt += file + fmt.Sprintf("(%v) ", line)
	}

	var index int = 1
	logsLen := len(logs)
	for _, val := range logs {
		lfmt += fmt.Sprint(val)
		index += 1
		if index < logsLen {
			lfmt += ","
		}
	}

	return lfmt
}

func Println(logs ...interface{}) {
	logwrite(logLevel, logs)
}

func Debug(logs ...interface{}) {
	logwrite(LOG_DEBUG, logs)
}

func Info(logs ...interface{}) {
	logwrite(LOG_INFO, logs)
}

func Warn(logs ...interface{}) {
	logwrite(LOG_WARN, logs)
}

func Error(logs ...interface{}) {
	logwrite(LOG_ERROR, logs)
}
