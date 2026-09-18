# cli

The ODIN command-line interface (Cobra + Bubble Tea).

## Run locally

```sh
go run .
```

If Make is installed, `make run`, `make test`, `make vet`, `make build`, and `make fmt` are also
available.

## Docker

A development `Dockerfile` runs the CLI with hot-reload (via [air](https://github.com/air-verse/air)).
Unlike the API services, `cli` has no HTTP server and exposes no port.

Build the image:

```sh
make docker-build
```

Run it locally with hot-reload (mounts the source code, interactive TTY):

```sh
make docker-dev
```

Editing any `.go` file rebuilds and restarts the CLI automatically inside the container.
