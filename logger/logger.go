package logger

import (
	"flag"
	"log"
	"os"
	"time"
)

var (
	Log *log.Logger
)

func init() {
	// set location of log file
	t := time.Now()
	var logpath = "./logs/" + t.Format("2006-01-02") + "logs.log"

	flag.Parse()
	f, err := os.OpenFile(logpath,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	defer f.Close()
	log.New(f, "", log.LstdFlags|log.Lshortfile)
}
