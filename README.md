# 🧩 Jigs

[![Test](https://github.com/branchard/jigs/actions/workflows/test.yaml/badge.svg)](https://github.com/branchard/jigs/actions/workflows/test.yaml)
[![GitHub Tag](https://img.shields.io/github/v/tag/branchard/jigs)](https://github.com/branchard/jigs/tags)

CLI tool for interactively managing dotenv files with no third-party dependencies.

Takes one or more template files (e.g. `.env.dist`, `.env.dev`), prompts you for any missing values, and writes a `.env` file. If `.env` already exists, only variables not yet present are prompted.

## Usage

```
jigs <file1> [file2 ...]
```

### Example

Given a `.env.dist`:

```
# Database
DB_HOST=localhost
DB_PORT=5432
DB_PASSWORD=

# App
APP_SECRET=
```

Run:

```
jigs .env.dist
```

The tool will prompt you for `DB_HOST` (default: `localhost`), `DB_PORT` (default: `5432`), `DB_PASSWORD`, and `APP_SECRET`, then write the results to `.env`.

Running again will skip any variables already defined in `.env`.

### Multiple source files

```
jigs .env.dist .env.dev
```

Variables are collected from all files in order. The first file to define a key sets the default value; duplicates in later files are ignored.

### With Docker

```
docker run --rm -it --volume ./:/mnt/ ghcr.io/branchard/jigs:latest .env.dist
```