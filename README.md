# Octi Sync Server - Easy Synchronization!

[![Coverage Status](https://coveralls.io/repos/github/jakob-moeller-cloud/octi-sync-server/badge.svg?branch=main)](https://coveralls.io/github/jakob-moeller-cloud/octi-sync-server?branch=main)
[![Go Report Card](https://goreportcard.com/badge/github.com/jakob-moeller-cloud/octi-sync-server)](https://goreportcard.com/report/github.com/jakob-moeller-cloud/octi-sync-server)
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fjakob-moeller-cloud%2Focti-sync-server.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2Fjakob-moeller-cloud%2Focti-sync-server?ref=badge_shield)

[![Tests](https://github.com/jakob-moeller-cloud/octi-sync-server/actions/workflows/test.yaml/badge.svg?branch=main)](https://github.com/jakob-moeller-cloud/octi-sync-server/actions/workflows/test.yaml)
[![Docker](https://github.com/jakob-moeller-cloud/octi-sync-server/actions/workflows/docker-publish.yml/badge.svg)](https://github.com/jakob-moeller-cloud/octi-sync-server/actions/workflows/docker-publish.yml)
[![CodeQL](https://github.com/jakob-moeller-cloud/octi-sync-server/actions/workflows/codeql.yml/badge.svg)](https://github.com/jakob-moeller-cloud/octi-sync-server/actions/workflows/codeql.yml)
[![golangci-lint](https://github.com/jakob-moeller-cloud/octi-sync-server/actions/workflows/golangci-lint.yaml/badge.svg)](https://github.com/jakob-moeller-cloud/octi-sync-server/actions/workflows/golangci-lint.yaml)


## License
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fjakob-moeller-cloud%2Focti-sync-server.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2Fjakob-moeller-cloud%2Focti-sync-server?ref=badge_large)


## Developing and Operating the Server

### Running the Server

#### In Kubernetes

Pre-Requisites:
- Kustomize
- A Kubernetes Cluster

Please see `deploy/kustomize` for a `kustomization.yaml` that you can use to deploy the Application.

By default, the Deployment creates a Redis to join the Server which is not bound to any PV.
This makes it *not ready for productive use* out of the box

#### From Docker

Pre-Requisites:
- Docker

```shell
docker run ghcr.io/jakob-moeller-cloud/octi-sync-server:latest
```

Make sure to open Port `8080` if you want the server to be reachable.
Also, you might want to bind in `config.yml` via volume to override configuration values.

#### From Source

```shell
go run .
```

Adjust configuration parameters in `config.yml` where necessary!

### Inspecting and Recreating The OpenAPI Definitions and Mocks

#### V1

You can access the active JSON-Formatted OpenAPI Definition under

`http://localhost:8080/v1/openapi`.

You can introspect the API with `Swagger-UI` using a static Path or from the hosted Server. 
For ease of development, you can introspect locally with 

```shell
# URL is where the server is hosted, 
# remember that by default it will be blocked by CORS!
docker run -it --rm \
    -p 80:8080 \
    -e URL=http://localhost:8080/v1/openapi \
    -v $(pwd)/api/v1:/v1 swaggerapi/swagger-ui
```
and then opening your browser on [http://localhost:80](http://localhost:80) while running the server.

You can recreate the Definitions and Mocks with

```shell
go generate ./...
```

### Running Tests

```shell
go test ./...
```

### Linting

```shell
docker run --rm -v $(pwd):/app -w /app golangci/golangci-lint:v1.49.0 golangci-lint run
```
