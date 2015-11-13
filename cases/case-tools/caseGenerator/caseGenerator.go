package main

import (
	"../../../lib/liboct"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path"
)

func createCase(caseName string) {
	sourceUrl := liboct.PreparePath(caseName, "source")
	if len(sourceUrl) == 0 {
		fmt.Println("Cannot prepare the case directory")
		return
	}

	var tc liboct.TestCase
	tc.Name = caseName
	tc.Summary = caseName + " test"
	tc.Version = "0.1.0"
	tc.License = "Apache 2.0"
	tc.Group = "yourgroup/yoursubgroup"
	tc.Owner = "you you@you.com"

	var req liboct.Require
	req.Class = "osA"
	req.Type = "os"
	req.Distribution = "openSUSE/CentOS/Ubuntu"
	req.Version = "12"
	req.Resource = liboct.OSResource{1, "1GB", "1GB"}
	tc.Requires = []liboct.Require{req}

	var dep liboct.Deploy
	dep.Object = "hostA"
	dep.Class = "osA"
	dep.Cmd = "./yourdeploycommand"
	tc.Deploys = []liboct.Deploy{dep}

	var run liboct.Deploy
	run.Object = "hostA"
	run.Cmd = "./yourruncommand -o case.log"
	tc.Run = []liboct.Deploy{run}

	var collect liboct.Collect
	collect.Object = "hostA"
	collect.Files = []string{"case.log"}
	tc.Collects = []liboct.Collect{collect}

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
