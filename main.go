package main

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	murmur "github.com/mono0x/murmur/lib"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"
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

				return murmur.Execute(config)
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

				r.Post("/jobs/exec", func(w http.ResponseWriter, r *http.Request) {
					config, err := murmur.LoadConfig(r.Body)
					if err != nil {
						log.Printf("%+v\n", err)
						w.WriteHeader(http.StatusBadRequest)
						return
					}

					if err := murmur.Execute(config); err != nil {
						log.Printf("%+v\n", err)
						w.WriteHeader(http.StatusInternalServerError)
						return
					}

					w.WriteHeader(http.StatusOK)
				})

				r.Post("/jobs/exec_bulk", func(w http.ResponseWriter, r *http.Request) {
					configs, err := murmur.LoadBulkConfig(r.Body)
					if err != nil {
						log.Printf("%+v\n", err)
						w.WriteHeader(http.StatusBadRequest)
						return
					}

					eg := errgroup.Group{}
					for _, config := range configs {
						config := config
						eg.Go(func() error {
							return murmur.Execute(config)
						})
					}

					if err := eg.Wait(); err != nil {
						log.Printf("%+v\n", err)
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

	if err := run(); err != nil {
		log.Fatalf("%+v\n", err)
	}
}
