# Octi Sync Server - Easy Synchronization!

[![Coverage Status](https://coveralls.io/repos/github/jakobmoellerdev/octi-sync-server/badge.svg?branch=main)](https://coveralls.io/github/jakobmoellerdev/octi-sync-server?branch=main)
[![Go Report Card](https://goreportcard.com/badge/github.com/jakobmoellerdev/octi-sync-server)](https://goreportcard.com/report/github.com/jakobmoellerdev/octi-sync-server)
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fjakobmoellerdev%2Focti-sync-server.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2Fjakobmoellerdev%2Focti-sync-server?ref=badge_shield)

[![Build](https://github.com/jakobmoellerdev/octi-sync-server/actions/workflows/build.yaml/badge.svg?branch=main)](https://github.com/jakobmoellerdev/octi-sync-server/actions/workflows/build.yaml)
[![CodeQL](https://github.com/jakobmoellerdev/octi-sync-server/actions/workflows/codeql.yml/badge.svg)](https://github.com/jakobmoellerdev/octi-sync-server/actions/workflows/codeql.yml)


## License

[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fjakobmoellerdev%2Focti-sync-server.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2Fjakobmoellerdev%2Focti-sync-server?ref=badge_large)


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
docker run ghcr.io/jakobmoellerdev/octi-sync-server:latest
```

You can verify build the build integrity with the provided `cosign.pub` Public Key:

```shell
docker run -it --rm -v $PWD:/repo gcr.io/projectsigstore/cosign verify --key /repo/cosign.pub ghcr.io/jakobmoellerdev/octi-sync-server:latest
```

Make sure to open Port `8080` if you want the server to be reachable.
Also, you might want to bind in `config.yml` via volume to override configuration values.

#### From Source

```shell
go run .
```

Adjust configuration parameters in `config.yml` where necessary!

For running a redis locally for testing, use

```shell
docker run -it --rm -p 6379:6379 --name octi-redis redis:latest
```

#### From Release

First download the artifact:
```shell
VERSION=0.2.3-alpha4 \
RELEASE=https://github.com/jakobmoellerdev/octi-sync-server/releases/download/v$VERSION; \
wget $RELEASE/octi-sync-server_${VERSION}_Linux_x86_64.tar.gz
```

Next download the signature:

To verify the integrity of the checksums of the remote build before downloading:
```shell
## note that PUBLIC_KEY is coming from the repository here, you can also download it before and use a local mount
PUBLIC_KEY=/repo/cosign.pub \
VERSION=0.2.3-alpha4 \
RELEASE=https://github.com/jakobmoellerdev/octi-sync-server/releases/download/v$VERSION; \
docker run -it --rm -v $PWD:/repo gcr.io/projectsigstore/cosign \
  verify-blob --key PUBLIC_KEY \
  --signature $RELEASE/checksums.txt.sig \
  $RELEASE/checksums.txt
```

Now verify the downloaded artifact from above

```shell
PUBLIC_KEY=/repo/cosign.pub \
VERSION=0.2.3-alpha4 \
RELEASE=https://github.com/jakobmoellerdev/octi-sync-server/releases/download/v$VERSION; \
echo "$(wget -qO /dev/stdout $RELEASE/checksums.txt | grep octi-sync-server_${VERSION}_Linux_x86_64.tar.gz)" | \
sha256sum --check
```

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
go install go.uber.org/mock/mockgen@v1.6.0
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
