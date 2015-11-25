package main

import (
	"../../liboct"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/drone/routes"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
)

type OCTDConfig struct {
	TSurl string
	Port  int
	Debug bool

	Class           string
	Distribution    string
	ContainerDaemon string
	ContainerClient string
}

const OCTDCacheDir = "/tmp/.octd_cache"

var pubConfig OCTDConfig

func GetTaskReport(w http.ResponseWriter, r *http.Request) {
	//TODO
	var realURL string
	filename := r.URL.Query().Get("File")
	id := r.URL.Query().Get(":ID")

	_, err := os.Stat(filename)
	if err == nil {
		//absolute path
		realURL = filename
	} else {
		realURL = path.Join(GetWorkingDir(id), filename)
	}
	file, err := os.Open(realURL)
	defer file.Close()
	if err != nil {
		//FIXME: add to head
		w.Write([]byte("Cannot open the file: " + realURL))
		return
	}

	buf := bytes.NewBufferString("")
	buf.ReadFrom(file)

	w.Write([]byte(buf.String()))
}

func ReceiveTask(w http.ResponseWriter, r *http.Request) {
	var ret liboct.HttpRet
	realURL, params := liboct.ReceiveFile(w, r, OCTDCacheDir)

	fmt.Println(params)

	//TODO: add id to a database?
	if _, ok := params["id"]; ok {
		//The real url may not be the test case format, could be only files
		if strings.HasSuffix(realURL, ".tar.gz") {
			liboct.UntarFile(realURL, strings.TrimSuffix(realURL, ".tar.gz"))
		}
		ret.Status = liboct.RetStatusOK
	} else {
		ret.Status = "Failed"
		ret.Message = "Cannot find the task id"
	}

	ret_string, _ := json.Marshal(ret)
	w.Write([]byte(ret_string))
	return
}

func RunCommand(action liboct.TestActionCommand, id string) bool {
	dir := GetWorkingDir(id)
	if pubConfig.Debug {
		fmt.Println("Run the command < ", action.Command, ">  in ", dir)
	}
	//check it since some case only has a config.json
	liboct.PreparePath(dir, "")

	if pubConfig.Class == "os" {
		var sh string
		//For 'os', the dir is the way to differentiate tasks
		switch liboct.TestAction(action.Action) {
		case liboct.TestActionDeploy:
			sh = action.Command
		case liboct.TestActionRun:
			sh = action.Command
		case liboct.TestActionCollect:
			return true
		case liboct.TestActionDestroy:
			//TODO: remove the dir
			return true
		}
		return liboct.ExecSH(sh, dir)
	} else if pubConfig.Class == "container" {
		var sh string
		//For 'container', the 'dir' is the way to store test files and use 'id' to differentiate tasks
		clientCommand := pubConfig.ContainerClient
		switch liboct.TestAction(action.Action) {
		case liboct.TestActionDeploy:
			//TODO: mount to /test dir, add to spec!
			sh = fmt.Sprintf("%s -v %s setID %s", clientCommand, dir, "/test", id)
		case liboct.TestActionRun:
			sh = fmt.Sprintf("%s start %s", clientCommand, id)
		case liboct.TestActionCollect:
			return true
		case liboct.TestActionDestroy:
			sh = fmt.Sprintf("%s remove %s", clientCommand, id)
		}

		return liboct.ExecSH(sh, dir)
	}

	return false
}

//TODO: the working dir should be defined in the spec.
func GetWorkingDir(id string) string {
	return path.Join("/tmp/.octd_cache", id)
}

func PostTaskAction(w http.ResponseWriter, r *http.Request) {
	var ret liboct.HttpRet
	result, _ := ioutil.ReadAll(r.Body)
	fmt.Println("Post task action ", string(result))
	r.Body.Close()
	action, err := liboct.ActionCommandFromString(string(result))
	if err != nil {
		ret.Status = liboct.RetStatusFailed
		ret.Message = "Invalid action command"
		retInfo, _ := json.MarshalIndent(ret, "", "\t")
		w.Write([]byte(retInfo))
		return
	}

	id := r.URL.Query().Get(":ID")

	if RunCommand(action, id) {
		ret.Status = liboct.RetStatusOK
	} else {
		ret.Status = liboct.RetStatusFailed
	}
	ret_string, _ := json.Marshal(ret)
	w.Write([]byte(ret_string))
}

func RegisterToTestServer() {
	post_url := pubConfig.TSurl + "/resource"

	//TODO
	//Seems there will be lots of coding while getting the system info
	//Using config now.

	content := liboct.ReadFile("./host.conf")
	fmt.Println(content)
	ret := liboct.SendCommand(post_url, []byte(content))
	fmt.Println(ret)
}

func init() {
	of, err := os.Open("octd.conf")
	if err != nil {
		fmt.Errorf("Cannot find octd.conf.")
		return
	}
	defer of.Close()

	if err = json.NewDecoder(of).Decode(&pubConfig); err != nil {
		fmt.Errorf("Error in parse octd.conf")
		return
	}

	if pubConfig.Class == "container" {
		cmd := exec.Command("/bin/sh", "-c", pubConfig.ContainerDaemon)
		cmd.Stdin = os.Stdin
		if _, err := cmd.CombinedOutput(); err != nil {
			fmt.Errorf("Error in running container daemon %s.", pubConfig.ContainerDaemon)
			return
		}
	}
	RegisterToTestServer()
}

func main() {
	var port string
	port = fmt.Sprintf(":%d", pubConfig.Port)

	mux := routes.New()
	mux.Post("/task", ReceiveTask)
	mux.Post("/task/:ID", PostTaskAction)
	mux.Get("/task/:ID/report", GetTaskReport)

	http.Handle("/", mux)
	fmt.Println("Start to listen ", port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
