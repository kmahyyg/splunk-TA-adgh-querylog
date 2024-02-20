package main

import (
	"adgh-querylog-preprocessor/comm"
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
	// read if progress exists
	if finfo, err := os.Stat(RECOVERY_FROM_PATH); err == nil && !finfo.IsDir() {
		lastSavedTSData, err := os.ReadFile(RECOVERY_FROM_PATH)
		if err != nil {
			log.Println("Error: cannot read last saved progress file: ", err.Error())
		}
		err = lastTS.UnmarshalBinary(lastSavedTSData)
		if err != nil {
			log.Println("Error: cannot read last saved progress file: ", err.Error())
		}
	} else {
		lastTS = time.Now()
	}
	// start check
	var lastTSChan = make(chan time.Time, 1)
	go func() {
		log.Println("Start Last-Fetch Timestamp Saver.")
		time.Sleep(2 * time.Second)
		fetchedCounter := 0
		for {
			// last timestamp saver
			currentTime := time.Now()
			i, isLast := <-lastTSChan
			lastTS = i
			fetchedCounter += 1
			// periodically save
			if isLast || currentTime.Sub(lastTS) > 300*time.Second || fetchedCounter > 100 {
				savedTS, _ := lastTS.MarshalBinary()
				err := os.WriteFile(RECOVERY_FROM_PATH, savedTS, 0644)
				log.Println("Last Progress save action triggered, progress: ", lastTS.String())
				if err != nil {
					log.Println("Save progress error: " + err.Error())
				}
				lastTS = currentTime
				if fetchedCounter > 100 {
					fetchedCounter = 0
				}
			}
		}
	}()
	// buffer 100 lines
	var logDataChan = make(chan *ext.ADGHLogEntry, 100)
	go func() {
		// maybe github.com/nxadm/tail
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
			if buf.Time.Sub(lastTS) <= 0 {
				logDataChan <- buf
				lastTSChan <- buf.Time
			}
		}
	}()
	tcpConnCli := &comm.TCPClient{}
	tcpConnCli.SetDest(conf.DestinationTcp)
	go func() {
		// tcp conn and duplicator and processor and data sender
		// always retry connection, for each object, parse log, remove answer, extend log, send final log to splunk
		err := tcpConnCli.Connect()
		if err != nil {
			panic(err)
		}
		for {
			log2Write := <-logDataChan
			n, err := ext.ParseAnswerInLog(log2Write)
			if err != nil {
				log.Println("Log Parser Error: ", err.Error())
			}
			if n >= 0 {
				err = ext.RemoveAnswerInLog(log2Write)
				if err != nil {
					log.Println("Log Filter Error: ", err.Error())
				}
			}
			// write to splunk
			fdata, err := json.Marshal(log2Write)
			if err != nil {
				log.Println("Marshal Log Error: ", err.Error())
			}
			_, err = tcpConnCli.Write(fdata)
			if err != nil {
				log.Println("Write To TCP Error: ", err.Error())
			}
		}
	}()
	var sigChan = make(chan os.Signal, 1)
	defer close(sigChan)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	<-sigChan
	close(lastTSChan)
	close(logDataChan)
	log.Println("Process Terminated.")
}
