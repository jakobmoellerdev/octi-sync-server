package config

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"

	"github.com/jakob-moeller-cloud/octi-sync-server/service"

	"github.com/go-redis/redis/v9"
	"gopkg.in/yaml.v3"
)

var ErrIsADirectory = errors.New("is a directory, not a normal file")

type LogSettingsFormat string

//goland:noinspection ALL
const (
	LogSettingsFormatJSON   LogSettingsFormat = "json"
	LogSettingsFormatPretty LogSettingsFormat = "pretty"
	LogSettingsFormatNone   LogSettingsFormat = ""
)

type LogSettings struct {
	Format LogSettingsFormat `yaml:"format"`
}

// Config struct for webapp config.
type Config struct {
	Server struct {
		// Host is the local machine IP Address to bind the HTTP Server to
		Host string `yaml:"host"`

		// Port is the local machine TCP Port to bind the HTTP Server to
		Port string `yaml:"port"`

		Timeout struct {
			// Server is the general server timeout to use
			// for graceful shutdowns
			Server time.Duration `yaml:"server"`

			// Write is the amount of time to wait until an HTTP server
			// write operation is cancelled
			Write time.Duration `yaml:"write"`

			// Read is the amount of time to wait until an HTTP server
			// read operation is cancelled
			Read time.Duration `yaml:"read"`

			// Read is the amount of time to wait
			// until an IDLE HTTP session is closed
			Idle time.Duration `yaml:"idle"`
		} `yaml:"timeout"`

		// MaxRequestBodySize can be specified as `4x` or `4xB`, where x is one of the multiple from K, M,
		// G, T or P.
		MaxRequestBodySize string `yaml:"maxRequestBodySize"`
	} `yaml:"server"`

	Redis struct {
		redis.Options `yaml:",inline"`
		Ping          struct {
			Enable   bool          `yaml:"enable"`
			Timeout  time.Duration `yaml:"timeout"`
			Interval time.Duration `yaml:"interval"`
		} `yaml:"ping"`
	} `yaml:"redis"`

	LogSettings `yaml:"log"`
	Logger      *zerolog.Logger `yaml:"-"`

	Services struct {
		service.Accounts
		service.Modules
		service.Devices
		service.Health
	} `yaml:"-"`
}

// NewConfig returns a new decoded Config struct.
func NewConfig(configPath string) (*Config, error) {
	// Create config structure
	config := &Config{}

	// Open config file
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Init new YAML decode
	d := yaml.NewDecoder(file)

	// Start YAML decoding from file
	if err := d.Decode(&config); err != nil {
		return nil, err
	}

	return config, nil
}

// ValidateConfigPath just makes sure, that the path provided is a file,
// that can be read.
func ValidateConfigPath(path string) error {
	s, err := os.Stat(path)
	if err != nil {
		return err
	}
	if s.IsDir() {
		return fmt.Errorf("'%s':%w", path, ErrIsADirectory)
	}
	return nil
}

// ParseFlags will create and parse the CLI flags
// and return the path to be used elsewhere.
func ParseFlags() (string, error) {
	// String that contains the configured configuration path
	var configPath string

	// Set up a CLI flag called "-config" to allow users
	// to supply the configuration file
	flag.StringVar(&configPath, "config", "./config.yml", "path to config file")

	debug := flag.Bool("debug", false, "sets global log level to debug, otherwise defaults to info")

	// Actually parse the flags
	flag.Parse()

	// Default level for this example is info, unless debug flag is present
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	// Validate the path first
	if err := ValidateConfigPath(configPath); err != nil {
		return "", err
	}

	// Return the configuration path
	return configPath, nil
}
