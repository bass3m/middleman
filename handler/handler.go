package handler

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/bass3m/mush/resource"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, "Index page\n")
}

func Status(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, "Status page\n")
}

func Push(rm resource.Balancer) func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		log.Infof("Push RM %v", rm)

		resource, err := resource.FindResource(rm, r.RemoteAddr, r.URL)
		if err != nil {
			log.Errorf("Error %v getting resource for url: %v\n", err, resource.URL)
			return
		}
		req, err := http.NewRequest(r.Method, resource.URL.String()+r.URL.String(), r.Body)
		if err != nil {
			log.Error("Error creating request:", err)
			return
		}

		client := resource.Client
		resp, err := client.Do(req)
		if err != nil {
			log.Error("Error sending to resource:", err)
			return
		}
		defer resp.Body.Close()
		if !(resp.StatusCode >= 200 && resp.StatusCode < 300) {
			log.Error("HTTP status %d", resp.StatusCode)
		}

	}
}

// XXX find which PG to delete from
func Delete(rm resource.Balancer) func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		log.Infof("DELETE job %v", rm)
		for index, param := range ps {
			log.Infof("del index %d param k: %s v: %s", index, param.Key, param.Value)
		}
		resource, err := resource.DeleteJob(rm, r.RemoteAddr, r.URL)
		if err != nil {
			log.Errorf("Error %v deleting resource for url: %v\n", err, resource.URL)
			return
		}
		req, err := http.NewRequest(r.Method, resource.URL.String()+r.URL.String(), r.Body)
		if err != nil {
			log.Error("Error creating request:", err)
			return
		}

		client := resource.Client
		resp, err := client.Do(req)
		if err != nil {
			log.Error("Error sending to resource:", err)
			return
		}
		defer resp.Body.Close()
		if !(resp.StatusCode >= 200 && resp.StatusCode < 300) {
			log.Error("HTTP status %d", resp.StatusCode)
		}
	}
}

func SetupRoutes(router *httprouter.Router, rm resource.Balancer, routePrefix string) {
	router.GET("/", Index)

	pushAPIPath := routePrefix + "/metrics"
	router.PUT(pushAPIPath+"/job/:job/*labels", Push(rm))
	router.POST(pushAPIPath+"/job/:job/*labels", Push(rm))
	router.DELETE(pushAPIPath+"/job/:job/*labels", Delete(rm))
	router.PUT(pushAPIPath+"/job/:job", Push(rm))
	router.POST(pushAPIPath+"/job/:job", Push(rm))
	router.DELETE(pushAPIPath+"/job/:job", Delete(rm))
	router.GET(routePrefix+"/status", Status)
}
