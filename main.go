package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"gopkg.in/yaml.v2"
)

var stack string

var networkIDMapping = make(map[string]string)

// DeployStruct is the place holder for the service
type DeployStruct struct {
	Mode     string `json:"mode,omitempty" yaml:"mode,omitempty"`
	Replicas uint64 `json:"replicas,omitempty" yaml:"replicas,omitempty"`
}

// ServiceStruct is the place holder for the service
type ServiceStruct struct {
	Image    string       `json:"image,omitempty" yaml:"image,omitempty"`
	Command  string       `json:"command,omitempty" yaml:"command,omitempty"`
	Ports    []string     `json:"ports,omitempty" yaml:"ports,omitempty"`
	Networks []string     `json:"networks,omitempty" yaml:"networks,omitempty"`
	Volumes  []string     `json:"volumes,omitempty" yaml:"volumes,omitempty"`
	Deploy   DeployStruct `json:"deploy,omitempty" yaml:"deploy,omitempty"`
}

// ServicesMap is the place holder for the service
type ServicesMap map[string]ServiceStruct

// ComposeFileStruct is the place holder for the service
type ComposeFileStruct struct {
	Version  string      `json:"version,omitempty" yaml:"version,omitempty"`
	Services ServicesMap `json:"services,omitempty" yaml:"services,omitempty"`
}

func main() {

	// CLI flags, -n & -o
	var namespace, output string
	flag.StringVar(&namespace, "n", "", "Stack Namespace")
	flag.StringVar(&output, "o", "", "Output File Name")

	flag.Parse()

	if namespace != "" {
		stack = namespace
	} else {
		fmt.Println("[ERROR] Namespace required, try: russell -n portainer")
		return
	}

	// Actual Logic
	newServiceMap := make(ServicesMap)
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}

	services, err := cli.ServiceList(context.Background(), types.ServiceListOptions{})
	if err != nil {
		panic(err)
	}

	for _, service := range services {

		stackNamespace := service.Spec.TaskTemplate.ContainerSpec.Labels["com.docker.stack.namespace"]

		if stackNamespace == stack {

			var newService ServiceStruct

			newService.Deploy.Replicas = 0
			newService.Deploy.Mode = "global"
			newService.Image = service.Spec.TaskTemplate.ContainerSpec.Image
			serviceName := strings.Replace(service.Spec.Annotations.Name, stack+"_", "", 1)

			if service.Spec.Mode.Replicated != nil {
				newService.Deploy.Mode = "replicated"
				newService.Deploy.Replicas = *service.Spec.Mode.Replicated.Replicas
			}

			// Volume Mounts
			for _, volume := range service.Spec.TaskTemplate.ContainerSpec.Mounts {
				if strings.HasPrefix(volume.Source, stack+"_") {
					volume.Source = strings.Replace(volume.Source, stack+"_", "", 1)
				}
				volumeString := volume.Source + ":" + volume.Target
				newService.Volumes = append(newService.Volumes, volumeString)
			}

			// Networks
			for _, network := range service.Spec.TaskTemplate.Networks {
				networkName := getNetworkName(network.Target, cli)

				if networkName != "n/a" && strings.HasPrefix(networkName, stack+"_") {
					networkName = strings.Replace(networkName, stack+"_", "", 1)
					newService.Networks = append(newService.Networks, networkName)
				}
			}

			// Ports
			for _, port := range service.Spec.EndpointSpec.Ports {
				portString := fmt.Sprint(port.PublishedPort) + ":" + fmt.Sprint(port.TargetPort)
				newService.Ports = append(newService.Ports, portString)
			}

			if len(service.Spec.TaskTemplate.ContainerSpec.Args) > 0 {
				newService.Command = strings.Join(service.Spec.TaskTemplate.ContainerSpec.Args, " ")
			}

			newServiceMap[serviceName] = newService
		}
	}

	newComposeFile := ComposeFileStruct{
		Version:  "3.2",
		Services: newServiceMap,
	}

	res, err := yaml.Marshal(newComposeFile)

	if err != nil {
		fmt.Println(err)
		return
	}

	if output != "" {
		err = ioutil.WriteFile(output, res, 0644)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("[INFO] DockerCompose file saved; try: cat " + output)
	} else {
		fmt.Println(string(res))
		return
	}
}

func getNetworkName(networkID string, cli *client.Client) string {
	networkName := "n/a"
	networks, err := cli.NetworkList(context.Background(), types.NetworkListOptions{})
	if err != nil {
		panic(err)
	}

	if networkIDMapping[networkID] != "" {
		networkName = networkIDMapping[networkID]
	} else {
		for _, network := range networks {
			if network.ID == networkID {
				networkName = network.Name
				networkIDMapping[networkID] = networkName
				break
			}
		}
	}
	return networkName
}
