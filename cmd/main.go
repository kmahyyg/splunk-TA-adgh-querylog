package main

import (
	"adgh-querylog-preprocessor/ext"
	"bufio"
	"encoding/json"
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
	_ = dstConn.Close()
	log.Println("TCP Connection can be established.")
	log.Println("Prestart check complete.")
	// others
	var lastTS time.Time
	//TODO: read if progress exists
	var lastTSChan = make(chan time.Time, 1)
	go func() {
		log.Println("Start Last-Fetch Timestamp Saver.")
		lastSaveTime := time.Now()
		for {
			// last timestamp saver
			currentTime := time.Now()
			i, isLast := <-lastTSChan
			lastTS = i
			// periodically save
			if isLast || currentTime.Sub(lastSaveTime) > 300*time.Second {
				savedTS, _ := lastTS.MarshalBinary()
				err := os.WriteFile(RECOVERY_FROM_PATH, savedTS, 0644)
				log.Println("Last Progress save action triggered, progress: ", lastTS.String())
				if err != nil {
					log.Println("Save progress error: " + err.Error())
				}
				lastSaveTime = currentTime
			}
		}
	}()
	// buffer 100 lines
	var logDataChan = make(chan *ext.ADGHLogEntry, 100)
	go func() {
		//todo: maybe github.com/nxadm/tail
		file, err := os.Open(conf.SourceFile)
		if err != nil {
			panic(err)
		}
		defer file.Close()
		lscanner := bufio.NewScanner(file)
		lscanner.Split(bufio.ScanLines)
		log.Println("Start File Reader.")
		for lscanner.Scan() {
			curLine := lscanner.Bytes()
			buf := &ext.ADGHLogEntry{}
			err := json.Unmarshal(curLine, buf)
			if err != nil {
				log.Println("Unmarshal Error: ", err.Error())
			}
			logDataChan <- buf
			lastTSChan <- buf.Time
		}
	}()
	go func() {
		// tcp conn and duplicator and processor and data sender
		//TODO: always retry connection, for each object, parse log, remove answer, extend log, send final log to splunk

	}()
	var sigChan = make(chan os.Signal, 1)
	defer close(sigChan)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	<-sigChan
	close(lastTSChan)
	close(logDataChan)
	log.Println("Process Terminated.")
}
