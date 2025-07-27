{
  description = "Elephant Dev Shell";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-25.05";

  outputs = { self, nixpkgs }:
    let
      system = "x86_64-linux";
      pkgs = import nixpkgs { inherit system; };
    in {
      devShells.${system}.default = pkgs.mkShell {
        name = "elephant-dev-shell";

        buildInputs = with pkgs; [
          go
          gcc
        ];
      };
    };
}