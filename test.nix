args@{
  self,
  system,
  nixpkgsFor,
  libFor,
  nixosLibFor,
  ldflags,
  ...
}:
let
  pkgs = nixpkgsFor.${system};
  lib = nixosLibFor.${system} { inherit system; };
  rootCert = builtins.readFile ./cert/minica.pem;
  rootKey = builtins.readFile ./cert/minica-key.pem;
  csCert = builtins.readFile ./cert/cs/cert.pem;
  csKey = builtins.readFile ./cert/cs/key.pem;

  autologin =
    { ... }:
    {
      services.getty.autologinUser = "root";
    };
  base =
    { ... }:
    {
      imports = [ autologin ];
      virtualisation.vlans = [ 1 ];
      environment.systemPackages = with pkgs; [ wireguard-tools ];
      services.logrotate.enable = false; # clogs up the logs
    };
in
{
  goal = lib.runTest (
    let
      peerPrivateKey = "kCtV08G5gyM/cGHToObIAtwRq/bqI2Jd3akIsAMXRXM=";
      peerPublicKey = "72zpXYpjSWnvyhwZTuRNwtghjxjzhWEVzUNRA82hoUA=";
      defaultPrivateKey = "eDq8aX08rF5cLG+NNi14Ae8TIudsMHiWCjsbBTDI1Ec=";
      defaultPublicKey = "+atCYz0YmiwBx4AZy5kDGr5WHqHs3RMbIuPfj593sRk=";
      etc = self.outputs.packages.${system}.etc;
      machine1 = pkgs.writeText "machine1.json" (builtins.toJSON { });
      machine2 = pkgs.writeText "machine2.json" (
        builtins.toJSON {
          Interfaces = [
            {
              Name = "wiring";
              PrivateKey = defaultPrivateKey;
              ListenPort = 51820;
              Addresses = [ "10.10.0.0/32" ];
              Peers = [
                {
                  Name = "peer";
                  PublicKey = peerPublicKey;
                  Endpoint = "peer:51820";
                  AllowedIPs = [ "10.10.0.1/32" ];
                  PersistentKeepalive = "30s";
                }
              ];
            }
          ];
        }
      );
      machine3 = pkgs.writeText "machine3.json" (
        builtins.toJSON {
          Interfaces = [
            {
              Name = "wiring";
              PrivateKey = defaultPrivateKey;
              ListenPort = 51820;
              Addresses = [ "10.10.0.0/32" ];
              Peers = [
                {
                  Name = "peer";
                  PublicKey = peerPublicKey;
                  Endpoint = "peer:51820";
                  AllowedIPs = [ "10.10.0.8/32" ];
                }
              ];
            }
          ];
        }
      );
      continuityPort = 51821;
      continuityServer = pkgs.writeText "continuityServer.py" ''
        import socketserver
        import os

        class MyTCPHandler(socketserver.BaseRequestHandler):
            def handle(self):
                counter = 0
                while True:
                    print("waiting...")
                    self.data = self.request.recv(1024).strip()
                    print(f"received from {self.client_address[0]}: {self.data}")
                    i = int(self.data.decode('ascii'))
                    print('a')
                    if i != counter + 1:
                        print("out of order")
                        exit(1)
                    counter = i
                    print('b')
                    self.request.sendall("ok\n".encode("ascii"))
                    print("sent.")

        if __name__ == "__main__":
            host, port = os.getenv("HOST"), int(os.getenv("PORT"))

            with socketserver.TCPServer((host, port), MyTCPHandler) as server:
                print(f"serving on {host}:{port}...")
                server.serve_forever()
      '';
      continuityClient = pkgs.writeText "continuityClient.py" ''
        import os
        import signal
        import socket
        import sys
        import time

        host, port = os.getenv("HOST"), int(os.getenv("PORT"))
        print(f"host {host}, port {port}")

        with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as sock:
            print("connecting...")
            sock.connect((host, port))
            print("connected.")
            counter = 1
            while True:
                if counter == 2:
                    print("setting signal handler for graceful exit...")
                    signal.signal(signal.SIGTERM, lambda signum, frame: exit(0))
                sock.sendall(bytes(str(counter) + "\n", "utf-8"))
                received = str(sock.recv(1024), "utf-8")
                print(f"received {len(received)}: {received}")
                if "ok" not in received:
                    raise RuntimeError("not ok")
                print(f"{counter} ok")
                time.sleep(0.3)
                counter += 1
      '';
    in
    {
      name = "goal";
      hostPkgs = pkgs;
      nodes.default =
        { pkgs, ... }:
        {
          imports = [ base ];
          environment.systemPackages = [ self.outputs.packages.${system}.etc ];
          systemd.services.goal1.script = ''
            QRYSTAL_LOGGING_CONFIG=development ${etc}/bin/test-goal -a-path ${machine1} -b-path ${machine2}
          '';
          systemd.services.goal1.path = [ pkgs.iputils ];
          systemd.services.goal2.script = ''
            QRYSTAL_LOGGING_CONFIG=development ${etc}/bin/test-goal -a-path ${machine2} -b-path ${machine3}
          '';
          systemd.services.goal2.path = [ pkgs.iputils ];
          systemd.services."continuityClient" = {
            environment.HOST = "peer";
            environment.PORT = builtins.toString continuityPort;
            script = "${pkgs.python3}/bin/python3 ${continuityClient}";
          };
        };
      nodes.peer =
        { pkgs, ... }:
        {
          imports = [ base ];
          #TODO: wireguard config
          networking.wireguard.interfaces = {
            wiring = {
              ips = [ "10.10.0.1/32" ];
              privateKey = peerPrivateKey;
              listenPort = 51820;
              peers = [
                {
                  publicKey = defaultPublicKey;
                  allowedIPs = [ "10.10.0.0/32" ];
                  endpoint = "default:51820";
                  persistentKeepalive = 30;
                }
              ];
            };
          };
          networking.firewall.allowedTCPPorts = [ continuityPort ];
          systemd.services."continuityServer" = {
            environment.HOST = "0.0.0.0";
            environment.PORT = builtins.toString continuityPort;
            script = "${pkgs.python3}/bin/python3 ${continuityServer}";
          };
        };
      testScript =
        { ... }:
        ''
          peer.start()
          default.start()

          peer.wait_until_succeeds("wg show wiring")
          default.systemctl("--wait start goal1.service")
          # TODO: verify goal1.service is actually done after this systemctl call
          print(default.execute("systemctl status goal1.service")[1])
          print(default.execute("systemctl show goal1.service --property=Result")[1])
          assert "Result=success" in default.execute("systemctl show goal1.service --property=Result")[1]
          print(default.succeed("ip link show"))
          print("default addr", default.succeed("ip addr"))
          print("peer addr", peer.succeed("ip addr"))
          print("peer: ", peer.succeed("wg show wiring"))
          print("default: ", default.succeed("wg show wiring"))
          print("default: ", default.succeed("ping -c 2 10.10.0.0"))
          print("default: ", default.succeed("wg show wiring"))
          print("default: ", default.succeed("ping -c 2 10.10.0.1"))
          print("default: ", default.succeed("wg show wiring"))
          peer.succeed("ping -c 2 10.10.0.0")
          peer.succeed("ping -c 2 10.10.0.1")
          peer.systemctl("start continuityServer.service")
          default.systemctl("start continuityClient.service")
          default.systemctl("--wait start goal2.service")
          print(default.execute("systemctl status goal2.service")[1])
          assert "Result=success" in default.execute("systemctl show goal2.service --property=Result")[1]
          print(default.succeed("ip link show"))
          print("default: ", default.succeed("wg show wiring"))
          peer.fail("ping -c 4 10.10.0.0")
          peer.succeed("ping -c 4 10.10.0.1")
          default.succeed("ping -c 4 10.10.0.0")
          default.fail("ping -c 4 10.10.0.1")
          default.fail("ping -c 4 10.10.0.8")
          print(peer.systemctl("status continuityServer.service")[1])
          print(default.systemctl("stop continuityClient.service")[1])
          print(default.systemctl("status continuityClient.service")[1])
          assert "Result=success" in peer.execute("systemctl show continuityServer.service --property=Result")[1]
          assert "Result=success" in default.execute("systemctl show continuityClient.service --property=Result")[1]
        '';
    }
  );
  coordServerIntegration-single = lib.runTest (
    let
      etc = self.outputs.packages.${system}.etc;
      serverPort = 39390;
      client1PrivateKey = "WMRg3WjzE0ZLuP4bAWtvcrh/Tw23MR3kv4VjpAQQz04=";
      client1PublicKey = "0LrS7ekXRHD8pLEimzLfeLlKyPprJR9oJwdAMOGhtU0=";
      client2PrivateKey = "AFiQ0ipcWrEluvCmaEoQ7PQeurOo3bVRXANAOXYny0s=";
      client2PublicKey = "J4nZeURCVbUmo5w/IBnaCU9M5tOMqGRZnPB2vAji4hE=";
      client1TokenHash = "qrystalcth_30e72874f2598c1ad8020507182a4f57a7304806947b677b69c7d76a88003bc6";
      client1Config = pkgs.writeText "client1config.json" (
        builtins.toJSON {
          MachineJSONPath = "/tmp/machine.json";
          BaseURL = "http://server:${builtins.toString serverPort}";
          Token = "qrystalct_ZXcX7NyjY2aqiy5bb7Oe952QSCsVxzh2FU2ahvaPiHZPJaWeN+Xi59HHvqTDT0Tyy7IOhzC9Uc3Nn7dQ+8GhbQ==";
          Network = "wiring";
          Device = "client1";
          PrivateKey = client1PrivateKey;
        }
      );
      tokens = {
        ${client1TokenHash} = {
          Identities = [
            [
              "wiring"
              "client1"
            ]
          ];
        };
      };
      spec.Networks = [
        {
          Name = "wiring";
          Devices = [
            {
              Name = "client1";
              Endpoints = [ "client1:51820" ];
              Addresses = [ "10.10.0.1/32" ];
              ListenPort = 51820;
              PersistentKeepalive = "1s";
              AccessAll = true;
            }
            {
              Name = "client2";
              Endpoints = [ "client2:51820" ];
              Addresses = [ "10.10.0.2/32" ];
              ListenPort = 51820;
              PublicKey = client2PublicKey;
              AccessAll = true;
            }
          ];
        }
      ];
      serverConfig = pkgs.writeText "serverConfig.json" (
        builtins.toJSON {
          Tokens = tokens;
          Spec = spec;
          Addr = "0.0.0.0:${builtins.toString serverPort}";
        }
      );
    in
    {
      name = "coordServerIntegration-single";
      hostPkgs = pkgs;
      nodes.server =
        { pkgs, ... }:
        {
          # not a wg peer
          imports = [ base ];
          systemd.services.qrystal-coord-server = {
            script = ''
              QRYSTAL_LOGGING_CONFIG=development ${etc}/bin/coord-server --config=${serverConfig}
            '';
            serviceConfig = {
              Type = "notify";
              NotifyAccess = "all";
            };
          };
          networking.firewall.allowedTCPPorts = [ serverPort ];
        };
      nodes.client1 =
        { pkgs, ... }:
        {
          # wg peer setup using test-device
          imports = [ base ];
          environment.systemPackages = [ pkgs.iputils ];
        };
      nodes.client2 =
        { pkgs, ... }:
        {
          # wg peer setup using NixOS's wireguard module
          imports = [ base ];
          networking.wireguard.interfaces = {
            wiring = {
              ips = [ "10.10.0.2/32" ];
              privateKey = client2PrivateKey;
              listenPort = 51820;
              peers = [
                {
                  publicKey = client1PublicKey;
                  allowedIPs = [ "10.10.0.1/32" ];
                  endpoint = "client1:51820";
                  persistentKeepalive = 1;
                }
              ];
            };
          };
        };
      testScript =
        { ... }:
        ''
          start_all()
          server.systemctl("start qrystal-coord-server.service")
          server.wait_until_succeeds("systemctl status qrystal-coord-server.service")
          client1.succeed("echo '{}' > /tmp/machine.json")
          client1.succeed("QRYSTAL_LOGGING_CONFIG=development ${etc}/bin/test-device --config-path ${client1Config}")
          server.succeed("systemctl status qrystal-coord-server.service")
          client1.succeed("ping -c 2 10.10.0.1")
          client1.succeed("ping -c 2 10.10.0.2")
          client2.succeed("ping -c 2 10.10.0.1")
          client2.succeed("ping -c 2 10.10.0.2")
        '';
    }
  );
  coordServerIntegration-double = lib.runTest (
    let
      etc = self.outputs.packages.${system}.etc;
      serverPort = 39390;
      client1PrivateKey = "WMRg3WjzE0ZLuP4bAWtvcrh/Tw23MR3kv4VjpAQQz04=";
      client1PublicKey = "0LrS7ekXRHD8pLEimzLfeLlKyPprJR9oJwdAMOGhtU0=";
      client2PrivateKey = "AFiQ0ipcWrEluvCmaEoQ7PQeurOo3bVRXANAOXYny0s=";
      client2PublicKey = "J4nZeURCVbUmo5w/IBnaCU9M5tOMqGRZnPB2vAji4hE=";
      client1TokenHash = "qrystalcth_30e72874f2598c1ad8020507182a4f57a7304806947b677b69c7d76a88003bc6";
      client1Config = pkgs.writeText "client1config.json" (
        builtins.toJSON {
          MachineJSONPath = "/tmp/machine.json";
          BaseURL = "http://server:${builtins.toString serverPort}";
          Token = "qrystalct_ZXcX7NyjY2aqiy5bb7Oe952QSCsVxzh2FU2ahvaPiHZPJaWeN+Xi59HHvqTDT0Tyy7IOhzC9Uc3Nn7dQ+8GhbQ==";
          Network = "wiring";
          Device = "client1";
          PrivateKey = client1PrivateKey;
        }
      );
      client2TokenHash = "qrystalcth_4f4a908fbcc2f13f45ee71b438efa4df982f99526b085a42fbe05b019056af9e";
      client2Config = pkgs.writeText "client2config.json" (
        builtins.toJSON {
          MachineJSONPath = "/tmp/machine.json";
          BaseURL = "http://server:${builtins.toString serverPort}";
          Token = "qrystalct_wrN7qG37s1KeewvlafCI7GXPC71Jx6DZQAexAJTcfMHRveN7CCMebxo5VIfZxP0YQqSd79rblAgkZkZjJENRMQ==";
          Network = "wiring";
          Device = "client2";
          PrivateKey = client2PrivateKey;
        }
      );
      tokens = {
        ${client1TokenHash} = {
          Identities = [
            [
              "wiring"
              "client1"
            ]
          ];
        };
        ${client2TokenHash} = {
          Identities = [
            [
              "wiring"
              "client2"
            ]
          ];
        };
      };
      spec.Networks = [
        {
          Name = "wiring";
          Devices = [
            {
              Name = "client1";
              Endpoints = [ "client1:51820" ];
              Addresses = [ "10.10.0.1/32" ];
              ListenPort = 51820;
              PersistentKeepalive = "1s";
              AccessAll = true;
            }
            {
              Name = "client2";
              Endpoints = [ "client2:51820" ];
              Addresses = [ "10.10.0.2/32" ];
              ListenPort = 51820;
              PersistentKeepalive = "1s";
              AccessAll = true;
            }
          ];
        }
      ];
      serverConfig = pkgs.writeText "serverConfig.json" (
        builtins.toJSON {
          Tokens = tokens;
          Spec = spec;
          Addr = "0.0.0.0:${builtins.toString serverPort}";
        }
      );
    in
    {
      name = "coordServerIntegration-double";
      hostPkgs = pkgs;
      nodes.server =
        { pkgs, ... }:
        {
          # not a wg peer
          imports = [ base ];
          systemd.services.qrystal-coord-server = {
            script = ''
              QRYSTAL_LOGGING_CONFIG=development ${etc}/bin/coord-server --config=${serverConfig}
            '';
            serviceConfig = {
              Type = "notify";
              NotifyAccess = "all";
            };
          };
          networking.firewall.allowedTCPPorts = [ serverPort ];
        };
      nodes.client1 =
        { pkgs, ... }:
        {
          # wg peer setup using test-device
          imports = [ base ];
          environment.systemPackages = [ pkgs.iputils ];
          networking.firewall.allowedUDPPorts = [ 51820 ];
        };
      nodes.client2 =
        { pkgs, ... }:
        {
          # wg peer setup using test-device
          imports = [ base ];
          networking.firewall.allowedUDPPorts = [ 51820 ];
        };
      testScript =
        { ... }:
        ''
          start_all()
          server.systemctl("start qrystal-coord-server.service")
          server.wait_until_succeeds("systemctl status qrystal-coord-server.service")
          client1.succeed("echo '{}' > /tmp/machine.json")
          client2.succeed("echo '{}' > /tmp/machine.json")
          client1.succeed("QRYSTAL_LOGGING_CONFIG=development ${etc}/bin/test-device --config-path ${client1Config}")
          print(client1.execute("cat /tmp/machine.json")[1])
          print("now client1's PublicKey is filled")
          client2.succeed("QRYSTAL_LOGGING_CONFIG=development ${etc}/bin/test-device --config-path ${client2Config}")
          print("now both PublicKeys are filled, and client2 is fully set up (i.e. is using client1's PublicKey)")
          client1.succeed("QRYSTAL_LOGGING_CONFIG=development ${etc}/bin/test-device --config-path ${client1Config}")
          print(client1.execute("cat /tmp/machine.json")[1])
          server.succeed("systemctl status qrystal-coord-server.service")
          client1.succeed("ping -c 2 10.10.0.1")
          client1.succeed("ping -c 2 10.10.0.2")
          client2.succeed("ping -c 2 10.10.0.1")
          client2.succeed("ping -c 2 10.10.0.2")
        '';
    }
  );
  coordServerIntegration-continuous-single = lib.runTest (
    let
      etc = self.outputs.packages.${system}.etc;
      serverPort = 39390;
      client1PrivateKey = "WMRg3WjzE0ZLuP4bAWtvcrh/Tw23MR3kv4VjpAQQz04=";
      client1PublicKey = "0LrS7ekXRHD8pLEimzLfeLlKyPprJR9oJwdAMOGhtU0=";
      client2PrivateKey = "AFiQ0ipcWrEluvCmaEoQ7PQeurOo3bVRXANAOXYny0s=";
      client2PublicKey = "J4nZeURCVbUmo5w/IBnaCU9M5tOMqGRZnPB2vAji4hE=";
      client1TokenHash = "qrystalcth_30e72874f2598c1ad8020507182a4f57a7304806947b677b69c7d76a88003bc6";
      client1Config = pkgs.writeText "client1config.json" (
        builtins.toJSON {
          Clients = {
            "server" = {
              BaseURL = "http://server:${builtins.toString serverPort}";
              Token = "qrystalct_ZXcX7NyjY2aqiy5bb7Oe952QSCsVxzh2FU2ahvaPiHZPJaWeN+Xi59HHvqTDT0Tyy7IOhzC9Uc3Nn7dQ+8GhbQ==";
              Network = "wiring";
              Device = "client1";
              PrivateKey = client1PrivateKey;
              MinimumInterval = "20s";
            };
          };
        }
      );
      tokens = {
        ${client1TokenHash} = {
          Identities = [
            [
              "wiring"
              "client1"
            ]
          ];
        };
      };
      spec.Networks = [
        {
          Name = "wiring";
          Devices = [
            {
              Name = "client1";
              Endpoints = [ "client1:51820" ];
              Addresses = [ "10.10.0.1/32" ];
              ListenPort = 51820;
              PersistentKeepalive = "1s";
              AccessAll = true;
            }
            {
              Name = "client2";
              Endpoints = [ "client2:51820" ];
              Addresses = [ "10.10.0.2/32" ];
              ListenPort = 51820;
              PublicKey = client2PublicKey;
              AccessAll = true;
            }
          ];
        }
      ];
      serverConfig = pkgs.writeText "serverConfig.json" (
        builtins.toJSON {
          Tokens = tokens;
          Spec = spec;
          Addr = "0.0.0.0:${builtins.toString serverPort}";
        }
      );
    in
    {
      name = "coordServerIntegration-continuous-single";
      hostPkgs = pkgs;
      nodes.server =
        { pkgs, ... }:
        {
          # not a wg peer
          imports = [ base ];
          systemd.services.qrystal-coord-server = {
            script = ''
              QRYSTAL_LOGGING_CONFIG=development ${etc}/bin/coord-server --config=${serverConfig}
            '';
            serviceConfig = {
              Type = "notify";
              NotifyAccess = "all";
            };
          };
          networking.firewall.allowedTCPPorts = [ serverPort ];
        };
      nodes.client1 =
        { pkgs, ... }:
        {
          # wg peer setup using test-device
          imports = [ base ];
          systemd.services.qrystal-device-client = {
            script = ''
              QRYSTAL_LOGGING_CONFIG=development ${etc}/bin/device-client --config=${client1Config}
            '';
            serviceConfig = {
              Type = "notify";
              NotifyAccess = "all";
              StateDirectory = [ "qrystal-device-client" ];
            };
            path = [ pkgs.iputils ];
          };
        };
      nodes.client2 =
        { pkgs, ... }:
        {
          # wg peer setup using NixOS's wireguard module
          imports = [ base ];
          networking.wireguard.interfaces = {
            wiring = {
              ips = [ "10.10.0.2/32" ];
              privateKey = client2PrivateKey;
              listenPort = 51820;
              peers = [
                {
                  publicKey = client1PublicKey;
                  allowedIPs = [ "10.10.0.1/32" ];
                  endpoint = "client1:51820";
                  persistentKeepalive = 1;
                }
              ];
            };
          };
        };
      testScript =
        { ... }:
        ''
          start_all()
          server.systemctl("start qrystal-coord-server.service")
          server.wait_until_succeeds("systemctl status qrystal-coord-server.service")
          client1.systemctl("start qrystal-device-client.service")
          client1.wait_until_succeeds("systemctl status qrystal-device-client.service")
          server.succeed("systemctl status qrystal-coord-server.service")
          client1.wait_until_succeeds("ping -c 2 10.10.0.1")
          client1.wait_until_succeeds("ping -c 2 10.10.0.2")
          client2.succeed("ping -c 2 10.10.0.1")
          client2.succeed("ping -c 2 10.10.0.2")
        '';
    }
  );
  coordServerIntegration-continuous-single-dns-module = lib.runTest (
    let
      module = self.outputs.nixosModules.${system}.default;
      etc = self.outputs.packages.${system}.etc;
      serverPort = 39390;
      client1PrivateKey = "WMRg3WjzE0ZLuP4bAWtvcrh/Tw23MR3kv4VjpAQQz04=";
      client1PublicKey = "0LrS7ekXRHD8pLEimzLfeLlKyPprJR9oJwdAMOGhtU0=";
      client2PrivateKey = "AFiQ0ipcWrEluvCmaEoQ7PQeurOo3bVRXANAOXYny0s=";
      client2PublicKey = "J4nZeURCVbUmo5w/IBnaCU9M5tOMqGRZnPB2vAji4hE=";
      client1TokenHash = "qrystalcth_30e72874f2598c1ad8020507182a4f57a7304806947b677b69c7d76a88003bc6";
      tokens = {
        ${client1TokenHash} = {
          Identities = [
            [
              "wiring"
              "client1"
            ]
          ];
        };
      };
      spec.Networks = [
        {
          Name = "wiring";
          Devices = [
            {
              Name = "client1";
              Endpoints = [ "client1:51820" ];
              Addresses = [ "10.10.0.1/32" ];
              ListenPort = 51820;
              PersistentKeepalive = "1s";
              AccessAll = true;
            }
            {
              Name = "client2";
              Endpoints = [ "client2:51820" ];
              Addresses = [ "10.10.0.2/32" ];
              ListenPort = 51820;
              PublicKey = client2PublicKey;
              AccessAll = true;
            }
          ];
        }
      ];
    in
    {
      name = "coordServerIntegration-continuous-single-dns-module";
      hostPkgs = pkgs;
      nodes.server =
        { pkgs, ... }:
        {
          # not a wg peer
          imports = [ base module ];
          services.qrystal-coord-server = {
            enable = true;
            openFirewall = true;
            bind.address = "0.0.0.0";
            bind.port = serverPort;
            config.Tokens = tokens;
            config.Spec = spec;
          };
        };
      nodes.client1 =
        { pkgs, ... }:        {
          # wg peer setup using test-device
          imports = [ base module ];
          services.qrystal-device-client = {
            enable = true;
            config.Clients.server = {
              BaseURL = "http://server:${builtins.toString serverPort}";
              TokenPath = "${pkgs.writeText "client1Token" "qrystalct_ZXcX7NyjY2aqiy5bb7Oe952QSCsVxzh2FU2ahvaPiHZPJaWeN+Xi59HHvqTDT0Tyy7IOhzC9Uc3Nn7dQ+8GhbQ=="}";
              Network = "wiring";
              Device = "client1";
              PrivateKeyPath = "${pkgs.writeText "client1PublicKey" client1PrivateKey}";
              MinimumInterval = "20s";
            };
            config.dns.enable = true;
            config.dns.Parents = [
              { Suffix = ".qrystal.internal"; }
              {
                Suffix = "client1.nyiyui.ca";
                Network = "wiring";
                Device = "client1";
              }
              {
                Suffix = "client2.nyiyui.ca";
                Network = "wiring";
                Device = "client2";
              }
            ];
            config.dns.Address = "127.0.0.39:53";
          };
        };
      nodes.client2 =
        { pkgs, ... }:
        {
          # wg peer setup using NixOS's wireguard module
          imports = [ base ];
          networking.wireguard.interfaces = {
            wiring = {
              ips = [ "10.10.0.2/32" ];
              privateKey = client2PrivateKey;
              listenPort = 51820;
              peers = [
                {
                  publicKey = client1PublicKey;
                  allowedIPs = [ "10.10.0.1/32" ];
                  endpoint = "client1:51820";
                  persistentKeepalive = 1;
                }
              ];
            };
          };
        };
      testScript =
        { ... }:
        ''
          start_all()
          server.wait_until_succeeds("systemctl start qrystal-coord-server.service")
          server.wait_until_succeeds("systemctl status qrystal-coord-server.service")
          client1.wait_until_succeeds("systemctl start qrystal-device-client.service")
          client1.wait_until_succeeds("systemctl status qrystal-device-client.service")
          server.succeed("systemctl status qrystal-coord-server.service")
          client1.wait_until_succeeds("ping -c 2 10.10.0.1")
          client1.wait_until_succeeds("ping -c 2 10.10.0.2")
          assert "10.10.0.1" in client1.succeed("host client1.wiring.qrystal.internal 127.0.0.39"), "client1.wiring.qrystal.internal"
          assert "10.10.0.1" in client1.succeed("host client1.nyiyui.ca 127.0.0.39"), "client1.nyiyui.ca"
          assert "10.10.0.2" in client1.succeed("host client2.wiring.qrystal.internal 127.0.0.39"), "client2.wiring.qrystal.internal"
          assert "10.10.0.2" in client1.succeed("host client2.nyiyui.ca 127.0.0.39"), "client2.nyiyui.ca"
          client2.succeed("ping -c 2 10.10.0.1")
          client2.succeed("ping -c 2 10.10.0.2")
        '';
    }
  );
  coordServerIntegration-continuous-single-dns = lib.runTest (
    let
      etc = self.outputs.packages.${system}.etc;
      serverPort = 39390;
      client1PrivateKey = "WMRg3WjzE0ZLuP4bAWtvcrh/Tw23MR3kv4VjpAQQz04=";
      client1PublicKey = "0LrS7ekXRHD8pLEimzLfeLlKyPprJR9oJwdAMOGhtU0=";
      client2PrivateKey = "AFiQ0ipcWrEluvCmaEoQ7PQeurOo3bVRXANAOXYny0s=";
      client2PublicKey = "J4nZeURCVbUmo5w/IBnaCU9M5tOMqGRZnPB2vAji4hE=";
      client1TokenHash = "qrystalcth_30e72874f2598c1ad8020507182a4f57a7304806947b677b69c7d76a88003bc6";
      client1Config = pkgs.writeText "client1config.json" (
        builtins.toJSON {
          Clients = {
            "server" = {
              BaseURL = "http://server:${builtins.toString serverPort}";
              Token = "qrystalct_ZXcX7NyjY2aqiy5bb7Oe952QSCsVxzh2FU2ahvaPiHZPJaWeN+Xi59HHvqTDT0Tyy7IOhzC9Uc3Nn7dQ+8GhbQ==";
              Network = "wiring";
              Device = "client1";
              PrivateKey = client1PrivateKey;
              MinimumInterval = "20s";
            };
          };
        }
      );
      client1DNSConfig = pkgs.writeText "client1config.json" (
        builtins.toJSON {
          Parents = [
            { Suffix = ".qrystal.internal"; }
            {
              Suffix = "client1.nyiyui.ca";
              Network = "wiring";
              Device = "client1";
            }
            {
              Suffix = "client2.nyiyui.ca";
              Network = "wiring";
              Device = "client2";
            }
          ];
        }
      );
      tokens = {
        ${client1TokenHash} = {
          Identities = [
            [
              "wiring"
              "client1"
            ]
          ];
        };
      };
      spec.Networks = [
        {
          Name = "wiring";
          Devices = [
            {
              Name = "client1";
              Endpoints = [ "client1:51820" ];
              Addresses = [ "10.10.0.1/32" ];
              ListenPort = 51820;
              PersistentKeepalive = "1s";
              AccessAll = true;
            }
            {
              Name = "client2";
              Endpoints = [ "client2:51820" ];
              Addresses = [ "10.10.0.2/32" ];
              ListenPort = 51820;
              PublicKey = client2PublicKey;
              AccessAll = true;
            }
          ];
        }
      ];
      serverConfig = pkgs.writeText "serverConfig.json" (
        builtins.toJSON {
          Tokens = tokens;
          Spec = spec;
          Addr = "0.0.0.0:${builtins.toString serverPort}";
        }
      );
    in
    {
      name = "coordServerIntegration-continuous-single-dns";
      hostPkgs = pkgs;
      nodes.server =
        { pkgs, ... }:
        {
          # not a wg peer
          imports = [ base ];
          systemd.services.qrystal-coord-server = {
            script = ''
              QRYSTAL_LOGGING_CONFIG=development ${etc}/bin/coord-server --config=${serverConfig}
            '';
            serviceConfig = {
              Type = "notify";
              NotifyAccess = "all";
              PrivateTmp = true;
              NoNewPrivileges = true;
              ProtectSystem = "strict";
              ProtectHome = true;
              ProtectDevices = true;
              ProtectKernelTunables = true;
              ProtectKernelModules = true;
              ProtectControlGroups = true;
              RestrictNamespaces = true;
              PrivateMounts = true;
              DynamicUser = true;
            };
          };
          networking.firewall.allowedTCPPorts = [ serverPort ];
        };
      nodes.client1 =
        { pkgs, ... }:
        let
          dnsRPCSocket = "/run/qrystal-device-dns.sock"; # NOTE that in a production environment, DNS RPC sockets must be in a private directory. (No authn/authz is performed by device-dns.
          dnsAddress = "127.0.0.39:53";
        in
        {
          # wg peer setup using test-device
          imports = [ base ];
          systemd.services.qrystal-device-dns = {
            script = ''
              QRYSTAL_LOGGING_CONFIG=development ${etc}/bin/device-dns --config=${client1DNSConfig} --rpc-listen=${dnsRPCSocket} --dns-listen=${dnsAddress}
            '';
            serviceConfig = {
              Type = "notify";
              NotifyAccess = "all";
              PrivateTmp = true;
              NoNewPrivileges = true;
              ProtectSystem = "strict";
              ProtectHome = true;
              ProtectDevices = true;
              ProtectKernelTunables = true;
              ProtectKernelModules = true;
              ProtectControlGroups = true;
              RestrictNamespaces = true;
              PrivateMounts = true;
            };
          };
          systemd.services.qrystal-device-client = {
            script = ''
              QRYSTAL_LOGGING_CONFIG=development ${etc}/bin/device-client --config=${client1Config} --dns-socket=${dnsRPCSocket}
            '';
            serviceConfig = {
              Type = "notify";
              NotifyAccess = "all";
              StateDirectory = [ "qrystal-device-client" ];
              PrivateTmp = true;
              NoNewPrivileges = true;
              ProtectSystem = "strict";
              ProtectHome = true;
              ProtectDevices = true;
              ProtectKernelTunables = true;
              ProtectKernelModules = true;
              ProtectControlGroups = true;
              RestrictNamespaces = true;
              PrivateMounts = true;
            };
            path = [ pkgs.iputils ];
          };
          networking.firewall.allowedTCPPorts = [ 51820 ];
        };
      nodes.client2 =
        { pkgs, ... }:
        {
          # wg peer setup using NixOS's wireguard module
          imports = [ base ];
          networking.wireguard.interfaces = {
            wiring = {
              ips = [ "10.10.0.2/32" ];
              privateKey = client2PrivateKey;
              listenPort = 51820;
              peers = [
                {
                  publicKey = client1PublicKey;
                  allowedIPs = [ "10.10.0.1/32" ];
                  endpoint = "client1:51820";
                  persistentKeepalive = 1;
                }
              ];
            };
          };
          networking.firewall.allowedTCPPorts = [ 51820 ];
        };
      testScript =
        { ... }:
        ''
          start_all()
          server.systemctl("start qrystal-coord-server.service")
          server.wait_until_succeeds("systemctl status qrystal-coord-server.service")
          client1.systemctl("start qrystal-device-dns.service")
          client1.wait_until_succeeds("systemctl status qrystal-device-dns.service")
          client1.systemctl("start qrystal-device-client.service")
          client1.wait_until_succeeds("systemctl status qrystal-device-client.service")
          server.succeed("systemctl status qrystal-coord-server.service")
          client1.wait_until_succeeds("ping -c 2 10.10.0.1")
          client1.wait_until_succeeds("ping -c 2 10.10.0.2")
          assert "10.10.0.1" in client1.succeed("host client1.wiring.qrystal.internal 127.0.0.39"), "client1.wiring.qrystal.internal"
          assert "10.10.0.1" in client1.succeed("host client1.nyiyui.ca 127.0.0.39"), "client1.nyiyui.ca"
          assert "10.10.0.2" in client1.succeed("host client2.wiring.qrystal.internal 127.0.0.39"), "client2.wiring.qrystal.internal"
          assert "10.10.0.2" in client1.succeed("host client2.nyiyui.ca 127.0.0.39"), "client2.nyiyui.ca"
          client2.succeed("ping -c 2 10.10.0.1")
          client2.succeed("ping -c 2 10.10.0.2")
        '';
    }
  );
  coordServerIntegration-continuous-triple-module-forwarding = lib.runTest (
    let
      module = self.outputs.nixosModules.${system}.default;
      etc = self.outputs.packages.${system}.etc;
      serverPort = 39390;
      client1PrivateKey = "WMRg3WjzE0ZLuP4bAWtvcrh/Tw23MR3kv4VjpAQQz04=";
      client1PublicKey = "0LrS7ekXRHD8pLEimzLfeLlKyPprJR9oJwdAMOGhtU0=";
      client2PrivateKey = "AFiQ0ipcWrEluvCmaEoQ7PQeurOo3bVRXANAOXYny0s=";
      client2PublicKey = "J4nZeURCVbUmo5w/IBnaCU9M5tOMqGRZnPB2vAji4hE=";
      client3PrivateKey = "8C/kJIoIzmJBKhQ4zHzjgVBNtdolVCPbChdb/b8zl3k=";
      client3PublicKey = "puL1ZQzr/Qon7Xm1EULGq3j2lxwo2bn7k7nasuQrLXk=";
      client1TokenHash = "qrystalcth_30e72874f2598c1ad8020507182a4f57a7304806947b677b69c7d76a88003bc6";
      client3Token = "qrystalct_S09ibZAXpMCuTw7ennAvNT4Hk53MdSGl5zgXEN8Nnf6lMq6xMaz5q/PK5yYm5WwXROepQaT1P57qZvqnvWNC8w==";
      client3TokenHash = "qrystalcth_0f50dd6c7ebe4803409f81c0489b01ee1ce5c2abaae4de966731f8cc3717102f";
      tokens = {
        ${client1TokenHash} = {
          Identities = [
            [
              "wiring"
              "client1"
            ]
          ];
        };
        ${client3TokenHash} = { Identities = [ ["wiring" "client3"] ]; };
      };
      spec.Networks = [
        {
          Name = "wiring";
          Devices = [
            {
              Name = "client1";
              Endpoints = [ "client1:51820" ];
              Addresses = [ "10.10.0.1/32" ];
              ListenPort = 51820;
              PersistentKeepalive = "1s";
              ForwardsFor = [ "client2" ];
              AccessAll = true;
            }
            {
              # client2 is manually configured, and purposefully does not have cryptokey routing for client3
              Name = "client2";
              Endpoints = [ "client2:51820" ];
              Addresses = [ "10.10.0.2/32" ];
              ListenPort = 51820;
              PublicKey = client2PublicKey;
              AccessAll = true;
            }
            {
              Name = "client3";
              Endpoints = [ "client3:51820" ];
              Addresses = [ "10.10.0.3/32" ];
              ListenPort = 51820;
              PersistentKeepalive = "1s";
              AccessAll = true;
            }
          ];
        }
      ];
    in
    {
      name = "coordServerIntegration-continuous-triple-module-forwarding";
      hostPkgs = pkgs;
      nodes.server =
        { pkgs, ... }:
        {
          # not a wg peer
          imports = [ base module ];
          services.qrystal-coord-server = {
            enable = true;
            openFirewall = true;
            bind.address = "0.0.0.0";
            bind.port = serverPort;
            config.Tokens = tokens;
            config.Spec = spec;
          };
        };
      nodes.client1 =
        { pkgs, ... }:        {
          # wg peer setup using test-device
          imports = [ base module ];
          services.qrystal-device-client = {
            enable = true;
            config.Clients.server = {
              BaseURL = "http://server:${builtins.toString serverPort}";
              TokenPath = "${pkgs.writeText "client1Token" "qrystalct_ZXcX7NyjY2aqiy5bb7Oe952QSCsVxzh2FU2ahvaPiHZPJaWeN+Xi59HHvqTDT0Tyy7IOhzC9Uc3Nn7dQ+8GhbQ=="}";
              Network = "wiring";
              Device = "client1";
              PrivateKeyPath = "${pkgs.writeText "client1PublicKey" client1PrivateKey}";
              MinimumInterval = "20s";
            };
            config.dns.enable = true;
            config.dns.Address = "127.0.0.39:53";
          };
        };
      nodes.client2 =
        { pkgs, ... }:
        {
          # wg peer setup using NixOS's wireguard module
          imports = [ base ];
          networking.wireguard.interfaces = {
            wiring = {
              ips = [ "10.10.0.2/32" ];
              privateKey = client2PrivateKey;
              listenPort = 51820;
              peers = [
                {
                  publicKey = client1PublicKey;
                  allowedIPs = [ "10.10.0.1/32" ];
                  endpoint = "client1:51820";
                  persistentKeepalive = 1;
                }
              ];
            };
          };
        };
      nodes.client3 =
        { pkgs, ... }:        {
          # wg peer setup using test-device
          imports = [ base module ];
          services.qrystal-device-client = {
            enable = true;
            config.Clients.server = {
              BaseURL = "http://server:${builtins.toString serverPort}";
              TokenPath = "${pkgs.writeText "client3Token" client3Token}";
              Network = "wiring";
              Device = "client3";
              PrivateKeyPath = "${pkgs.writeText "client3PublicKey" client3PrivateKey}";
              MinimumInterval = "20s";
            };
            config.dns.enable = true;
            config.dns.Address = "127.0.0.39:53";
          };
        };
      testScript =
        { ... }:
        ''
          start_all()
          client3.wait_until_succeeds("ping -c 1 client2")
          client3_ip = client3.execute("ip -4 addr show eth1 | grep -oP '(?<=inet\s)\d+(\.\d+){3}'")[1].strip()
          #client2.succeed(f"iptables -A INPUT -s {client3_ip} -p icmp --icmp-type echo-request -j DROP")
          client2.succeed(f"iptables -A INPUT -s {client3_ip} -i eth1 -j DROP")
          client3.fail("ping -c 1 client2")
          raise RuntimeError("test is not implemented yet")
          #server.wait_until_succeeds("systemctl start qrystal-coord-server.service")
          #server.wait_until_succeeds("systemctl status qrystal-coord-server.service")
          #client1.wait_until_succeeds("systemctl start qrystal-device-client.service")
          #client1.wait_until_succeeds("systemctl status qrystal-device-client.service")
          #server.succeed("systemctl status qrystal-coord-server.service")
          #client1.wait_until_succeeds("ping -c 2 10.10.0.1")
          #client1.wait_until_succeeds("ping -c 2 10.10.0.2")
          #assert "10.10.0.1" in client1.succeed("host client1.wiring.qrystal.internal 127.0.0.39"), "client1.wiring.qrystal.internal"
          #assert "10.10.0.1" in client1.succeed("host client1.nyiyui.ca 127.0.0.39"), "client1.nyiyui.ca"
          #assert "10.10.0.2" in client1.succeed("host client2.wiring.qrystal.internal 127.0.0.39"), "client2.wiring.qrystal.internal"
          #assert "10.10.0.2" in client1.succeed("host client2.nyiyui.ca 127.0.0.39"), "client2.nyiyui.ca"
          #client2.succeed("ping -c 2 10.10.0.1")
          #client2.succeed("ping -c 2 10.10.0.2")
        '';
    }
  );
}
