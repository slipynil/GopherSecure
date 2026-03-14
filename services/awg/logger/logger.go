package logger

import (
	"log"
	"log/syslog"
	"os"
)

type logger struct {
	writter *syslog.Writer
}

func New() *logger {
	writter, err := syslog.New(syslog.LOG_ERR, "awg http service")
	if err != nil {
		log.Fatal(err)
	}

	return &logger{
		writter: writter,
	}
}

func (l *logger) Fatal(err error) {
	l.writter.Err("Fatal Error: " + err.Error())
	os.Exit(1)
}
