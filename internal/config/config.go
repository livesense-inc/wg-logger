package config

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	// EventLogPath is the output path for wirerguard access log.
	EventLogPath string `toml:"event_log_path"`
	// DaemonLogPathr is the output path for wg-logger internal log.
	DaemonLogPath string `toml:"daemon_log_path"`
	// LogMaxMB is the maximum size in megabytes of the log file before it gets rotated.
	LogMaxMB int `toml:"log_max_mb"`
	// LogMaxDays is the maximum number of days to retain old log files based on the
	// timestamp encoded in their filename.
	LogMaxDays int `toml:"log_max_days"` // keepdays
	// LogLevel is string, choosen from 'error', 'warn', 'info', 'debug'
	LogLevel string `toml:"log_level"`
	// WGConf is the path to wireguard config file
	WGConf string `toml:"wg_conf"`
	// Database is the path to database file (peristent data)
	Database string `toml:"database"`
	// Interval is the interval time in seconds to check wireguard status
	Interval int64 `toml:"interval"`
	// SuspectedInactiveThreshold is the threshold time in minutes to detect event 'suspected inactive'
	SuspectedInactiveThreshold int64 `toml:"suspected_inactive_threshold"`
	// WGToolsPath is the path to wg-tools(wg) command
	WGToolsPath string `toml:"wg_tools_path"`
}

// PrintConfig prints current config parameters as TOML
func (c *Config) PrintConfig() {
	fmt.Printf("\n")
	var buffer bytes.Buffer
	encoder := toml.NewEncoder(&buffer)
	err := encoder.Encode(c)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%v\n", buffer.String())
}

func GetDefault() *Config {
	return &Config{
		EventLogPath:               "/var/log/wg-logger/wg.log",
		DaemonLogPath:              "/var/log/wg-logger/wg-logger.log",
		LogMaxMB:                   100,
		LogMaxDays:                 7,
		LogLevel:                   "info",
		WGConf:                     "/etc/wireguard/wg0.conf",
		Database:                   "/var/log/wg-logger/wg-logger.db",
		Interval:                   30,
		SuspectedInactiveThreshold: 30,
		WGToolsPath:                "wg",
	}
}

// GetConfig loads config file.
func GetConfig(conf string) (*Config, error) {
	if len(conf) <= 0 {
		return GetDefault(), fmt.Errorf("you must set 'conf' option")
	}
	if _, err := os.Stat(conf); os.IsNotExist(err) {
		return GetDefault(), fmt.Errorf("config file '%s' is not found", conf)
	}

	f, err := os.Open(conf)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	buf, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	// Set default values
	config := GetDefault()
	if _, err = toml.Decode(string(buf), config); err != nil {
		return nil, err
	}

	return config, nil
}
