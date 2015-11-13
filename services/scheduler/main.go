package main

import (
	"../../lib/libocit"
	"../../lib/routes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

const SchedularCacheDir = "/tmp/.test_schedular_cache"

type SchedulerConfig struct {
	Port           int
	ServerListFile string
	CacheDir       string
	Debug          bool
}

func ReceiveTaskCommand(w http.ResponseWriter, r *http.Request) {
	var ret libocit.HttpRet
	id := r.URL.Query().Get(":ID")
	sInterface, ok := libocit.DBGet(libocit.DBResource, id)
	if !ok {
		ret.Status = libocit.RetStatusFailed
		ret.Message = "Invalid task id"
		ret_string, _ := json.MarshalIndent(ret, "", "\t")
		w.Write([]byte(ret_string))
		return
	}
	s, _ := libocit.SchedulerFromString(sInterface.String())

	var cmd libocit.TestActionCommand
	result, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	json.Unmarshal([]byte(result), &cmd)
	action, ok := libocit.TestActionFromString(cmd.Action)
	if !ok {
		ret.Status = libocit.RetStatusFailed
		ret.Message = "Invalid action"
	} else {
		if s.Command(action) {
			ret.Status = libocit.RetStatusOK
		} else {
			ret.Status = libocit.RetStatusFailed
			ret.Message = "Failed to run the action"
		}
	}

	ret_string, _ := json.MarshalIndent(ret, "", "\t")
	w.Write([]byte(ret_string))
}

func ReceiveTask(w http.ResponseWriter, r *http.Request) {
	var ret libocit.HttpRet
	var tc libocit.TestCase
	realURL, _ := libocit.ReceiveFile(w, r, SchedularCacheDir)
	tc, err := libocit.CaseFromTar(realURL, "")
	if err != nil {
		ret.Status = libocit.RetStatusFailed
		ret.Message = err.Error()
		ret_string, _ := json.MarshalIndent(ret, "", "\t")
		w.Write([]byte(ret_string))
		return
	} else {
	}

	s, _ := libocit.SchedulerNew(tc)

	if s.Command(libocit.TestActionApply) {
		if id, ok := libocit.DBAdd(libocit.DBScheduler, s); ok {
			ret.Status = libocit.RetStatusOK
			ret.Message = id
			ret_string, _ := json.MarshalIndent(ret, "", "\t")
			w.Write([]byte(ret_string))
			return
		}
	}
	ret.Status = libocit.RetStatusFailed
	ret.Message = "Failed in allocate the resource"
	ret_string, _ := json.MarshalIndent(ret, "", "\t")
	w.Write([]byte(ret_string))
	return
}

// Will use DB in the future, (mongodb for example)
func init() {
	sf, err := os.Open("scheduler.conf")
	if err != nil {
		fmt.Errorf("Cannot find scheduler.conf.")
		return
	}
	defer sf.Close()

	if err = json.NewDecoder(sf).Decode(&pubConfig); err != nil {
		fmt.Errorf("Error in parse scheduler.conf")
		return
	}

	libocit.DBRegist(libocit.DBResource)
	if len(pubConfig.ServerListFile) == 0 {
		return
	}

	slf, err := os.Open(pubConfig.ServerListFile)
	if err != nil {
		return
	}
	defer slf.Close()

	var rl []libocit.Resource
	if err = json.NewDecoder(slf).Decode(&rl); err != nil {
		return
	}

	for index := 0; index < len(rl); index++ {
		if _, ok := libocit.DBAdd(libocit.DBResource, rl[index]); ok {
			fmt.Println(rl[index])
		}
	}
}

var pubConfig SchedulerConfig

func main() {
	return

	mux := routes.New()

	mux.Get("/resource", GetResource)
	mux.Post("/resource", PostResource)
	mux.Get("/resource/:ID/status", GetResourceStatus)

	mux.Post("/task", ReceiveTask)
	mux.Post("/task/:ID", ReceiveTaskCommand)

	http.Handle("/", mux)
	port := fmt.Sprintf(":%d", pubConfig.Port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
