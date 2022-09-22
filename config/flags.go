//nolint:gochecknoglobals
package config

import "flag"

var (
	// String that contains the configured configuration path.
	configPath string
	debug      bool
)

//nolint:gochecknoinits
func init() {
	// Set up a CLI flag called "-config" to allow users
	// to supply the configuration file
	flag.StringVar(&configPath, "config", "./config.yml", "path to config file")
	flag.BoolVar(&debug, "debug", false, "sets global log level to debug, otherwise defaults to info")
}
