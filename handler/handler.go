package handler

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/bass3m/middleman/resource"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, "Index page\n")
}

func Status(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, "Status page\n")
}

func Push(m *resource.Manager) func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		resource, err := m.FindResource(r.RemoteAddr, r.URL)
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

func Delete(m *resource.Manager) func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		log.Infof("DELETE job")
		resource, err := m.DeleteJob(r.RemoteAddr, r.URL)
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

func SetupRoutes(router *httprouter.Router, m *resource.Manager, routePrefix string) {
	router.GET("/", Index)

	pushAPIPath := routePrefix + "/metrics"
	router.PUT(pushAPIPath+"/job/:job/*labels", Push(m))
	router.POST(pushAPIPath+"/job/:job/*labels", Push(m))
	router.DELETE(pushAPIPath+"/job/:job/*labels", Delete(m))
	router.PUT(pushAPIPath+"/job/:job", Push(m))
	router.POST(pushAPIPath+"/job/:job", Push(m))
	router.DELETE(pushAPIPath+"/job/:job", Delete(m))
	router.GET(routePrefix+"/status", Status)
}
