# Start by building the application.
FROM golang:1.19 as build

WORKDIR /go/src/app
COPY . .

ENV CGO_ENABLED=0

RUN go mod download

RUN go vet -v

RUN go test -v

RUN go build -o /go/bin/app

# Now copy it into our base image.
FROM gcr.io/distroless/static-debian11

ENV GIN_MODE=release

COPY --from=build /go/bin/app /
COPY --from=build /go/src/app/config.yml /

CMD ["/app"]