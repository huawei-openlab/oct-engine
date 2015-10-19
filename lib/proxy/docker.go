package proxy

import (
	"fmt"
	"os"
	"os/exec"
)

type DockerContainer struct {
	ContainerCommon
}

func (dc DockerContainer) Build() bool {
	cmd := fmt.Sprintf("docker build -t %s", dc.Name)
	fmt.Println("Docker build: ", cmd)

	os.Chdir(dc.BuildDir)
	c := exec.Command("/bin/sh", "-c", cmd)
	c.Run()

	return true
}

func (dc DockerContainer) Pull() bool {
	cmd := fmt.Sprintf("docker pull %s", dc.Name)
	fmt.Println("Docker build: ", cmd)

	os.Chdir(dc.BuildDir)
	c := exec.Command("/bin/sh", "-c", cmd)
	c.Run()

	return true
}

func (dc DockerContainer) Deploy() bool {
	return true
}

func (dc DockerContainer) Run() bool {
	return true
}

func (dc DockerContainer) Collect() bool {
	return true
}

func (dc DockerContainer) Destroy() bool {
	return true
}

func (dc DockerContainer) Status() string {
	return "{\"status\": \"ok\"}"
}
