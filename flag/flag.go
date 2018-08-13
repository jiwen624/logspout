package flag

import "flag"

var (
	// ConfigPath defines the file name and path of the configuration file
	ConfigPath string

	// LogLevel defines the log level in string format
	LogLevel string
)

func init() {
	flag.StringVar(&ConfigPath, "f", "logspout.json",
		"specify the config file in json format")
	flag.StringVar(&LogLevel, "v", "info",
		"Log level: debug, info, warning, error")

	flag.Parse()
}
