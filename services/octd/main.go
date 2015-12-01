package main

import (
	"../../liboct"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/drone/routes"
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
	filename := r.URL.Query().Get("File")
	id := r.URL.Query().Get(":ID")

	db := liboct.GetDefaultDB()
	taskInterface, err := db.Get(liboct.DBTask, id)
	if err != nil {
		logrus.Warnf("Cannot find the test job %v", id)
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Cannot find the test job " + id))
		return
	}

	task, _ := liboct.TaskFromString(taskInterface.String())
	workingDir := strings.TrimSuffix(task.BundleURL, ".tar.gz")
	realURL := path.Join(workingDir, filename)
	file, err := os.Open(realURL)
	defer file.Close()
	if err != nil {
		logrus.Warnf("Cannot open the file %v", filename)
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Cannot open the file: " + realURL))
		return
	}

	buf := bytes.NewBufferString("")
	buf.ReadFrom(file)

	w.Write([]byte(buf.String()))
}

func ReceiveTask(w http.ResponseWriter, r *http.Request) {
	db := liboct.GetDefaultDB()
	realURL, params := liboct.ReceiveFile(w, r, OCTDCacheDir)

	logrus.Debugf("ReceiveTask %v", realURL)

	if id, ok := params["id"]; ok {
		//The real url may not be the test case format, could be only files
		if strings.HasSuffix(realURL, ".tar.gz") {
			liboct.UntarFile(realURL, strings.TrimSuffix(realURL, ".tar.gz"))
		}
		var task liboct.TestTask
		task.ID = id
		task.BundleURL = realURL
		db.Update(liboct.DBTask, id, task)

		liboct.RenderOK(w, "", nil)
	} else {
		liboct.RenderErrorf(w, fmt.Sprintf("Cannot find the task id: %d", id))
	}
}

func RunCommand(action liboct.TestActionCommand, id string) bool {
	db := liboct.GetDefaultDB()
	taskInterface, err := db.Get(liboct.DBTask, id)
	if err != nil {
		logrus.Warnf("Cannot find the test job %v", id)
		return false
	}
	task, _ := liboct.TaskFromString(taskInterface.String())
	workingDir := strings.TrimSuffix(task.BundleURL, ".tar.gz")
	logrus.Debugf("Run the command <%v> in %v ", action.Command, workingDir)

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
		return liboct.ExecSH(sh, workingDir)
	} else if pubConfig.Class == "container" {
		var sh string
		//For 'container', the 'dir' is the way to store test files and use 'id' to differentiate tasks
		clientCommand := pubConfig.ContainerClient
		switch liboct.TestAction(action.Action) {
		case liboct.TestActionDeploy:
			return true
		case liboct.TestActionRun:
			//docker run  -w=/test -v /tmp/.OCT/1234/bundle:/test ubuntu sh exe.sh
			//TODO: the ResName is the container name, so need to query if it exists. then need to pull in the apply session
			//Now , use ubuntu as the default
			sh = fmt.Sprintf("%s run -w=/octtest -v %s:/octtest  %s %s", clientCommand, workingDir, action.ResName, action.Command)
		case liboct.TestActionCollect:
			return true
		case liboct.TestActionDestroy:
			return true
		}

		logrus.Debugf(sh)
		return liboct.ExecSH(sh, workingDir)
	}

	return false
}

func PostTaskAction(w http.ResponseWriter, r *http.Request) {
	result, _ := ioutil.ReadAll(r.Body)
	logrus.Debugf("Post task action %v", string(result))
	r.Body.Close()
	action, err := liboct.ActionCommandFromString(string(result))
	if err != nil {
		liboct.RenderError(w, err)
		return
	}

	id := r.URL.Query().Get(":ID")

	if RunCommand(action, id) {
		liboct.RenderOK(w, "", nil)
	} else {
		liboct.RenderErrorf(w, fmt.Sprintf("Failed in run command: %s", string(result)))
	}
}

func RegisterToTestServer() {
	post_url := pubConfig.TSurl + "/resource"

	//Seems there will be lots of coding while getting the system info
	//Using config now.

	content := liboct.ReadFile("./host.conf")
	logrus.Debugf(content)
	liboct.SendCommand(post_url, []byte(content))
}

func init() {
	of, err := os.Open("octd.conf")
	if err != nil {
		logrus.Fatal(err)
		return
	}
	defer of.Close()

	if err = json.NewDecoder(of).Decode(&pubConfig); err != nil {
		logrus.Fatal(err)
		return
	}

	if pubConfig.Class == "container" {
		cmd := exec.Command("/bin/sh", "-c", pubConfig.ContainerDaemon)
		cmd.Stdin = os.Stdin
		if _, err := cmd.CombinedOutput(); err != nil {
			logrus.Fatal(err)
			return
		}
	}

	if pubConfig.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	db := liboct.GetDefaultDB()
	db.RegistCollect(liboct.DBTask)

	RegisterToTestServer()
}

func GetClair(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get(":ID")
	realURL := path.Join("/tmp/", id)

	file, err := os.Open(realURL)
	defer file.Close()
	if err != nil {
		logrus.Warnf("Cannot file the %v", realURL)
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Cannot open the file: " + realURL))
		return
	}

	buf := bytes.NewBufferString("")
	buf.ReadFrom(file)

	w.Write([]byte(buf.String()))
}

func main() {
	var port string
	port = fmt.Sprintf(":%d", pubConfig.Port)

	mux := routes.New()
	mux.Post("/task", ReceiveTask)
	mux.Post("/task/:ID", PostTaskAction)
	mux.Get("/task/:ID/report", GetTaskReport)

	mux.Get("/:ID", GetClair)
	http.Handle("/", mux)
	logrus.Infof("Start to listen %v", port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
