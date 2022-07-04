let pkgs = import <nixpkgs> { }; in
pkgs.mkShell {
  packages = [
    pkgs.elmPackages.elm
    pkgs.gcc
    pkgs.go_1_18
    pkgs.nodejs-16_x
  ];
}
