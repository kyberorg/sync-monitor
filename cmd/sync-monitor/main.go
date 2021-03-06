package main

import (
	"bufio"
	"fmt"
	"github.com/kyberorg/sync-monitor/cmd/sync-monitor/config"
	"github.com/kyberorg/sync-monitor/cmd/sync-monitor/constants"
	"github.com/kyberorg/sync-monitor/cmd/sync-monitor/state"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

var (
	metric = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "repo_sync_seconds",
		Help: "Second after last sync",
	})
)

func main() {
	go checkSyncDelta()
	if config.ShouldRunStateChecker() {
		go state.GetChecker().RunChecks()
	}

	if config.GetAppConfig().NoMetrics {
		fmt.Printf("Press Ctrl+C to end\n")
		waitForCtrlC()
		fmt.Printf("\n")
	} else {
		portString := strconv.Itoa(int(config.GetAppConfig().MetricsPort))
		http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
			serveMainPage(writer)
		})
		http.Handle("/metrics", promhttp.Handler())
		log.Printf("Sync Monitor metrics server listening at port :%s.", portString)
		err := http.ListenAndServe(":"+portString, nil)
		log.Fatal(err)
	}
}

func checkSyncDelta() {
	log.Println("Interval is", config.GetAppConfig().Interval)
	var delta int64
	var err error
	for {
		delta, err = readSyncDelta()
		if err != nil {
			log.Printf(err.Error())
			delta = -1
		}
		if config.GetAppConfig().NoMetrics {
			log.Printf("Last sync was %d seconds ago", delta)
		} else {
			metric.Set(float64(delta))
		}
		time.Sleep(config.GetAppConfig().Interval)
	}
}

func readSyncDelta() (int64, error) {
	//read
	line, fileReadError := readSyncTimestamp()
	if fileReadError != nil {
		return constants.WrongTimeStamp, fileReadError
	}

	//convert to ts
	timeStampAsInt, convertErr := strconv.ParseInt(line, 10, 0)
	if convertErr != nil {
		return constants.WrongTimeStamp, convertErr
	}
	return time.Now().Unix() - timeStampAsInt, nil
}

func readSyncTimestamp() (string, error) {
	//read file
	lastSyncFileContent, fileReadErr := os.Open(config.GetAppConfig().LastsyncFile)
	if fileReadErr != nil {
		return constants.EmptyString, fileReadErr
	}
	defer lastSyncFileContent.Close()
	reader := bufio.NewReader(lastSyncFileContent)
	var line string

	//reading first line
	line, fileReadErr = reader.ReadString('\n')
	if fileReadErr != nil && fileReadErr != io.EOF {
		return constants.EmptyString, fileReadErr
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

func serveMainPage(w http.ResponseWriter) {
	_, err := w.Write([]byte(`<html>
			<head><title>Sync Monitor</title></head>
			<body>
			<h1>Sync Monitor Metrics Server</h1>
			<p><a href="/metrics">Metrics</a></p>
			</body>
			</html>`))
	if err != nil {
		log.Fatalln("Failed to server main page")
	}
}
