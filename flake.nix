{
  description = "atk-tool nix flake";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs =
    { self, nixpkgs }:
    let
      supportedSystems = [
        "x86_64-linux"
        "aarch64-linux"
      ];
      forAllSystems = nixpkgs.lib.genAttrs supportedSystems;
      nixpkgsFor = forAllSystems (system: import nixpkgs { inherit system; });
    in
    {
      packages = forAllSystems (system: {
        default = nixpkgsFor.${system}.buildGoModule {
          pname = "atk-tool";
          version = "0.2.0";
          src = ./.;
          vendorHash = "sha256-Imyr7hUSD3Zec0SzPw7DBN65S3feCeddmhlySoeIG0Q=";
          nativeBuildInputs = [ nixpkgsFor.${system}.pkg-config ];
          buildInputs = [ nixpkgsFor.${system}.udev ];
          subPackages = [ "cmd/atk-tool" ];
        };
      });

      nixosModules.default =
        {
          config,
          lib,
          pkgs,
          ...
        }:
        {
          options.services.atk-tool.enable = lib.mkEnableOption "atk-tool HID configuration";

          config = lib.mkIf config.services.atk-tool.enable {
            environment.systemPackages = [ self.packages.${pkgs.stdenv.hostPlatform.system}.default ];

            services.udev.extraRules = ''
              SUBSYSTEM=="hidraw", ATTRS{idVendor}=="373b", TAG+="uaccess"
            '';
          };
        };

      devShells = forAllSystems (
        system:
        let
          pkgs = nixpkgsFor.${system};
        in
        {
          default = pkgs.mkShell {
            buildInputs = [
              pkgs.go
              pkgs.gopls
              pkgs.udev
            ];
          };
        }
      );
    };
}
