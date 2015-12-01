package main

import (
	"../../liboct"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/drone/routes"
)

type SchedulerConfig struct {
	Port           int
	ServerListFile string
	Debug          bool
}

func GetTaskStatus(w http.ResponseWriter, r *http.Request) {
	//TODO
}

func GetTaskReport(w http.ResponseWriter, r *http.Request) {
	db := liboct.GetDefaultDB()
	id := r.URL.Query().Get(":ID")
	sInterface, err := db.Get(liboct.DBScheduler, id)
	if err != nil {
		liboct.RenderError(w, err)
		return
	}
	s, _ := liboct.SchedulerFromString(sInterface.String())

	//Send the collect command to the octd
	if err := s.Command(liboct.TestActionCollect); err != nil {
		db.Update(liboct.DBScheduler, id, s)
		liboct.RenderError(w, err)
		return
	} else {
		db.Update(liboct.DBScheduler, id, s)
	}

	//Tar the reports in the cacheDir
	reportURL := path.Join(liboct.SchedulerCacheDir, s.ID, "reports.tar.gz")
	logrus.Debugf("Get reportURL %s", reportURL)
	_, err = os.Stat(reportURL)
	if err != nil {
		logrus.Warn(err)
		var reports []string
		for index := 0; index < len(s.Case.Units); index++ {
			reports = append(reports, s.Case.Units[index].ReportURL)
		}
		reportURL = liboct.TarFileList(reports, path.Join(liboct.SchedulerCacheDir, s.ID), "reports")
	}

	file, err := os.Open(reportURL)
	defer file.Close()
	if err != nil {
		liboct.RenderError(w, err)
	} else {
		buf := bytes.NewBufferString("")
		buf.ReadFrom(file)

		//Different write back, not json
		w.Write([]byte(buf.String()))
	}
}

func ReceiveTaskCommand(w http.ResponseWriter, r *http.Request) {
	db := liboct.GetDefaultDB()
	id := r.URL.Query().Get(":ID")
	sInterface, err := db.Get(liboct.DBScheduler, id)
	if err != nil {
		liboct.RenderError(w, err)
		return
	}
	s, _ := liboct.SchedulerFromString(sInterface.String())

	result, _ := ioutil.ReadAll(r.Body)
	logrus.Debugf("Receive task Command %s", string(result))
	r.Body.Close()
	/* Donnot use this now FIXME
	var cmd liboct.TestActionCommand
	json.Unmarshal([]byte(result), &cmd)
	*/
	action, err := liboct.TestActionFromString(string(result))
	if err != nil {
		liboct.RenderError(w, err)
		return
	}

	err = s.Command(action)
	db.Update(liboct.DBScheduler, id, s)
	if err != nil {
		liboct.RenderError(w, err)
	} else {
		liboct.RenderOK(w, "", nil)
	}
}

// Since the sheduler ID is got after receiving the test files
// we need to move it to a better place
// /tmp/.../A.tar.gz --> /tmp/.../id/A.tar.gz
func updateSchedulerBundle(id string, oldURL string) {
	db := liboct.GetDefaultDB()
	sInterface, _ := db.Get(liboct.DBScheduler, id)
	s, _ := liboct.SchedulerFromString(sInterface.String())

	dir := path.Dir(path.Dir(oldURL))
	newURL := fmt.Sprintf("%s/%s", path.Join(dir, id), path.Base(oldURL))
	liboct.PreparePath(path.Join(dir, id), "")
	logrus.Debugf("Old URL %s, New URL %s", oldURL, newURL)
	os.Rename(strings.TrimSuffix(oldURL, ".tar.gz"), strings.TrimSuffix(newURL, ".tar.gz"))
	os.Rename(oldURL, newURL)
	//bundleURL is the directory of the bundle
	s.Case.SetBundleURL(strings.TrimSuffix(newURL, ".tar.gz"))
	db.Update(liboct.DBScheduler, id, s)
}

func ReceiveTask(w http.ResponseWriter, r *http.Request) {
	logrus.Debugf("ReceiveTask begin")
	var tc liboct.TestCase
	db := liboct.GetDefaultDB()
	realURL, _ := liboct.ReceiveFile(w, r, liboct.SchedulerCacheDir)
	tc, err := liboct.CaseFromTar(realURL, "")
	if err != nil {
		liboct.RenderError(w, err)
		return
	}

	s, _ := liboct.SchedulerNew(tc)

	err = s.Command(liboct.TestActionApply)
	if err == nil {
		if id, e := db.Add(liboct.DBScheduler, s); e == nil {
			updateSchedulerBundle(id, realURL)
			liboct.RenderOK(w, id, nil)
		} else {
			liboct.RenderError(w, e)
		}
	} else {
		liboct.RenderError(w, err)
	}
}

// Will use DB in the future, (mongodb for example)
func init() {
	sf, err := os.Open("scheduler.conf")
	if err != nil {
		logrus.Fatal(err)
		return
	}
	defer sf.Close()

	if err = json.NewDecoder(sf).Decode(&pubConfig); err != nil {
		logrus.Fatal(err)
		return
	}

	if pubConfig.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.WarnLevel)
	}
	db := liboct.GetDefaultDB()
	db.RegistCollect(liboct.DBResource)
	db.RegistCollect(liboct.DBScheduler)
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
		if _, e := db.Add(liboct.DBResource, rl[index]); e != nil {
			logrus.Warn(e)
		}
	}
}

var pubConfig SchedulerConfig

func main() {
	mux := routes.New()

	mux.Get("/resource", GetResource)
	mux.Post("/resource", PostResource)
	mux.Get("/resource/:ID", GetResourceStatus)

	mux.Post("/task", ReceiveTask)
	mux.Get("/task/:ID", GetTaskStatus)
	mux.Post("/task/:ID", ReceiveTaskCommand)
	mux.Get("/task/:ID/report", GetTaskReport)

	logrus.Infof("Start to listen :%d", pubConfig.Port)
	http.Handle("/", mux)
	port := fmt.Sprintf(":%d", pubConfig.Port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		logrus.Fatal(err)
	}
}
