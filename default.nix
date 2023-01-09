{ pkgs ? import (import ./nix/sources.nix).nixpkgs { } }:
pkgs.buildGo119Module {
  pname = "kartusche";
  version = pkgs.lib.fileContents ./version.txt;
  src = ./.;
  vendorSha256 = "sha256-uZdxIFvrudro0e1V9TmZvACaSJku76G0AE4YI4IFrfc=";
}
