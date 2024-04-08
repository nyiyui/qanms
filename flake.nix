{
  inputs.nixpkgs.url = "nixpkgs/nixpkgs-unstable";
  inputs.flake-utils.url = "github:numtide/flake-utils";

  outputs = { self, nixpkgs, flake-utils }:
    let
      # to work with older version of flakes
      lastModifiedDate =
        self.lastModifiedDate or self.lastModified or "19700101";

      # Generate a user-friendly version number.
      version = (builtins.substring 0 8 lastModifiedDate) + "-"
        + (if (self ? rev) then self.rev else "dirty");

      # System types to support.
      supportedSystems = [ "x86_64-linux" "aarch64-linux" ];

      # Helper function to generate an attrset '{ x86_64-linux = f "x86_64-linux"; ... }'.
      forAllSystems = nixpkgs.lib.genAttrs supportedSystems;

      # Nixpkgs instantiated for supported system 
      nixpkgsFor = forAllSystems (system: import nixpkgs { inherit system; });
      libFor = forAllSystems (system: import (nixpkgs + "/lib"));
      nixosLibFor = forAllSystems (system: import (nixpkgs + "/nixos/lib"));
    in flake-utils.lib.eachSystem supportedSystems (system:
      let
        pkgs = import nixpkgs { inherit system; };
        lib = import (nixpkgs + "/lib") { inherit system; };
        nixosLib = import (nixpkgs + "/nixos/lib") { inherit system; };
        ldflags = pkgs: [
          "-X github.com/nyiyui/qrystal/mio.CommandBash=${pkgs.bash}/bin/bash"
          "-X github.com/nyiyui/qrystal/mio.CommandWg=${pkgs.wireguard-tools}/bin/wg"
          "-X github.com/nyiyui/qrystal/mio.CommandWgQuick=${pkgs.wireguard-tools}/bin/wg-quick"
          "-X github.com/nyiyui/qrystal/node.CommandIp=${pkgs.iproute2}/bin/ip"
          "-X github.com/nyiyui/qrystal/node.CommandIptables=${pkgs.iptables}/bin/iptables"
          "-race"
        ];
      in rec {
        devShells = let pkgs = nixpkgsFor.${system};
        in {
          default = pkgs.mkShell {
            buildInputs = with pkgs; [
              bash
              go
              git
              protobuf
              protoc-gen-go
              protoc-gen-go-grpc
              nixfmt
              govulncheck
            ];
          };
        };
        packages = let
          pkgs = nixpkgsFor.${system};
          lib = libFor.${system};
          common = {
            inherit version;
            src = ./.;

            ldflags = ldflags pkgs;

            tags = [ "nix" "sdnotify" ];

            #vendorHash = pkgs.lib.fakeSha256;
            vendorHash =
              "sha256-T/HPyY+A1o6oMw5NEMg8ui5IHitOOyGF6DAIjfLmNRI=";
          };
        in {
          runner = pkgs.buildGoModule (common // {
            pname = "runner";
            subPackages = [
              "cmd/runner"
              "cmd/runner-mio"
              "cmd/runner-node"
              "cmd/runner-hokuto"
            ];
            ldflags = (ldflags pkgs) ++ [
              "-X github.com/nyiyui/qrystal/runner.NodeUser=qrystal-node"
            ];
            postInstall = ''
              mkdir $out/lib
              cp $src/mio/dev-add.sh $out/lib
              cp $src/mio/dev-remove.sh $out/lib
            '';
          });
          cs = pkgs.buildGoModule (common // {
            pname = "cs";
            subPackages = [ "cmd/cs" ];
          });
          etc = pkgs.buildGoModule (common // {
            name = "etc";
            # NOTE: specifying subPackages makes buildGoModule not test other packages :(
          });
          sd-notify-test = pkgs.buildGoModule (common // {
            pname = "sd-notify-test";
            subPackages = [ "cmd/sd-notify-test" ];
          });
        };
        checks = (import ./test.nix) {
          inherit self system nixpkgsFor libFor nixosLibFor ldflags;
        };
        nixosModules = ((import ./nixos-modules.nix) {
          inherit self system nixpkgsFor libFor nixosLibFor ldflags packages;
        }).nixosModules;
      });
}
