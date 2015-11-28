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

	liboct.RenderOK(w, fmt.Sprintf("%d repos founded", len(rl)), rl)
}

func GetRepo(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get(":ID")
	db := liboct.GetDefaultDB()
	if repo, err := db.Get(liboct.DBRepo, id); err != nil {
		liboct.RenderError(w, err)
	} else {
		liboct.RenderOK(w, "", repo)
	}
}

func AddRepo(w http.ResponseWriter, r *http.Request) {
	action := r.URL.Query().Get("Action")
	db := liboct.GetDefaultDB()
	//Add and refresh
	if action == "Add" {
		var repo liboct.TestCaseRepo
		result, _ := ioutil.ReadAll(r.Body)
		r.Body.Close()
		err := json.Unmarshal([]byte(result), &repo)
		if err != nil {
			liboct.RenderError(w, err)
		} else {
			if err := repo.IsValid(); err == nil {
				if id, e := db.Add(liboct.DBRepo, repo); e != nil {
					liboct.RenderError(w, err)
				} else {
					RefreshRepo(id)
					liboct.RenderOK(w, "", nil)
				}
			} else {
				liboct.RenderError(w, err)
			}
		}
	} else if action == "Refresh" {
		ids := db.Lookup(liboct.DBRepo, liboct.DBQuery{})
		for index := 0; index < len(ids); index++ {
			RefreshRepo(ids[index])
		}
		liboct.RenderOK(w, "", nil)
	} else {
		liboct.RenderErrorf(w, "The action in AddRepo is limited to Add/Refresh")
	}
}

func ModifyRepo(w http.ResponseWriter, r *http.Request) {
	repoID := r.URL.Query().Get(":ID")
	action := r.URL.Query().Get("Action")
	db := liboct.GetDefaultDB()
	val, err := db.Get(liboct.DBRepo, repoID)
	if err != nil {
		liboct.RenderError(w, err)
		return
	}
	if action == "Modify" {
		var newRepo liboct.TestCaseRepo
		result, _ := ioutil.ReadAll(r.Body)
		r.Body.Close()
		err := json.Unmarshal([]byte(result), &newRepo)
		if err != nil {
			liboct.RenderError(w, err)
		} else {
			oldRepo, _ := liboct.RepoFromString(val.String())
			oldRepo.Modify(newRepo)
			db.Update(liboct.DBRepo, repoID, oldRepo)
			RefreshRepo(repoID)
			liboct.RenderOK(w, "", nil)
		}
	} else if action == "Refresh" {
		RefreshRepo(repoID)
		liboct.RenderOK(w, "", nil)
	} else {
		liboct.RenderErrorf(w, "The action in ModifyRepo is limited to Add/Refresh")
	}
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
	db := liboct.GetDefaultDB()
	ids := db.Lookup(liboct.DBRepo, liboct.DBQuery{})
	for index := 0; index < len(ids); index++ {
		RefreshRepo(ids[index])
	}
	liboct.RenderOK(w, "", nil)
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

	liboct.RenderOK(w, fmt.Sprintf("%d cases founded", len(caseList)), caseList)
}

func GetCase(w http.ResponseWriter, r *http.Request) {
	//TODO: support another query method : repo/group/name
	db := liboct.GetDefaultDB()
	id := r.URL.Query().Get(":ID")
	if val, err := db.Get(liboct.DBCase, id); err != nil {
		liboct.RenderError(w, err)
	} else {
		tc, _ := liboct.CaseFromString(val.String())
		value := tc.GetBundleContent()

		if len(value) > 0 {
			w.Write([]byte(value))
		} else {
			liboct.RenderErrorf(w, "Case is empty.")
		}
	}
}

func GetCaseReport(w http.ResponseWriter, r *http.Request) {
	db := liboct.GetDefaultDB()
	id := r.URL.Query().Get(":ID")
	if val, err := db.Get(liboct.DBCase, id); err != nil {
		liboct.RenderError(w, err)
	} else {
		tc, _ := liboct.CaseFromString(val.String())
		content := tc.GetReportContent()
		if len(content) > 0 {
			liboct.RenderOK(w, "", content)
		} else {
			liboct.RenderErrorf(w, "Case report is empty.")
		}
	}
}
