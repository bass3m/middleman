package main

import (
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	log "github.com/Sirupsen/logrus"
	"github.com/bass3m/middleman/handler"
	"github.com/bass3m/middleman/resource"
	"github.com/julienschmidt/httprouter"
	"gopkg.in/alecthomas/kingpin.v2"
)

func main() {
	var (
		app = kingpin.New(filepath.Base(os.Args[0]), "middleman")

		listenAddress = app.Flag("web.listen-address", "Address to listen on for the web interface and API.").Default(":9723").String()
		routePrefix   = app.Flag("web.route-prefix", "Prefix for the internal routes of web endpoints.").Default("").String()
		configPath    = app.Flag("cfg.path", "Path to YAML configuration file.").Default("/etc/middleman/mush.yml").String()
	)
	app.HelpFlag.Short('h')
	kingpin.MustParse(app.Parse(os.Args[1:]))

	if *routePrefix == "/" {
		*routePrefix = ""
	}
	if *routePrefix != "" {
		*routePrefix = "/" + strings.Trim(*routePrefix, "/")
	}

	log.Infoln("Starting middleman")
	log.Debugf("Prefix path is '%s'", *routePrefix)

	flags := map[string]string{}
	for _, f := range app.Model().Flags {
		flags[f.Name] = f.Value.String()
	}

	var c Config
	c.ReadConfig(*configPath)

	// create resource manager
	m := resource.Create(c.Resources.Uris, c.Middleman.Algorithm)

	router := httprouter.New()
	handler.SetupRoutes(router, m, *routePrefix)
	l, err := net.Listen("tcp", *listenAddress)
	if err != nil {
		log.Fatal(err)
	}
	go interruptHandler(l)
	err = (&http.Server{Addr: *listenAddress, Handler: router}).Serve(l)
	log.Errorln("Middleman HTTP server stopped:", err)
}

func interruptHandler(l net.Listener) {
	notifier := make(chan os.Signal)
	signal.Notify(notifier, os.Interrupt, syscall.SIGTERM)
	<-notifier
	log.Info("Middleman Received SIGINT/SIGTERM; exiting ...")
	l.Close()
}
