# Qrystal

[Website/Docs](https://nyiyui.ca/qrystal) /
[Github.com](https://github.com/nyiyui/qrystal)

Qrystal /kristl/ sets up several WireGuard tunnels between servers.
In addition, it provides centralised configuration management.
Nodes and tokens can be dynamically added (and removed, in a future
version).

## Installation

1. [Install Go](https://go.dev/dl)
2. Download the source: https://github.com/nyiyui/qrystal/archive/refs/heads/next2goal.tar.gz
3. cd-into the source code

### Installing for a Device Client

4. `make device-client gen-keys`
5. `sudo make install-device`
  - If you see errors such as 'user not found', make sure systemd-sysusersd has run after the Makefile ran.
6. Edit config files at `/etc/qrystal-device/` (`gen-keys` will be useful here!)
7. `systemctl enable --now qrystal-device-client.service`

### Installing for a Coordiation Server

4. `make coord-server`
5. `sudo make install-coord`
  - If you see errors such as 'user not found', make sure systemd-sysusersd has run after the Makefile ran.
6. Edit config files at `/etc/qrystal-coord/`
7. `systemctl enable --now qrystal-coord-server.service`

## Contributing

Using [Nix](https://nixos.org/download/) and [direnv](https://direnv.net/) is recommended. To set up, install Nix and direnv, cd into this repo, then run `direnv allow`. This will setup your `$PATH` to have all the tools needed (and with the right versions) to develop.

Testing should be done using `go test ./...` for Go tests and `nix flake check` for NixOS tests. Note that `nix flake check` downloads a lot of files and is fairly slow/expensive (involves starting multiple VMs for testing).

Additionally, individual NixOS tests can be run:
```shell
# Example for running `goal` test:
nix build --print-build-logs .#checks.x86_64-linux.goal
# Run an interactive test:
nix build --print-build-logs .#checks.x86_64-linux.goal.driverInteractive && ./result/bin/nixos-test-driver
# Opens a Python REPL; run `test_script()` to run the test itself. See <https://wiki.nixos.org/wiki/NixOS_VM_tests> for details.
```

## TODO

- node: test node backport (in test.nix)
- confine qrystal-node and qrystal-cs (using systemd's options)
- configure existing interfaces without disrupting connections (as much as possible)
- support multiple hosts
  - e.g. specify VPC network IP address first, and then public IP address
  - heuristics for a successful wg connection?
- test all fails on `host cs` but after waiting a few hours, `host cs` works so I'll have to figure that out...
- if azusa contains configuration for a network that isn't in config.cs.networks, warn about this (possible misconfiguration)
- SRV records
