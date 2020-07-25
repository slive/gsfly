/*
 * Author:slive
 * DATE:2020/7/24
 */
package logger

import (
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
	log "github.com/sirupsen/logrus"
	"gsfly/config"
	"time"

	//"log"
	"os"
	"path"
)

const (
	LOG_DEBUG = "DEBUG"
	LOG_INFO  = "INFO"
	LOG_WARN  = "WARN"
	LOG_ERROR = "ERROR"
)

func init() {
	//log.SetFlags(log.Ldate | log.Lmicroseconds | log.Lshortfile)
	logdir := config.Global_Conf.LogConf.LogDir
	//创建日志文件夹
	_ = mkdirLog(logdir)

	filePath := path.Join(logdir, config.Global_Conf.LogConf.LogFile)
	log.Println("filePath:", filePath)
	var logfile *os.File = nil
	if checkFileExist(filePath) {
		//文件存在，打开
		logfile, _ = os.OpenFile(filePath, os.O_APPEND|os.O_RDWR, 0644)
	} else {
		//文件不存在，创建
		logfile, _ = os.OpenFile(filePath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	}

	//log.SetOutput(os.Stdout)
	//log.SetOutput(logfile)
	log.SetFormatter(&log.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})
	log.AddHook(newRlfHook(filePath, 10))
	log.SetOutput(logfile)
	log.SetOutput(os.Stdout)
}

func newRlfHook(logName string, maxRemainCont uint) log.Hook {
	writer, err := rotatelogs.New(logName+"%Y%m%d%H",
		rotatelogs.WithLinkName(logName),
		rotatelogs.WithRotationTime(time.Minute),
		rotatelogs.WithRotationCount(maxRemainCont))
	if err != nil {
		log.Errorf("config local file system for logger error: %v", err)
	}

	log.SetLevel(log.DebugLevel)

	lfsHook := lfshook.NewHook(lfshook.WriterMap{
		log.DebugLevel: writer,
		log.InfoLevel:  writer,
		log.WarnLevel:  writer,
		log.ErrorLevel: writer,
		log.FatalLevel: writer,
		log.PanicLevel: writer,
	}, &log.TextFormatter{DisableColors: true})

	return lfsHook
}

func checkFileExist(filename string) bool {
	exist := true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		exist = false
	}
	return exist
}

func mkdirLog(dir string) (e error) {
	defer log.Println("error:", e)
	_, er := os.Stat(dir)
	b := er == nil || os.IsExist(er)
	if !b {
		if err := os.MkdirAll(dir, 0775); err != nil {
			if os.IsPermission(err) {
				e = err
			}
		}
	}
	return
}

func logwrite(level string, logs ...interface{}) {
	log.Println("["+level+"]", logs)
}

func Debug(logs ...interface{}) {
	logwrite(LOG_INFO, logs)
}

func Info(logs ...interface{}) {
	logwrite(LOG_WARN, logs)
}

func Warn(logs ...interface{}) {
	logwrite(LOG_ERROR, logs)
}

func Error(logs ...interface{}) {
	logwrite(LOG_DEBUG, logs)
}
