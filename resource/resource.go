package resource

import (
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
)

type Balancer interface {
	Balance(Job) (Resource, error)
	GetResources() []Resource
	SetResources([]Resource)
}

// Units of work that we hand over to our resources
type Job struct {
	addr string
	URL  *url.URL
}

type Resource struct {
	Client   *http.Client
	URL      *url.URL
	Jobs     []Job
	JobsSent int
}

type LeastResourceManager struct {
	resources []Resource
}
type RandomResourceManager struct {
	resources []Resource
}

func NewResourceManager(balancer string) (Balancer, error) {
	switch balancer {
	case "least":
		return &LeastResourceManager{}, nil
	case "random":
		return &RandomResourceManager{}, nil
	default:
		log.Errorf("Unrecognize balancer option %v", balancer)
		return nil, errors.New(fmt.Sprintf("Unrecognized balancer option %d\n", balancer))

	}
}

func (r *Resource) FindJob(ra string, u *url.URL) (int, error) {
	for i, job := range r.Jobs {
		if strings.Compare(job.addr, ra) == 0 && strings.Compare(job.URL.String(), u.String()) == 0 {
			log.Infof("Found Job %v at index %d", job, i)
			return i, nil
		}
	}
	return -1, fmt.Errorf("Job: Remote %v URL %v not found", ra, u.String())
}

func FindResource(rm Balancer, remoteAddr string, u *url.URL) (Resource, error) {
	rs := rm.GetResources()
	host := strings.Split(remoteAddr, ":")[0]
	log.Infof("Find Resource Current resources %v", rs)
	for _, r := range rs {
		// remoteAddr is host:port
		log.Infof("Find Resource host %v", host)
		if _, err := r.FindJob(host, u); err == nil {
			log.Infof("Found existing resource %v for host %v", r, host)
			return r, nil
		}
	}
	// otherwise find a resource to handle job
	job := Job{addr: host, URL: u}
	log.Infof("Find Resource rm %v", rm)
	if r, err := rm.Balance(job); err == nil {
		log.Infof("Found resource %v for new job: %v", r, job)
		return r, nil
	}
	return Resource{}, fmt.Errorf("No resource found for Job %v", job)
}

func AddResource(rm Balancer, r Resource) {
	rs := rm.GetResources()
	rs = append(rs, r)
	rm.SetResources(rs)
	log.Infof("Added resource: %v Now %v", r, rm.GetResources())
}

func Create(uris []string, algo string) Balancer {
	rm, err := NewResourceManager(algo)
	if err != nil {
		log.Fatal(err)
	}

	for _, rawUrl := range uris {
		u, err := url.Parse(rawUrl)
		if err != nil {
			log.Fatal(err)
		}
		AddResource(rm, Resource{Client: &http.Client{},
			URL:      u,
			Jobs:     []Job{},
			JobsSent: 0,
		})

	}
	return rm
}

func (rm *RandomResourceManager) Balance(j Job) (Resource, error) {
	i := rand.Intn(len(rm.resources))
	rm.resources[i].JobsSent++
	jobs := append(rm.resources[i].Jobs, j)
	rm.resources[i].Jobs = jobs
	return rm.resources[i], nil
}

func (rm *LeastResourceManager) GetResources() []Resource {
	return rm.resources
}
func (rm *LeastResourceManager) SetResources(rs []Resource) {
	rm.resources = rs
	return
}
func (rm *RandomResourceManager) GetResources() []Resource {
	return rm.resources
}
func (rm *RandomResourceManager) SetResources(rs []Resource) {
	rm.resources = rs
	return
}

func (rm *LeastResourceManager) Balance(j Job) (Resource, error) {
	// initilize min to len of first resource's jobs
	log.Infof("Balance Resources %v", rm)
	log.Infof("Balance Resources %v len %v", rm.resources, len(rm.resources))
	minIdx := 0
	min := len(rm.resources[minIdx].Jobs)
	log.Infof("Balance min %v minIdx %v", min, minIdx)
	minResource := rm.resources[0]
	for i, r := range rm.resources {
		// count jobs belonging to this resource
		log.Infof("Balance i %v r %v", i, r)
		if len(r.Jobs) <= min {
			minIdx = i
			min = len(r.Jobs)
			minResource = r
			log.Infof("Found Min resource i %v r %v", minIdx, minResource)
		}
	}
	// add job to resource
	jobs := append(rm.resources[minIdx].Jobs, j)
	rm.resources[minIdx].Jobs = jobs
	rm.resources[minIdx].JobsSent++
	log.Infof("Least used resource: %v Now jobs has %v", minResource, rm.resources[minIdx].Jobs)
	return minResource, nil
}
