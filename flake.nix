{
  description = "kartusche";

  outputs = { self, nixpkgs }:
    let
      forAllSystems = nixpkgs.lib.genAttrs nixpkgs.lib.systems.flakeExposed;
    in
    {
      # Loaded with `nix run`
      packages = forAllSystems
        (system: {
          default = import ./. {
            pkgs = nixpkgs.legacyPackages.${system};
          };
        });

      # Loaded with `nix develop`
      devShells = forAllSystems (system: {
        default = import ./shell.nix {
          pkgs = nixpkgs.legacyPackages.${system};
        };
      });
    };
}
