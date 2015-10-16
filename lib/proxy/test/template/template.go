package main

import (
	"fmt"
)

type TemplateContainer struct {
	ContainerCommon
}

func (temp TemplateContainer) Build() bool {
	fmt.Println("Template build", temp.Distribution)
	return true
}

func (temp TemplateContainer) Pull() bool {
	return true
}

func (temp TemplateContainer) Deploy() bool {
	return true
}

func (temp TemplateContainer) Run() bool {
	return true
}

func (temp TemplateContainer) Collect() bool {
	return true
}

func (temp TemplateContainer) Destroy() bool {
	return true
}

func (temp TemplateContainer) Status() string {
	return "{\"status\": \"ok\"}"
}
