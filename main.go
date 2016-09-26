package main

import (
	"errors"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
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

	configData, err := ioutil.ReadFile(configFile)
	if err != nil {
		return err
	}

	config, err := murmur.LoadConfig(configData)
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
	if err := notifier.Notify(); err != nil {
		return err
	}

	return nil
}

func main() {
	log.SetFlags(log.Lshortfile)

	if err := run(); err != nil {
		log.Fatal(err)
	}
}
