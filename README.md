# cache-sweep

A small CLI that finds `node_modules` directories and Python virtualenvs
(directories containing `pyvenv.cfg`) under a given directory tree, shows
how much disk space they're using, and deletes them after you confirm.

## Install

Download a prebuilt binary from the [releases page](https://github.com/mattfirtion/cache-sweep/releases)
for your platform (macOS, Linux, and Windows; amd64/arm64), or build from source:

```sh
git clone https://github.com/mattfirtion/cache-sweep.git
cd cache-sweep
go build -o cache-sweep .
```

## Usage

```sh
cache-sweep [directory]
```

If no directory is given, cache-sweep scans the current directory. It walks
the tree looking for `node_modules` directories and Python virtualenvs,
stopping its search at each match (so nested matches inside a `node_modules`
or venv aren't double-counted), then lists what it found before deleting
anything:

```
$ cache-sweep ~/code
Found 3 directories under /Users/matt/code:

  1. 128.4MiB  /Users/matt/code/api/node_modules
  2.  45.2MiB  /Users/matt/code/api/.venv
  3. 512.7MiB  /Users/matt/code/web/node_modules

Total reclaimable: 686.3MiB

Delete all 3 directories (686.3MiB)? [y/N] y
removed /Users/matt/code/api/node_modules
removed /Users/matt/code/api/.venv
removed /Users/matt/code/web/node_modules

Reclaimed 686.3MiB.
```

Answering anything other than `y`/`yes` aborts without deleting anything.

## Development

```sh
go test ./...
```

Commits follow [Conventional Commits](https://www.conventionalcommits.org/)
and are enforced by commitlint on pull requests; releases are cut
automatically from `main` via semantic-release and GoReleaser.

## License

[MIT](LICENSE)
