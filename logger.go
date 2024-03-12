package esql

import (
	"log"
)

type Logger interface {
	Debugf(format string, v ...interface{})
	Infof(format string, v ...interface{})
	Errorf(format string, v ...interface{})
	Output(query string, err error, v ...interface{})
}

type logger struct {
	level    int
	debugLog *log.Logger
	errorLog *log.Logger
	infoLog  *log.Logger
}

const (
	Disabled = iota
	DebugLevel
	ErrorLevel
	InfoLevel
)

func newDefaultLogger() Logger {
	return &logger{
		level: ErrorLevel,
	}
}

func (l *logger) Debugf(format string, v ...interface{}) {
	log.Printf(format, v...)
}

func (l *logger) Infof(format string, v ...interface{}) {
	log.Printf(format, v...)
}

func (l *logger) Errorf(format string, v ...interface{}) {
	log.Printf(format, v...)
}

func (l *logger) Output(query string, err error, v ...interface{}) {

	switch {
	case l.level == DebugLevel:
		l.Debugf("sql: %s value:%v \n", query, v)
		if err != nil {
			l.Debugf("error:%+v \n", err)
		}
	case err != nil && l.level >= ErrorLevel:
		l.Errorf("sql: %s value:%v error:%+v \n", query, v, err)
	case l.level >= InfoLevel:
		l.Infof("sql %s value:%+v \n", query, v)
	}

}
