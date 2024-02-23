package main

import (
	"adgh-querylog-preprocessor/comm"
	"adgh-querylog-preprocessor/ext"
	"encoding/json"
	"github.com/nxadm/tail"
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

var (
	VersionNum = "v0.0.0-unknown"
)

type PreProcessorConfig struct {
	DestinationTcp string
	SourceFile     string
}

func main() {
	var err error
	var dstConn *net.TCPConn
	defer func() {
		log.Println("Sleep 3 seconds for cleanup, then exit.")
		time.Sleep(3 * time.Second)
	}()
	// init
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.LUTC | log.Lmicroseconds)
	log.Println("Start Initialization.")
	log.Println("Version: ", VersionNum)
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
			i, hasNext := <-lastTSChan
			if hasNext {
				lastTS = i
			}
			fetchedCounter += 1
			// periodically save
			if !hasNext || currentTime.Sub(lastTS) > 300*time.Second || fetchedCounter > 100 {
				savedTS, _ := lastTS.MarshalBinary()
				err := os.WriteFile(RECOVERY_FROM_PATH, savedTS, 0644)
				log.Println("Last Progress save action triggered, progress: ", lastTS.String())
				if err != nil {
					log.Println("Save progress error: " + err.Error())
				}
				if fetchedCounter > 100 {
					fetchedCounter = 0
				}
				if !hasNext {
					break
				}
			}
		}
	}()
	// buffer 100 lines
	var logDataChan = make(chan *ext.ADGHLogEntry, 100)
	go func() {
		file, err := tail.TailFile(conf.SourceFile, tail.Config{
			ReOpen:        true,
			MustExist:     true,
			Follow:        true,
			CompleteLines: true,
		})
		if err != nil {
			panic(err)
		}
		defer file.Cleanup()
		defer file.Stop()
		for line := range file.Lines {
			curLine := []byte(line.Text)
			buf := &ext.ADGHLogEntry{}
			err := json.Unmarshal(curLine, buf)
			if err != nil {
				log.Println("Unmarshal Error: ", err.Error())
			}
			log.Println("Read one line of Logs.")
			if buf.Time.Sub(lastTS) >= 0 {
				logDataChan <- buf
				//log.Printf("Written to LogDataChan: %v \n", buf)
			} else {
				log.Println("Current Log has been skipped after unmarshal.")
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
			log2Write, hasNext := <-logDataChan
			//log.Printf("Read From LogDataChan: %v \n", log2Write)
			if !hasNext {
				log.Println("logDataChan closed.")
				break
			}
			n, err := ext.ParseAnswerInLog(log2Write)
			if err != nil {
				log.Println("Log Parser Error: ", err.Error())
			}
			if n > 0 {
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
			//log.Println("Write To TCP: ", fdata)
			_, err = tcpConnCli.Write(fdata)
			if err != nil {
				log.Println("Write To TCP Error: ", err.Error())
			}
			lastTSChan <- log2Write.Time
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
