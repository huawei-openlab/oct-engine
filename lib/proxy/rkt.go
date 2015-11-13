package proxy

import (
	"fmt"
)

type RKTContainer struct {
	ContainerCommon
}

func (rkt RKTContainer) Build() string {
	fmt.Println("RKT build", rkt.Distribution)
	return "ok"
}

func (rkt RKTContainer) Pull() bool {
	return true
}

func (rkt RKTContainer) Deploy() bool {
	return true
}

func (rkt RKTContainer) Run() bool {
	return true
}

func (rkt RKTContainer) Collect() bool {
	return true
}

func (rkt RKTContainer) Destroy() bool {
	return true
}

func (rkt RKTContainer) Status() string {
	return "{\"status\": \"ok\"}"
}
