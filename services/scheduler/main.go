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
	"path"
	"strings"
)

type SchedulerConfig struct {
	Port           int
	ServerListFile string
	Debug          bool
}

func GetTaskReport(w http.ResponseWriter, r *http.Request) {
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

	//Send the collect command to the octd
	if err := s.Command(liboct.TestActionCollect); err != nil {
		ret.Status = liboct.RetStatusFailed
		ret.Message = err.Error()
		ret_string, _ := json.MarshalIndent(ret, "", "\t")
		w.Write([]byte(ret_string))
		liboct.DBUpdate(liboct.DBScheduler, id, s)
		return
	} else {
		liboct.DBUpdate(liboct.DBScheduler, id, s)
	}

	//Tar the reports in the cacheDir
	reportURL := path.Join(liboct.SchedulerCacheDir, s.ID, "reports.tar.gz")
	fmt.Println("Get reportURL ", reportURL)
	_, err = os.Stat(reportURL)
	if err != nil {
		var reports []string
		for index := 0; index < len(s.Case.Units); index++ {
			reports = append(reports, s.Case.Units[index].ReportURL)
		}
		reportURL = liboct.TarFileList(reports, path.Join(liboct.SchedulerCacheDir, s.ID), "reports")
	}

	file, err := os.Open(reportURL)
	defer file.Close()
	if err != nil {
		ret.Status = liboct.RetStatusFailed
		ret.Message = err.Error()
		ret_string, _ := json.MarshalIndent(ret, "", "\t")
		w.Write([]byte(ret_string))
	} else {
		buf := bytes.NewBufferString("")
		buf.ReadFrom(file)

		w.Write([]byte(buf.String()))
	}
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
		liboct.DBUpdate(liboct.DBScheduler, id, s)
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
	realURL, _ := liboct.ReceiveFile(w, r, liboct.SchedulerCacheDir)
	tc, err := liboct.CaseFromTar(realURL, "")
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
	mux.Get("/task/:ID/report", GetTaskReport)

	http.Handle("/", mux)
	port := fmt.Sprintf(":%d", pubConfig.Port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
