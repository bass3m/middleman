package dockerapi

import (
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/bass3m/middleman/config"
	"github.com/fsouza/go-dockerclient"
	"strconv"
	"time"
)

type Action int

const (
	Start = Action(iota)
	Stop
)

type Event struct {
	action string
	name   string
	id     string
	uri    string
}

// XXX add atomic counter for id ?

func dockerEventListener(cfg *config.Config, resourceChan chan<- *Event) {
	events := make(chan *docker.APIEvents)
	client := cfg.Client
	label := cfg.FileConfig.Resources.Docker.Label
	network := cfg.FileConfig.Resources.Docker.Network
	uri := ""

	err := client.AddEventListener(events)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		err := client.RemoveEventListener(events)
		if err != nil {
			log.Fatal(err)
		}
		close(events)
	}()
EventLoop:
	for {
		select {
		case event := <-events:
			if event != nil {
				val, ok := event.Actor.Attributes["middleman.resource"]
				if ok == true && val == label {
					log.Debugf("Got docker event: %+v", event)
					id := ""
					switch event.Action {
					case "start":
						id = event.Actor.ID
						uri, err = GetDockerResourceByID(client, network, id)
						if err != nil {
							log.Errorf("Failed to get resource of container %+v err %+v",
								id, err)
							continue EventLoop

						}
					case "die":
						id = event.Actor.ID
					}
					if id != "" {
						resourceChan <- &Event{action: event.Action,
							name: event.Actor.Attributes["name"],
							uri:  uri, id: id}
					}
				}
			}
		}
	}
	return
}

func SetupDocker(c *config.Config, resourceChan chan<- *Event) error {

	if c.FileConfig.Resources.Docker.Endpoint == "" {
		log.Fatal("Docker is enabled but no endpoint for docker API specified")
		return errors.New("Docker enabled but no endpoint for docker API specified")
	}
	endpoint := c.FileConfig.Resources.Docker.Endpoint
	client, err := docker.NewClient(endpoint)
	if err != nil {
		panic(err)
	}

	c.Client = client
	go dockerEventListener(c, resourceChan)
	return nil
}

// given docker id, get the uri for the resource, should be useful when adding/deleting a resource
// returns the uri
func GetDockerResourceByID(client *docker.Client, network string, id string) (string, error) {
	c, err := client.InspectContainer(id)
	if err != nil {
		log.Errorf("Failed to inspect container %v err: %v", id, err)
		return "", err
	}
	for _, value := range c.NetworkSettings.Ports {
		if len(value) != 0 {
			ip := c.NetworkSettings.Networks[network].IPAddress
			port := value[0].HostPort
			uri := "http://" + ip + ":" + port
			return uri, nil
		}
	}
	return "", nil
}

func GetContainerUris(client *docker.Client, label, network string) (map[string]string, error) {
	uris := map[string]string{}
	cs, err := client.ListContainers(docker.ListContainersOptions{})
	if err != nil {
		log.Warnf("No containers found. error: %v", err)
		return uris, err
	}
	for _, c := range cs {
		for key, value := range c.Labels {
			if key == "middleman.resource" && value == label {
				log.Infof("Found middleman docker resource %+v", c)
				ip := c.Networks.Networks[network].IPAddress
				if ip == "" {
					log.Warnf("IP addr not set yet for container ID %+v", c.ID)
					return map[string]string{}, fmt.Errorf("IP addr not ready for ID %+v", c.ID)
				}
				var port int64 = 0
				for _, p := range c.Ports {
					if p.PublicPort != 0 {
						port = p.PublicPort
					}
				}
				log.Infof("middleman container IP %+v", ip)
				if port != 0 {
					uris["http://"+ip+":"+strconv.FormatInt(port, 10)] = c.ID
				}
			}
		}
	}
	return uris, nil
}

func GetResources(cfg config.FileConfig, client *docker.Client) (map[string]string, error) {
	label := cfg.Resources.Docker.Label
	network := cfg.Resources.Docker.Network
	uris := map[string]string{}
	var err error

	log.Infof("Getting docker resources for label %v network %v", label, network)
	// wait for containers to be available
	for i := 0; i < cfg.Resources.Docker.Retries; i++ {
		uris, err = GetContainerUris(client, label, network)
		if err != nil {
			log.Infof("containers not ready yet, waiting for %+v secs. %+v", cfg.Resources.Docker.RetryTimeout, err)
			time.Sleep(cfg.Resources.Docker.RetryTimeout * time.Second)
		} else {
			break
		}
	}

	return uris, err
}
