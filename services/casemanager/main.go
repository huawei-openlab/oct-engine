package main

import (
	"../../lib/libocit"
	"../../lib/routes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

type CaseManagerConf struct {
	Repos        []libocit.TestCaseRepo
	SchedulerURL string
	Port         int
	Debug        bool
}

var pubConfig CaseManagerConf

func init() {
	cmf, err := os.Open("casemanager.conf")
	if err != nil {
		fmt.Errorf("Cannot find casemanager.conf.")
		return
	}
	defer cmf.Close()

	if err = json.NewDecoder(cmf).Decode(&pubConfig); err != nil {
		fmt.Errorf("Error in parse casemanager.conf")
		return
	}

	libocit.DBRegist(libocit.DBCase)
	libocit.DBRegist(libocit.DBRepo)
	libocit.DBRegist(libocit.DBTask)

	repos := pubConfig.Repos
	for index := 0; index < len(repos); index++ {
		if err := repos[index].IsValid(); err != nil {
			if pubConfig.Debug == true {
				fmt.Println("The repo ", repos[index], " is invalid. ", err)
				continue
			}
		}
		fmt.Println(repos[index])
		if id, ok := libocit.DBAdd(libocit.DBRepo, repos[index]); ok {
			RefreshRepo(id)
		}
	}
}

//TODO: is there any usefull Restful help document lib?
func GetHelp(w http.ResponseWriter, r *http.Request) {
	var ret libocit.HttpRet
	ret.Status = "OK"
	ret.Message = fmt.Sprintf("The following APIs are supported.")
	ret.Data = "case, repo and task"
	retInfo, _ := json.MarshalIndent(ret, "", "\t")
	w.Write([]byte(retInfo))
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

	//Add a new task by the case id
	mux.Get("/task", ListTasks)
	mux.Post("/task", AddTask)
	mux.Get("/task/:ID", GetTaskStatus)
	mux.Post("/task/:ID", PostTaskAction)

	http.Handle("/", mux)
	fmt.Println("Listen to port ", port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}