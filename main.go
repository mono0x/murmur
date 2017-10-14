package main

import (
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/mono0x/murmur/lib"
)

func run() error {
	if transport, ok := http.DefaultTransport.(*http.Transport); ok {
		transport.TLSHandshakeTimeout = 30 * time.Second
	} else {
		return errors.New("Change TLSHandshakeTimeout failed")
	}

	var configFile string
	flag.StringVar(&configFile, "config", "config.yaml", "string config")
	flag.Parse()

	f, err := os.Open(configFile)
	if err != nil {
		return err
	}
	defer f.Close()

	config, err := murmur.LoadConfig(f)
	if err != nil {
		return err
	}

	source, err := config.NewSource()
	if err != nil {
		return err
	}

	sink, err := config.NewSink()
	if err != nil {
		return err
	}
	defer sink.Close()

	notifier := murmur.NewNotifier(source, sink)
	return notifier.Notify()
}

func main() {
	log.SetFlags(log.Lshortfile)

	if err := run(); err != nil {
		log.Fatal(err)
	}
}
