package main

import (
	"../../lib/libocit"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

const CMTKey = "case manager task"

func AddTask(w http.ResponseWriter, r *http.Request) {
	var ret libocit.HttpRet
	result, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()

	caseID := string(result)
	caseInterface, ok := libocit.DBGet(libocit.DBCase, caseID)
	if !ok {
		ret.Status = libocit.RetStatusFailed
		ret.Message = "Invalid case id"
		retInfo, _ := json.MarshalIndent(ret, "", "\t")
		w.Write([]byte(retInfo))
		return
	}
	tc, _ := libocit.CaseFromString(caseInterface.String())

	bundleURL := tc.GetBundleURL()
	postURL := pubConfig.SchedulerURL
	task := libocit.TestTaskNew(postURL, bundleURL, libocit.SchedularDefaultPrio)
	if id, ok := libocit.DBAdd(libocit.DBTask, task); ok {
		ret.Status = libocit.RetStatusOK
		ret.Data = id
	} else {
		ret.Status = libocit.RetStatusFailed
	}
	retInfo, _ := json.MarshalIndent(ret, "", "\t")
	w.Write([]byte(retInfo))
}

func GetTaskStatus(w http.ResponseWriter, r *http.Request) {
	var ret libocit.HttpRet
	id := r.URL.Query().Get(":ID")
	if taskInterface, ok := libocit.DBGet(libocit.DBTask, id); ok {
		ret.Status = libocit.RetStatusOK
		ret.Data = taskInterface
	} else {
		ret.Status = libocit.RetStatusFailed
		ret.Message = "Cannot find the task, wrong ID provided"
	}
	retInfo, _ := json.MarshalIndent(ret, "", "\t")
	w.Write([]byte(retInfo))
}

func PostTaskAction(w http.ResponseWriter, r *http.Request) {
	var ret libocit.HttpRet
	id := r.URL.Query().Get(":ID")
	if _, ok := libocit.DBGet(libocit.DBTask, id); ok {
		//TODO: send to schedular
		ret.Status = libocit.RetStatusOK
	} else {
		ret.Status = libocit.RetStatusFailed
		ret.Message = "Cannot find the task, wrong ID provided"
	}
	retInfo, _ := json.MarshalIndent(ret, "", "\t")
	w.Write([]byte(retInfo))
}

func ListTasks(w http.ResponseWriter, r *http.Request) {
	//TODO status
	var query libocit.DBQuery
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

	ids := libocit.DBLookup(libocit.DBTask, query)

	var tl []libocit.DBInterface
	for index := 0; index < len(ids); index++ {
		task, _ := libocit.DBGet(libocit.DBTask, ids[index])
		tl = append(tl, task)
	}
	var ret libocit.HttpRet
	ret.Status = libocit.RetStatusOK
	ret.Message = fmt.Sprintf("%d tasks founded", len(tl))
	ret.Data = tl
	retInfo, _ := json.MarshalIndent(ret, "", "\t")
	w.Write([]byte(retInfo))
}
