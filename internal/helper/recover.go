package helper

import (
	"github.com/sirupsen/logrus"
)

type Result struct {
	Item    interface{}
	Success bool
	Index   int64
}

func HandleOutOfNilResponse(logger *logrus.Entry) {
	if r := recover(); r != nil {
		logger.Debugln("Recovering from panic:", r)
	}
}
