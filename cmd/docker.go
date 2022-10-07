package cmd

import (
	"context"
	"io"
	"io/ioutil"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/docker/go-units"
	"github.com/docker/docker/api/types/filters"

	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"
     "path/filepath"
"os"
)

func init() {
	log.SetHandler(cli.Default)
	
}
func startContainer(client *client.Client, containername string, servername string, networkName string, version string, kubeconfig string) {



	context := context.Background()

	imageName := "rancher/k3s:" + version

	ctx := log.WithFields(log.Fields{
		"container":containername,
		"networkName":networkName,
		"version":version,
		"kubeconfig":kubeconfig,
		"image":imageName,
	})

	filter := filters.NewArgs()
	filter.Add("reference", imageName)

	imgs, err := client.ImageList(context, types.ImageListOptions{Filters: filter})
	if err != nil {
		ctx.Fatalf("Can't list images", err)
}

	
	if len(imgs) == 0 {
		ctx.Info("pulling " + imageName)
	reader, err := client.ImagePull(context, imageName, types.ImagePullOptions{})
	if err != nil {
		ctx.Fatalf("Can't pull image", err)
		}
	defer  reader.Close()

	io.Copy(ioutil.Discard, reader)

	}



	var config *container.Config
	var hostConfig *container.HostConfig
	if servername == containername {
		config = &container.Config{
			Image:        imageName,
			AttachStdout: false,
			Tty:          false,
			Hostname:     containername,
			Domainname:   containername,
			Cmd:          []string{"server", "--disable", "traefik", "--disable", "metrics-server", "--node-label=ingress-ready=true"},
			ExposedPorts: nat.PortSet{
				"6443/tcp": struct{}{},
				"443/tcp":  struct{}{},
			},
			Env: []string{
				"K3S_TOKEN=test",
				"K3S_KUBECONFIG_OUTPUT=/output/kubeconfig.yaml",
				"K3S_KUBECONFIG_MODE=666",
			},
		}
		hostConfig = &container.HostConfig{
			Binds: []string{
				kubeconfig + ":/output/kubeconfig.yaml",
			},
			Tmpfs: map[string]string{
				"/run":     "",
				"/var/run": "",
			},
			PortBindings: nat.PortMap{
				"6443/tcp": []nat.PortBinding{
					{
						HostIP:   "0.0.0.0",
						HostPort: "6443",
					},
				},
				"443/tcp": []nat.PortBinding{
					{
						HostIP:   "0.0.0.0",
						HostPort: "443",
					},
				},
			},
			Resources: container.Resources{
				Ulimits: []*units.Ulimit{
					{Name: "nofile", Hard: 65535, Soft: 65535},
					{Name: "nproc", Hard: 65535, Soft: 65535},
				},
			},
			Privileged: true,
		}
	} else {
		config = &container.Config{
			AttachStdout: false,
			AttachStdin:  false,
			AttachStderr: false,
			OpenStdin:    false,
			Tty:          false,
			Image:        imageName,
			Hostname:     containername,
			Domainname:   containername,
			Env: []string{
				"K3S_TOKEN=test",
				"K3S_URL=https://" + servername + ":6443",
			},
		}
		hostConfig = &container.HostConfig{
			Tmpfs: map[string]string{
				"/run":     "",
				"/var/run": "",
			},
			Resources: container.Resources{
				Ulimits: []*units.Ulimit{
					{Name: "nofile", Hard: 65535, Soft: 65535},
					{Name: "nproc", Hard: 65535, Soft: 65535},
				},
			},
			Privileged: true,
		}
	}
	networkConfig := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			networkName: {},
		},
	}
	ctx.Debug(containername+" create started")
	
	resp, err := client.ContainerCreate(context, config, hostConfig, networkConfig, nil, containername)
	if err != nil {
		ctx.Fatalf("Can't create container", err)

		}
	ctx.Debug(containername+" created")

	if err := client.ContainerStart(context, resp.ID, types.ContainerStartOptions{}); err != nil {
		ctx.Fatalf("Can't start container", err)
		}
	ctx.Info(containername+" started")


}
func stopAndRemoveContainer(client *client.Client, containername string)  {
	ctx := log.WithFields(log.Fields{
		"container":containername,
	})
	context := context.Background()
	if err := client.ContainerStop(context, containername, nil); err != nil {
		ctx.Fatalf("Can't stop container", err)

	}
	removeOptions := types.ContainerRemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	}
	if err := client.ContainerRemove(context, containername, removeOptions); err != nil {
		ctx.Fatalf("Can't remove container", err)
	}
}
func listContainers(client *client.Client) []string {
	ctx := log.WithFields(log.Fields{
	})
	context := context.Background()
	containers, err := client.ContainerList(context, types.ContainerListOptions{All: true})
	if err!=nil{
		ctx.Fatalf("Can't list containers", err)
	}
	var arr []string
	for _, container := range containers {
		arr = append(arr, container.Names[0])
	}
	return arr
}
func createNetwork(client *client.Client, networkName string) string {
	ctx := log.WithFields(log.Fields{
		"networkName":networkName,
	})

	networkList, err := client.NetworkList(context.Background(), types.NetworkListOptions{})
	if err!=nil{
		ctx.Fatalf("Can't list networks", err)
	}
	var networkID string
	for _, net := range networkList {
		if net.Name == networkName {
			networkID = net.ID
		}
	}
	if networkID == "" {
		resp, err := client.NetworkCreate(context.Background(), networkName, types.NetworkCreate{
			CheckDuplicate: true,
		})
		if err!=nil{
		ctx.Fatalf("Can't create network", err)
		}
		ctx.Info("Created "+networkName+" network")

		networkID = resp.ID
	}else{
		ctx.Info("Reusing "+networkName+" network")

	}
	return networkID
}

func clusterUp(name string, kubeconfig string, version string) {
	networkName := name + "-net"
	server_name := name + "-server"
	worker_name := name + "-worker"
		ctx := log.WithFields(log.Fields{
		"name": name,
		"kubeconfig":kubeconfig,
		"version":version,
		"networkName":networkName,
	})
	// client, err := client.NewEnvClient()

	client, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())


	if err != nil {
		ctx.Fatalf("Docker client not created", err)

	}
	createNetwork(client,networkName)
	if Contains(listContainers(client), "/"+server_name) {
		ctx.Info(server_name + " already exist (use --recreate or k3dev down to remove)")
	}else{
		if _, err := os.Stat(kubeconfig); err == nil {
			ctx.Warn("File exists")
			removeFile(kubeconfig)
		}

		if err := os.MkdirAll(filepath.Dir(kubeconfig), os.ModePerm); err != nil {
			ctx.Fatalf("could not create folder: %v", err)
	
		}

		_, err := os.Create(kubeconfig)
		if err != nil {
			ctx.Fatalf("%s", err)

		} else {
			ctx.Debug("Empty file created")
		}


		ctx.Info("Starting server " + server_name)

	startContainer(client, server_name, server_name, networkName, version, kubeconfig)

	if Contains(listContainers(client), "/"+worker_name){
		stopAndRemoveContainer(client , worker_name)
	}

	}
	if Contains(listContainers(client), "/"+worker_name) {
		ctx.Info(worker_name + " already exist (use --recreate or k3dev down to remove)")

	}else{
		ctx.Info("Starting server " + worker_name)
		startContainer(client, worker_name, server_name, networkName, version, kubeconfig)

	}
}
