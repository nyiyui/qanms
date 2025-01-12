args@{
  self,
  system,
  nixpkgsFor,
  libFor,
  nixosLibFor,
  ldflags,
  packages,
  ...
}:
{
  default =
    {
      config,
      lib,
      pkgs,
      ...
    }:
    with lib;
    with types;
    let
      coordCfg = config.services.qrystal-coord-server;
      deviceCfg = config.services.qrystal-device-client;
      networkName = addCheck str (s: (builtins.stringLength s) <= 15);
      tokenHashType = addCheck str (hasPrefix "qrystalcth_");
      tokenType = addCheck str (hasPrefix "qrystalct_");
      specType = submodule {
        options.Networks = mkOption {
          type = listOf networkType;
          default = [ ];
        };
      };
      networkType = submodule {
        options.Name = mkOption { type = networkName; };
        options.Devices = mkOption {
          type = listOf deviceType;
          default = [ ];
        };
      };
      deviceTypeRaw = submodule {
        options.Name = mkOption { type = str; };
        options.Endpoints = mkOption {
          type = listOf str;
          default = [ ];
          description = "Unordered list of endpoints on which the peer is available on. Leave blank if the peer is not accessible from any other peer (e.g. behind a NAT).";
        };
        options.Addresses = mkOption {
          type = addCheck (listOf str) (l: (builtins.length l) > 0);
          description = "List of IP networks that this peer represents.";
        };
        options.ListenPort = mkOption {
          type = port;
          default = 0;
          description = "The port that WireGuard will listen on. Set to 0 to not specify.";
        };
        options.PublicKey = mkOption {
          type = nullOr str;
          default = null;
          description = "Base64 public key. Leave empty string to allow peer to set it automatically (using the peer's private key).";
        };
        options.PresharedKeyPath = mkOption {
          type = nullOr path;
          default = null;
          description = "Path to Base64 pre-shared key. Leave null to allow peer to set it automatically.";
        };
        options.PersistentKeepalive = mkOption {
          type = str;
          default = "0s";
          description = "Specifies how oftan a packet is sent by WireGuard to keep make sure the connection is seen as alive. Leave zero to disable.";
        };
        options.ForwardsFor = mkOption {
          type = listOf str;
          default = [ ];
          description = "List of devices (in the same network) that this peer has access to, and can fowrard packets to.";
        };
        options.AccessAll = mkOption {
          type = bool;
          default = true;
          description = "If true, this device can access all devices on the network. TODO: note that some devices may not have access backwards, leading to no useful connection.";
        };
        options.AccessOnly = mkOption {
          type = listOf str;
          default = [ ];
          description = "List of devices this device can access.";
        };
      };
      deviceType = addCheck deviceTypeRaw (d: (d.AccessAll == true) != ((length d.AccessControl) > 0));
      clientConfig = submodule {
        options.BaseURL = mkOption {
          type = str;
          description = "Qrystal coordination server base URL.";
        };
        options.TokenPath = mkOption {
          type = path;
          description = "Token to use with coordination server.";
        };
        options.Network = mkOption { type = networkName; };
        options.Device = mkOption { type = str; };
        options.PrivateKeyPath = mkOption {
          type = nullOr str;
          default = null;
          description = "Path to Base64 privateKey for this WireGuard interface. Leave blank to autogenerate.";
        };
        options.MinimumInterval = mkOption {
          type = str;
          default = "2m";
          description = "minimum interval to poll for updates to coordination server.";
        };
        options.CertPath = mkOption {
          type = nullOr path;
          default = null;
          description = "TLS certificate to use with server.";
        };
      };
      dnsParent = submodule {
        options.Suffix = mkOption {
          type = str;
          description = "DNS suffix. Precede with a dot if this suffix does not specify a network and device.";
        };
        options.Network = mkOption {
          type = networkName;
          default = "";
          description = "Preset network for this parent.";
        };
        options.Device = mkOption {
          type = str;
          default = "";
          description = "Preset device for this parent.";
        };
      };
    in
    {
      options.services.qrystal-coord-server = {
        enable = mkEnableOption "Qrystal coordination server to centrally manage network configurations.";
        openFirewall = mkOption {
          type = bool;
          description = "Opens the respective port in the firewall.";
          default = false;
        };
        bind = mkOption {
          type = submodule {
            options.address = mkOption {
              type = str;
              default = "0.0.0.0";
            };
            options.port = mkOption {
              type = port;
              default = 39390;
            };
          };
        };
        certPath = mkOption {
          type = nullOr path;
          default = null;
          description = "TLS certificate for HTTPS server. Using this enables HTTPS and disables HTTP.";
        };
        keyPath = mkOption {
          type = nullOr path;
          default = null;
          description = "TLS private key for HTTPS server. Using this enables HTTPS and disables HTTP.";
        };
        config = mkOption {
          type = submodule {
            options = {
              Spec = mkOption {
                type = specType;
                description = "specification for this server to provide";
              };
              Tokens = mkOption {
                type = attrsOf (submodule {
                  options.Identities = mkOption {
                    type = listOf (addCheck (listOf str) (l: (length l) == 2));
                    description = "The devices that this token can identify as (i.e. perform actions as). Tuple with two values, network and then device.";
                  };
                });
                description = "token hashes and their authorized actions.";
              };
            };
          };
        };
      };
      options.services.qrystal-device-client = {
        enable = mkEnableOption "Qrystal on-device client for WireGuard configuration.";
        config = mkOption {
          type = submodule {
            options = {
              Clients = mkOption {
                type = attrsOf clientConfig;
                default = [ ];
              };
              CanForward = mkOption {
                type = bool;
                default = true;
                description = "Enables the device to forward packets between peers.";
              };
              AssumeProc = mkOption {
                type = bool;
                default = true;
                description = "Instead of writing to procfs to change options, assume they are already set and do not write to procfs.";
              };
              dns = mkOption {
                type = submodule {
                  options.enable = mkEnableOption "Qrystal on-device DNS server.";
                  options.Parents = mkOption {
                    type = listOf dnsParent;
                    description = "List of DNS names to respond to.";
                    default = [ { Suffix = ".qrystal.internal"; } ];
                  };
                  options.Address = mkOption {
                    type = str;
                    description = "Address DNS server listens on.";
                    default = "127.0.0.39:53";
                  };
                };
                default.enable = false;
              };
            };
          };
        };
      };
      config =
        let
          baseServiceConfig = {
            PrivateTmp = true;
            NoNewPrivileges = true;
            ProtectSystem = "strict";
            ProtectHome = true;
            ProtectKernelTunables = true;
            ProtectKernelModules = true;
            ProtectControlGroups = true;
            RestrictNamespaces = true;
            PrivateMounts = true;
          };
        in
        mkMerge [
          (mkIf coordCfg.enable {
            systemd.services.qrystal-coord-server = {
              script = ''
                ${packages.coord-server}/bin/coord-server --config=${pkgs.writeText "qrystal-coord-server-config.json" (builtins.toJSON coordCfg.config)} --addr=${coordCfg.bind.address}:${builtins.toString coordCfg.bind.port}
              '';
              serviceConfig = {
                Type = "notify";
                NotifyAccess = "all";
                DynamicUser = true;
              } // baseServiceConfig;
              wantedBy = [ "multi-user.target" ];
            };
          })
          (mkIf (coordCfg.enable && coordCfg.openFirewall) {
            networking.firewall.allowedTCPPorts = [ coordCfg.bind.port ];
          })
          (mkIf deviceCfg.enable {
            users.groups.qrystal-device = { };
            users.users.qrystal-device = {
              isSystemUser = true;
              description = "Qrystal on-device services";
              group = "qrystal-device";
            };
            systemd.services.qrystal-device-client = {
              requires = [ "network.target" ];
              serviceConfig = {
                ExecStart = "${packages.device}/bin/device-client --config=${pkgs.writeText "qrystal-device-client-config.json" (builtins.toJSON deviceCfg.config)} --dns-config=${pkgs.writeText "qrystal-device-dns-config.json" (builtins.toJSON deviceCfg.config.dns)} --dns-self=true";
                Type = "notify";
                NotifyAccess = "all";
                StateDirectory = [ "qrystal-device-client" ];
                AmbientCapabilities = [
                  "CAP_NET_ADMIN"
                  "CAP_NET_BIND_SERVICE"
                ];
                User = "qrystal-device";
              } // baseServiceConfig;
              wantedBy = [ "multi-user.target" ];
              path = [ pkgs.iputils ];
            };
          })
        ];
    };
}
