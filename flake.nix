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

  outputs = { self, nixpkgs, flake-utils, ... }:
    flake-utils.lib.eachDefaultSystem (system: let
      pkgs = nixpkgs.legacyPackages.${system};
      
      devx = pkgs.buildGoModule {
        pname = "devx";
        version = "0.1.0";
        src = ./.;
        vendorHash = null;
      };
    in {
      packages = {
        devx = devx;
        default = devx;
      };

      devShells.${system}.default = pkgs.mkShell {
        buildInputs = with pkgs; [
          go_1_23
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
