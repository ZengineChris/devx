{
  description = "DevX development tool";

  # GitHub URLs for the Nix inputs we're using
  inputs = {
    # Simply the greatest package repository on the planet
    nixpkgs.url = "github:NixOS/nixpkgs";
    # A set of helper functions for using flakes
    flake-utils.url = "github:numtide/flake-utils";

    # For shell.nix compatibility
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
      pkgs = import nixpkgs {inherit system;};

      # Define your Go package
      devx = pkgs.buildGoModule {
        pname = "devx";
        version = "0.1.0";
        src = ./.; # For local development
        vendorSha256 = null; # Will be computed automatically
      };
    in {
      packages.${system} = rec {
        devx = devx;
        default = devx;
      };
      devShells = {
        default = pkgs.mkShell {
          # Packages included in the environment
          buildInputs = with pkgs; [
            go_1_23

            # goimports, godoc, etc.
            gotools

            # https://github.com/golangci/golangci-lint
            golangci-lint

            # The Go language server (for IDEs and such)
            gopls

            # https://pkg.go.dev/github.com/ramya-rao-a/go-outline
            go-outline

            # https://github.com/uudashr/gopkgs
            gopkgs
          ];

          # Run when the shell is started up
          shellHook = ''
            ${pkgs.go_1_23}/bin/go version
          '';
        };
      };
      overlays.default = final: prev: {
        devx = self.packages.${prev.system}.default;
      };
    });
}
