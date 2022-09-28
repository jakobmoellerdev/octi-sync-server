# Start by building the application.
FROM golang:1.19 as build

WORKDIR /go/src/app

COPY main.go main.go
COPY router.go router.go

COPY middleware/ middleware/
COPY api/ api/
COPY config/ config/
COPY service/ service/
COPY server/ server/

COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /go/bin/app

# Now copy it into our base image.
FROM gcr.io/distroless/static-debian11

COPY --from=build /go/bin/app /app
COPY config.yml /

CMD ["/app", "-config", "config.yml"]