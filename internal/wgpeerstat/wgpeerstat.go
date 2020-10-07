package wgpeerstat

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

type PeerStat struct {
	//Interface string
	PublicKey string
	//PreSharedKey        string
	Endpoint string
	//AllowedIPs      string
	LatestHandshake time.Time
	TransferRX      uint64
	TransferTX      uint64
	//PersistentKeepalive string
}

func (s PeerStat) MarshalZerologObject(e *zerolog.Event) {
	e.Str("public_key", s.PublicKey).
		Str("endpoint", s.Endpoint).
		Time("latest_handshake", s.LatestHandshake)
}

func parsePeerStat(line string) (peerStat PeerStat, err error) {
	values := strings.SplitN(line, "\t", 9)
	if len(values) != 9 {
		return peerStat, fmt.Errorf("Parse Error '%s'", line)
	}
	unixtime, err := strconv.ParseInt(values[5], 10, 64)
	if err != nil {
		return peerStat, fmt.Errorf("Parse Error at LatestHandshake '%s'", values[5])
	}
	transferRX, err := strconv.ParseUint(values[6], 10, 64)
	if err != nil {
		return peerStat, fmt.Errorf("Parse Error at TransferRX '%s'", values[6])
	}
	transferTX, err := strconv.ParseUint(values[7], 10, 64)
	if err != nil {
		return peerStat, fmt.Errorf("Parse Error at TransferTX '%s'", values[7])
	}

	return PeerStat{
		//Interface: values[0],
		PublicKey: values[1],
		//PreSharedKey:        values[2],
		Endpoint: values[3],
		//AllowedIPs:          values[4],
		LatestHandshake: time.Unix(unixtime, 0),
		TransferRX:      transferRX,
		TransferTX:      transferTX,
		//PersistentKeepalive: values[8],
	}, nil
}

var readWGDump = _readWGDump

func _readWGDump(wgCommand string) (lines []string, err error) {
	if wgCommand == "" {
		wgCommand = "wg"
	}
	cmd := exec.Command(wgCommand, "show", "all", "dump")
	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		return []string{""}, err
	}
	return strings.Split(out.String(), "\n"), nil
}

func GetPeerStats(wgCommand string) ([]PeerStat, error) {
	var peerStats []PeerStat
	lines, err := readWGDump(wgCommand)
	if err != nil {
		return peerStats, err
	}

	// wg-tools returns these lines:
	//   1: private-key,  public-key, listen-port, fwmark.
	//   2-: public-key, preshared-key, endpoint, allowed-ips, latest-handshake, transfer-rx, transfer-tx, persistent-keepalive.
	for n, l := range lines {
		if n == 0 {
			continue
		}
		line := strings.Trim(l, "\r\n \t")
		if len(l) <= 0 {
			continue
		}
		peerStat, err := parsePeerStat(line)
		if err != nil {
			continue
		}
		peerStats = append(peerStats, peerStat)
	}

	return peerStats, nil
}
