package services

import (
	"github.com/fsouza/go-dockerclient"
	"log"
)

func CheckNephNginx(endpoint string) error {
	client, _ := docker.NewClient(endpoint)
	//cntnrs, _ := client.ListContainers(docker.ListContainersOptions{All: false})
	// NB Need to specify the container name in a more configurable way!
	cntnrs, _ := client.ListContainers(docker.ListContainersOptions{Filters: map[string][]string{"Names": []string{"/neph-nginx"}}})

	for _, c := range cntnrs {
		log.Println(" Found Neph NGINX, container ID: ", c.ID)
		log.Println("                         status: ", c.Status)
	}

	return nil
}

func ReloadNGINXConfig(endpoint string) error {
	client, _ := docker.NewClient(endpoint)
	//cntnrs, _ := client.ListContainers(docker.ListContainersOptions{All: false})
	// NB Need to specify the container name in a more configurable way!
	cntnrs, _ := client.ListContainers(docker.ListContainersOptions{Filters: map[string][]string{"Names": []string{"/neph-nginx"}}})

	for _, c := range cntnrs {
		client.KillContainer(docker.KillContainerOptions{ID: c.ID, Signal: docker.SIGHUP})
		log.Printf(" Sent SIGHUP to container: %s", c.Names[0])
		log.Printf("                   status: %s", c.Status)
	}
	/*
		for _, img := range imgs {
			log.Println("ID: ", img.ID)
			log.Println("RepoTags: ", img.RepoTags)
			log.Println("Created: ", img.Created)
			log.Println("Size: ", img.Size)
			log.Println("VirtualSize: ", img.VirtualSize)
			log.Println("ParentId: ", img.ParentID)
		}
	*/

	return nil
}
