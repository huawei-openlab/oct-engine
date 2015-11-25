package main

import (
	"../../liboct"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

func ListRepos(w http.ResponseWriter, r *http.Request) {
	var query liboct.DBQuery
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

	var rl []liboct.DBInterface
	db := liboct.GetDefaultDB()
	ids := db.Lookup(liboct.DBRepo, query)
	for index := 0; index < len(ids); index++ {
		repo, _ := db.Get(liboct.DBRepo, ids[index])
		rl = append(rl, repo)
	}

	var ret liboct.HttpRet
	ret.Status = liboct.RetStatusOK
	ret.Message = fmt.Sprintf("%d repos founded", len(rl))
	ret.Data = rl
	retInfo, _ := json.MarshalIndent(ret, "", "\t")
	w.Write([]byte(retInfo))
}

func GetRepo(w http.ResponseWriter, r *http.Request) {
	var ret liboct.HttpRet
	id := r.URL.Query().Get(":ID")
	db := liboct.GetDefaultDB()
	if repo, err := db.Get(liboct.DBRepo, id); err == nil {
		ret.Status = liboct.RetStatusOK
		ret.Data = repo
	} else {

		ret.Status = liboct.RetStatusFailed
		ret.Message = err.Error()
	}
	retInfo, _ := json.MarshalIndent(ret, "", "\t")
	w.Write([]byte(retInfo))
}

func AddRepo(w http.ResponseWriter, r *http.Request) {
	var ret liboct.HttpRet
	action := r.URL.Query().Get("Action")
	db := liboct.GetDefaultDB()
	//Add and refresh
	if action == "Add" {
		var repo liboct.TestCaseRepo
		result, _ := ioutil.ReadAll(r.Body)
		r.Body.Close()
		err := json.Unmarshal([]byte(result), &repo)
		if err != nil {
			ret.Status = liboct.RetStatusFailed
			ret.Message = fmt.Sprintf("The repo is invalid.")
		} else {
			if err := repo.IsValid(); err == nil {
				if id, e := db.Add(liboct.DBRepo, repo); e == nil {
					ret.Status = liboct.RetStatusOK
					RefreshRepo(id)
				} else {
					ret.Status = liboct.RetStatusFailed
					ret.Message = e.Error()
				}
			} else {
				ret.Status = liboct.RetStatusFailed
				ret.Message = err.Error()
			}
		}
	} else if action == "Refresh" {
		ids := db.Lookup(liboct.DBRepo, liboct.DBQuery{})
		for index := 0; index < len(ids); index++ {
			RefreshRepo(ids[index])
		}
		ret.Status = liboct.RetStatusOK
	} else {
		ret.Status = liboct.RetStatusFailed
		ret.Message = "The action in AddRepo is limited to Add/Refresh"
	}

	retInfo, _ := json.MarshalIndent(ret, "", "\t")
	w.Write([]byte(retInfo))
}

func ModifyRepo(w http.ResponseWriter, r *http.Request) {
	var ret liboct.HttpRet
	repoID := r.URL.Query().Get(":ID")
	action := r.URL.Query().Get("Action")
	db := liboct.GetDefaultDB()
	val, err := db.Get(liboct.DBRepo, repoID)
	if err != nil {
		ret.Status = liboct.RetStatusFailed
		ret.Message = fmt.Sprintf("The repo %s is not exist.", repoID)
		retInfo, _ := json.MarshalIndent(ret, "", "\t")
		w.Write([]byte(retInfo))
		return
	}
	if action == "Modify" {
		//FIXME: The ID should not be changed.
		var newRepo liboct.TestCaseRepo
		result, _ := ioutil.ReadAll(r.Body)
		r.Body.Close()
		err := json.Unmarshal([]byte(result), &newRepo)
		if err != nil {
			ret.Status = liboct.RetStatusFailed
			ret.Message = fmt.Sprintf("The modified information is invalid.")
		} else {
			oldRepo, _ := liboct.RepoFromString(val.String())
			oldRepo.Modify(newRepo)
			db.Update(liboct.DBRepo, repoID, oldRepo)
			RefreshRepo(repoID)
			ret.Status = liboct.RetStatusOK
		}
	} else if action == "Refresh" {
		RefreshRepo(repoID)
		ret.Status = liboct.RetStatusOK
	} else {
		ret.Status = liboct.RetStatusFailed
		ret.Message = "The action in ModifyRepo is limited to Add/Refresh"
	}

	retInfo, _ := json.MarshalIndent(ret, "", "\t")
	w.Write([]byte(retInfo))
}

func CleanRepo(repo liboct.TestCaseRepo) {
	var query liboct.DBQuery
	db := liboct.GetDefaultDB()
	query.Params = make(map[string]string)
	query.Params["RepoID"] = repo.GetID()
	ids := db.Lookup(liboct.DBRepo, query)
	for index := 0; index < len(ids); index++ {
		db.Remove(liboct.DBCase, ids[index])
	}
}

//This refresh the 'cache' maintained in casemanager
func RefreshRepo(id string) {
	db := liboct.GetDefaultDB()
	val, err := db.Get(liboct.DBRepo, id)
	if err != nil {
		return
	}
	repo, _ := liboct.RepoFromString(val.String())
	if repo.Refresh() {
		CleanRepo(repo)
		cases := repo.GetCases()
		for index := 0; index < len(cases); index++ {
			fmt.Println("case loaded ", cases[index])
			db.Add(liboct.DBCase, cases[index])
		}
	}
}

func RefreshRepos(w http.ResponseWriter, r *http.Request) {
	var ret liboct.HttpRet

	db := liboct.GetDefaultDB()
	ids := db.Lookup(liboct.DBRepo, liboct.DBQuery{})
	for index := 0; index < len(ids); index++ {
		RefreshRepo(ids[index])
	}
	ret.Status = liboct.RetStatusOK
	retInfo, _ := json.MarshalIndent(ret, "", "\t")
	w.Write([]byte(retInfo))
}

func ListCases(w http.ResponseWriter, r *http.Request) {
	//Need better explaination of 'status', currently, only hasReport/isUpdated
	db := liboct.GetDefaultDB()
	var query liboct.DBQuery
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
	ids := db.Lookup(liboct.DBCase, query)

	var caseList []liboct.TestCase
	for index := 0; index < len(ids); index++ {
		if val, err := db.Get(liboct.DBCase, ids[index]); err == nil {
			tc, _ := liboct.CaseFromString(val.String())
			caseList = append(caseList, tc)
		}
	}

	var ret liboct.HttpRet
	ret.Status = liboct.RetStatusOK
	ret.Message = fmt.Sprintf("%d cases founded", len(caseList))
	ret.Data = caseList

	retInfo, _ := json.MarshalIndent(ret, "", "\t")
	w.Write([]byte(retInfo))
}

func GetCase(w http.ResponseWriter, r *http.Request) {
	//TODO: support another query method : repo/group/name
	db := liboct.GetDefaultDB()
	id := r.URL.Query().Get(":ID")
	if val, err := db.Get(liboct.DBCase, id); err == nil {
		tc, _ := liboct.CaseFromString(val.String())
		value := tc.GetBundleContent()

		if len(value) > 0 {
			//FIXME: add the error to head
			w.Write([]byte(value))
			return
		}
	}
	w.WriteHeader(404)
	w.Write([]byte("Cannot get the case."))
}

func GetCaseReport(w http.ResponseWriter, r *http.Request) {
	var ret liboct.HttpRet
	db := liboct.GetDefaultDB()
	id := r.URL.Query().Get(":ID")
	if val, err := db.Get(liboct.DBCase, id); err == nil {
		tc, _ := liboct.CaseFromString(val.String())
		content := tc.GetReportContent()
		if len(content) > 0 {
			ret.Status = liboct.RetStatusOK
			ret.Data = content
			retInfo, _ := json.MarshalIndent(ret, "", "\t")
			w.Write([]byte(retInfo))
			return
		}
	}
	ret.Status = liboct.RetStatusFailed
	ret.Message = "Cannot find the report"

	retInfo, _ := json.MarshalIndent(ret, "", "\t")
	w.Write([]byte(retInfo))
}
