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

//type ResourceManager struct {
//	resources []Resource
//	balancer  Balancer
//}

type LeastResourceManager struct {
	//RM ResourceManager
	//resources []Resource
	//	balancer  Balancer
	resources []Resource
	//Balancer
}
type RandomResourceManager struct {
	//RM ResourceManager
	//resources []Resource
	//	balancer  Balancer
	resources []Resource
	//Balancer
}

//func NewResourceManager(balancer int) (*ResourceManager, error) {
//	rm := &ResourceManager{}
//	switch balancer {
//	case Least:
//		rm.balancer = &LeastResourceManager{}
//		return rm, nil
//	case Random:
//		rm.balancer = &RandomResourceManager{}
//		return rm, nil
//	default:
//		log.Errorf("Unrecognize balancer option %v", balancer)
//		return nil, errors.New(fmt.Sprintf("Unrecognized balancer option %d\n", balancer))
//
//	}
//}
func NewResourceManager(balancer int) (Balancer, error) {
	switch balancer {
	case Least:
		return &LeastResourceManager{}, nil
	case Random:
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
	for _, r := range rs {
		if _, err := r.FindJob(remoteAddr, u); err == nil {
			return r, nil
		}
	}
	// otherwise find a resource to handle job
	job := Job{addr: remoteAddr, URL: u}
	log.Infof("Find Resource rm %v", rm)
	//log.Infof("Find Resource balancer  %v", rm.balancer)
	if r, err := rm.Balance(job); err != nil {
		log.Infof("Found resource %v for new job: %v", r, job)
		return r, nil
	}
	return Resource{}, fmt.Errorf("No resource found for Job %v", job)
}

//func (rm *ResourceManager) FindResource(remoteAddr string, u *url.URL) (Resource, error) {
//	for _, r := range rm.balancer.resources {
//		if _, err := r.FindJob(remoteAddr, u); err == nil {
//			return r, nil
//		}
//	}
//	// otherwise find a resource to handle job
//	job := Job{addr: remoteAddr, URL: u}
//	log.Infof("Find Resource rm %v", rm)
//	log.Infof("Find Resource balancer  %v", rm.balancer)
//	if r, err := rm.balancer.Balance(job); err != nil {
//		log.Infof("Found resource %v for new job: %v", r, job)
//		return r, nil
//	}
//	return Resource{}, fmt.Errorf("No resource found for Job %v", job)
//}

func AddResource(rm Balancer, r Resource) {
	//rs := append(rm.resources, r)
	rs := rm.GetResources()
	rs = append(rs, r)
	rm.SetResources(rs)
	//rm.resources = rs
	log.Infof("Added resource: %v Now %v", r, rm.GetResources())
}

//func (rm *ResourceManager) AddResource(r Resource) {
//	rs := append(rm.balancer.resources, r)
//	rm.balancer.resources = rs
//	log.Infof("Added resource: %v Now %v", r, rm)
//}

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
