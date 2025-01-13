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
            QRYSTAL_LOGGING_CONFIG=development ${etc}/bin/test-goal -m-path ${machine1}
          '';
          systemd.services.goal1.serviceConfig.Type = "oneshot";
          systemd.services.goal1.path = [ pkgs.iputils ];
          systemd.services.goal2.script = ''
            QRYSTAL_LOGGING_CONFIG=development ${etc}/bin/test-goal -m-path ${machine2}
          '';
          systemd.services.goal2.serviceConfig.Type = "oneshot";
          systemd.services.goal2.path = [ pkgs.iputils ];
          systemd.services.goal3.script = ''
            QRYSTAL_LOGGING_CONFIG=development ${etc}/bin/test-goal -m-path ${machine3}
          '';
          systemd.services.goal3.serviceConfig.Type = "oneshot";
          systemd.services.goal3.path = [ pkgs.iputils ];
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
          # sanity check
          default.systemctl("--wait start goal1.service")
          assert "Result=success" in default.execute("systemctl show goal1.service --property=Result")[1]

          # run goal1
          default.systemctl("--wait start goal2.service")
          assert "Result=success" in default.execute("systemctl show goal2.service --property=Result")[1]
          print(default.succeed("ip link show"))
          print("default addr", default.succeed("ip addr"))
          print("peer addr", peer.succeed("ip addr"))
          print("peer: ", peer.succeed("wg show wiring"))
          print("default: ", default.succeed("wg show wiring"))
          print("default: ", default.wait_until_succeeds("ping -c 2 10.10.0.0"))
          print("default: ", default.wait_until_succeeds("ping -c 2 10.10.0.1"))
          peer.wait_until_succeeds("ping -c 2 10.10.0.0")
          peer.wait_until_succeeds("ping -c 2 10.10.0.1")

          peer.systemctl("start continuityServer.service")
          default.systemctl("start continuityClient.service")
          default.systemctl("--wait start goal3.service")
          assert "Result=success" in default.execute("systemctl show goal3.service --property=Result")[1]
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
}
