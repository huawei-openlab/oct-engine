package main

import (
	"../../../lib/libocit"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path"
)

func createCase(caseName string) {
	sourceUrl := libocit.PreparePath(caseName, "source")
	if len(sourceUrl) == 0 {
		fmt.Println("Cannot prepare the case directory")
		return
	}

	var tc libocit.TestCase
	tc.Name = caseName
	tc.Summary = caseName + " test"
	tc.Version = "0.1.0"
	tc.License = "Apache 2.0"
	tc.Group = "yourgroup/yoursubgroup"
	tc.Owner = "you you@you.com"

	var req libocit.Require
	req.Class = "osA"
	req.Type = "os"
	req.Distribution = "openSUSE/CentOS/Ubuntu"
	req.Version = "12"
	req.Resource = libocit.OSResource{1, "1GB", "1GB"}
	tc.Requires = []libocit.Require{req}

	var dep libocit.Deploy
	dep.Object = "hostA"
	dep.Class = "osA"
	dep.Cmd = "./yourdeploycommand"
	tc.Deploys = []libocit.Deploy{dep}

	var run libocit.Deploy
	run.Object = "hostA"
	run.Cmd = "./yourruncommand -o case.log"
	tc.Run = []libocit.Deploy{run}

	var collect libocit.Collect
	collect.Object = "hostA"
	collect.Files = []string{"case.log"}
	tc.Collects = []libocit.Collect{collect}

	content, _ := json.MarshalIndent(tc, "", "\t")

	f, err := os.Create(path.Join(caseName, "config.json"))
	if err != nil {
		fmt.Println("Cannot create the file ", path.Join(caseName, "config.json"))
		return
	}
	f.Write([]byte(content))
	f.Close()
	fmt.Println("Generate a \"", caseName, "\" bundle.")
}

func main() {
	var caseName = flag.String("n", "", "input the 'case name'")
	flag.Parse()

	if len(*caseName) > 0 {
		createCase(*caseName)
	} else {
		fmt.Println("Please input the case name.")
		return
	}
}
