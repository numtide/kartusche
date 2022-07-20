{ pkgs ? import <nixpkgs> { } }:
pkgs.buildGo118Module {
  pname = "kartusche";
  version = pkgs.lib.fileContents ./version.txt;
  src = ./.;
  vendorSha256 = "sha256-e8FN/NTLR56dvZ1V+n2wEdU5oN/UyfFczkqbXVuNvKI=";
}
