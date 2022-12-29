{ pkgs ? import (import ./nix/sources.nix).nixpkgs { }  }:
pkgs.buildGo119Module {
  pname = "kartusche";
  version = pkgs.lib.fileContents ./version.txt;
  src = ./.;
  vendorSha256 = "sha256-k+Qwuw2bsE7ttoEAul6LEyi9mNIeK+kTJWUjLbmW88k=";
}
