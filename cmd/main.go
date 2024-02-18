package main

import (
	"adgh-querylog-preprocessor/ext"
	"bufio"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	RECOVERY_FROM_PATH = "/etc/adgh-log-preproc/.recover.prog"
	DEST_DEF           = "DEST_TCP"
	SRC_DEF            = "SRC_LOG"
)

type PreProcessorConfig struct {
	DestinationTcp string
	SourceFile     string
}

func main() {
	var err error
	var dstConn *net.TCPConn
	defer func() {
		log.Println("Sleep 5 seconds for cleanup, then exit.")
		time.Sleep(5 * time.Second)
	}()
	// init
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.LUTC | log.Lmicroseconds)
	log.Println("Start Initialization.")
	conf := &PreProcessorConfig{
		DestinationTcp: os.Getenv(DEST_DEF),
		SourceFile:     os.Getenv(SRC_DEF),
	}
	if conf.DestinationTcp == "" || conf.SourceFile == "" {
		panic(ext.ErrConfigInvalid)
	}
	log.Println("Config read.")
	// validate conf by check file existence and try tcp conn
	if _, err = os.Stat(conf.SourceFile); err != nil {
		panic(err)
	}
	log.Println("Source file exists.")
	dialDest, err := net.ResolveTCPAddr("tcp", conf.DestinationTcp)
	if err != nil {
		panic(err)
	}
	dstConn, err = net.DialTCP("tcp", nil, dialDest)
	if err != nil {
		panic(err)
	}
	err = dstConn.SetKeepAlive(true)
	if err != nil {
		panic(err)
	}
	_ = dstConn.Close()
	log.Println("TCP Connection can be established.")
	log.Println("Prestart check complete.")
	// others
	var lastTS time.Time
	var lastTSChan = make(chan time.Time, 1)
	go func() {
		log.Println("Start Last-Fetch Timestamp Saver.")
		// last timestamp saver
		for {
			i, isLast := <-lastTSChan
			lastTS = i
			if isLast {
				log.Println("LastTSChan closed.")
				savedTS, _ := lastTS.MarshalBinary()
				_ = os.WriteFile(RECOVERY_FROM_PATH, savedTS, 0644)
				log.Println("Last Progress saved, progress: ", lastTS.String())
			}
		}
	}()
	// buffer 100 lines
	var logDataChan = make(chan string, 100)
	go func() {
		file, err := os.Open(conf.SourceFile)
		if err != nil {
			panic(err)
		}
		defer file.Close()
		lscanner := bufio.NewScanner(file)
		lscanner.Split(bufio.ScanLines)
		log.Println("Start File Reader.")
		for lscanner.Scan() {
			curLine := lscanner.Text()
			logDataChan <- curLine
		}
	}()
	go func() {
		// tcp conn and duplicator and processor and data sender
		//TODO: always retry connection, parse log, extend log, send timestamp to lastTS, send final log to splunk

	}()
	var sigChan = make(chan os.Signal, 1)
	defer close(sigChan)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	<-sigChan
	close(lastTSChan)
	close(logDataChan)
	log.Println("Process Terminated.")
}
