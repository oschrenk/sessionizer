{
  description = "Sessionizer - fuzzy finder for creating and switching between tmux sessions";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";

  # Offer prebuilt binaries from the Cachix cache so `nix profile install`
  # downloads instead of compiling. Consumers are prompted to trust these.
  nixConfig = {
    extra-substituters = [ "https://oschrenk.cachix.org" ];
    extra-trusted-public-keys = [
      "oschrenk.cachix.org-1:3JOMfkq2vFiLw4UsCVwzu8kWFBkuS/3DD5AojcO9pks="
    ];
  };

  outputs =
    { self, nixpkgs }:
    let
      # Single source of truth for the version: ./VERSION holds a bare semver
      # (e.g. 0.13.0); the "v" prefix is added here and by the taskfile release
      # flow, so both builds stamp cmd.Version with the same string.
      version = nixpkgs.lib.fileContents ./VERSION;

      # aarch64-darwin only: it is the only system that runs sessionizer, and
      # so the only one the binary cache is ever populated for.
      systems = [
        "aarch64-darwin"
      ];
      forAllSystems = f: nixpkgs.lib.genAttrs systems (system: f nixpkgs.legacyPackages.${system});
    in
    {
      packages = forAllSystems (pkgs: rec {
        sessionizer = pkgs.buildGoModule {
          pname = "sessionizer";
          inherit version;
          src = self;

          # Regenerate after changing go.mod/go.sum: set to lib.fakeHash,
          # run `nix build`, then paste the expected hash from the error.
          vendorHash = "sha256-CD+LB7OPl/P2g26oLMVW7YN6MzqxUa+btS0j891trOg=";

          # No subPackages: the entrypoint is main.go at the repo root.
          ldflags = [
            "-s"
            "-w"
            "-X github.com/oschrenk/sessionizer/cmd.Version=v${version}"
          ];

          # `completion` is hidden (cmd/root.go) but not disabled. The fatal
          # config read lives in initConfig, which only start and search call,
          # and the root PersistentPreRun read is non-fatal — so this runs fine
          # in the sandbox, where there is no config file.
          nativeBuildInputs = [ pkgs.installShellFiles ];
          postInstall = ''
            installShellCompletion --cmd sessionizer \
              --bash <($out/bin/sessionizer completion bash) \
              --zsh <($out/bin/sessionizer completion zsh) \
              --fish <($out/bin/sessionizer completion fish)
          '';

          meta = {
            description = "Fuzzy finder for creating and switching between tmux sessions";
            homepage = "https://github.com/oschrenk/sessionizer";
            mainProgram = "sessionizer";
            platforms = nixpkgs.lib.platforms.darwin;
          };
        };
        default = sessionizer;
      });

      apps = forAllSystems (pkgs: rec {
        sessionizer = {
          type = "app";
          program = "${self.packages.${pkgs.stdenv.hostPlatform.system}.sessionizer}/bin/sessionizer";
        };
        default = sessionizer;
      });

      homeModules = rec {
        sessionizer = import ./nix/home-manager.nix self;
        default = sessionizer;
      };

      devShells = forAllSystems (pkgs: {
        default = pkgs.mkShell {
          packages = with pkgs; [
            go # go, language
            golangci-lint # go, linter runner
            gopls # go, lsp
          ];
        };
      });
    };
}
