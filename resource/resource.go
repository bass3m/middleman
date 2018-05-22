package resource

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

type Job struct {
	addr string
	URL  *url.URL
}

type SvrResource struct {
	URI string
	ID  string
}

type Resource struct {
	// XXX want to add an id so it's easier to delete
	Client   *http.Client
	URL      *url.URL
	Jobs     []Job
	JobsSent int
	ID       string
}

type Balancer interface {
	Balance([]*Resource, Job) (Resource, error)
}

type Manager struct {
	Balancer  Balancer
	Resources []*Resource
	mux       sync.Mutex
}

func (m *Manager) Balance(job Job) (Resource, error) {
	return m.Balancer.Balance(m.Resources, job)
}

type LeastManager struct{}

func (r *Resource) JobExists(ra string, u *url.URL) bool {
	i, _ := r.FindJobIdx(ra, u)
	return i > -1
}

func (r *Resource) FindJobIdx(ra string, u *url.URL) (int, error) {
	for i, job := range r.Jobs {
		if strings.Compare(job.addr, ra) == 0 && strings.Compare(job.URL.String(), u.String()) == 0 {
			r.JobsSent++
			log.Debugf("Found Job %v at index %d JobsSent %d", job, i, r.JobsSent)
			return i, nil
		}
	}
	return -1, fmt.Errorf("Job: Remote %v URL %v not found", ra, u.String())
}

func (m Manager) JobExists(ra string, u *url.URL) bool {
	for _, r := range m.Resources {
		if r.JobExists(ra, u) == true {
			return true
		}
	}
	return false
}

func (m *Manager) FindResource(remoteAddr string, u *url.URL) (Resource, error) {
	host := strings.Split(remoteAddr, ":")[0]
	m.mux.Lock()
	defer m.mux.Unlock()
	for _, r := range m.Resources {
		// remoteAddr is host:port
		if r.JobExists(host, u) == true {
			log.Debugf("Found existing resource %v for host %v", r, host)
			return *r, nil
		}
	}
	// otherwise find a resource to handle job
	job := Job{addr: host, URL: u}
	if r, err := m.Balance(job); err == nil {
		log.Debugf("Found new resource %v for new job: %v", r, job)
		return r, nil
	}
	return Resource{}, fmt.Errorf("No resource found for Job %v", job)
}

func (j Job) Print() {
	fmt.Printf("\tJob: Addr: %v URL: %v\n", j.addr, j.URL.String())
}
func (r Resource) Print() {
	for i, j := range r.Jobs {
		fmt.Printf("\tResource: URL %v\n", r.URL.String())
		fmt.Printf("\tJobs at %d:\n", i)
		fmt.Printf("\t===============\n")
		j.Print()
	}

}
func (m Manager) Print() {
	for i, r := range m.Resources {
		fmt.Printf("Resource at %d:\n", i)
		fmt.Printf("===============\n")
		r.Print()
	}
}

func (m *Manager) DeleteJob(remoteAddr string, u *url.URL) (Resource, error) {
	rs := m.Resources
	host := strings.Split(remoteAddr, ":")[0]
	for i, r := range rs {
		// remoteAddr is host:port
		if j, err := r.FindJobIdx(host, u); err == nil {
			log.Debugf("Deleting found existing resource %v:%d at idx %d for host %v\n", r, i, j, host)
			jobs := append(r.Jobs[:j], r.Jobs[j+1:]...)
			m.Resources[i].Jobs = jobs
			m.Print()
			return *r, nil
		}
	}
	return Resource{}, fmt.Errorf("No resource found for remoteAddr %v url %v", remoteAddr, u.String())
}

func (m *Manager) AddResource(strURL string, id string) {
	u, err := url.Parse(strURL)
	if err != nil {
		log.Fatal(err)
	}

	r := &Resource{Client: &http.Client{},
		URL:      u,
		ID:       id,
		Jobs:     []Job{},
		JobsSent: 0}
	rs := append(m.Resources, r)
	m.Resources = rs
	log.Debugf("Added resource: %v Now %v", r, m.Resources)
}

func CreateBalancer(uris map[string]string, algo string) *Manager {
	var m *Manager
	switch algo {
	case "least":
		m = &Manager{Balancer: &LeastManager{}}
		break
	default:
		log.Fatalf("Unrecognized balancer option %v", algo)
	}

	for u, i := range uris {
		m.AddResource(u, i)
	}
	return m
}

func (LeastManager) Balance(resources []*Resource, j Job) (Resource, error) {
	// initilize min to len of first resource's jobs
	minIdx := 0
	min := len(resources[minIdx].Jobs)
	for i, r := range resources {
		// count jobs belonging to this resource
		if len(r.Jobs) <= min {
			minIdx = i
			min = len(r.Jobs)
		}
	}
	// add job to resource
	jobs := append(resources[minIdx].Jobs, j)
	resources[minIdx].Jobs = jobs
	resources[minIdx].JobsSent++
	log.Debugf("Least used resource: %v", resources[minIdx])
	return *resources[minIdx], nil
}
