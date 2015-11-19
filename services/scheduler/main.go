package main

import (
	"../../lib/liboct"
	"../../lib/routes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
)

const SchedularCacheDir = "/tmp/.test_schedular_cache"

type SchedulerConfig struct {
	Port           int
	ServerListFile string
	CacheDir       string
	Debug          bool
}

func ReceiveTaskCommand(w http.ResponseWriter, r *http.Request) {
	var ret liboct.HttpRet
	id := r.URL.Query().Get(":ID")
	sInterface, err := liboct.DBGet(liboct.DBScheduler, id)
	if err != nil {
		ret.Status = liboct.RetStatusFailed
		ret.Message = err.Error()
		ret_string, _ := json.MarshalIndent(ret, "", "\t")
		w.Write([]byte(ret_string))
		return
	}
	s, _ := liboct.SchedulerFromString(sInterface.String())

	result, _ := ioutil.ReadAll(r.Body)
	fmt.Println("Receive task Command ", string(result))
	r.Body.Close()
	/* Donnot use this now FIXME
	var cmd liboct.TestActionCommand
	json.Unmarshal([]byte(result), &cmd)
	*/
	action, err := liboct.TestActionFromString(string(result))
	if err != nil {
		ret.Status = liboct.RetStatusFailed
		ret.Message = err.Error()
	} else {
		if e := s.Command(action); e == nil {
			ret.Status = liboct.RetStatusOK
		} else {
			ret.Status = liboct.RetStatusFailed
			ret.Message = e.Error()
		}
	}

	ret_string, _ := json.MarshalIndent(ret, "", "\t")
	w.Write([]byte(ret_string))
}

// Since the sheduler ID is got after receiving the test files
// we need to move it to a better place
// /tmp/.../A.tar.gz --> /tmp/.../id/A.tar.gz
func updateSchedulerBundle(id string, oldURL string) {
	sInterface, _ := liboct.DBGet(liboct.DBScheduler, id)
	s, _ := liboct.SchedulerFromString(sInterface.String())
	dir := path.Dir(oldURL)
	newURL := fmt.Sprintf("%s/%s", path.Join(dir, id), path.Base(oldURL))
	liboct.PreparePath(path.Join(dir, id), "")
	os.Rename(strings.TrimSuffix(oldURL, ".tar.gz"), strings.TrimSuffix(newURL, ".tar.gz"))
	os.Rename(oldURL, newURL)
	//bundleURL is the directory of the bundle
	s.Case.SetBundleURL(strings.TrimSuffix(newURL, ".tar.gz"))
	liboct.DBUpdate(liboct.DBScheduler, id, s)
}

func ReceiveTask(w http.ResponseWriter, r *http.Request) {
	fmt.Println("ReceiveTask begin")
	var ret liboct.HttpRet
	var tc liboct.TestCase
	realURL, _ := liboct.ReceiveFile(w, r, SchedularCacheDir)
	tc, err := liboct.CaseFromTar(realURL, "")
	fmt.Println("Receive task after file", realURL, tc)
	if err != nil {
		ret.Status = liboct.RetStatusFailed
		ret.Message = err.Error()
		ret_string, _ := json.MarshalIndent(ret, "", "\t")
		w.Write([]byte(ret_string))
		return
	}

	s, _ := liboct.SchedulerNew(tc)

	err = s.Command(liboct.TestActionApply)
	if err == nil {
		if id, e := liboct.DBAdd(liboct.DBScheduler, s); e == nil {
			updateSchedulerBundle(id, realURL)
			fmt.Println("Add ok ", id)
			ret.Status = liboct.RetStatusOK
			ret.Message = id
		} else {
			ret.Status = liboct.RetStatusFailed
			ret.Message = e.Error()
		}
	} else {
		ret.Status = liboct.RetStatusFailed
		ret.Message = err.Error()
	}
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

	liboct.DBRegist(liboct.DBResource)
	liboct.DBRegist(liboct.DBScheduler)
	if len(pubConfig.ServerListFile) == 0 {
		return
	}

	slf, err := os.Open(pubConfig.ServerListFile)
	if err != nil {
		return
	}
	defer slf.Close()

	var rl []liboct.Resource
	if err = json.NewDecoder(slf).Decode(&rl); err != nil {
		return
	}

	for index := 0; index < len(rl); index++ {
		if _, e := liboct.DBAdd(liboct.DBResource, rl[index]); e == nil {
			fmt.Println(rl[index])
		}
	}
}

var pubConfig SchedulerConfig

func main() {
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
