package state

import (
	"bufio"
	"errors"
	"github.com/kyberorg/sync-monitor/cmd/sync-monitor/config"
	"github.com/kyberorg/sync-monitor/cmd/sync-monitor/constants"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
)

const (
	NoReposErr               = "nothing to check"
	ReadRepoDirErr           = "failed to read repo dir"
	EmptyRepoDirErr          = "repository dir is empty"
	NoSuchRepoErr            = "repo: not found"
	EmptyRepoErr             = "repo: got empty repo"
	NoStateFileErr           = "repo: state file not found"
	TimeStampLineNotFoundErr = "timestamp line not found"
)

const (
	stateFileName       = "state"
	invalidNum          = -1
	timeStampLinePrefix = "date="
)

type Checker struct {
	repositories []repository
}

type repository struct {
	stateFile string
	metric    prometheus.Gauge
}

func GetChecker() *Checker {
	return &Checker{
		repositories: make([]repository, 0),
	}
}

func (stateChecker *Checker) RunChecks() {
	initErr := stateChecker.verifyRepos()
	if initErr != nil {
		log.Printf(initErr.Error())
		return
	}

	//iterate thru checked repos
	var delta int64
	var err error
	for {
		for _, repo := range stateChecker.repositories {
			delta, err = readStateDelta(repo.stateFile)
			if err != nil {
				log.Printf(err.Error())
				delta = invalidNum
			}
			if config.GetAppConfig().NoMetrics {
				log.Printf("State file is %d seconds old", delta)
			} else {
				repo.metric.Set(float64(delta))
			}
		}
		time.Sleep(config.GetAppConfig().Interval)
	}
}

func readStateDelta(stateFile string) (int64, error) {
	//read timestamp
	timeStampStr, readErr := readStateTimestamp(stateFile)
	if readErr != nil {
		return constants.WrongTimeStamp, readErr
	} else if timeStampStr == constants.EmptyString {
		return constants.WrongTimeStamp, errors.New(TimeStampLineNotFoundErr + " in " + stateFile)
	}
	//convert
	timeStamp, convertErr := time.Parse(time.RFC3339, timeStampStr)
	if convertErr != nil {
		return constants.WrongTimeStamp, convertErr
	}

	return time.Now().Unix() - timeStamp.Unix(), nil
}

func readStateTimestamp(stateFileName string) (string, error) {
	//read file
	stateFile, fileReadErr := os.Open(stateFileName)
	if fileReadErr != nil {
		return "", fileReadErr
	}
	defer stateFile.Close()

	reader := bufio.NewReader(stateFile)
	var line = constants.EmptyString
	var lineFound = false
	var reachedEndOfFile = false
	for {
		//reading line by line
		fileLine, fileReadErr := reader.ReadString('\n')
		if fileReadErr != nil && fileReadErr != io.EOF {
			return "", fileReadErr
		} else if fileReadErr == io.EOF {
			reachedEndOfFile = true
		}

		//searching for line with date
		if strings.HasPrefix(fileLine, timeStampLinePrefix) {
			line = strings.TrimPrefix(fileLine, timeStampLinePrefix)
			lineFound = true
		}
		if lineFound || reachedEndOfFile {
			break
		}
	}
	//remove line delimiter, if any
	line = strings.TrimSuffix(line, "\n")
	return line, nil
}

func (stateChecker *Checker) verifyRepos() error {
	if len(config.GetAppConfig().ManjaroRepos) == 0 {
		return errors.New(NoReposErr)
	}

	//read repo dir and got num of dirs
	numOfRepos, err := stateChecker.countRepos()
	if err != nil {
		return err
	} else if numOfRepos == 0 {
		return errors.New(EmptyRepoDirErr)
	}

	for _, repoName := range config.GetAppConfig().ManjaroRepos {
		//repo exists
		repoDir := config.GetAppConfig().ManjaroRepoPath + string(os.PathSeparator) + repoName
		repo, err := ioutil.ReadDir(repoDir)
		if err != nil {
			log.Printf("%s %s", repoName, NoSuchRepoErr)
			continue
		}
		//repo not empty
		if len(repo) <= 0 {
			log.Printf("%s %s", repoName, EmptyRepoErr)
			continue
		}
		var stateFile = constants.EmptyString
		//repo's state file here
		for _, file := range repo {
			if file.Name() == stateFileName {
				stateFile = file.Name()
				break
			}
		}
		//add to repo list
		if len(stateFile) > 0 {
			repository := repository{
				stateFile: repoDir + string(os.PathSeparator) + stateFile,
				metric: promauto.NewGauge(
					prometheus.GaugeOpts{
						Name: "repo_" + repoName + "_state_seconds_old",
						Help: "Seconds after timestamp in '" + repoName + "' Repo State File",
					}),
			}
			stateChecker.repositories = append(stateChecker.repositories, repository)
		} else {
			log.Printf("%s %s", repoName, NoStateFileErr)
		}
	}
	return nil
}

func (stateChecker *Checker) countRepos() (int, error) {
	files, err := ioutil.ReadDir(config.GetAppConfig().ManjaroRepoPath)
	if err != nil {
		return invalidNum, errors.New(ReadRepoDirErr)
	}
	var counter int
	for _, file := range files {
		if file.IsDir() {
			counter++
		}
	}
	return counter, nil
}
