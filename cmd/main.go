package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"
	"time"

	"github.com/livesense-inc/wg-logger/internal/config"
	"github.com/livesense-inc/wg-logger/internal/kvs"
	"github.com/livesense-inc/wg-logger/internal/logger"
	"github.com/livesense-inc/wg-logger/internal/wgconf"
	"github.com/livesense-inc/wg-logger/internal/wgpeerstat"
	"github.com/rs/zerolog"
	"github.com/urfave/cli/v2"
)

func bytesReadable(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%dB", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%ciB", float64(b)/float64(div), "KMGTPE"[exp])
}

type WGPeerStatLog struct {
	wgpeerstat.PeerStat
	EndpointIP                string
	TransferedRXPerEndpoint   uint64
	TransferedTXPerEndpoint   uint64
	TransferedRXPerEndpointIP uint64
	TransferedTXPerEndpointIP uint64
	SuspectedInactive         bool
}

func (s WGPeerStatLog) MarshalZerologObject(e *zerolog.Event) {
	e.Str("public_key", s.PublicKey).
		Str("endpoint", s.Endpoint).
		Str("endpoint_ip", s.EndpointIP).
		Time("latest_handshake", s.LatestHandshake).
		Str("transfered_rx_per_endpoint", bytesReadable(s.TransferedRXPerEndpoint)).
		Str("transfered_tx_per_endpoint", bytesReadable(s.TransferedTXPerEndpoint)).
		Str("transfered_rx_per_endpoint_ip", bytesReadable(s.TransferedRXPerEndpointIP)).
		Str("transfered_tx_per_endpoint_ip", bytesReadable(s.TransferedTXPerEndpointIP))
}

type WGLogger struct {
	Cache                      *kvs.KVS
	WGConf                     *wgconf.WGConf
	EventLogger                *zerolog.Logger
	DaemonLogger               *zerolog.Logger
	Interval                   int64
	SuspectedInactiveThreshold int64
	WgCommandPath              string
}

func (wgl *WGLogger) check() (err error) {
	wgl.DaemonLogger.Debug().
		Msg("check")
	names, err := wgl.WGConf.GetFriendlyNameMap()
	if err != nil {
		wgl.DaemonLogger.Error().
			Err(err).
			Msgf("Cannot read '%s'", wgl.WGConf.Path)
		return
	}

	stats, err := wgpeerstat.GetPeerStats(wgl.WgCommandPath)
	if err != nil {
		wgl.DaemonLogger.Error().
			Err(err).
			Msg("Cannot check wg-tool")
		return
	}

	for _, stat := range stats {
		var lastStat WGPeerStatLog
		v := wgl.Cache.Get(stat.PublicKey)
		if v == nil {
			lastStat = WGPeerStatLog{
				PeerStat: wgpeerstat.PeerStat{
					LatestHandshake: time.Unix(0, 0),
				},
			}
		} else {
			if err = json.Unmarshal(v, &lastStat); err != nil {
				wgl.DaemonLogger.Error().
					Err(err).
					Msgf("Invalid data was inserted into database '%s'", wgl.Cache.DBPath)
				return err
			}
		}

		endpointIP := "(none)"
		if i := strings.LastIndex(stat.Endpoint, ":"); i > 0 {
			endpointIP = stat.Endpoint[0:i]
		}

		// if counter was rotated, use current value as transfered bytes.
		transferedRX := stat.TransferRX - lastStat.TransferRX
		if stat.TransferRX < lastStat.TransferRX {
			wgl.DaemonLogger.Info().
				Object("peer", stat).
				Msg("TransferRX counter was rotated")
			transferedRX = stat.TransferRX
		}
		transferedTX := stat.TransferTX - lastStat.TransferTX
		if stat.TransferTX < lastStat.TransferTX {
			wgl.DaemonLogger.Info().
				Object("peer", stat).
				Msg("TransferTX counter was rotated")
			transferedTX = stat.TransferTX
		}

		curStat := WGPeerStatLog{
			PeerStat:                  stat,
			EndpointIP:                endpointIP,
			TransferedRXPerEndpoint:   lastStat.TransferedRXPerEndpoint + transferedRX,
			TransferedTXPerEndpoint:   lastStat.TransferedTXPerEndpoint + transferedTX,
			TransferedRXPerEndpointIP: lastStat.TransferedRXPerEndpointIP + transferedRX,
			TransferedTXPerEndpointIP: lastStat.TransferedTXPerEndpointIP + transferedTX,
		}

		if curStat.EndpointIP != lastStat.EndpointIP {
			// EndpointIP changed

			// output final stat of last endpoint
			finalStat := curStat
			finalStat.Endpoint = lastStat.Endpoint
			finalStat.EndpointIP = lastStat.EndpointIP
			finalStat.LatestHandshake = lastStat.LatestHandshake
			wgl.EventLogger.Log().
				Str("event", "statistics").
				Str("friendly_name", names[curStat.PublicKey]).
				Time("event_time", curStat.LatestHandshake).
				Object("peer", finalStat).
				Msg("endpoint statistics")

			// initialize stat and output first information
			curStat.TransferedRXPerEndpoint = 0
			curStat.TransferedTXPerEndpoint = 0
			curStat.TransferedRXPerEndpointIP = 0
			curStat.TransferedTXPerEndpointIP = 0
			curStat.SuspectedInactive = false
			wgl.EventLogger.Log().
				Str("event", "endpoint_ip updated").
				Str("friendly_name", names[curStat.PublicKey]).
				Time("event_time", curStat.LatestHandshake).
				Object("peer", curStat).
				Msg("status update")
		} else if curStat.Endpoint != lastStat.Endpoint {
			// Endpoint changed

			// output final stat of last endpoint
			finalStat := curStat
			finalStat.Endpoint = lastStat.Endpoint
			finalStat.EndpointIP = lastStat.EndpointIP
			finalStat.LatestHandshake = lastStat.LatestHandshake
			wgl.EventLogger.Log().
				Str("event", "statistics").
				Str("friendly_name", names[curStat.PublicKey]).
				Time("event_time", curStat.LatestHandshake).
				Object("peer", finalStat).
				Msg("endpoint statistics")

			// initialize stat and output first information
			curStat.TransferedRXPerEndpoint = 0
			curStat.TransferedTXPerEndpoint = 0
			curStat.SuspectedInactive = false
			wgl.EventLogger.Log().
				Str("event", "endpoint updated").
				Str("friendly_name", names[curStat.PublicKey]).
				Time("event_time", curStat.LatestHandshake).
				Object("peer", curStat).
				Msg("status update")
		} else if curStat.LatestHandshake != lastStat.LatestHandshake {
			// Handshake occured
			wgl.EventLogger.Log().
				Str("event", "handshake").
				Str("friendly_name", names[curStat.PublicKey]).
				Time("event_time", curStat.LatestHandshake).
				Object("peer", curStat).
				Msg("status update")
		} else if curStat.LatestHandshake != time.Unix(0, 0) &&
			time.Since(curStat.LatestHandshake) > time.Minute*time.Duration(wgl.SuspectedInactiveThreshold) &&
			curStat.TransferRX == lastStat.TransferRX &&
			curStat.TransferTX == lastStat.TransferTX {
			// suspect connection was inactive
			if !lastStat.SuspectedInactive {
				wgl.EventLogger.Log().
					Str("event", "suspected inactive").
					Str("friendly_name", names[curStat.PublicKey]).
					Time("event_time", curStat.LatestHandshake).
					Object("peer", curStat).
					Msgf("last handshake was %d minutes ago.", int64(time.Since(curStat.LatestHandshake).Minutes()))
			}
			curStat.SuspectedInactive = true
		}

		// save current stat
		data, err := json.Marshal(curStat)
		if err != nil {
			wgl.DaemonLogger.Error().
				Err(err).
				Msgf("Cannot marshal '%s' data", curStat.PublicKey)
		}
		if err = wgl.Cache.Set(curStat.PublicKey, data); err != nil {
			wgl.DaemonLogger.
				Err(err).
				Msgf("Cannot store '%s' data", curStat.PublicKey)
		}
	}

	return nil
}

func Action(c *cli.Context) error {
	fmt.Printf("initializing wg-logger %s (rev:%s)...\n", version, gitcommit)
	configPath := c.String("config")
	conf, err := config.GetConfig(configPath)
	if err != nil {
		msg := fmt.Sprintf("Cannot read config file: %s\n", configPath)
		fmt.Println(msg)
		return fmt.Errorf("%s", msg)
	}

	// override configs
	if c.String("loglevel") != "" {
		conf.LogLevel = c.String("loglevel")
	}
	if c.String("wireguard-config") != "" {
		conf.WGConf = c.String("wireguard-config")
	}
	if c.String("database") != "" {
		conf.Database = c.String("database")
	}
	if c.Int64("interval") > 0 &&
		c.Int64("interval") != conf.Interval {
		conf.Interval = c.Int64("interval")
	}
	if c.Int64("inactive-threshold") > 0 &&
		c.Int64("inactive-threshold") != conf.SuspectedInactiveThreshold {
		conf.SuspectedInactiveThreshold = c.Int64("inactive-threshold")
	}
	if c.String("wg-tools-path") != "" {
		conf.WGToolsPath = c.String("wg-tools-path")
	}

	if c.Bool("config-dump") {
		conf.PrintConfig()
		return nil
	}

	var EventLogger, DaemonLogger *zerolog.Logger
	if c.Bool("daemon") {
		EventLogger, DaemonLogger = logger.NewFileLogger(conf)
	} else {
		EventLogger, DaemonLogger = logger.NewConsoleLogger(conf)
	}

	if conf.Database == "" {
		msg := fmt.Sprintf("database path '%s' is invalid", conf.Database)
		DaemonLogger.Error().
			Msg(msg)
		return fmt.Errorf("%s", msg)
	}
	if err := os.MkdirAll(path.Dir(conf.Database), 0770); err != nil {
		DaemonLogger.Error().
			Err(err).
			Msgf("mkdir to database path '%s' failed", conf.Database)
		return err
	}

	cache, err := kvs.Open(conf.Database, "main")
	if err != nil {
		DaemonLogger.Error().
			Err(err).
			Msgf("open database '%s' failed", conf.Database)
		return err
	}
	defer cache.Close()

	wgConf, err := wgconf.New(conf.WGConf)
	if err != nil {
		DaemonLogger.Error().
			Err(err).
			Msgf("loading wireguard config file '%s' failed", conf.WGConf)
		return err
	}

	wglogger := WGLogger{
		Cache:                      cache,
		WGConf:                     wgConf,
		EventLogger:                EventLogger,
		DaemonLogger:               DaemonLogger,
		Interval:                   conf.Interval,
		SuspectedInactiveThreshold: conf.SuspectedInactiveThreshold,
		WgCommandPath:              conf.WGToolsPath,
	}

	DaemonLogger.Warn().
		Msg("wg-logger start")

	if err = wglogger.check(); err != nil {
		DaemonLogger.Warn().
			Err(err).
			Msg("Cannot check WireGuard status")
	}
	ch := make(chan os.Signal, 1)
	defer close(ch)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	for {
		select {
		case <-time.After(time.Duration(wglogger.Interval) * time.Second):
			go func() {
				if err = wglogger.check(); err != nil {
					DaemonLogger.Warn().
						Err(err).
						Msg("Cannot check WireGuard status")
				}
			}()
		case s := <-ch:
			DaemonLogger.Warn().
				Msgf("Signal '%s' received, shutting down wg-logger", s.String())
			return nil
		}
	}
}

var Flags = []cli.Flag{
	&cli.StringFlag{
		Name:    "config",
		Aliases: []string{"c"},
		Usage:   "specify the config file path",
		Value:   "/etc/wg-logger.conf",
	},
	&cli.StringFlag{
		Name:    "wireguard-config",
		Aliases: []string{"n"},
		Usage:   "specify the wireguard config file path",
	},
	&cli.StringFlag{
		Name:  "database",
		Usage: "specify the cache database file path",
	},
	&cli.BoolFlag{
		Name:    "daemon",
		Aliases: []string{"d"},
		Usage:   "run like daemon. the outputs will be written to logfiles",
	},
	&cli.StringFlag{
		Name:  "loglevel",
		Usage: "set loglevel (override config file)",
	},
	&cli.Int64Flag{
		Name:    "interval",
		Aliases: []string{"i"},
		Usage:   "set interval time in sedonds to check wireguard status",
	},
	&cli.Int64Flag{
		Name:  "inactive-threshold",
		Usage: "set threshold time in minutes to detect event 'suspected inactive'",
	},
	&cli.StringFlag{
		Name:    "wg-tools-path",
		Aliases: []string{"W"},
		Usage:   "specify the wg-tools(wg) path",
	},
	&cli.BoolFlag{
		Name:  "config-dump",
		Usage: "dump config parameters loaded. without '-c' option, dump default parameters",
	},
}

func main() {
	app := cli.NewApp()
	app.Name = "wireguard-logger"
	app.Version = fmt.Sprintf("%s (rev:%s)", version, gitcommit)
	app.Flags = Flags
	app.Action = Action

	if err := app.Run(os.Args); err != nil {
		fmt.Println("wg-logger stopped abnormaly.")
		os.Exit(1)
	}
	os.Exit(0)
}
