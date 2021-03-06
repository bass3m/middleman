package main

import (
	"github.com/bass3m/middleman/handler"
	"github.com/bass3m/middleman/resource"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

const checkMark = "\u2713"
const ballotX = "\u2717"

var m *resource.Manager
var router *httprouter.Router

func setup(uris []string, algo string) {
	m = resource.Create(uris, algo)

	router = httprouter.New()
	handler.SetupRoutes(router, m, "")
}

func TestConfig(t *testing.T) {
	t.Log("Given the need to test reading config.")
	var c Config
	c.ReadConfig("./middleman_test.yml")

	// create resource manager
	m = resource.Create(c.Resources.Uris, c.Middleman.Algorithm)

	rs := m.Resources
	if len(rs) != 8 {
		t.Fatal("\tShould have created 8 resources", ballotX)
	}
	t.Log("\tShould have created 8 resources", checkMark)
}

func TestSinglePushStatusCode(t *testing.T) {
	setup([]string{"http://localhost:9091"}, "least")
	w := httptest.NewRecorder()
	t.Log("Given the need to test the PUSH endpoint.")
	req, err := http.NewRequest("PUT", "/metrics/job/nodeexporter/instance/myhostname", nil)
	if err != nil {
		t.Fatal("\tShould be able to create a PUT request", ballotX, err)
	}
	t.Log("\tShould be able to create a PUT request", checkMark)
	router.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatal("\tShould receive \"200\"", ballotX, w.Code)
	}
	t.Log("\tShould receive \"200\"", checkMark)
}

func TestBalance1(t *testing.T) {
	setup([]string{
		"http://localhost:9091",
		"http://localhost:1909",
		//"http://localhost:9092",
		//"http://localhost:19092",
		//"http://localhost:9093",
		//"http://localhost:19093",
		//"http://localhost:9094",
		//"http://localhost:19094"
	}, "least")

	w := httptest.NewRecorder()
	t.Log("Given the need to test resource balancing.")
	requests := []string{
		"/metrics/job/nodeexporter/instance/myhostname1",
		"/metrics/job/cadvisor/instance/myhostname1",
		"/metrics/job/nodeexporter/instance/myhostname2",
		"/metrics/job/cadvisor/instance/myhostname2",
		"/metrics/job/nodeexporter/instance/myhostname3",
		"/metrics/job/cadvisor/instance/myhostname3",
		"/metrics/job/nodeexporter/instance/myhostname4",
		"/metrics/job/cadvisor/instance/myhostname4",
	}
	for _, u := range requests {
		req, err := http.NewRequest("PUT", u, nil)
		if err != nil {
			t.Fatal("\tShould be able to create a PUT request", ballotX, err)
		}
		router.ServeHTTP(w, req)
		if w.Code != 200 {
			t.Fatal("\tShould receive \"200\"", ballotX, w.Code)
		}

	}
	rs := m.Resources
	t.Log("Check resources should have 2 resources with 4 jobs each")
	for _, r := range rs {
		if len(r.Jobs) != 4 {
			t.Fatal("\tShould have created 4 jobs", ballotX)
		}

	}
}

func TestBalance2(t *testing.T) {
	setup([]string{
		"http://localhost:9091",
		"http://localhost:1909",
		"http://localhost:9092",
		"http://localhost:19092",
		"http://localhost:9093",
		"http://localhost:19093",
		"http://localhost:9094",
		"http://localhost:19094",
	}, "least")

	w := httptest.NewRecorder()
	t.Log("Given the need to test resource balancing.")
	requests := []string{
		"/metrics/job/nodeexporter/instance/myhostname1",
		"/metrics/job/cadvisor/instance/myhostname1",
		"/metrics/job/nodeexporter/instance/myhostname2",
		"/metrics/job/cadvisor/instance/myhostname2",
		"/metrics/job/nodeexporter/instance/myhostname3",
		"/metrics/job/cadvisor/instance/myhostname3",
		"/metrics/job/nodeexporter/instance/myhostname4",
		"/metrics/job/cadvisor/instance/myhostname4",
	}
	for _, u := range requests {
		req, err := http.NewRequest("PUT", u, nil)
		if err != nil {
			t.Fatal("\tShould be able to create a PUT request", ballotX, err)
		}
		router.ServeHTTP(w, req)
		if w.Code != 200 {
			t.Fatal("\tShould receive \"200\"", ballotX, w.Code)
		}

	}
	rs := m.Resources
	t.Log("Check resources should have 8 resources with 1 job each")
	for _, r := range rs {
		if len(r.Jobs) != 1 {
			t.Fatal("\tShould have created 1 job", ballotX)
		}

	}
}

func TestJobsSent(t *testing.T) {
	setup([]string{
		"http://localhost:9091",
		"http://localhost:1909",
	}, "least")

	w := httptest.NewRecorder()
	t.Log("Given the need to test resource balancing.")
	requests := []string{
		"/metrics/job/nodeexporter/instance/myhostname1",
		"/metrics/job/cadvisor/instance/myhostname1",
		"/metrics/job/nodeexporter/instance/myhostname1",
		"/metrics/job/cadvisor/instance/myhostname1",
		"/metrics/job/nodeexporter/instance/myhostname1",
		"/metrics/job/cadvisor/instance/myhostname1",
	}
	for _, u := range requests {
		req, err := http.NewRequest("PUT", u, nil)
		if err != nil {
			t.Fatal("\tShould be able to create a PUT request", ballotX, err)
		}
		router.ServeHTTP(w, req)
		if w.Code != 200 {
			t.Fatal("\tShould receive \"200\"", ballotX, w.Code)
		}

	}
	rs := m.Resources
	t.Log("Check resources should have 2 resources with 1 job each")
	for _, r := range rs {
		if len(r.Jobs) != 1 {
			t.Fatal("\tShould have created 1 job", ballotX)
		}
		t.Log("Check resources should have 2 resources with 1 job each", checkMark)
		// check jobs sent should be 3
		t.Log("Check resources should have 2 resources with 1 job each with 3 jobs sent")
		if r.JobsSent != 3 {
			t.Fatal("\tShould have 3 jobs sent to each resource", ballotX)
		}

	}
}

func TestDeleteJob(t *testing.T) {
	setup([]string{
		"http://localhost:9091",
		"http://localhost:1909",
	}, "least")

	w := httptest.NewRecorder()
	t.Log("Given the need to test deleting jobs.")
	requests := []string{
		"/metrics/job/nodeexporter/instance/myhostname1",
		"/metrics/job/cadvisor/instance/myhostname1",
	}
	for _, u := range requests {
		req, err := http.NewRequest("PUT", u, nil)
		if err != nil {
			t.Fatal("\tShould be able to create a PUT request", ballotX, err)
		}
		router.ServeHTTP(w, req)
		if w.Code != 200 {
			t.Fatal("\tShould receive \"200\"", ballotX, w.Code)
		}

	}
	req, err := http.NewRequest("DELETE", "/metrics/job/nodeexporter/instance/myhostname1", nil)
	if err != nil {
		t.Fatal("\tShould be able to create a DELETE request", ballotX, err)
	}
	router.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatal("\tShould receive \"200\"", ballotX, w.Code)
	}

	u, err := url.Parse("/metrics/job/nodeexporter/instance/myhostname1")
	if err != nil {
		t.Fatal(err)
	}
	if m.JobExists(req.RemoteAddr, u) {
		t.Fatal("\tJob not deleted", ballotX, req.RemoteAddr, "/metrics/job/nodeexporter/instance/myhostname1")
	}
	t.Log("Was able to delete job successfully", checkMark)
}
