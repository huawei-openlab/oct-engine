package main

import (
	"../../lib/libocit"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

func ListRepos(w http.ResponseWriter, r *http.Request) {
	var query libocit.DBQuery
	page_string := r.URL.Query().Get("Page")
	page, err := strconv.Atoi(page_string)
	if err == nil {
		query.Page = page
	}
	pageSizeString := r.URL.Query().Get("PageSize")
	pageSize, err := strconv.Atoi(pageSizeString)
	if err == nil {
		query.PageSize = pageSize
	}

	var rl []libocit.DBInterface
	ids := libocit.DBLookup(libocit.DBRepo, query)
	for index := 0; index < len(ids); index++ {
		repo, _ := libocit.DBGet(libocit.DBRepo, ids[index])
		rl = append(rl, repo)
	}

	var ret libocit.HttpRet
	ret.Status = libocit.RetStatusOK
	ret.Message = fmt.Sprintf("%d repos founded", len(rl))
	ret.Data = rl
	retInfo, _ := json.MarshalIndent(ret, "", "\t")
	w.Write([]byte(retInfo))
}

func GetRepo(w http.ResponseWriter, r *http.Request) {
	var ret libocit.HttpRet
	id := r.URL.Query().Get(":ID")

	if repo, ok := libocit.DBGet(libocit.DBRepo, id); ok {
		ret.Status = libocit.RetStatusOK
		ret.Data = repo
	} else {

		ret.Status = libocit.RetStatusFailed
		ret.Message = "Cannot find the repo, wrong ID provided"
	}
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
			ret.Status = libocit.RetStatusFailed
			ret.Message = fmt.Sprintf("The repo is invalid.")
		} else {
			if err := repo.IsValid(); err == nil {
				if id, ok := libocit.DBAdd(libocit.DBRepo, repo); ok {
					ret.Status = libocit.RetStatusOK
					RefreshRepo(id)
				} else {
					ret.Status = libocit.RetStatusFailed
					ret.Message = fmt.Sprintf("Cannot add repo: ", err.Error())
				}
			} else {
				ret.Status = libocit.RetStatusFailed
				ret.Message = fmt.Sprintf("The repo has error: ", err.Error())
			}
		}
	} else if action == "Refresh" {
		ids := libocit.DBLookup(libocit.DBRepo, libocit.DBQuery{})
		for index := 0; index < len(ids); index++ {
			RefreshRepo(ids[index])
		}
		ret.Status = libocit.RetStatusOK
	} else {
		ret.Status = libocit.RetStatusFailed
		ret.Message = "The action in AddRepo is limited to Add/Refresh"
	}

	retInfo, _ := json.MarshalIndent(ret, "", "\t")
	w.Write([]byte(retInfo))
}

func ModifyRepo(w http.ResponseWriter, r *http.Request) {
	var ret libocit.HttpRet
	repoID := r.URL.Query().Get(":ID")
	action := r.URL.Query().Get("Action")

	val, ok := libocit.DBGet(libocit.DBRepo, repoID)
	if !ok {
		ret.Status = libocit.RetStatusFailed
		ret.Message = fmt.Sprintf("The repo %s is not exist.", repoID)
		retInfo, _ := json.MarshalIndent(ret, "", "\t")
		w.Write([]byte(retInfo))
		return
	}
	if action == "Modify" {
		//FIXME: The ID should not be changed.
		var newRepo libocit.TestCaseRepo
		result, _ := ioutil.ReadAll(r.Body)
		r.Body.Close()
		err := json.Unmarshal([]byte(result), &newRepo)
		if err != nil {
			ret.Status = libocit.RetStatusFailed
			ret.Message = fmt.Sprintf("The modified information is invalid.")
		} else {
			oldRepo, _ := libocit.RepoFromString(val.String())
			oldRepo.Modify(newRepo)
			libocit.DBModify(libocit.DBRepo, repoID, oldRepo)
			RefreshRepo(repoID)
			ret.Status = libocit.RetStatusOK
		}
	} else if action == "Refresh" {
		RefreshRepo(repoID)
		ret.Status = libocit.RetStatusOK
	} else {
		ret.Status = libocit.RetStatusFailed
		ret.Message = "The action in ModifyRepo is limited to Add/Refresh"
	}

	retInfo, _ := json.MarshalIndent(ret, "", "\t")
	w.Write([]byte(retInfo))
}

func CleanRepo(repo libocit.TestCaseRepo) {
	var query libocit.DBQuery
	query.Params = make(map[string]string)
	query.Params["RepoID"] = repo.GetID()
	ids := libocit.DBLookup(libocit.DBRepo, query)
	for index := 0; index < len(ids); index++ {
		libocit.DBRemove(libocit.DBCase, ids[index])
	}
}

//This refresh the 'cache' maintained in casemanager
func RefreshRepo(id string) {
	val, ok := libocit.DBGet(libocit.DBRepo, id)
	if !ok {
		return
	}
	repo, _ := libocit.RepoFromString(val.String())
	if repo.Refresh() {
		CleanRepo(repo)
		cases := repo.GetCases()
		for index := 0; index < len(cases); index++ {
			fmt.Println("case loaded ", cases[index])
			libocit.DBAdd(libocit.DBCase, cases[index])
		}
	}
}

func RefreshRepos(w http.ResponseWriter, r *http.Request) {
	var ret libocit.HttpRet

	ids := libocit.DBLookup(libocit.DBRepo, libocit.DBQuery{})
	for index := 0; index < len(ids); index++ {
		RefreshRepo(ids[index])
	}
	ret.Status = libocit.RetStatusOK
	retInfo, _ := json.MarshalIndent(ret, "", "\t")
	w.Write([]byte(retInfo))
}

func ListCases(w http.ResponseWriter, r *http.Request) {
	//Need better explaination of 'status', currently, only hasReport/isUpdated
	var query libocit.DBQuery
	page_string := r.URL.Query().Get("Page")
	page, err := strconv.Atoi(page_string)
	if err == nil {
		query.Page = page
	}
	pageSizeString := r.URL.Query().Get("PageSize")
	pageSize, err := strconv.Atoi(pageSizeString)
	if err != nil {
		query.PageSize = pageSize
	}

	status := r.URL.Query().Get("Status")
	if len(status) > 0 {
		query.Params = make(map[string]string)
		query.Params["Status"] = status
	}
	ids := libocit.DBLookup(libocit.DBCase, query)

	var caseList []libocit.TestCase
	for index := 0; index < len(ids); index++ {
		if val, ok := libocit.DBGet(libocit.DBCase, ids[index]); ok {
			tc, _ := libocit.CaseFromString(val.String())
			caseList = append(caseList, tc)
		}
	}

	var ret libocit.HttpRet
	ret.Status = libocit.RetStatusOK
	ret.Message = fmt.Sprintf("%d cases founded", len(caseList))
	ret.Data = caseList

	retInfo, _ := json.MarshalIndent(ret, "", "\t")
	w.Write([]byte(retInfo))
}

func GetCase(w http.ResponseWriter, r *http.Request) {
	//TODO: support another query method : repo/group/name
	id := r.URL.Query().Get(":ID")
	if val, ok := libocit.DBGet(libocit.DBCase, id); ok {
		tc, _ := libocit.CaseFromString(val.String())
		value := tc.GetBundleContent()

		if len(value) > 0 {
			//FIXME: add the error to head
			w.Write([]byte(value))
			return
		}
	}
	w.Write([]byte("Cannot get the case."))
}

func GetCaseReport(w http.ResponseWriter, r *http.Request) {
	var ret libocit.HttpRet
	id := r.URL.Query().Get(":ID")
	if val, ok := libocit.DBGet(libocit.DBCase, id); ok {
		tc, _ := libocit.CaseFromString(val.String())
		content := tc.GetReportContent()
		if len(content) > 0 {
			ret.Status = libocit.RetStatusOK
			ret.Data = content
			retInfo, _ := json.MarshalIndent(ret, "", "\t")
			w.Write([]byte(retInfo))
			return
		}
	}
	ret.Status = libocit.RetStatusFailed
	ret.Message = "Cannot find the report"

	retInfo, _ := json.MarshalIndent(ret, "", "\t")
	w.Write([]byte(retInfo))
}
