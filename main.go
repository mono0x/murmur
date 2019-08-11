package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	murmur "github.com/mono0x/murmur/lib"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func run() error {
	app := cli.NewApp()
	app.Name = "murmur"

	app.Commands = []cli.Command{
		{
			Name: "update",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "config", Value: "config.yaml", Usage: "config file"},
			},
			Action: func(c *cli.Context) error {
				configFile := c.String("config")
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
			},
		},
		{
			Name: "serve",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "listen"},
				cli.IntFlag{Name: "port", Value: 8080, EnvVar: "PORT", Usage: "listen port"},
			},
			Action: func(c *cli.Context) error {
				r := chi.NewRouter()
				r.Use(middleware.Recoverer)

				r.Post("/update", func(w http.ResponseWriter, r *http.Request) {
					config, err := murmur.LoadConfig(r.Body)
					if err != nil {
						log.Println(err)
						w.WriteHeader(http.StatusBadRequest)
						return
					}

					source, err := config.NewSource()
					if err != nil {
						log.Println(err)
						w.WriteHeader(http.StatusInternalServerError)
						return
					}

					sink, err := config.NewSink()
					if err != nil {
						log.Println(err)
						w.WriteHeader(http.StatusInternalServerError)
						return
					}

					defer sink.Close()

					notifier := murmur.NewNotifier(source, sink)
					if err := notifier.Notify(); err != nil {
						log.Println(err)
						w.WriteHeader(http.StatusInternalServerError)
						return
					}

					w.WriteHeader(http.StatusOK)

				})

				listen := c.String("listen")
				if listen == "" {
					listen = ":" + strconv.Itoa(c.Int("port"))
				}

				return http.ListenAndServe(listen, r)
			},
		},
	}

	return app.Run(os.Args)
}

func main() {
	log.SetFlags(log.Lshortfile)

	http.DefaultTransport.(*http.Transport).TLSHandshakeTimeout = 30 * time.Second

	if err := run(); err != nil {
		log.Fatalf("%+v\n", err)
	}
}
