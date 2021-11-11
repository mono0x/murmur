FROM golang:1.17-buster AS builder

WORKDIR /go/src/github.com/mono0x/murmur

ADD go.mod go.sum Makefile ./
RUN make download

ADD . ./
RUN make build-linux

FROM gcr.io/distroless/static-debian10

COPY --from=builder /go/src/github.com/mono0x/murmur/murmur.linux /app
CMD ["/app", "serve"]
