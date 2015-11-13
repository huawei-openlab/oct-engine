package main

import (
	"../"
	"../../liboct"
	"fmt"
)

func main() {
	containerFile := "container.conf"
	value := liboct.ReadFile(containerFile)
	container, ok := proxy.ContainerNew(value)
	if ok == false {
		fmt.Println("Cannot get container content.")
		return
	}
	fmt.Println(container)
	fmt.Println("--------")
	//	container.Hook()
	container.Build()
	container.Name = "name"
	fmt.Println("--------")
	fmt.Println(container)
	/*
		container.Status()

		container.Pull()
		container.Deploy()
		container.Run()

		container.Collect()
		container.Destroy()
	*/
}
