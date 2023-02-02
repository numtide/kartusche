let
  unstable = import (fetchTarball https://nixos.org/channels/nixos-unstable/nixexprs.tar.xz) { };
in
{ pkgs ? import <nixpkgs> { } }:
pkgs.mkShell {
  packages = [
    pkgs.gcc
    unstable.go_1_19
    unstable.gotools
    unstable.gopls
    unstable.go-outline
    unstable.gocode
    unstable.gopkgs
    unstable.gocode-gomod
    unstable.godef
    unstable.golint
    unstable.gh
    unstable.delve
    unstable.go-tools
  ];
  hardeningDisable = [ "all" ];  # to build the cross-compiler
}