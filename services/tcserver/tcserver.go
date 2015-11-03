package main

import (
	"../../lib/libocit"
	"../../lib/routes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"strconv"
)

type TCServerConf struct {
	Repos []libocit.TestCaseRepo
	Port  int
	Debug bool
}

var pub_config TCServerConf

//TODO, need to add a lock
var repoStore map[string]libocit.TestCaseRepo
var caseStore map[string]libocit.TestCase

func ListRepos(w http.ResponseWriter, r *http.Request) {
	page_string := r.URL.Query().Get("Page")
	page, err := strconv.Atoi(page_string)
	if err != nil {
		page = 0
	}
	pageSizeString := r.URL.Query().Get("PageSize")
	pageSize, err := strconv.Atoi(pageSizeString)
	if err != nil {
		pageSize = 10
	}

	var rl []libocit.TestCaseRepo
	cur_num := 0
	for _, repoInStore := range repoStore {
		cur_num += 1
		if (cur_num >= page*pageSize) && (cur_num < (page+1)*pageSize) {
			rl = append(rl, repoInStore)
		}
	}

	var ret libocit.HttpRet
	ret.Status = "OK"
	ret.Message = fmt.Sprintf("%d repos founded", len(rl))
	ret.Data = rl
	retInfo, _ := json.MarshalIndent(ret, "", "\t")
	w.Write([]byte(retInfo))
}

func GetRepo(w http.ResponseWriter, r *http.Request) {
	var ret libocit.HttpRet
	repoID := r.URL.Query().Get(":ID")

	for _, repoInStore := range repoStore {
		if repoID == repoInStore.GetID() {
			ret.Status = "OK"
			ret.Data = repoInStore
			retInfo, _ := json.MarshalIndent(ret, "", "\t")
			w.Write([]byte(retInfo))
			return
		}
	}

	ret.Status = "Failed"
	ret.Message = "Cannot find the repo, wrong ID provided"
	retInfo, _ := json.MarshalIndent(ret, "", "\t")
	w.Write([]byte(retInfo))
}

func AddRepo(w http.ResponseWriter, r *http.Request) {
	var ret libocit.HttpRet
	action := r.URL.Query().Get("Action")

	//Add and refresh
	if action == "Add" {
		var repo libocit.TestCaseRepo
		result, _ := ioutil.ReadAll(r.Body)
		r.Body.Close()
		err := json.Unmarshal([]byte(result), &repo)
		if err != nil {
			ret.Status = "Failed"
			ret.Message = fmt.Sprintf("The repo is invalid.")
		} else {
			if msgs, ok := repo.IsValid(); ok {
				repo.SetID(libocit.MD5(path.Join(repo.Name, repo.CaseFolder)))
				repoStore[repo.GetID()] = repo
				ret.Status = "OK"
				RefreshRepo(repo)
			} else {
				ret.Status = "Failed"
				ret.Message = fmt.Sprintf("The repo has %d error(s).", len(msgs))
				ret.Data = msgs
			}
		}
	} else if action == "Refresh" {
		for _, repoInStore := range repoStore {
			RefreshRepo(repoInStore)
		}
		ret.Status = "OK"
	} else {
		ret.Status = "Failed"
		ret.Message = "The action in AddRepo is limited to Add/Refresh"
	}

	retInfo, _ := json.MarshalIndent(ret, "", "\t")
	w.Write([]byte(retInfo))
}

func ModifyRepo(w http.ResponseWriter, r *http.Request) {
	var ret libocit.HttpRet
	var repoPoint *libocit.TestCaseRepo
	repoID := r.URL.Query().Get(":ID")
	action := r.URL.Query().Get("Action")

	repoPoint = nil
	for id, repoInStore := range repoStore {
		if repoID == id {
			repoPoint = &repoInStore
			break
		}
	}

	if repoPoint == nil {
		ret.Status = "Failed"
		ret.Message = fmt.Sprintf("The repo %s is not exist.", repoID)
		retInfo, _ := json.MarshalIndent(ret, "", "\t")
		w.Write([]byte(retInfo))
		return
	}
	repo := *repoPoint
	if action == "Modify" {
		//FIXME: The ID should not be changed.
		var newRepo libocit.TestCaseRepo
		result, _ := ioutil.ReadAll(r.Body)
		r.Body.Close()
		err := json.Unmarshal([]byte(result), &newRepo)
		if err != nil {
			ret.Status = "Failed"
			ret.Message = fmt.Sprintf("The modified information is invalid.")
		} else {
			repo.Modify(newRepo)
			repo.SetID(libocit.MD5(path.Join(repo.Name, repo.CaseFolder)))
			RefreshRepo(repo)
			ret.Status = "OK"
		}
	} else if action == "Refresh" {
		RefreshRepo(repo)
		ret.Status = "OK"
	} else {
		ret.Status = "Failed"
		ret.Message = "The action in ModifyRepo is limited to Add/Refresh"
	}

	retInfo, _ := json.MarshalIndent(ret, "", "\t")
	w.Write([]byte(retInfo))
}

func CleanRepo(repo libocit.TestCaseRepo) {
	for caseID, tc := range caseStore {
		caseRepoID := tc.GetRepoID()
		if len(caseRepoID) > 0 && caseRepoID == repo.GetID() {
			delete(caseStore, caseID)
		}
	}
}

//This refresh the 'cache' maintained in tcserver
func RefreshRepo(repo libocit.TestCaseRepo) {
	if repo.Refresh() {
		CleanRepo(repo)
		cases := repo.GetCases()
		for index := 0; index < len(cases); index++ {
			fmt.Println("case loaded ", cases[index])
			cases[index].SetID(libocit.MD5(cases[index].GetBundleURL()))
			caseStore[cases[index].GetID()] = cases[index]
		}
	}
}

func RefreshRepos(w http.ResponseWriter, r *http.Request) {
	var ret libocit.HttpRet

	for _, repoInStore := range repoStore {
		RefreshRepo(repoInStore)
	}
	ret.Status = "OK"
	retInfo, _ := json.MarshalIndent(ret, "", "\t")
	w.Write([]byte(retInfo))
}

func ListCases(w http.ResponseWriter, r *http.Request) {
	//Need better explaination of 'status', currently, only hasReport/isUpdated
	status := r.URL.Query().Get("Status")
	page_string := r.URL.Query().Get("Page")
	page, err := strconv.Atoi(page_string)
	if err != nil {
		page = 0
	}
	pageSizeString := r.URL.Query().Get("PageSize")
	pageSize, err := strconv.Atoi(pageSizeString)
	if err != nil {
		pageSize = 10
	}

	var caseList []libocit.TestCase
	cur_num := 0

	for _, caseInStore := range caseStore {
		if caseInStore.MatchStatus(status) {
			cur_num += 1
			if (cur_num >= page*pageSize) && (cur_num < (page+1)*pageSize) {
				caseList = append(caseList, caseInStore)
			}
		}
	}

	var ret libocit.HttpRet
	ret.Status = "OK"
	ret.Message = fmt.Sprintf("%d cases founded", len(caseList))
	ret.Data = caseList

	retInfo, _ := json.MarshalIndent(ret, "", "\t")
	w.Write([]byte(retInfo))
}

func GetCase(w http.ResponseWriter, r *http.Request) {
	//TODO: support another query method : repo/group/name
	id := r.URL.Query().Get(":ID")
	tc := caseStore[id]

	value := tc.GetBundleContent()

	if len(value) > 0 {
		//FIXME: add the error to head
		w.Write([]byte(value))
	} else {
		w.Write([]byte("Cannot get the case."))
	}
}

func GetCaseReport(w http.ResponseWriter, r *http.Request) {
	var ret libocit.HttpRet
	id := r.URL.Query().Get(":ID")
	tc := caseStore[id]

	content := tc.GetReportContent()
	if len(content) > 0 {
		ret.Status = "OK"
		ret.Data = content
	} else {
		ret.Status = "Failed"
		ret.Message = "Cannot find the report"
	}

	retInfo, _ := json.MarshalIndent(ret, "", "\t")
	w.Write([]byte(retInfo))
}

func init() {
	repoStore = make(map[string]libocit.TestCaseRepo)
	caseStore = make(map[string]libocit.TestCase)
	content := libocit.ReadFile("./tcserver.conf")
	fmt.Println(content)
	json.Unmarshal([]byte(content), &pub_config)
	fmt.Println(pub_config)
	repos := pub_config.Repos
	for index := 0; index < len(repos); index++ {
		msgs, ok := repos[index].IsValid()
		if (ok == false) && (pub_config.Debug == true) {
			fmt.Println("The repo ", repos[index], " is invalid. ", msgs)
			continue
		}
		repos[index].SetID(libocit.MD5(path.Join(repos[index].Name, repos[index].CaseFolder)))
		repoStore[repos[index].GetID()] = repos[index]
	}

	for _, repoInStore := range repoStore {
		RefreshRepo(repoInStore)
	}
}

func main() {
	port := fmt.Sprintf(":%d", pub_config.Port)
	mux := routes.New()
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
	http.Handle("/", mux)
	fmt.Println("Listen to port ", port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
