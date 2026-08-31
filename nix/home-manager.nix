# home-manager module for sessionizer.
#
#   inputs.sessionizer.url = "github:oschrenk/sessionizer";
#   imports = [ inputs.sessionizer.homeModules.sessionizer ];
#
#   programs.sessionizer = {
#     enable = true;
#     startup = {
#       name = "config";
#       path = "$HOME/.config";
#     };
#     search = {
#       directories = [ "$HOME/Projects" ];
#       entries = [
#         "$HOME/.local/share/chezmoi"
#         { path = "$HOME/Obsidian/memex"; name = "self/memex"; layout = "two-columns"; }
#       ];
#     };
#     layouts.two-columns = {
#       windows = [ { layout = "even-horizontal"; panes = [ { focus = true; } { focus = false; } ]; } ];
#     };
#   };
#
# Generates $XDG_CONFIG_HOME/sessionizer/config.toml, the first path
# cmd/root.go's configureViper searches, plus one file per named layout under
# $XDG_CONFIG_HOME/sessionizer/layouts/. The two have to be siblings: a named
# layout resolves to `filepath.Dir(viper.ConfigFileUsed())/layouts/<name>.yml`
# (core/search.go, resolveLayoutPath).
#
# `base.tmuxPath` points at this flake's pkgs.tmux, so set
# `inputs.sessionizer.inputs.nixpkgs.follows = "nixpkgs"` in the consuming flake
# to avoid pulling in a second tmux build.
self:
{
  config,
  lib,
  pkgs,
  ...
}:

let
  inherit (lib)
    literalExpression
    mkEnableOption
    mkIf
    mkOption
    types
    ;

  cfg = config.programs.sessionizer;

  tomlFormat = pkgs.formats.toml { };
  yamlFormat = pkgs.formats.yaml { };

  # Drop unset keys so the generated TOML holds only what was configured.
  withoutNull = lib.filterAttrs (_: v: v != null);

  # An entry is a path string, coerced to the { path, name, layout } table
  # cmd/search.go's parseSearchEntries accepts. Both shapes are valid TOML for
  # sessionizer; the coercion keeps the common case a bare string in Nix.
  entryType = types.coercedTo types.str (path: { inherit path; }) (
    types.submodule {
      options = {
        path = mkOption {
          type = types.str;
          example = "$HOME/.local/share/chezmoi";
          description = ''
            Directory the session opens in. Expanded by sessionizer through
            `os.ExpandEnv`, so `$HOME` may be used literally.
          '';
        };

        name = mkOption {
          type = types.nullOr types.str;
          default = null;
          description = ''
            Label shown in the fuzzy finder, and the tmux session name. Defaults
            to the basename of {option}`path`.
          '';
        };

        layout = mkOption {
          type = types.nullOr types.str;
          default = null;
          description = ''
            Name of a layout from {option}`programs.sessionizer.layouts`, applied
            when the session is first created. A `.sessionizer.yml` in the
            directory itself takes precedence.
          '';
        };
      };
    }
  );

  settings =
    {
      base = withoutNull {
        ignore = cfg.base.ignore;
        rooter_patterns = cfg.base.rooterPatterns;
        socket_name = cfg.base.socketName;
        tmux_path = cfg.base.tmuxPath;
      };

      search = {
        directories = cfg.search.directories;
        entries = map withoutNull cfg.search.entries;
      };
    }
    // lib.optionalAttrs (cfg.startup.name != null) {
      # core.BuildEntries only offers the startup entry when the name is set,
      # so an unset name omits the section entirely.
      startup = withoutNull {
        inherit (cfg.startup) name path;
        layout_path = cfg.startup.layoutPath;
      };
    };

  layoutFiles = lib.mapAttrs' (
    name: layout:
    lib.nameValuePair "sessionizer/layouts/${name}.yml" {
      source = yamlFormat.generate "sessionizer-layout-${name}.yml" layout;
    }
  ) cfg.layouts;
in
{
  options.programs.sessionizer = {
    enable = mkEnableOption "sessionizer, a fuzzy finder for tmux sessions";

    package = mkOption {
      type = types.package;
      default = self.packages.${pkgs.stdenv.hostPlatform.system}.sessionizer;
      defaultText = literalExpression "sessionizer.packages.\${system}.sessionizer";
      description = "The sessionizer package to install.";
    };

    base = {
      ignore = mkOption {
        type = types.listOf types.str;
        default = [ ];
        example = [
          ".build"
          "node_modules"
        ];
        description = ''
          Directory names skipped while walking {option}`search.directories`.
          Only applies to directories that are not themselves a project.
        '';
      };

      rooterPatterns = mkOption {
        type = types.listOf types.str;
        default = [ ".git" ];
        example = [
          "build.sbt"
          ".git"
          "README.md"
        ];
        description = ''
          Files or directories that mark a directory as a project. The walk stops
          descending as soon as one matches, so nested repositories below a match
          are not offered separately.
        '';
      };

      socketName = mkOption {
        type = types.nullOr types.str;
        default = null;
        example = "primary";
        description = ''
          tmux socket every command targets, as `tmux -L <name>`. Null uses the
          default server. Overridden by `--socket-name` and
          `SESSIONIZER_SOCKET_NAME`.
        '';
      };

      tmuxPath = mkOption {
        type = types.str;
        default = "${pkgs.tmux}/bin/tmux";
        defaultText = literalExpression ''"''${pkgs.tmux}/bin/tmux"'';
        description = ''
          tmux executable that sessionizer shells out to. An absolute path is
          what makes it work under launchd, whose default PATH excludes the nix
          profile directories. Set to `"tmux"` to resolve through PATH instead.
        '';
      };
    };

    startup = {
      name = mkOption {
        type = types.nullOr types.str;
        default = null;
        example = "config";
        description = ''
          Name of the session `sessionizer start` opens, also offered first in
          the finder. Null drops the startup entry altogether.
        '';
      };

      path = mkOption {
        type = types.nullOr types.str;
        default = null;
        example = "$HOME/.config";
        description = ''
          Directory the startup session opens in. Required when
          {option}`startup.name` is set. Expanded through `os.ExpandEnv`.
        '';
      };

      layoutPath = mkOption {
        type = types.nullOr types.str;
        default = null;
        example = "$HOME/.config/sessionizer/startup.yml";
        description = ''
          Layout file applied to the startup session, given as a direct path
          rather than a name from {option}`layouts`.
        '';
      };
    };

    search = {
      directories = mkOption {
        type = types.listOf types.str;
        default = [ ];
        example = [ "$HOME/Projects" ];
        description = ''
          Directories walked recursively for projects, matched by
          {option}`base.rooterPatterns`. The label is the path relative to the
          search directory.
        '';
      };

      entries = mkOption {
        type = types.listOf entryType;
        default = [ ];
        example = literalExpression ''
          [
            "$HOME/.local/share/chezmoi"
            { path = "$HOME/Obsidian/memex"; name = "self/memex"; layout = "two-columns"; }
          ]
        '';
        description = ''
          Directories offered regardless of {option}`search.directories`, as a
          bare path or a table with `name` and `layout`.
        '';
      };
    };

    layouts = mkOption {
      type = types.attrsOf yamlFormat.type;
      default = { };
      example = literalExpression ''
        {
          two-columns.windows = [
            {
              layout = "even-horizontal";
              panes = [ { focus = true; } { focus = false; } ];
            }
          ];
        }
      '';
      description = ''
        Named layouts in tmuxp format, written to
        `$XDG_CONFIG_HOME/sessionizer/layouts/<name>.yml` and referenced by the
        `layout` key of a {option}`search.entries` object.
      '';
    };
  };

  config = mkIf cfg.enable {
    assertions = [
      {
        assertion = cfg.startup.name == null || cfg.startup.path != null;
        message = "programs.sessionizer: `startup.path` is required when `startup.name` is set.";
      }
    ];

    home.packages = [ cfg.package ];

    # Written unconditionally: initConfig prints "Config file not found" and
    # exits 1, so an enabled sessionizer without this file cannot search or start.
    xdg.configFile = {
      "sessionizer/config.toml".source = tomlFormat.generate "sessionizer-config.toml" settings;
    } // layoutFiles;
  };
}
