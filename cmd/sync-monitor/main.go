package main

import (
	"bufio"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/alecthomas/kingpin.v2"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"time"
)

const wrongTimeStamp = -1
const defaultInterval = 5 * time.Minute

var (
	lastsyncFile = kingpin.Flag("file", "Path to lastsync file").Required().ExistingFile()
	metricsPort  = kingpin.Flag("port", "Port to run metrics at").Uint16()
	interval     = kingpin.Flag("interval", "Time between checks").Duration()
)

var (
	noMetrics bool
	metric    = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "repo_sync_seconds",
		Help: "Second after last sync",
	})
)

func main() {
	kingpin.Parse()
	noMetrics = *metricsPort == 0
	go checkSyncDelta()

	if noMetrics {
		fmt.Printf("Press Ctrl+C to end\n")
		waitForCtrlC()
		fmt.Printf("\n")
	} else {
		portString := strconv.Itoa(int(*metricsPort))
		http.Handle("/metrics", promhttp.Handler())
		log.Printf("FTP Monitor metrics server listening at port :%s.", portString)
		err := http.ListenAndServe(":"+portString, nil)
		log.Fatal(err)
	}
}

func checkSyncDelta() {
	if interval == nil || *interval == 0 {
		*interval = defaultInterval
	}
	log.Println("Interval is", *interval)
	var delta int64
	var err error
	for {
		delta, err = readSyncDelta()
		if err != nil {
			log.Printf(err.Error())
			delta = -1
		}
		if noMetrics {
			log.Printf("Last sync was %d seconds ago", delta)
		} else {
			metric.Set(float64(delta))
		}
		time.Sleep(*interval)
	}
}

func readSyncDelta() (int64, error) {
	//read
	line, fileReadError := readSyncTimestamp()
	if fileReadError != nil {
		return wrongTimeStamp, fileReadError
	}

	//convert to ts
	timeStampAsInt, convertErr := strconv.ParseInt(line, 10, 0)
	if convertErr != nil {
		return wrongTimeStamp, convertErr
	}
	return time.Now().Unix() - timeStampAsInt, nil
}

func readSyncTimestamp() (string, error) {
	//read file
	lastSyncFileContent, fileReadErr := os.Open(*lastsyncFile)
	if fileReadErr != nil {
		return "", fileReadErr
	}
	defer lastSyncFileContent.Close()
	reader := bufio.NewReader(lastSyncFileContent)
	var line string

	//reading first line
	line, fileReadErr = reader.ReadString('\n')
	if fileReadErr != nil && fileReadErr != io.EOF {
		return "", fileReadErr
	}
	line = strings.TrimSuffix(line, "\n")
	return line, nil
}

func waitForCtrlC() {
	var endWaiter sync.WaitGroup
	endWaiter.Add(1)
	var signalChannel chan os.Signal
	signalChannel = make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt)
	go func() {
		<-signalChannel
		endWaiter.Done()
	}()
	endWaiter.Wait()
}
