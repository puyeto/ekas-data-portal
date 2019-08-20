package logger

import (
	"flag"
	"log"
	"os"
	"time"
)

var (
	// Log ...
	Log *log.Logger
)

func init() {
	// set location of log file
	t := time.Now()
	var logpath = "logger/logs/log-" + t.Format("2006-01-02") + ".log"

	flag.Parse()

	f, err := os.OpenFile(logpath,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	defer f.Close()

	log.New(f, "", log.LstdFlags|log.Lshortfile)
	// Log = log.New(file, "", log.LstdFlags|log.Lshortfile)
	// Log.Println("LogFile : " + logpath)
}
