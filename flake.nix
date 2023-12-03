{
  description = "Wander";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let pkgs = nixpkgs.legacyPackages.${system};
      in {
        packages.default = self.packages.${system}.wander;

        packages.wander = pkgs.buildGoModule {
          name = "wander";
          vendorHash = "sha256-SqDGXV8MpvEQFAkcE1NWvWjdzYsvbO5vA6k+hpY0js0=";
          src = ./.;

          nativeBuildInputs = [ pkgs.installShellFiles ];

          postInstall = ''
            installShellCompletion --cmd wander \
               --fish <($out/bin/wander completion fish) \
               --bash <($out/bin/wander completion bash) \
               --zsh <($out/bin/wander completion zsh)
          '';
        };
      });
}
