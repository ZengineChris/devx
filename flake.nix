{
  description = "DevX development env";

  # GitHub URLs for the Nix inputs we're using
  inputs = {
    # Simply the greatest package repository on the planet
    nixpkgs.url = "github:NixOS/nixpkgs";
    # A set of helper functions for using flakes
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = {
    self,
    nixpkgs,
    flake-utils,
  }:
    flake-utils.lib.eachDefaultSystem (system: let
      pkgs = import nixpkgs {inherit system;};
    in {
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
    });
}
