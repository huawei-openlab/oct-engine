package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/drone/routes"
	"github.com/huawei-openlab/oct-engine/liboct"
)

type CaseManagerConf struct {
	Repos        []liboct.TestCaseRepo
	SchedulerURL string
	Port         int
	Debug        bool
}

var pubConfig CaseManagerConf

func init() {
	cmf, err := os.Open("casemanager.conf")
	if err != nil {
		logrus.Fatal(err)
		return
	}
	defer cmf.Close()

	if err = json.NewDecoder(cmf).Decode(&pubConfig); err != nil {
		logrus.Fatal(err)
		return
	}

	if pubConfig.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.WarnLevel)
	}

	db := liboct.GetDefaultDB()
	db.RegistCollect(liboct.DBCase)
	db.RegistCollect(liboct.DBRepo)
	db.RegistCollect(liboct.DBTask)

	repos := pubConfig.Repos
	for index := 0; index < len(repos); index++ {
		if err := repos[index].IsValid(); err != nil {
			logrus.Warnf("The repo ", repos[index], " is invalid. ", err.Error())
			continue
		}
		if id, err := db.Add(liboct.DBRepo, repos[index]); err == nil {
			RefreshRepo(id)
		}
	}
}

//TODO: is there any usefull Restful help document lib?
func GetHelp(w http.ResponseWriter, r *http.Request) {
	liboct.RenderOK(w, fmt.Sprintf("The following APIs are supported."), "case, repo and task")
}

func main() {
	port := fmt.Sprintf(":%d", pubConfig.Port)
	mux := routes.New()
	mux.Get("/", GetHelp)
	mux.Get("/repo", ListRepos)
	mux.Get("/repo/:ID", GetRepo)
	//either refresh or add
	mux.Post("/repo", AddRepo)
	//either refresh or modify (especially enable/disable)
	mux.Post("/repo/:ID", ModifyRepo)

	mux.Get("/case", ListCases)
	//TODO: add group/name in the future
	mux.Get("/case/:ID", GetCase)
	mux.Get("/case/:ID/report", GetCaseReport)

	//The task api is not necessary now, just used for web in the future.
	//Add a new task by the case id
	mux.Get("/task", ListTasks)
	//Add task : turn a case into a task, but donnot send to scheduler
	mux.Post("/task", AddTask)
	mux.Get("/task/:TaskID", GetTaskStatus)
	mux.Get("/task/:TaskID/report", GetTaskReport)
	// apply/deploy/run/collect/destroy
	mux.Post("/task/:TaskID", PostTaskAction)

	http.Handle("/", mux)
	logrus.Infof("Listen to port ", port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
