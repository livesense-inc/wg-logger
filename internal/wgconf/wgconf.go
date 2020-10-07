package wgconf

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

type WGConf struct {
	// Path to WireGuard config file
	Path string
	// Modtime is modified time of config file
	ModTime         time.Time
	friendryNameMap map[string]string
}

func New(confPath string) (wgconf *WGConf, err error) {
	stat, err := os.Stat(confPath)
	if os.IsNotExist(err) {
		return wgconf, fmt.Errorf("%s is not found", confPath)
	}

	return &WGConf{
		Path:    confPath,
		ModTime: stat.ModTime(),
	}, nil
}

// GetFriendlyNameMap returns a map with peer's public key as key, peer's friendly name as value.
func (c *WGConf) GetFriendlyNameMap() (names map[string]string, err error) {
	stat, err := os.Stat(c.Path)
	if os.IsNotExist(err) {
		return names, fmt.Errorf("%s is not found", c.Path)
	}

	if stat.ModTime() == c.ModTime && len(c.friendryNameMap) > 0 {
		// return cache
		return c.friendryNameMap, nil
	}

	names = make(map[string]string)
	file, err := os.Open(c.Path)
	if err != nil {
		return
	}
	defer file.Close()

	publicKey := ""
	friendlyName := ""
	lineNumber := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		lineNumber++

		switch {
		case strings.HasPrefix(line, "[Peer]"):
			fallthrough
		case strings.HasPrefix(line, "[peer]"):
			if publicKey != "" {
				if friendlyName != "" {
					names[publicKey] = friendlyName
				} else {
					names[publicKey] = "(no name)"
				}
			}
			publicKey = ""
			friendlyName = ""
			lineNumber = 0
			continue
		case strings.HasPrefix(strings.ToLower(strings.TrimLeft(line, " \t")), "publickey"):
			// parse public key
			kv := strings.SplitN(line, "=", 2)
			v := strings.SplitN(kv[1], "#", 2)
			publicKey = strings.TrimSpace(v[0])
			continue
		case lineNumber == 1:
			// the comment at 1st line is treated as friendly name
			if strings.HasPrefix(line, "#") {
				friendlyName = strings.TrimLeft(line, "# ")
			}
			continue
		}
	}

	if publicKey != "" {
		if friendlyName != "" {
			names[publicKey] = friendlyName
		} else {
			names[publicKey] = "no name"
		}
	}

	if err = scanner.Err(); err != nil {
		return
	}

	c.friendryNameMap = names
	return
}
