# 🧩 Jigs

[![Test](https://github.com/branchard/jigs/actions/workflows/test.yaml/badge.svg)](https://github.com/branchard/jigs/actions/workflows/test.yaml)
[![GitHub Tag](https://img.shields.io/github/v/tag/branchard/jigs)](https://github.com/branchard/jigs/tags)

CLI tool for interactively managing dotenv files with no third-party dependencies.

Takes one or more template files (e.g. `.env.dist`, `.env.dev`), prompts you for any missing values, and writes a `.env`
file. If `.env` already exists, only variables not yet present are prompted.

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
