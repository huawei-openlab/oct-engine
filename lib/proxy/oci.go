package proxy

import (
	"fmt"
)

type OCIContainer struct {
	ContainerCommon
}

func (oci OCIContainer) Build() bool {
	fmt.Println("OCI build", oci.Distribution)
	return true
}

func (oci OCIContainer) Pull() bool {
	return true
}

func (oci OCIContainer) Deploy() bool {
	return true
}

func (oci OCIContainer) Run() bool {
	return true
}

func (oci OCIContainer) Collect() bool {
	return true
}

func (oci OCIContainer) Destroy() bool {
	return true
}

func (oci OCIContainer) Status() string {
	return "{\"status\": \"ok\"}"
}
