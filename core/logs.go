package core

import (
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

// Create a new instance of the Logger. You can have any number of instances.
var (
	Logger = logrus.New()
)

// InitLogger Initialize Logger
func InitLogger() {
	// The API for setting attributes is a little different than the package level
	// exported Logger. See Godoc.
	Logger.Out = os.Stdout

	if os.Getenv("GO_ENV") == "production" {
		// You could set this to any `io.Writer` such as a file
		t := time.Now()
		name := "data_" + t.Format("2006-01-02")
		file, err := os.OpenFile("/go/logs/"+name+".log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err == nil {
			Logger.Out = file
		} else {
			Logger.Info("Failed to log to file, using default stderr")
		}
	}
}
