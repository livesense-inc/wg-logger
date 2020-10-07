# wg-logger (WireGuard Logger)

wg-logger is a daemon program for logging [WireGuard](https://www.wireguard.com/) usage for debugging and auditing purposes.

## Features

Log status of WireGuard peers by using polling. The example is below.

```bash
$ grep 'endpoint updated' /var/log/wg-logger/wg.log | jq .

{
  "event": "endpoint updated",
  "friendly_name": "1st person",
  "event_time": "2020-09-24T18:12:54+09:00",
  "peer": {
    "public_key": "i+VdaJmF7mSlQlDQnEuFbo1JFicB2X054uN0DF5MICA=",
    "endpoint": "1.2.3.4:52978",
    "endpoint_ip": "1.2.3.4",
    "latest_handshake": "2020-09-24T18:12:54+09:00",
    "transfered_rx_per_endpoint": "0B",
    "transfered_tx_per_endpoint": "0B",
    "transfered_rx_per_endpoint_ip": "5.2MiB",
    "transfered_tx_per_endpoint_ip": "15.1MiB"
  },
  "time": "2020-09-24T18:12:58+09:00",
  "message": "status update"
}
```

Log format is JSON. And each log contains:

* event:
  * `handshake`: Invoked handshake. It means 'connection is active'.
  * `endpoint_ip updated`: Peer's IP Address was changed.
  * `endpoint updated`: Peer's UDP port number was changed.
  * `suspected inactive`: It hasn't handshaken for a long time, so it's probably been inactive.
  * `statistics`: Peer's information.
* event_time: timestamp of event occurs.
* friendly_name: human-readable peer name.
* peer: Peer's statistics.
* time: logging time.

## Requires

You must install following package(s).

* wireguard-tools (`wg` command)
  * on debian, 'wireguard'
  * see details at [WireGuard Official Site](https://www.wireguard.com/install/)

## Usage

### Quick start

Create 'wg-logger.conf' config file first. A sample is [in configs directory](configs/sample.conf).

wg-logger require WireGuard config file path for **Friendly Name** feature. At the minimum, please include the `wg_conf` setting. More information on Friendly Name is provided below.

You can use `--config-dump` option to see config parameters. `wg-logger --config-dump` outputs default parameters when config file does not exist.

```bash
$ wg-logger --config-dump
initializing wg-logger xxxx (rev:xxxx)...
Cannot read config file: /etc/wg-logger.conf

event_log_path = "/var/log/wg-logger/wg.log"
daemon_log_path = "/var/log/wg-logger/wg-logger.log"
log_max_mb = 100
log_max_days = 7
log_level = "info"
wg_conf = "/etc/wireguard/wg0.conf"
database = "/var/log/wg-logger/wg-logger.db"
interval = 30
suspected_inactive_threshold = 30
wg_tools_path = "wg"
```

Place the config file, run.

```bash
$ sudo wg-logger -c /etc/wg-logger.conf -d
```

You need systemd, nohup or etc to run wg-logger in background.

### Friendly Name

WireGuard uses base64-encoded public keys to distinguish between peers. This is not familiar with human. So wg-logger appends human-readable text for each messages. It's called 'Friendly Name'.

For this feature, you need to add comments to your WireGuard config file below the `[Peer]` definition. For example this is how you edit your WireGuard config file:

before:

```
[Peer]
PublicKey = i+VdaJmF7mSlQlDQnEuFbo1JFicB2X054uN0DF5MICA=
AllowedIPs = 192.168.100.1/32

[Peer]
PublicKey = 63clN7mNlJ7ckYH7VirX1VyAfXwR4t9DP9DRp2qMu0o=  #test
AllowedIPs = 192.168.100.2/32
```

after:

```
[Peer]
# 1st person
PublicKey = i+VdaJmF7mSlQlDQnEuFbo1JFicB2X054uN0DF5MICA=
AllowedIPs = 192.168.100.1/32

[Peer]
# 2nd person
PublicKey = 63clN7mNlJ7ckYH7VirX1VyAfXwR4t9DP9DRp2qMu0o=
AllowedIPs = 192.168.100.2/32
```

## Note

* wg-logger was born because WireGuard does not output access logs. (2020/09)
  * WireGuard is connection-less protocol. So there is no 'session start/end' time.
  * wg-logger detects that the peer status has changed. We call this 'event'.
* You must run this tool as root permission (because `wg` command needs root permission).
* Friendly Name comment is compatible with [Prometheus WireGuard Exporter](https://github.com/MindFlavor/prometheus_wireguard_exporter).

## License

See [LICENSE](LICENSE).

```
wg-logger
Copyright 2020 Livesense Inc.
```
