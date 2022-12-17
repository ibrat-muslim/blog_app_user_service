package logger

import "github.com/sirupsen/logrus"

func New() *logrus.Logger {
	log := logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{PrettyPrint: true})
	log.SetLevel(logrus.InfoLevel)

	return log
}
