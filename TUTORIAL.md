# Qrystal Tutorial

Qrystal has a "client," which runs on each machine that you want to configura a VPN, and a "server," which runs on a single (generally publicly-available) server.

## Coordination Server

### Installing the Coordination Server

First, install the Coordination Server.

1. [Install Go](https://go.dev/dl)
2. Download the source: https://github.com/nyiyui/qrystal/archive/refs/heads/next2goal.tar.gz
3. cd-into the source code
4. `make coord-server`
5. `sudo make install-coord`
  - If you see errors such as 'user not found', make sure systemd-sysusersd has run after the Makefile ran.
6. Edit config files at `/etc/qrystal-coord/`
  - See the "Configuring the Coordination Server" section below.
7. `systemctl enable --now qrystal-coord-server.service`

## Configuring the Coordination Server

The Coordination Server configuration (at "/etc/qrystal-coord/config.json") has the following general process:

```json
{
  "Spec": "JSON object",
  "Tokens": {
  },
  "Addr": "0.0.0.0:39390",
  "CertPath": "",
  "KeyPath": ""
}
```

`Spec` is the network spec. Let's say we want to make a network called `qrystal0` (in 10.10.0.0/24) with three devices: the server, a desktop, and a mobile device (which does not run Qrystal's device client).

Assume the desktop has a static local IP address at 192.168.0.1.

Assume `qrystalcth_xxx` is the token hash for the token `qrystalct_xxx`.

The spec will be like:
```json
{
  "Spec": {
    "Networks": [
      {
        "Name": "qrystal0",
        "Devices": [
          {
            "Name": "server",
            "Endpoints": ["server.example.com:51820"],
            "Addresses": ["10.10.0.1/32"],
            "ListenPort": 51820,
            "PersistentKeepalive": "30s",
            "PublicKey": "oHy1MHcvxKcly2BKy7cg6cmrKNOCt4m7fijY2bMuVAQ="
          },
          {
            "Name": "desktop",
            "Endpoints": ["192.168.0.1"],
            "Addresses": ["10.10.0.2/32"],
            "ListenPort": 51820
          },
          {
            "Name": "mobile",
            "Addresses": ["10.10.0.3/32"],
            "ListenPort": 51820,
            "PublicKey": "UMT2E3W0iFVB+gctOPh1gFlPa93qlMH3F61N+P4LViI="
          }
        ],
      }
    ]
  },
  "Tokens": {
    "qrystalcth_xxx": {
      "Identities": [
        ["qrystal0", "server"]
      ]
    },
    "qrystalcth_yyy": {
      "Identities": [
        ["qrystal0", "desktop"]
      ]
    }
  }
}
```

### Securing the Coordination Server using TLS

If securing the server using TLS, specify `CertPath` and `KeyPath` to the TLS certificate and key paths.
Make sure these paths are readable by the server process.

## Device Client (WIP)

### Configuring the Mobile Device

Since the mobile device doesn't run Qrystal, we need to manually configure it:

```ini
[Interface]
PrivateKey=SIKhGomAeDywVwd/Rqox9iBZ4JtbIbO6YsNWcTxKgVY=
Addresses=10.10.0.3/32

[Peer]
AllowedIPs=10.10.0.1/32
PublicKey=oHy1MHcvxKcly2BKy7cg6cmrKNOCt4m7fijY2bMuVAQ=
```

Notice how the PublicKey= etc are the ones automatically set in the Coordination Server configuration above.

### Configuring the Desktop and Server

1. [Install Go](https://go.dev/dl)
2. Download the source: https://github.com/nyiyui/qrystal/archive/refs/heads/next2goal.tar.gz
3. cd-into the source code
4. `make coord-server`
5. `sudo make install-coord`
  - If you see errors such as 'user not found', make sure systemd-sysusersd has run after the Makefile ran.
6. See below for configuration.
7. `systemctl enable --now qrystal-coord-server.service`


The Device Client configuration (at "/etc/qrystal-coord/config.json") has the following general process:

```json
{
  "Clients": {
    "server": {
      "BaseURL": "https://server:39390",
      "Token": "qrystalct_xxx or qrystalct_yyy",
      "TokenPath": "/etc/qrystal-device/path-to-file-containing-token",
      "Network": "qrystal0",
      "Device": "server or desktop",
      "PrivateKey": "yCLuNHlGueV+V+af5LB3GMPqpfX1Bt76sOPtOA3J10U= if server, leave blank (will be automatically generated) if desktop",
      "PrivateKeyPath": "or the path to a file containing the private key",
      "MinimumInterval": "1m",
      "CertPath": "tls cert path if applicable"
    }
  }
}
```

MinimumInterval specifies the minimum amount of time until the device client contacts the server to check for an updated spec.
