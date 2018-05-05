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
	// XXX can we get rid of this ?
	GetResources() []Resource
	SetResources([]Resource)
	DeleteJob(string, *url.URL) (Resource, error)
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
	//	work     chan Job
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
			r.JobsSent++
			log.Infof("Found Job %v at index %d JobsSent %d", job, i, r.JobsSent)
			return i, nil
		}
	}
	return -1, fmt.Errorf("Job: Remote %v URL %v not found", ra, u.String())
}

func (r *Resource) DeleteJobAt(i int) {
	log.Infof("Jobs is %v", r.Jobs)
	jobs := append(r.Jobs[:i], r.Jobs[i+1:]...)
	r.Jobs = jobs
}

func FindResource(rm Balancer, remoteAddr string, u *url.URL) (Resource, error) {
	rs := rm.GetResources()
	host := strings.Split(remoteAddr, ":")[0]
	for _, r := range rs {
		// remoteAddr is host:port
		if _, err := r.FindJob(host, u); err == nil {
			log.Infof("Found existing resource %v for host %v", r, host)
			return r, nil
		}
	}
	// otherwise find a resource to handle job
	job := Job{addr: host, URL: u}
	if r, err := rm.Balance(job); err == nil {
		log.Infof("Found new resource %v for new job: %v", r, job)
		return r, nil
	}
	return Resource{}, fmt.Errorf("No resource found for Job %v", job)
}

func (rm *LeastResourceManager) DeleteJob(remoteAddr string, u *url.URL) (Resource, error) {
	rs := rm.GetResources()
	host := strings.Split(remoteAddr, ":")[0]
	for i, r := range rs {
		// remoteAddr is host:port
		if j, err := r.FindJob(host, u); err == nil {
			log.Infof("Found existing resource %v for host %v", r, host)
			log.Infof("Deleting job %v at index %d", r, j)
			r.DeleteJobAt(j)
			log.Infof("Jobs now %v", r.Jobs)
			//rm.SetResources(rs)
			rm.resources[i].Jobs = r.Jobs
			log.Infof("Resources now %v", rm.GetResources())
			return r, nil
		}
	}
	return Resource{}, fmt.Errorf("No resource found for remoteAddr %v url %v", remoteAddr, u.String())
}

func DeleteJob(rm Balancer, remoteAddr string, u *url.URL) (Resource, error) {
	return rm.DeleteJob(remoteAddr, u)
	//rs := rm.GetResources()
	//host := strings.Split(remoteAddr, ":")[0]
	//for _, r := range rs {
	//	// remoteAddr is host:port
	//	if i, err := r.FindJob(host, u); err == nil {
	//		log.Infof("Found existing resource %v for host %v", r, host)
	//		log.Infof("Deleting job %v at index %d", r, i)
	//		r.DeleteJobAt(i)
	//		log.Infof("Jobs now %v", r.Jobs)
	//		//rm.SetResources(rs)
	//		log.Infof("Resource now %v", rm.GetResources())
	//		return r, nil
	//	}
	//}
	//	return Resource{}, fmt.Errorf("No resource found for remoteAddr %v url %v", remoteAddr, u.String())
}

func AddResource(rm Balancer, r Resource) {
	rs := rm.GetResources()
	rs = append(rs, r)
	rm.SetResources(rs)
	log.Debugf("Added resource: %v Now %v", r, rm.GetResources())
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

func (rm *RandomResourceManager) GetResources() []Resource {
	return rm.resources
}

func (rm *RandomResourceManager) SetResources(rs []Resource) {
	rm.resources = rs
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
}

func (rm *LeastResourceManager) Balance(j Job) (Resource, error) {
	// initilize min to len of first resource's jobs
	minIdx := 0
	min := len(rm.resources[minIdx].Jobs)
	for i, r := range rm.resources {
		// count jobs belonging to this resource
		log.Debugf("Balance i %v r %v", i, r)
		if len(r.Jobs) <= min {
			minIdx = i
			min = len(r.Jobs)
		}
	}
	// add job to resource
	jobs := append(rm.resources[minIdx].Jobs, j)
	rm.resources[minIdx].Jobs = jobs
	rm.resources[minIdx].JobsSent++
	log.Infof("Least used resource: %v", rm.resources[minIdx])
	return rm.resources[minIdx], nil
}

func (rm *RandomResourceManager) DeleteJob(remoteAddr string, u *url.URL) (Resource, error) {
	rs := rm.GetResources()
	host := strings.Split(remoteAddr, ":")[0]
	for _, r := range rs {
		// remoteAddr is host:port
		if i, err := r.FindJob(host, u); err == nil {
			log.Infof("Found existing resource %v for host %v", r, host)
			log.Infof("Deleting job %v at index %d", r, i)
			r.DeleteJobAt(i)
			log.Infof("Jobs now %v", r.Jobs)
			//rm.SetResources(rs)
			log.Infof("Resource now %v", rm.GetResources())
			return r, nil
		}
	}
	return Resource{}, fmt.Errorf("No resource found for remoteAddr %v url %v", remoteAddr, u.String())
}
