{
  description = "kartusche";

  outputs = { self, nixpkgs }: {
    # Loaded with `nix run`
    packages.x86_64-linux.default = import ./. {
      pkgs = nixpkgs.legacyPackages.x86_64-linux;
    };

    # Loaded with `nix develop`
    devShells.x86_64-linux.default = import ./shell.nix {
      pkgs = nixpkgs.legacyPackages.x86_64-linux;
    };
  };
}
