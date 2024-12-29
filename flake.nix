{
  description = "DevX development tool";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs";
    flake-utils.url = "github:numtide/flake-utils";
    flake-compat = {
      url = "github:edolstra/flake-compat";
      flake = false;
    };
  };

  outputs = {
    self,
    nixpkgs,
    flake-utils,
    ...
  }:
    flake-utils.lib.eachDefaultSystem (system: let
      pkgs = nixpkgs.legacyPackages.${system};

      devx = pkgs.buildGo123Module {
        pname = "devx";
        version = "0.1.0";
        src = ./.;
        vendorHash = null;
        CGO_ENABLED = 1;
        subPackages = ["cmd/devx"];

        # `nix-build` has .git folder but `nix build` does not, this caters for both cases
        preConfigure = ''
          export VERSION="$(git describe --tags --always || echo nix-build-at-"$(date +%s)")"
          export REVISION="$(git rev-parse HEAD || echo nix-unknown)"
          ldflags="-X github.com/zenginechris/devx/config.appVersion=$VERSION
                    -X github.com/zenginechris/devx/config.revision=$REVISION"
        '';

      };
    in {
      packages = {
        devx = devx;
        default = devx;
      };

      devShells.${system}.default = pkgs.mkShell {
        buildInputs = with pkgs; [
          go_1_23
          git
          gotools
          golangci-lint
          gopls
          go-outline
          gopkgs
        ];

        shellHook = ''
          ${pkgs.go_1_23}/bin/go version
        '';
      };

      overlays.default = final: prev: {
        devx = self.packages.${prev.system}.default;
      };
    });
}
