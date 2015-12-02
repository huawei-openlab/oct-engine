package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/drone/routes"
	"github.com/huawei-openlab/oct-engine/liboct"
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
	//The real working dir is under 'source'
	realURL := path.Join(workingDir, "source", filename)
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
		if strings.HasSuffix(realURL, ".tar.gz") {
			liboct.UntarFile(realURL, strings.TrimSuffix(realURL, ".tar.gz"))
		}
		var task liboct.TestTask
		task.ID = id
		task.BundleURL = realURL
		if name, ok := params["name"]; ok {
			task.Name = name
		} else {
			task.Name = id
			logrus.Warnf("Cannot find the name of the task.")
		}
		db.Update(liboct.DBTask, id, task)

		liboct.RenderOK(w, "", nil)
	} else {
		liboct.RenderErrorf(w, fmt.Sprintf("Cannot find the task id: %d", id))
	}
}

func RunCommand(action liboct.TestActionCommand, workingDir string) ([]byte, error) {
	logrus.Debugf("Run the command <%v> in %v ", action.Command, workingDir)
	if pubConfig.Class == "os" {
		return liboct.ExecSH(action.Command, workingDir)
	} else if pubConfig.Class == "container" {
		//For 'container', the 'dir' is the way to store test files and use 'id' to differentiate tasks
		clientCommand := pubConfig.ContainerClient
		//For container, the action of deploy and run is merged.
		if liboct.TestAction(action.Action) == liboct.TestActionRun {
			//docker run  -w=/test -v /tmp/.OCT/1234/bundle:/test ubuntu sh exe.sh
			//TODO: the ResName is the container name, so need to query if it exists. then need to pull in the apply session
			//Now , use ubuntu as the default
			sh := fmt.Sprintf("%s run -w=/octtest -v %s:/octtest  %s %s", clientCommand, workingDir, action.ResName, action.Command)
			logrus.Debugf(sh)
			return liboct.ExecSH(sh, workingDir)
		}

	}

	return nil, nil
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

	db := liboct.GetDefaultDB()
	taskInterface, err := db.Get(liboct.DBTask, id)
	if err != nil {
		liboct.RenderError(w, err)
		return
	}
	task, _ := liboct.TaskFromString(taskInterface.String())
	workingDir := path.Join(strings.TrimSuffix(task.BundleURL, ".tar.gz"), "source")
	if _, err := os.Stat(workingDir); err != nil {
		//Create in the case which has no 'source' files
		os.MkdirAll(workingDir, 0777)
	}

	switch action.Action {
	case liboct.TestActionDeploy:
		fallthrough
	case liboct.TestActionRun:
		val, err := RunCommand(action, workingDir)

		//Save the logs
		logFile := fmt.Sprintf("%s/%s.log", workingDir, task.Name)
		ioutil.WriteFile(logFile, val, 0644)
		if err != nil {
			liboct.RenderErrorf(w, fmt.Sprintf("Failed in run command: %s", string(result)))
		} else {
			liboct.RenderOK(w, "", string(val))
		}
		return
	}

	liboct.RenderErrorf(w, fmt.Sprintf("Action %s is not support yet", action.Action))
}

func RegisterToTestServer() {
	post_url := pubConfig.TSurl + "/resource"

	//Seems there will be lots of coding while getting the system info
	//Using config now.

	file, err := os.Open("./host.conf")
	defer file.Close()
	if err != nil {
		logrus.Info(err)
		return
	}
	buf := bytes.NewBufferString("")
	buf.ReadFrom(file)
	liboct.SendCommand(post_url, buf.Bytes())
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
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}

	db := liboct.GetDefaultDB()
	db.RegistCollect(liboct.DBTask)

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
	logrus.Infof("Start to listen %v", port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		logrus.Fatalf("ListenAndServe: %v ", err)
	}
}
