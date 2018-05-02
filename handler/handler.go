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
		//for i, param := range ps {
		//	log.Infof("push index %d param k: %s v: %s", i, param.Key, param.Value)
		//}
		//log.Infof("request: %s URL %s", r, r.URL)
		//log.Infof("Header %s", r.Header)
		//log.Infof("Push RM %v", rm)

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
			log.Error("Error sending to PG:", err)
			return
		}
		defer resp.Body.Close()
		if !(resp.StatusCode >= 200 && resp.StatusCode < 300) {
			log.Error("HTTP status %d", resp.StatusCode)
		}

	}
}

// XXX find which PG to delete from
func Delete(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fmt.Fprint(w, "Delete\n")
	for index, param := range ps {
		log.Infof("del index %d param k: %s v: %s", index, param.Key, param.Value)
	}
}
