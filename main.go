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
	"github.com/bass3m/middleman/config"
	"github.com/bass3m/middleman/dockerapi"
	"github.com/bass3m/middleman/handler"
	"github.com/bass3m/middleman/resource"
	"github.com/julienschmidt/httprouter"
	"gopkg.in/alecthomas/kingpin.v2"
)

func GetResources(c config.Config) (map[string]string, error) {
	if c.FileConfig.Resources.Docker.Enabled == true {
		log.Infof("Getting resources from docker")
		uris, err := dockerapi.GetResources(c.FileConfig, c.Client)
		if err != nil {
			return map[string]string{}, err
		}
		return uris, nil
	} else {
		rs := map[string]string{}
		for _, u := range c.FileConfig.Resources.Uris {
			rs[u] = ""
		}
		return rs, nil
	}
}

func main() {
	var (
		app = kingpin.New(filepath.Base(os.Args[0]), "middleman")

		listenAddress = app.Flag("web.listen-address", "Address to listen on for the web interface and API.").Default(":9723").String()
		routePrefix   = app.Flag("web.route-prefix", "Prefix for the internal routes of web endpoints.").Default("").String()
		configPath    = app.Flag("cfg.path", "Path to YAML configuration file.").Default("/etc/middleman/middleman.yml").String()
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

	c, err := config.ReadConfig(*configPath)
	if err != nil {
		log.Fatal(err)
	}

	// if using docker, set it up, pass it channel so it sends us events ?
	if c.FileConfig.Resources.Docker.Enabled == true {
		resourceChan := make(chan *dockerapi.Event)
		defer func() {
			close(resourceChan)
		}()
		dockerapi.SetupDocker(&c, resourceChan)
		go handleResourceEvents(resourceChan)
	}

	uris, err := GetResources(c)
	if err != nil {
		log.Fatal(err)
	}

	log.Infof("Found the following resources: %v", uris)
	// create resource balancer
	m := resource.CreateBalancer(uris, c.FileConfig.Middleman.Algorithm)

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

func handleResourceEvents(resourceChan <-chan *dockerapi.Event) {
	log.Infof("handleResourceEvents go routine")
	for {
		select {
		case event := <-resourceChan:
			log.Infof("Got resource event: %+v", event)
			// XXX add switch to handle msg
			//resourceChan <- &Event{action: event.Action, name: event.Actor.Atttributes.name, id: event.Actor.ID}
			// should be adding or deleting resources here
			//		m.AddResource(r)
		}
	}
	return
}

func interruptHandler(l net.Listener) {
	notifier := make(chan os.Signal)
	signal.Notify(notifier, os.Interrupt, syscall.SIGTERM)
	<-notifier
	log.Info("Middleman Received SIGINT/SIGTERM; exiting ...")
	l.Close()
}
