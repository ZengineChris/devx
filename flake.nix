{
  description = "DevX development tool";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    flake-compat = {
      url = "github:edolstra/flake-compat";
      flake = false;
    };
    nur = {
      url = "github:nix-community/NUR";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = {
    self,
    nixpkgs,
    flake-utils,
    nur,
    ...
  }:
    flake-utils.lib.eachDefaultSystem (system: let
      pkgs = nixpkgs.legacyPackages.${system};

      devx = pkgs.buildGo124Module {
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
          go_1_24
          git
          gotools
          golangci-lint
          go-task
          gopls
          go-outline
          gopkgs
          nur.repos.goreleaser.goreleaser
        ];

        shellHook = ''
          ${pkgs.go_1_24}/bin/go version
        '';
      };

      overlays.default = final: prev: {
        devx = self.packages.${prev.system}.default;
      };
    });
}
