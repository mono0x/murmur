# murmur

[![Build Status](https://travis-ci.org/mono0x/murmur.svg?branch=master)](https://travis-ci.org/mono0x/murmur)

## How to Build

```
make setup
make
```

## Usage

### CLI mode

```
murmur update --config config.yaml
```

### Server mode

#### Server

```
murmur serve --listen localhost:8080
```

#### Client

```
curl -X POST http://localhost:8080/update --data @config.yaml
```