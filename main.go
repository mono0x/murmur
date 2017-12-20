package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/mono0x/murmur/lib"
	"github.com/pkg/errors"
)

func run() error {
	var configFile string
	flag.StringVar(&configFile, "config", "config.yaml", "string config")
	flag.Parse()

	f, err := os.Open(configFile)
	if err != nil {
		return errors.WithStack(err)
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

	http.DefaultTransport.(*http.Transport).TLSHandshakeTimeout = 30 * time.Second

	if err := run(); err != nil {
		log.Fatalf("%+v\n", err)
	}
}
