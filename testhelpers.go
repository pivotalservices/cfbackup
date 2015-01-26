package cfbackup

import (
	"github.com/pivotal-golang/lager"
	"github.com/pivotal-golang/lager/lagertest"
)

var logger = getLogger()

func Logger() lager.Logger {
	return logger
}

func getLogger() lager.Logger {
	return lagertest.NewTestLogger("TestLogger")
}
