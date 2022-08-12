{ pkgs ? import (import ./nix/sources.nix).nixpkgs { }  }:
pkgs.buildGo119Module {
  pname = "kartusche";
  version = pkgs.lib.fileContents ./version.txt;
  src = ./.;
  vendorSha256 = "sha256-H2BvWn4F7gCmkXBZVJdC2t65OPc2w7cyogP69xbB/Yo=";
}
