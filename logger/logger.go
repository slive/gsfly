/*
 * Author:slive
 * DATE:2020/7/24
 */
package logger

import (
	"fmt"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
	logx "github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/writer"
	"gsfly/util"
	"io/ioutil"
	"log"
	"os"
	"path"
	"runtime"
	"strings"
	"time"
)

const (
	LOG_DEBUG = logx.DebugLevel
	LOG_INFO  = logx.InfoLevel
	LOG_WARN  = logx.WarnLevel
	LOG_ERROR = logx.ErrorLevel
)

type LogConf struct {
	// LogFile 日志文件
	LogFile string

	// LogDir 日志路径
	LogDir string

	Level int
}

func NewDefaultLogConf() *LogConf {
	logConf := &LogConf{
		LogFile: "log-gsfly.log",
		LogDir:  util.GetPwd() + "/log",
	}
	return logConf
}

var logLevel logx.Level

func init() {
	logConf := NewDefaultLogConf()
	logdir := logConf.LogDir
	// 创建日志文件夹
	_ = mkdirLog(logdir)

	filePath := path.Join(logdir, logConf.LogFile)
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
	logx.SetOutput(ioutil.Discard) // Send all logs to nowhere by default
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

	logx.AddHook(&writer.Hook{ // Send logs with level higher than warning to stderr
		Writer: os.Stderr,
		LogLevels: []logx.Level{
			logx.PanicLevel,
			logx.FatalLevel,
			logx.ErrorLevel,
			logx.WarnLevel,
		},
	})

	tf := logx.TextFormatter{
		DisableTimestamp: false,
		FullTimestamp:    true,
		TimestampFormat:  "2006-01-02 15:04:05.999999999",
		DisableSorting:   false,
	}
	logx.SetFormatter(&tf)
	logx.AddHook(newRlfHook(10, filePath, &tf))
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

func newRlfHook(maxRemainCont uint, logName string, tf *logx.TextFormatter) logx.Hook {
	writer, err := rotatelogs.New(logName+"%Y%m%d",
		rotatelogs.WithLinkName(logName),
		rotatelogs.WithRotationTime(time.Hour*24),
		rotatelogs.WithRotationCount(maxRemainCont))
	if err != nil {
		logx.Errorf("config local file system for logger error: %v", err)
	}

	lfsHook := lfshook.NewHook(lfshook.WriterMap{
		logx.DebugLevel: writer,
		logx.InfoLevel:  writer,
		logx.WarnLevel:  writer,
		logx.ErrorLevel: writer,
		logx.FatalLevel: writer,
		logx.PanicLevel: writer,
	}, tf)

	return lfsHook
}

func logwrite(level logx.Level, logs []interface{}) {
	if logs != nil {
		// finalLog := convertFinalLog(logs)
		finalLog := logs

		switch level {
		case LOG_DEBUG:
			logx.WithField("file", caller()).Debug(finalLog)
			break
		case LOG_INFO:
			logx.WithField("file", caller()).Info(finalLog)
			break
		case LOG_WARN:
			logx.WithField("file", caller()).Warn(finalLog)
			break
		case LOG_ERROR:
			logx.WithField("file", caller()).Debug(finalLog)
			break
		default:
			logx.WithField("file", caller()).Println(finalLog)
		}
	}
}

func convertFinalLog(logs []interface{}) string {
	lfmt := caller()

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

func caller() string {
	lfmt := ""
	_, file, line, ok := runtime.Caller(3)
	if ok {
		sps := strings.Split(file, "/")
		if sps != nil {
			file = sps[len(sps)-1]
		}
		lfmt += file + fmt.Sprintf("(%v) ", line)
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
