package main

import (
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	log "github.com/Sirupsen/logrus"
	"github.com/bass3m/mush/handler"
	"github.com/bass3m/mush/resource"
	"github.com/julienschmidt/httprouter"
	"gopkg.in/alecthomas/kingpin.v2"
)

func main() {
	var (
		app = kingpin.New(filepath.Base(os.Args[0]), "mush")

		listenAddress = app.Flag("web.listen-address", "Address to listen on for the web interface and API.").Default(":9723").String()
		routePrefix   = app.Flag("web.route-prefix", "Prefix for the internal routes of web endpoints.").Default("").String()
	)
	app.HelpFlag.Short('h')
	kingpin.MustParse(app.Parse(os.Args[1:]))

	if *routePrefix == "/" {
		*routePrefix = ""
	}
	if *routePrefix != "" {
		*routePrefix = "/" + strings.Trim(*routePrefix, "/")
	}

	log.Infoln("Starting mush")
	log.Debugf("Prefix path is '%s'", *routePrefix)

	flags := map[string]string{}
	for _, f := range app.Model().Flags {
		flags[f.Name] = f.Value.String()
	}

	// create resource manager
	u, err := url.Parse("http://localhost:9091")
	if err != nil {
		log.Fatal(err)
	}
	rm, err := resource.NewResourceManager(resource.Least)
	if err != nil {
		log.Fatal(err)
	}
	rm.AddResource(resource.Resource{Client: &http.Client{},
		URL:      u,
		Jobs:     []resource.Job{}, // do i need these ?
		JobsSent: 0,                // same as above
	})
	log.Infof("After add Resources %v ", rm)

	router := httprouter.New()
	router.GET("/", handler.Index)

	pushAPIPath := *routePrefix + "/metrics"
	router.PUT(pushAPIPath+"/job/:job/*labels", handler.Push(rm))
	router.POST(pushAPIPath+"/job/:job/*labels", handler.Push(rm))
	router.DELETE(pushAPIPath+"/job/:job/*labels", handler.Delete)
	router.PUT(pushAPIPath+"/job/:job", handler.Push(rm))
	router.POST(pushAPIPath+"/job/:job", handler.Push(rm))
	router.DELETE(pushAPIPath+"/job/:job", handler.Delete)
	router.GET(*routePrefix+"/status", handler.Status)
	l, err := net.Listen("tcp", *listenAddress)
	if err != nil {
		log.Fatal(err)
	}
	go interruptHandler(l)
	err = (&http.Server{Addr: *listenAddress, Handler: router}).Serve(l)
	log.Errorln("Mush HTTP server stopped:", err)
}

func interruptHandler(l net.Listener) {
	notifier := make(chan os.Signal)
	signal.Notify(notifier, os.Interrupt, syscall.SIGTERM)
	<-notifier
	log.Info("Mush Received SIGINT/SIGTERM; exiting ...")
	l.Close()
}
