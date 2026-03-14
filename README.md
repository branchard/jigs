# 🧩 Jigs

[![Test](https://github.com/branchard/jigs/actions/workflows/test.yaml/badge.svg)](https://github.com/branchard/jigs/actions/workflows/test.yaml)
[![GitHub Release](https://img.shields.io/github/v/release/branchard/jigs)](https://github.com/branchard/jigs/releases)
[![Go](https://img.shields.io/github/go-mod/go-version/branchard/jigs)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://github.com/branchard/jigs/blob/main/LICENSE)

<img src="https://raw.githubusercontent.com/branchard/jigs/master/docs/demo.gif" alt="Jigs in action"/>

CLI tool for interactively managing dotenv files with no third-party dependencies.

Takes one or more template files (e.g. `.env.dist`, `.env.dev`), prompts you for any missing values, and writes a `.env`
file in the current directory. If `.env` already exists, only variables not yet present are prompted.

Key behaviors:

- **Idempotent** — running jigs again won't re-prompt for variables that are already defined in `.env`.
- **Non-destructive** — existing entries in `.env` are never modified or reordered. New variables are appended.
- **Structure-preserving** — comments and blank lines in `.env` are preserved through round-trips.
- **First-occurrence-wins** — when multiple template files define the same variable, the default value from the first file is used.

## Usage

### With the binary

```shell
curl -L -o jigs https://github.com/branchard/jigs/releases/latest/download/jigs-linux-amd64
chmod +x jigs
./jigs --help
./jigs .env.dist
```

### With Docker

```shell
docker run --rm -it --volume ./:/mnt/ ghcr.io/branchard/jigs:latest .env.dist
```

Docker images are available on [GitHub Container Registry](https://github.com/branchard/jigs/pkgs/container/jigs) for `linux/amd64` and `linux/arm64`.

## Supported syntax

Template files and `.env` files use the standard dotenv format:

```shell
# Comments start with #
KEY=value

# Spaces around = are allowed
KEY = value

# Values can be quoted (quotes are stripped)
KEY="hello world"
KEY='hello world'

# Empty values
KEY=

# Blank lines are preserved
```

Values containing spaces, tabs, `#`, quotes, backslashes, `$`, backticks, or `!` are automatically quoted when written.

Not supported: variable interpolation (`$VAR`), multiline values, `export` prefix, escape sequences within quoted values.

## Building from source

Requires Go 1.26+.

```shell
git clone https://github.com/branchard/jigs.git
cd jigs
make build       # build binary to ./build/jigs
make test        # run unit and end-to-end tests
make run         # run directly with go run
make dev         # start a development container (Docker)
```

## License

[MIT](LICENSE)
