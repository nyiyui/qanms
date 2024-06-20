args@{ self, system, nixpkgsFor, libFor, nixosLibFor, ldflags, ... }:
let
  pkgs = nixpkgsFor.${system};
  lib = nixosLibFor.${system} { inherit system; };
  node1Token =
    "qrystalct_/TTOsqg6hUeuODtIUj1z4aXDiU1ckks9T7/Eqod2mVrsgFC8eFdlS4fZXLBwggKO1MvI6oqoAWkiMZbHjLdP/w==";
  node1Hash =
    "qrystalcth_a2f29c49f4e3e520413f71ac2b42b5b66c0b9cc70bd757a543754d83e94ccfd8";
  node2Token =
    "qrystalct_jv4Abw0LouLeiq8GStjOsacArU56b77yyJ/XM0Nij/AoeSU7nlBFBFY87g05KCiuanyCdehtXZYg3MLxeFTI7Q==";
  node2Hash =
    "qrystalcth_75b2eb7d0cac7a796362115b5b0f267ee08eff7a87012fd4334082bba141c018";
  rootCert = builtins.readFile ./cert/minica.pem;
  rootKey = builtins.readFile ./cert/minica-key.pem;
  csCert = builtins.readFile ./cert/cs/cert.pem;
  csKey = builtins.readFile ./cert/cs/key.pem;

  autologin = { ... }: { services.getty.autologinUser = "root"; };
  base = { ... }: {
    imports = [ autologin ];
    virtualisation.vlans = [ 1 ];
    environment.systemPackages = with pkgs; [ wireguard-tools ];
    services.logrotate.enable = false; # clogs up the logs
  };
  networkBase = {
    keepalive = "10s";
    listenPort = 58120;
    ips = [ "10.123.0.1/16" ];
  };
  nodeToken = name: hash: networkNames: {
    inherit name;
    inherit hash;
    canPull = true;
    networks = builtins.foldl' (a: b: a // b) { }
      (map (networkName: { ${networkName} = name; }) networkNames);
  };
  adminTokenRaw =
    "qrystalct_0a3XVoDo0Q4Ni4b47tqSURZACuoqG0A79+LmfvkZQZsMco5P+OL/L6cbnPCKDe12Fj2kUkHWpHhw6eRypRgr8Q==";
  adminToken = {
    name = "admin";
    hash =
      "qrystalcth_98e2781b6a908f179e6df385b096decf5abde8ff8655dd30b5e55c7c4d81bb90";
    networks = null;
    canPull = true;
    canPush.any = true;
    canAdminTokens = {
      canPull = true;
      canPush = true;
    };
  };
  csTls = {
    certPath = builtins.toFile "testing-insecure-cert.pem" csCert;
    keyPath = builtins.toFile "testing-insecure-key.pem" csKey;
  };
in let
  csConfig = networkNames: token: {
    enable = true;
    config.cs = {
      comment = "cs";
      endpoint = "cs:39252";
      tls.certPath = builtins.toFile "testing-insecure-node-cert.pem"
        (rootCert + "\n" + csCert);
      networks = networkNames;
      tokenPath = builtins.toFile "token" token;
    };
  };
in {
  sd-notify-baseline = lib.runTest ({
    name = "sd-notify-baseline";
    hostPkgs = pkgs;
    nodes.machine = { pkgs, ... }: {
      systemd.services.sd-notify-test = {
        serviceConfig = {
          Type = "notify";
          ExecStart =
            "${pkgs.bash}/bin/bash -c '${pkgs.coreutils}/bin/echo notifying; ${pkgs.systemd}/bin/systemd-notify --ready & ${pkgs.coreutils}/bin/echo notified; while true; do sleep 1; done'";
        };
      };
    };
    testScript = ''
      machine.start()
      machine.systemctl("start sd-notify-test.service")
      machine.wait_for_unit("sd-notify-test.service")
    '';
  });
  sd-notify = lib.runTest ({
    name = "sd-notify";
    hostPkgs = pkgs;
    nodes.machine = { pkgs, ... }: {
      systemd.services.sd-notify-test = {
        serviceConfig = {
          Type = "notify";
          ExecStart = "${
              self.outputs.packages.${system}.sd-notify-test
            }/bin/sd-notify-test";
        };
      };
    };
    testScript = ''
      machine.start()
      machine.systemctl("start sd-notify-test.service")
      machine.wait_for_unit("sd-notify-test.service")
    '';
  });
  cs = lib.runTest {
    name = "cs";
    hostPkgs = pkgs;
    nodes = {
      cs = { pkgs, ... }: {
        imports = [ self.outputs.nixosModules.${system}.cs ];

        environment.systemPackages = with pkgs;
          [ self.outputs.packages.${system}.etc ];

        qrystal.services.cs = {
          enable = true;
          config = {
            tls = csTls;
            tokens = [
              (nodeToken "node1" node1Hash [ "testnet" ])
              (nodeToken "node2" node2Hash [ "testnet" ])
              adminToken
            ];
            central.networks.testnet = networkBase // {
              peers.node1 = { allowedIPs = [ "10.123.0.1/16" ]; };
            };
          };
        };
      };
    };
    testScript = { nodes, ... }: ''
      import json

      cs.start()
      cs.wait_for_unit("qrystal-cs.service")
      cs.succeed("cs-admin -server 'cs:39253' -token '${adminTokenRaw}' -cert '${
        builtins.toFile "testing-insecure-cert.pem" csCert
      }' token-rm -token-hash '${node1Hash}'")
      q = json.dumps(dict(
        overwrite=True,
        name='node1new',
        hash='${node1Hash}',
        canPull=dict(
          testnet='node1',
        ),
      ))
      cs.succeed(f"echo '{q}' | cs-admin -server 'cs:39253' -token '${adminTokenRaw}' -cert '${
        builtins.toFile "testing-insecure-cert.pem" csCert
      }' token-add")
      # TODO test this actually works
    '';
  };
  all = let
    networkName = "testnet";
    networkName2 = "othernet";
    testDomain = "cs";
  in let
    node = { token }:
      { pkgs, ... }: {
        imports = [
          base
          self.outputs.nixosModules.${system}.node
          {
            qrystal.services.node.config.srvList = pkgs.writeText "srvlist.json" (builtins.toJSON {
              ${networkName} = [{
                Service = "_testservice";
                Protocol = "_tcp";
                Priority = "10";
                Weight = "10";
                Port = "123";
              }];
            });
          }
        ];

        networking.firewall.allowedTCPPorts = [ 39251 ];
        qrystal.services.node = csConfig [ networkName networkName2 ] token;
        systemd.services.qrystal-node.wantedBy = [ ];
      };
  in lib.runTest ({
    name = "all";
    hostPkgs = pkgs;
    nodes = {
      node1 = node { token = node1Token; };
      node2 = node { token = node2Token; };
      cs = { pkgs, ... }: {
        imports = [ base self.outputs.nixosModules.${system}.cs ];

        networking.firewall.allowedTCPPorts = [ 39252 ];
        qrystal.services.cs = {
          enable = true;
          config = {
            tls = csTls;
            tokens = [
              (nodeToken "node1" node1Hash [ networkName networkName2 ])
              (nodeToken "node2" node2Hash [ networkName networkName2 ])
            ];
            central.networks.${networkName} = networkBase // {
              peers.node1 = {
                host = "node1:58120";
                allowedIPs = [ "10.123.0.1/32" ];
                canSee.only = [ "node2" ];
              };
              peers.node2 = {
                host = "node2:58120";
                allowedIPs = [ "10.123.0.2/32" ];
                canSee.only = [ "node1" ];
              };
            };
            central.networks.${networkName2} = networkBase // {
              keepalive = "10s";
              listenPort = 58121;
              ips = [ "10.45.0.1/16" ];
              peers.node1 = {
                host = "node1:58121";
                allowedIPs = [ "10.45.0.1/32" ];
                canSee.only = [ "node2" ];
              };
              peers.node2 = {
                host = "node2:58121";
                allowedIPs = [ "10.45.0.2/32" ];
                canSee.only = [ "node1" ];
              };
            };
          };
        };
      };
    };
    testScript = { nodes, ... }: ''
      nodes = [node1, node2]
      addrs = ["10.123.0.2", "10.123.0.1"]
      cs.start()
      cs.wait_for_unit("qrystal-cs.service")
      for node in nodes:
        node.start()
        node.wait_until_succeeds("host ${testDomain}") # test dnsmasq settings work
        node.systemctl("start qrystal-node.service")
        node.wait_for_unit("qrystal-node.service", timeout=20)
      print("all nodes started")
      # NOTE: there is a race condition where the peers' pubkeys could not be
      # set yet when pinged (so that's why we're using wait_until_*
      for i, node in enumerate(nodes):
        print(node.wait_until_succeeds("wg show"))
        print(node.wait_until_succeeds("wg show ${networkName}"))
        print(node.wait_until_succeeds("wg show ${networkName2}"))
        print(node.execute("cat /etc/wireguard/${networkName}.conf")[1])
        print(node.execute("ip route show")[1])
        for addr in addrs:
          print(node.execute(f"ip route get {addr}")[1])
      for i, node in enumerate(nodes):
        print(node.execute(f"ping -c 1 {addrs[i]}")[1])
        node.wait_until_succeeds(f"ping -c 1 {addrs[i]}")
      def pp(value):
        print("pp", value)
        return value
      assert "node2.testnet.qrystal.internal has address 10.123.0.2" in pp(node1.succeed("host node2.testnet.qrystal.internal"))
      assert "node1.testnet.qrystal.internal has address 10.123.0.1" in pp(node2.succeed("host node1.testnet.qrystal.internal"))
      for node in nodes:
        assert pp(node.execute("host idkpeer.testnet.qrystal.internal 127.0.0.39"))[0] == 1
        assert pp(node.execute("host node1.idknet.qrystal.internal 127.0.0.39"))[0] == 1
        assert 'has SRV record' not in pp(node.execute("host _testservice._tcp.idkpeer.testnet.qrystal.internal 127.0.0.39"))[1]
      # TODO: test network level queries
    '';
  });
  azusa = let
    networkName = "testnet";
    testDomain = "cs";
    node = { token, name, allowedIPs, canSee }:
      { pkgs, ... }: {
        imports = [
          base
          self.outputs.nixosModules.${system}.node
          {
            qrystal.services.node.config.srvList = pkgs.writeText "srvlist.json" (builtins.toJSON {
              ${networkName} = [{
                Service = "_testservice";
                Protocol = "_tcp";
                Priority = "10";
                Weight = "10";
                Port = "123";
              }];
            });
          }
          {
            qrystal.services.node.config.cs.azusa.networks.${networkName} = {
              inherit name;
              host = "${name}:58120";
              inherit allowedIPs;
              inherit canSee;
            };
          }
        ];

        networking.firewall.allowedTCPPorts = [ 39251 ];
        qrystal.services.node = csConfig [ networkName ] token;
        systemd.services.qrystal-node.wantedBy = [ ];
      };
  in lib.runTest ({
    name = "azusa";
    hostPkgs = pkgs;
    nodes = {
      node1 = node { token = node1Token; name = "node1"; allowedIPs = [ "10.123.0.1/32" ]; canSee.only = [ "node2" ]; };
      node2 = node { token = node2Token; name = "node2"; allowedIPs = [ "10.123.0.2/32" ]; canSee.only = [ "node1" ]; };
      cs = { pkgs, ... }: {
        imports = [ base self.outputs.nixosModules.${system}.cs ];

        networking.firewall.allowedTCPPorts = [ 39252 ];
        qrystal.services.cs = {
          enable = true;
          config = {
            tls = csTls;
            tokens = [
              ((nodeToken "node1" node1Hash [ networkName ]) // {
                canPush.networks.${networkName} = {
                  name = "node1";
                  canSeeElement = [ "node2" ];
                };
                canPull = true;
                networks.${networkName} = "node1";
              })
              ((nodeToken "node2" node2Hash [ networkName ]) // {
                canPush.networks.${networkName} = {
                  name = "node2";
                  canSeeElement = [ "node1" ];
                };
                canPull = true;
                networks.${networkName} = "node2";
              })
            ];
            central.networks.${networkName} = networkBase;
          };
        };
      };
    };
    testScript = { nodes, ... }: ''
      nodes = [node1, node2]
      addrs = ["10.123.0.2", "10.123.0.1"]
      cs.start()
      cs.wait_for_unit("qrystal-cs.service")
      for node in nodes:
        node.start()
        node.wait_until_succeeds("host ${testDomain}") # test dnsmasq settings work
        node.systemctl("start qrystal-node.service")
        node.wait_for_unit("qrystal-node.service", timeout=20)
      print("all nodes started")
      # NOTE: there is a race condition where the peers' pubkeys could not be
      # set yet when pinged (so that's why we're using wait_until_*
      for i, node in enumerate(nodes):
        print(node.wait_until_succeeds("wg show"))
        print(node.wait_until_succeeds("wg show ${networkName}"))
        print(node.execute("cat /etc/wireguard/${networkName}.conf")[1])
        print(node.execute("ip route show")[1])
        for addr in addrs:
          print(node.execute(f"ip route get {addr}")[1])
      for i, node in enumerate(nodes):
        print(node.execute(f"ping -c 1 {addrs[i]}")[1])
        node.wait_until_succeeds(f"ping -c 1 {addrs[i]}")
      def pp(value):
        print("pp", value)
        return value
      assert "node2.testnet.qrystal.internal has address 10.123.0.2" in pp(node1.succeed("host node2.testnet.qrystal.internal"))
      assert "node1.testnet.qrystal.internal has address 10.123.0.1" in pp(node2.succeed("host node1.testnet.qrystal.internal"))
      for node in nodes:
        assert pp(node.execute("host idkpeer.testnet.qrystal.internal 127.0.0.39"))[0] == 1
        assert pp(node.execute("host node1.idknet.qrystal.internal 127.0.0.39"))[0] == 1
        assert 'has SRV record' not in pp(node.execute("host _testservice._tcp.idkpeer.testnet.qrystal.internal 127.0.0.39"))[1]
      # TODO: test network level queries
    '';
  });
  eo = let
    networkName = "testnet";
    eoPath = pkgs.writeShellScript "eo.sh" ''
      echo '{"endpoint":"1.2.3.4:5678"}'
    '';
    node = { token }:
      { pkgs, ... }: {
        imports = [
          base
          self.outputs.nixosModules.${system}.node
          ({ ... }: { qrystal.services.node.config.endpointOverride = eoPath; })
        ];

        networking.firewall.allowedTCPPorts = [ 39251 ];
        qrystal.services.node = csConfig [ networkName ] token;
      };
  in lib.runTest ({
    name = "eo";
    hostPkgs = pkgs;
    nodes = {
      node1 = node { token = node1Token; };
      node2 = node { token = node2Token; };
      cs = { pkgs, ... }: {
        imports = [ base self.outputs.nixosModules.${system}.cs ];

        networking.firewall.allowedTCPPorts = [ 39252 ];
        qrystal.services.cs = {
          enable = true;
          config = {
            tls = csTls;
            tokens = [
              (nodeToken "node1" node1Hash [ networkName ])
              (nodeToken "node2" node2Hash [ networkName ])
            ];
            central.networks.${networkName} = networkBase // {
              peers.node1 = {
                host = "node1:58120";
                allowedIPs = [ "10.123.0.1/32" ];
                canSee.only = [ "node2" ];
              };
              peers.node2 = {
                host = "node2:58120";
                allowedIPs = [ "10.123.0.2/32" ];
                canSee.only = [ "node1" ];
              };
            };
          };
        };
      };
    };
    testScript = { nodes, ... }: ''
      import time

      nodes = [node1, node2]
      start_all()
      for node in nodes:
        node.wait_for_unit("qrystal-node.service")
      for node in nodes:
        start = time.time()
        while True:
          now = time.time()
          if now-start > 10:
            raise RuntimeError("timeout")
          print(node.wait_until_succeeds("wg show ${networkName}"))
          endpoints = node.wait_until_succeeds("wg show ${networkName} endpoints")
          print('endpoints', endpoints)
          if ':' not in endpoints:
            # wait until sync is done
            time.sleep(1)
            continue
          assert "1.2.3.4:5678" in endpoints, "endpoint check"
          break
    '';
  });
  trace = let
    networkName = "testnet";
    tracePath = "/etc/qrystal-trace";
    node = { token }:
      { pkgs, ... }: {
        imports = [
          base
          self.outputs.nixosModules.${system}.node
          ({ ... }: { 
            environment.etc."qrystal-trace" = {
              text = "not yet";
              mode = "0666";
            };
            qrystal.services.node.config.trace = {
              outputPath = tracePath;
              waitUntilCNs = [ networkName ];
            };
            environment.systemPackages = with pkgs; [ go ];
          })
        ];

        networking.firewall.allowedTCPPorts = [ 39251 ];
        qrystal.services.node = csConfig [ networkName ] token;
      };
  in lib.runTest ({
    name = "trace";
    hostPkgs = pkgs;
    nodes = {
      node1 = node { token = node1Token; };
      cs = { pkgs, ... }: {
        imports = [ base self.outputs.nixosModules.${system}.cs ];

        networking.firewall.allowedTCPPorts = [ 39252 ];
        qrystal.services.cs = {
          enable = true;
          config = {
            tls = csTls;
            tokens = [
              (nodeToken "node1" node1Hash [ networkName ])
            ];
            central.networks.${networkName} = networkBase // {
              peers.node1 = {
                host = "node1:58120";
                allowedIPs = [ "10.123.0.1/32" ];
                canSee.only = [ "node1" ];
              };
            };
          };
        };
      };
    };
    testScript = { nodes, ... }: ''
      import time

      cs.start()
      cs.wait_for_unit("qrystal-cs.service")
      node1.start()
      node1.wait_for_unit("qrystal-node.service", timeout=20)
      start = time.time()
      node1.wait_until_succeeds("cat ${tracePath}", timeout=20)
      # verify trace is valid (e.g. if it finished writing correctly)
      node1.succeed("go tool trace -pprof=net ${tracePath}")
      # pprof type doesn't matter
    '';
  });
  node-backport = let
    networkName = "testnet";
    networkName2 = "othernet";
  in let
    node = { token }:
      { pkgs, ... }: {
        imports = [
          base
          self.outputs.nixosModules.${system}.node
        ];

        networking.firewall.allowedTCPPorts = [ 39251 ];
        qrystal.services.node = csConfig [ networkName networkName2 ] token;
        systemd.services.qrystal-node.wantedBy = [ ];
      };
  in lib.runTest ({
    name = "node-backport";
    hostPkgs = pkgs;
    nodes = {
      node1 = node { token = node1Token; };
      node2 = node { token = node2Token; };
      cs = { pkgs, ... }: {
        imports = [ base self.outputs.nixosModules.${system}.cs ];

        networking.firewall.allowedTCPPorts = [ 39252 ];
        qrystal.services.cs = {
          enable = true;
          config = {
            tls = csTls;
            tokens = [
              (nodeToken "node1" node1Hash [ networkName networkName2 ])
              (nodeToken "node2" node2Hash [ networkName networkName2 ])
            ];
            central.networks.${networkName} = networkBase // {
              peers.node1 = {
                host = "node1:58120";
                allowedIPs = [ "10.123.0.1/32" ];
                canSee.only = [ "node2" ];
              };
              peers.node2 = {
                host = "node2:58120";
                allowedIPs = [ "10.123.0.2/32" ];
                canSee.only = [ "node1" ];
              };
            };
            central.networks.${networkName2} = networkBase // {
              keepalive = "10s";
              listenPort = 58121;
              ips = [ "10.45.0.1/16" ];
              peers.node1 = {
                host = "node1:58121";
                allowedIPs = [ "10.45.0.1/32" ];
                canSee.only = [ "node2" ];
              };
              peers.node2 = {
                host = "node2:58121";
                allowedIPs = [ "10.45.0.2/32" ];
                canSee.only = [ "node1" ];
              };
            };
          };
        };
      };
    };
    testScript = { nodes, ... }: ''
      nodes = [node1, node2]
      addrs = ["10.123.0.2", "10.123.0.1"]
      cs.start()
      cs.wait_for_unit("qrystal-cs.service")
      for node in nodes:
        node.start()
        node.systemctl("start qrystal-node.service")
        node.wait_for_unit("qrystal-node.service", timeout=20)
      print("all nodes started")
      # NOTE: there is a race condition where the peers' pubkeys could not be
      # set yet when pinged (so that's why we're using wait_until_*
      for i, node in enumerate(nodes):
        print(node.wait_until_succeeds("wg show"))
        print(node.wait_until_succeeds("wg show ${networkName}"))
        print(node.wait_until_succeeds("wg show ${networkName2}"))
        print(node.execute("cat /etc/wireguard/${networkName}.conf")[1])
        print(node.execute("ip route show")[1])
        for addr in addrs:
          print(node.execute(f"ip route get {addr}")[1])
      for i, node in enumerate(nodes):
        print(node.execute(f"ping -c 1 {addrs[i]}")[1])
        node.wait_until_succeeds(f"ping -c 1 {addrs[i]}")
      cs.crash() # bye bye
      # 1st, nodes must survive CS crashing
      for i, node in enumerate(nodes):
        print(node.execute(f"ping -c 1 {addrs[i]}")[1])
        node.wait_until_succeeds(f"ping -c 1 {addrs[i]}")
      # 2nd, nodes must survive CS crashing + restart
      for i, node in enumerate(nodes):
        node.systemctl("restart qrystal-node.service")
      for i, node in enumerate(nodes):
        print(node.execute(f"ping -c 1 {addrs[i]}")[1])
        node.wait_until_succeeds(f"ping -c 1 {addrs[i]}")
    '';
  });
  eo-udptunnel = let
    networkName = "testnet";
    testDomain = "cs";
    tunserverPort = 12345;
  in let
    node = { token }:
      { pkgs, ... }: {
        imports = [
          base
          self.outputs.nixosModules.${system}.node
          {
            qrystal.services.node.config.udptunnel = {
              enable = true;
              servers = {
                "${networkName}"."node1" = "tunserver:${toString tunserverPort}";
              };
            };
          }
        ];

        networking.firewall.allowedTCPPorts = [ 39251 ];
        qrystal.services.node = csConfig [ networkName ] token;
      };
  in lib.runTest ({
    name = "eo-udptunnel";
    hostPkgs = pkgs;
    nodes = {
      tunserver = { pkgs, ... }: {
        imports = [ base self.outputs.nixosModules.${system}.udptunnel-server ];

        networking.firewall.allowedUDPPorts = [ tunserverPort ];
        qrystal.services.udptunnel-server = {
          enable = true;
          listen = "0.0.0.0:${toString tunserverPort}";
          destination = "node1:58120";
        };
      };
      node1 = node { token = node1Token; };
      node2 = node { token = node2Token; };
      cs = { pkgs, ... }: {
        imports = [ base self.outputs.nixosModules.${system}.cs ];

        networking.firewall.allowedTCPPorts = [ 39252 ];
        qrystal.services.cs = {
          enable = true;
          config = {
            tls = csTls;
            tokens = [
              (nodeToken "node1" node1Hash [ networkName ])
              (nodeToken "node2" node2Hash [ networkName ])
            ];
            central.networks.${networkName} = networkBase // {
              peers.node1 = {
                host = "node1:58120";
                allowedIPs = [ "10.123.0.1/32" ];
                canSee.only = [ "node2" ];
              };
              peers.node2 = {
                host = "node2:58120";
                allowedIPs = [ "10.123.0.2/32" ];
                canSee.only = [ "node1" ];
              };
            };
          };
        };
      };
    };
    testScript = { nodes, ... }: ''
      def pp(value):
        print("pp", value)
        return value

      nodes = [node1, node2]
      addrs = ["10.123.0.2", "10.123.0.1"]
      start_all()
      cs.wait_for_unit("qrystal-cs.service")
      for node in nodes:
        node.wait_until_succeeds("host ${testDomain}")
        node.wait_for_unit("qrystal-node.service", timeout=20)
      print("all nodes started")
      # NOTE: there is a race condition where the peers' pubkeys could not be
      # set yet when pinged (so that's why we're using wait_until_*
      for i, node in enumerate(nodes):
        print(node.wait_until_succeeds("wg show"))
        print(node.wait_until_succeeds("wg show ${networkName}"))
        print(node.execute("cat /etc/wireguard/${networkName}.conf")[1])
        print(node.execute("ip route show")[1])
        for addr in addrs:
          print(node.execute(f"ip route get {addr}")[1])
      assert ":${toString tunserverPort}" in pp(node2.succeed("wg show ${networkName}"))
      for i, node in enumerate(nodes):
        for j in range(len(nodes)):
          print(f'pinging {j} from {i}')
          node.wait_until_succeeds(f"ping -c 1 {addrs[j]}")
    '';
  });
  udptunnel = lib.runTest ({
    name = "udptunnel";
    hostPkgs = pkgs;
    nodes = {
      client = { pkgs, ... }: let
        portal = { host = "127.0.0.1"; port = "12345"; };
      in {
        imports = [ base self.outputs.nixosModules.${system}.udptunnel-client ];
        qrystal.services.udptunnel-client = {
          enable = true;
          portal = "${portal.host}:${portal.port}";
          server = "server:443";
          verbose = true;
        };
        systemd.services.udp-send = let
          pythonScript = pkgs.writeText "udp-send.py" ''
            print('starting...')
            import itertools, socket, time

            s = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
            for i in itertools.count():
                print(f'sending {i}...')
                s.sendto(f"ping {i}".encode("ascii"), ("${portal.host}", ${portal.port}))
                time.sleep(1)
          '';
        in {
          script = ''
            ${pkgs.python3}/bin/python3 -Wd ${pythonScript}
          '';
        };
      };
      server = { pkgs, ... }: let
        destination = { host = "127.0.0.1"; port = "12345"; };
      in {
        imports = [ base self.outputs.nixosModules.${system}.udptunnel-server ];
        qrystal.services.udptunnel-server = {
          enable = true;
          listen = "0.0.0.0:443";
          destination = "${destination.host}:${destination.port}";
          verbose = true;
        };
        systemd.services.udp-receive = let
          pythonScript = pkgs.writeText "udp-receive.py" ''
            print('starting...')
            import itertools, socket, sys

            s = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
            s.bind(("${destination.host}", ${destination.port}))
            while 1:
                data, source = s.recvfrom(1024)
                print(f'received: {data} from {source}')
                if source != "${destination.host}": continue
                if "ping" not in data: continue
                print('correct message received')
                break
          '';
        in {
          serviceConfig.Type = "oneshot";
          script = ''
            ${pkgs.python3}/bin/python3 -Wd ${pythonScript}
          '';
        };
      };
    };
    testScript = { nodes, ... }: ''
      server.start()
      client.start()

      client.systemctl("start udp-send.service")
      client.wait_for_unit("udp-send.service")

      server.systemctl("start udp-receive.service")
      server.wait_for_unit("udp-receive.service")
    '';
  });
  goal = lib.runTest (let
    peerPrivateKey = "kCtV08G5gyM/cGHToObIAtwRq/bqI2Jd3akIsAMXRXM=";
    peerPublicKey = "72zpXYpjSWnvyhwZTuRNwtghjxjzhWEVzUNRA82hoUA=";
    defaultPrivateKey = "eDq8aX08rF5cLG+NNi14Ae8TIudsMHiWCjsbBTDI1Ec=";
    defaultPublicKey = "+atCYz0YmiwBx4AZy5kDGr5WHqHs3RMbIuPfj593sRk=";
    etc = self.outputs.packages.${system}.etc;
    machine1 = pkgs.writeText "machine1.json" (builtins.toJSON {
    });
    machine2 = pkgs.writeText "machine2.json" (builtins.toJSON {
      Interfaces = [
        {
          Name = "wiring";
          PrivateKey = defaultPrivateKey;
          ListenPort = 51820;
          Addresses = [ "10.10.0.0/32" ];
          Peers = [
            { Name = "peer"; PublicKey = peerPublicKey; Endpoint = "peer:51820"; AllowedIPs = [ "10.10.0.1/32" ]; PersistentKeepalive = 30; }
          ];
        }
      ];
    });
    machine3 = pkgs.writeText "machine3.json" (builtins.toJSON {
      Interfaces = [
        {
          Name = "wiring";
          PrivateKey = defaultPrivateKey;
          ListenPort = 51820;
          Addresses = [ "10.10.0.0/32" ];
          Peers = [
            { Name = "peer"; PublicKey = peerPublicKey; Endpoint = "peer:51820"; AllowedIPs = [ "10.10.0.8/32" ]; }
          ];
        }
      ];
    });
    continuityPort = 51821;
    continuityServer = pkgs.writeText "continuityServer.py" ''
import socketserver
import os

class MyTCPHandler(socketserver.BaseRequestHandler):
    counter = 0
    def handle(self):
        while True:
            print("waiting...")
            self.data = self.request.recv(1024).strip()
            print(f"received from {self.client_address[0]}: {self.data}")
            i = int(self.data)
            if i != counter + 1:
                print("out of order")
                exit(1)
            self.request.sendall("ok\n".encode("ascii"))
            print("sent.")
    def finish(self):
        print("done")
        exit(0)

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
  in {
    name = "goal";
    hostPkgs = pkgs;
    nodes.default = { pkgs, ... }: {
      imports = [ base ];
      environment.systemPackages = [ self.outputs.packages.${system}.etc ];
      systemd.services.goal1.script = ''
        QRYSTAL_LOGGING_CONFIG=development ${etc}/bin/test-goal -a-path ${machine1} -b-path ${machine2}
      '';
      systemd.services.goal2.script = ''
        QRYSTAL_LOGGING_CONFIG=development ${etc}/bin/test-goal -a-path ${machine2} -b-path ${machine3}
      '';
      systemd.services."continuityClient" = {
        environment.HOST = "peer";
        environment.PORT = builtins.toString continuityPort;
        script = "${pkgs.python3}/bin/python3 ${continuityClient}";
      };
    };
    nodes.peer = { pkgs, ... }: {
      imports = [ base ];
      #TODO: wireguard config
      networking.wireguard.interfaces = {
        wiring = {
          ips = [ "10.10.0.1/32" ];
          privateKey = peerPrivateKey;
          listenPort = 51820;
          peers = [{
            publicKey = defaultPublicKey;
            allowedIPs = [ "10.10.0.0/32" ];
            endpoint = "default:51820";
            persistentKeepalive = 30;
          }];
        };
      };
      networking.firewall.allowedTCPPorts = [ continuityPort ];
      systemd.services."continuityServer" = {
        environment.HOST = "0.0.0.0";
        environment.PORT = builtins.toString continuityPort;
        script = "${pkgs.python3}/bin/python3 ${continuityServer}";
      };
    };
    testScript = { ... }: ''
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
  });
  coordServerIntegration-single = lib.runTest (let
    etc = self.outputs.packages.${system}.etc;
    serverPort = 39390;
    client1PrivateKey = "WMRg3WjzE0ZLuP4bAWtvcrh/Tw23MR3kv4VjpAQQz04=";
    client1PublicKey = "0LrS7ekXRHD8pLEimzLfeLlKyPprJR9oJwdAMOGhtU0=";
    client2PrivateKey = "AFiQ0ipcWrEluvCmaEoQ7PQeurOo3bVRXANAOXYny0s=";
    client2PublicKey = "J4nZeURCVbUmo5w/IBnaCU9M5tOMqGRZnPB2vAji4hE=";
    client1TokenHash = "qrystalcth_30e72874f2598c1ad8020507182a4f57a7304806947b677b69c7d76a88003bc6";
    client1Config = pkgs.writeText "client1config.json" (builtins.toJSON {
      MachineJSONPath = "/tmp/machine.json";
      BaseURL = "http://server:${builtins.toString serverPort}";
      Token = "qrystalct_ZXcX7NyjY2aqiy5bb7Oe952QSCsVxzh2FU2ahvaPiHZPJaWeN+Xi59HHvqTDT0Tyy7IOhzC9Uc3Nn7dQ+8GhbQ==";
      Network = "wiring";
      Device = "client1";
      PrivateKey = client1PrivateKey;
    });
    tokens = {
      ${client1TokenHash} = {
        Identities = [ ["wiring" "client1"] ];
      };
    };
    spec.Networks = [{
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
    }];
    serverConfig = pkgs.writeText "serverConfig.json" (builtins.toJSON {
      Tokens = tokens;
      Spec = spec;
      Addr = "0.0.0.0:${builtins.toString serverPort}";
    });
  in {
    name = "coordServerIntegration-single";
    hostPkgs = pkgs;
    nodes.server = { pkgs, ... }: { # not a wg peer
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
    nodes.client1 = { pkgs, ... }: { # wg peer setup using test-device
      imports = [ base ];
    };
    nodes.client2 = { pkgs, ... }: { # wg peer setup using NixOS's wireguard module
      imports = [ base ];
      networking.wireguard.interfaces = {
        wiring = {
          ips = [ "10.10.0.2/32" ];
          privateKey = client2PrivateKey;
          listenPort = 51820;
          peers = [{
            publicKey = client1PublicKey;
            allowedIPs = [ "10.10.0.1/32" ];
            endpoint = "client1:51820";
            persistentKeepalive = 1;
          }];
        };
      };
    };
    testScript = { ... }: ''
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
  });
  coordServerIntegration-double = lib.runTest (let
    etc = self.outputs.packages.${system}.etc;
    serverPort = 39390;
    client1PrivateKey = "WMRg3WjzE0ZLuP4bAWtvcrh/Tw23MR3kv4VjpAQQz04=";
    client1PublicKey = "0LrS7ekXRHD8pLEimzLfeLlKyPprJR9oJwdAMOGhtU0=";
    client2PrivateKey = "AFiQ0ipcWrEluvCmaEoQ7PQeurOo3bVRXANAOXYny0s=";
    client2PublicKey = "J4nZeURCVbUmo5w/IBnaCU9M5tOMqGRZnPB2vAji4hE=";
    client1TokenHash = "qrystalcth_30e72874f2598c1ad8020507182a4f57a7304806947b677b69c7d76a88003bc6";
    client1Config = pkgs.writeText "client1config.json" (builtins.toJSON {
      MachineJSONPath = "/tmp/machine.json";
      BaseURL = "http://server:${builtins.toString serverPort}";
      Token = "qrystalct_ZXcX7NyjY2aqiy5bb7Oe952QSCsVxzh2FU2ahvaPiHZPJaWeN+Xi59HHvqTDT0Tyy7IOhzC9Uc3Nn7dQ+8GhbQ==";
      Network = "wiring";
      Device = "client1";
      PrivateKey = client1PrivateKey;
    });
    client2TokenHash = "qrystalcth_4f4a908fbcc2f13f45ee71b438efa4df982f99526b085a42fbe05b019056af9e";
    client2Config = pkgs.writeText "client2config.json" (builtins.toJSON {
      MachineJSONPath = "/tmp/machine.json";
      BaseURL = "http://server:${builtins.toString serverPort}";
      Token = "qrystalct_wrN7qG37s1KeewvlafCI7GXPC71Jx6DZQAexAJTcfMHRveN7CCMebxo5VIfZxP0YQqSd79rblAgkZkZjJENRMQ==";
      Network = "wiring";
      Device = "client2";
      PrivateKey = client2PrivateKey;
    });
    tokens = {
      ${client1TokenHash} = {
        Identities = [ ["wiring" "client1"] ];
      };
      ${client2TokenHash} = {
        Identities = [ ["wiring" "client2"] ];
      };
    };
    spec.Networks = [{
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
    }];
    serverConfig = pkgs.writeText "serverConfig.json" (builtins.toJSON {
      Tokens = tokens;
      Spec = spec;
      Addr = "0.0.0.0:${builtins.toString serverPort}";
    });
  in {
    name = "coordServerIntegration-double";
    hostPkgs = pkgs;
    nodes.server = { pkgs, ... }: { # not a wg peer
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
    nodes.client1 = { pkgs, ... }: { # wg peer setup using test-device
      imports = [ base ];
      networking.firewall.allowedUDPPorts = [ 51820 ];
    };
    nodes.client2 = { pkgs, ... }: { # wg peer setup using test-device
      imports = [ base ];
      networking.firewall.allowedUDPPorts = [ 51820 ];
    };
    testScript = { ... }: ''
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
  });
}
