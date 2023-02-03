{ pkgs ? import <nixpkgs> { } }:
pkgs.buildGo119Module {
  pname = "kartusche";
  version = pkgs.lib.fileContents ./version.txt;
  src = ./.;
  vendorSha256 = "sha256-0LZGx1iVZ218ZXyXs5dHvACX8xOJJDlh3chixtOlC0o=";
}
