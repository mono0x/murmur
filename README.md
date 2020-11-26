# murmur

![test](https://github.com/mono0x/murmur/workflows/test/badge.svg)

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
curl -X POST http://localhost:8080/jobs/exec --data-binary @config.yaml
```
