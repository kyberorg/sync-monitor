package config

import (
	"gopkg.in/alecthomas/kingpin.v2"
	"strings"
	"time"
)

const defaultInterval = 5 * time.Minute

var (
	lastsyncFile = kingpin.Flag("file", "Path to lastsync file").Required().ExistingFile()
	metricsPort  = kingpin.Flag("port", "Port to run metrics at").Uint16()
	interval     = kingpin.Flag("interval", "Time between checks").Duration()

	repoLocation = kingpin.Flag("manjaro.repo.path", "Path to directory with Manjaro repo").ExistingDir()
	repoList     = kingpin.Flag("manjaro.repo.list", "Comma separated list of repositories to check").String()
)

//internal vars
var (
	appConfig *applicationConfiguration
)

//applicationConfiguration application configuration values
type applicationConfiguration struct {
	//Lastsync file full name with path to
	LastsyncFile string
	//Port for running metrics at
	MetricsPort uint16
	//No Metrics flag
	NoMetrics bool
	//Time between checks
	Interval time.Duration
	//Path to dir with Manjaro Repo
	ManjaroRepoPath string
	//Repo names to check state file at
	ManjaroRepos []string
}

func init() {
	//parse flags
	kingpin.Parse()

	appConfig = &applicationConfiguration{
		LastsyncFile:    *lastsyncFile,
		MetricsPort:     *metricsPort,
		NoMetrics:       *metricsPort == 0,
		Interval:        *interval,
		ManjaroRepoPath: *repoLocation,
		ManjaroRepos:    csv2slice(*repoList),
	}

	if &appConfig.Interval == nil {
		appConfig.Interval = defaultInterval
	}
}

//GetAppConfig returns application configuration object
func GetAppConfig() *applicationConfiguration {
	return appConfig
}

func ShouldRunStateChecker() bool {
	return &appConfig.ManjaroRepoPath != nil && &appConfig.ManjaroRepos != nil && len(appConfig.ManjaroRepos) > 0
}

func csv2slice(csv string) []string {
	stringSlice := make([]string, 0)
	values := strings.Split(csv, ",")
	for _, value := range values {
		stringSlice = append(stringSlice, strings.TrimSpace(value))
	}
	return stringSlice
}
