package config

import (
	"reflect"
	"testing"
)

func Test_GetDefault(t *testing.T) {
	config := GetDefault()

	configTests := []struct {
		Name string
		Want interface{}
	}{
		{"EventLogPath", "/var/log/wg-logger/wg.log"},
		{"DaemonLogPath", "/var/log/wg-logger/wg-logger.log"},
		{"WGConf", "/etc/wireguard/wg0.conf"},
		{"Database", "/var/log/wg-logger/wg-logger.db"},
		{"LogMaxMB", 100},
		{"LogMaxDays", 7},
		{"LogLevel", "info"},
		{"Interval", int64(30)},
		{"SuspectedInactiveThreshold", int64(30)},
		{"WGToolsPath", "wg"},
	}

	v := reflect.Indirect(reflect.ValueOf(config))
	for _, tt := range configTests {
		if out := v.FieldByName(tt.Name).Interface(); !reflect.DeepEqual(out, tt.Want) {
			t.Errorf("%s: \n out:  %#v\n want: %#v", tt.Name, out, tt.Want)
		}
	}
}

func Test_GetConfig(t *testing.T) {
	config, _ := GetConfig("../../configs/sample.conf")
	configTests := []struct {
		Name string
		Want interface{}
	}{
		{"EventLogPath", "/var/log/wg-logger/wg.log"},
		{"DaemonLogPath", "/var/log/wg-logger/wg-logger.log"},
		{"WGConf", "/etc/wireguard/wg0.conf"},
		{"Database", "/var/tmp/wg-logger.db"},
		{"EventLogPath", "/var/log/wg-logger/wg.log"},
		{"DaemonLogPath", "/var/log/wg-logger/wg-logger.log"},
		{"LogMaxMB", 256},
		{"LogMaxDays", 3},
		{"LogLevel", "debug"},
		{"Interval", int64(10)},
		{"SuspectedInactiveThreshold", int64(15)},
		{"WGToolsPath", "/usr/bin/wg"},
	}
	v := reflect.Indirect(reflect.ValueOf(config))
	for _, tt := range configTests {
		if out := v.FieldByName(tt.Name).Interface(); !reflect.DeepEqual(out, tt.Want) {
			t.Errorf("%s: \n out:  %#v\n want: %#v", tt.Name, out, tt.Want)
		}
	}
}
