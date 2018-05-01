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

const (
	Random = iota
	Least
)

type Balancer interface {
	Balance(Job) (Resource, error)
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

type ResourceManager struct {
	resources []Resource
	balancer  Balancer
}

type LeastResourceManager struct {
	RM ResourceManager
}
type RandomResourceManager struct {
	RM ResourceManager
}

func NewResourceManager(balancer int) (*ResourceManager, error) {
	rm := &ResourceManager{}
	switch balancer {
	case Least:
		rm.balancer = &LeastResourceManager{}
		return rm, nil
	case Random:
		rm.balancer = &RandomResourceManager{}
		return rm, nil
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

func (rm *ResourceManager) FindResource(remoteAddr string, u *url.URL) (Resource, error) {
	for _, r := range rm.resources {
		if _, err := r.FindJob(remoteAddr, u); err == nil {
			return r, nil
		}
	}
	// otherwise find a resource to handle job
	job := Job{addr: remoteAddr, URL: u}
	log.Infof("Find Resource rm %v", rm)
	log.Infof("Find Resource balancer  %v", rm.balancer)
	if r, err := rm.balancer.Balance(job); err != nil {
		log.Infof("Found resource %v for new job: %v", r, job)
		return r, nil
	}
	return Resource{}, fmt.Errorf("No resource found for Job %v", job)
}

func (rm *ResourceManager) AddResource(r Resource) {
	rs := append(rm.resources, r)
	rm.resources = rs
	log.Infof("Added resource: %v Now %v", r, rm)
}

func (rm *RandomResourceManager) Balance(j Job) (Resource, error) {
	i := rand.Intn(len(rm.RM.resources))
	rm.RM.resources[i].JobsSent++
	jobs := append(rm.RM.resources[i].Jobs, j)
	rm.RM.resources[i].Jobs = jobs
	return rm.RM.resources[i], nil
}

func (rm *LeastResourceManager) Balance(j Job) (Resource, error) {
	// initilize min to len of first resource's jobs
	log.Infof("Balance Resources %v", rm)
	log.Infof("Balance Resources RM %v", rm.RM)
	log.Infof("Balance Resources %v len %v", rm.RM.resources, len(rm.RM.resources))
	minIdx := 0
	min := len(rm.RM.resources[minIdx].Jobs)
	log.Infof("Balance min %v minIdx %v", min, minIdx)
	minResource := rm.RM.resources[0]
	for i, r := range rm.RM.resources {
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
	jobs := append(rm.RM.resources[minIdx].Jobs, j)
	rm.RM.resources[minIdx].Jobs = jobs
	rm.RM.resources[minIdx].JobsSent++
	log.Infof("Least used resource: %v Now jobs has %v", minResource, rm.RM.resources[minIdx].Jobs)
	return minResource, nil
}
