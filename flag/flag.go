package flag

import "flag"

var (
	// ConfigPath defines the file name and path of the configuration file
	ConfigPath string

	// LogLevel defines the log level in string format
	LogLevel string

	// LogMode can be either dev (development mode) or prod (production)
	LogMode string
)

func init() {
	flag.StringVar(&ConfigPath, "f", "logspout.json",
		"specify the config file in json format.")
	flag.StringVar(&LogLevel, "v", "warn",
		"debug, info, warn, error.")
	flag.StringVar(&LogMode, "log", "dev",
		"specify the log mode. it can be either dev or prod.")
	flag.Parse()
}
