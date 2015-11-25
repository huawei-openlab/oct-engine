package main

import (
	"../../liboct"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

func AddTask(w http.ResponseWriter, r *http.Request) {
	var ret liboct.HttpRet
	result, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()

	db := liboct.GetDefaultDB()
	caseID := string(result)
	caseInterface, err := db.Get(liboct.DBCase, caseID)
	if err != nil {
		ret.Status = liboct.RetStatusFailed
		ret.Message = err.Error()
		retInfo, _ := json.MarshalIndent(ret, "", "\t")
		w.Write([]byte(retInfo))
		return
	}
	tc, _ := liboct.CaseFromString(caseInterface.String())

	bundleURL := tc.GetBundleTarURL()
	postURL := pubConfig.SchedulerURL
	if task, err := liboct.TestTaskNew(postURL, bundleURL, liboct.SchedulerDefaultPrio); err == nil {
		id, _ := db.Add(liboct.DBTask, task)
		ret.Status = liboct.RetStatusOK
		ret.Message = id
	} else {
		ret.Status = liboct.RetStatusFailed
		ret.Message = err.Error()
	}
	retInfo, _ := json.MarshalIndent(ret, "", "\t")
	w.Write([]byte(retInfo))
}

func GetTaskStatus(w http.ResponseWriter, r *http.Request) {
	var ret liboct.HttpRet
	db := liboct.GetDefaultDB()
	id := r.URL.Query().Get(":TaskID")
	if taskInterface, err := db.Get(liboct.DBTask, id); err == nil {
		ret.Status = liboct.RetStatusOK
		ret.Data = taskInterface
	} else {
		ret.Status = liboct.RetStatusFailed
		ret.Message = "Cannot find the task, wrong ID provided"
	}
	retInfo, _ := json.MarshalIndent(ret, "", "\t")
	w.Write([]byte(retInfo))
}

// redirect to local to test
func GetTaskReport(w http.ResponseWriter, r *http.Request) {
	var ret liboct.HttpRet
	db := liboct.GetDefaultDB()
	id := r.URL.Query().Get(":TaskID")
	taskInterface, err := db.Get(liboct.DBTask, id)
	if err != nil {
		ret.Status = liboct.RetStatusFailed
		ret.Message = "Cannot find the task, wrong ID provided"
		retInfo, _ := json.MarshalIndent(ret, "", "\t")
		w.Write([]byte(retInfo))
		return
	}

	task, _ := liboct.TaskFromString(taskInterface.String())
	// Here the task.PostURL is: http://ip:port/task/id
	getURL := fmt.Sprintf("%s/report", task.PostURL)
	fmt.Println("Get url ", getURL)
	resp, err := http.Get(getURL)
	if err != nil {
		ret.Status = liboct.RetStatusFailed
		ret.Message = "Cannot find get the report by the url"
		retInfo, _ := json.MarshalIndent(ret, "", "\t")
		w.Write([]byte(retInfo))
		return
	}
	defer resp.Body.Close()
	resp_body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		ret.Status = liboct.RetStatusFailed
		ret.Message = "Cannot find read the report by the url"
		retInfo, _ := json.MarshalIndent(ret, "", "\t")
		w.Write([]byte(retInfo))
		return
	}

	w.Write([]byte(resp_body))
}

func PostTaskAction(w http.ResponseWriter, r *http.Request) {
	var ret liboct.HttpRet
	db := liboct.GetDefaultDB()
	result, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	action, err := liboct.TestActionFromString(string(result))
	if err != nil {
		ret.Status = liboct.RetStatusFailed
		ret.Message = err.Error()
		retInfo, _ := json.MarshalIndent(ret, "", "\t")
		w.Write([]byte(retInfo))
		return
	}
	id := r.URL.Query().Get(":TaskID")
	taskInterface, err := db.Get(liboct.DBTask, id)
	if err != nil {
		ret.Status = liboct.RetStatusFailed
		ret.Message = err.Error()
		retInfo, _ := json.MarshalIndent(ret, "", "\t")
		w.Write([]byte(retInfo))
		return
	}

	task, _ := liboct.TaskFromString(taskInterface.String())
	if e := task.Command(action); e == nil {
		ret.Status = liboct.RetStatusOK
	} else {
		ret.Status = liboct.RetStatusFailed
		ret.Message = e.Error()
	}
	retInfo, _ := json.MarshalIndent(ret, "", "\t")
	w.Write([]byte(retInfo))
}

func ListTasks(w http.ResponseWriter, r *http.Request) {
	//TODO status
	var query liboct.DBQuery
	db := liboct.GetDefaultDB()
	pageStr := r.URL.Query().Get("Page")
	page, err := strconv.Atoi(pageStr)
	if err == nil {
		query.Page = page
	}
	pageSizeStr := r.URL.Query().Get("PageSize")
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err == nil {
		query.PageSize = pageSize
	}

	ids := db.Lookup(liboct.DBTask, query)

	var tl []liboct.DBInterface
	for index := 0; index < len(ids); index++ {
		task, _ := db.Get(liboct.DBTask, ids[index])
		tl = append(tl, task)
	}
	var ret liboct.HttpRet
	ret.Status = liboct.RetStatusOK
	ret.Message = fmt.Sprintf("%d tasks founded", len(tl))
	ret.Data = tl
	retInfo, _ := json.MarshalIndent(ret, "", "\t")
	w.Write([]byte(retInfo))
}
