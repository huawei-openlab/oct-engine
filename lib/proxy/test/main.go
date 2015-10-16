package main

import (
	"../"
	"../../libocit"
	"fmt"
)

func main() {
	containerFile := "container.conf"
	value := libocit.ReadFile(containerFile)
	container, ok := proxy.ContainerNew(value)
	if ok == false {
		fmt.Println("Cannot get container content.")
		return
	}
	container.Build()
	/*
		container.Status()

		container.Pull()
		container.Deploy()
		container.Run()

		container.Collect()
		container.Destroy()
	*/
}
