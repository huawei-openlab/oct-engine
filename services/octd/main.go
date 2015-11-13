package main

import (
	"../../lib/liboct"
	"../../lib/routes"
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
)

type OCTDConfig struct {
	TSurl string
	Port  int
	Debug bool
}

const OCTDCacheDir = "/tmp/.octd_cache"

var pubConfig OCTDConfig

func GetResult(w http.ResponseWriter, r *http.Request) {
	//TODO
	var realURL string
	filename := r.URL.Query().Get("File")
	id := r.URL.Query().Get("ID")

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
	real_url, params := liboct.ReceiveFile(w, r, OCTDCacheDir)

	fmt.Println(params)

	if val, ok := params["id"]; ok {
		//The real url may not be the test case format, could be only files
		if strings.HasSuffix(real_url, ".tar.gz") {
			liboct.UntarFile(real_url, path.Join(OCTDCacheDir, val))
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

func RunCommand(cmd string, dir string) bool {
	if pubConfig.Debug {
		fmt.Println("Run the command < ", cmd, ">  in ", dir)
	}
	//check it since some case only has a config.json
	liboct.PreparePath(dir, "")
	os.Chdir(dir)

	c := exec.Command("/bin/sh", "-c", cmd)
	c.Run()
	return true
}

func GetWorkingDir(id string) string {
	return path.Join("/tmp/.octd_cache", id)
}

func PostTaskAction(w http.ResponseWriter, r *http.Request) {
	var ret liboct.HttpRet
	result, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	cmd, err := liboct.ActionCommandFromString(string(result))
	if err != nil {
		ret.Status = liboct.RetStatusFailed
		ret.Message = "Invalid action command"
		retInfo, _ := json.MarshalIndent(ret, "", "\t")
		w.Write([]byte(retInfo))
		return
	}

	id := r.URL.Query().Get(":ID")

	wdir := GetWorkingDir(id)
	//TODO: Just handle the deploy work now
	if RunCommand(cmd.Command, wdir) {
		ret.Status = liboct.RetStatusOK
	} else {
		ret.Status = liboct.RetStatusFailed
	}
	ret_string, _ := json.Marshal(ret)
	w.Write([]byte(ret_string))
}

func RegisterToTestServer() {
	post_url := pubConfig.TSurl + "/os"

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

	RegisterToTestServer()
}

func main() {

	var port string
	port = fmt.Sprintf(":%d", pubConfig.Port)

	mux := routes.New()
	mux.Get("/result", GetResult)
	mux.Post("/task", ReceiveTask)
	mux.Post("/task/:ID", PostTaskAction)

	http.Handle("/", mux)
	fmt.Println("Start to listen ", port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
