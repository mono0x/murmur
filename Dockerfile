FROM golang:1.17-buster AS builder

WORKDIR /go/src/github.com/mono0x/murmur

COPY go.mod go.sum Makefile ./
RUN make download

COPY . ./
RUN make build-linux

# hadolint ignore=DL3006
FROM gcr.io/distroless/static-debian10

COPY --from=builder /go/src/github.com/mono0x/murmur/murmur.linux /app
CMD ["/app", "serve"]
