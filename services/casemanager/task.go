package main

import (
	"../../lib/liboct"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

const CMTKey = "case manager task"

func AddTask(w http.ResponseWriter, r *http.Request) {
	var ret liboct.HttpRet
	result, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()

	caseID := string(result)
	caseInterface, ok := liboct.DBGet(liboct.DBCase, caseID)
	if !ok {
		ret.Status = liboct.RetStatusFailed
		ret.Message = "Invalid case id"
		retInfo, _ := json.MarshalIndent(ret, "", "\t")
		w.Write([]byte(retInfo))
		return
	}
	tc, _ := liboct.CaseFromString(caseInterface.String())

	bundleURL := tc.GetBundleURL()
	postURL := pubConfig.SchedulerURL
	if task, ok := liboct.TestTaskNew(postURL, bundleURL, liboct.SchedulerDefaultPrio); ok {
		id, _ := liboct.DBAdd(liboct.DBTask, task)
		ret.Status = liboct.RetStatusOK
		ret.Message = id
	} else {
		ret.Status = liboct.RetStatusFailed
	}
	retInfo, _ := json.MarshalIndent(ret, "", "\t")
	w.Write([]byte(retInfo))
}

func GetTaskStatus(w http.ResponseWriter, r *http.Request) {
	var ret liboct.HttpRet
	id := r.URL.Query().Get(":ID")
	if taskInterface, ok := liboct.DBGet(liboct.DBTask, id); ok {
		ret.Status = liboct.RetStatusOK
		ret.Data = taskInterface
	} else {
		ret.Status = liboct.RetStatusFailed
		ret.Message = "Cannot find the task, wrong ID provided"
	}
	retInfo, _ := json.MarshalIndent(ret, "", "\t")
	w.Write([]byte(retInfo))
}

func PostTaskAction(w http.ResponseWriter, r *http.Request) {
	var ret liboct.HttpRet
	result, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	action, ok := liboct.TestActionFromString(string(result))
	if !ok {
		ret.Status = liboct.RetStatusFailed
		ret.Message = "Invalid action, should limit to 'deploy,run,collect and destroy'"
		retInfo, _ := json.MarshalIndent(ret, "", "\t")
		w.Write([]byte(retInfo))
		return
	}
	id := r.URL.Query().Get(":ID")
	taskInterface, ok := liboct.DBGet(liboct.DBTask, id)
	if !ok {
		ret.Status = liboct.RetStatusFailed
		ret.Message = "Cannot find the task, wrong ID provided"
		retInfo, _ := json.MarshalIndent(ret, "", "\t")
		w.Write([]byte(retInfo))
		return
	}

	task, _ := liboct.TaskFromString(taskInterface.String())
	if task.Command(action) {
		ret.Status = liboct.RetStatusOK
	} else {
		ret.Status = liboct.RetStatusFailed
		ret.Message = "Failed to call the action"
	}
	retInfo, _ := json.MarshalIndent(ret, "", "\t")
	w.Write([]byte(retInfo))
}

func ListTasks(w http.ResponseWriter, r *http.Request) {
	//TODO status
	var query liboct.DBQuery
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

	ids := liboct.DBLookup(liboct.DBTask, query)

	var tl []liboct.DBInterface
	for index := 0; index < len(ids); index++ {
		task, _ := liboct.DBGet(liboct.DBTask, ids[index])
		tl = append(tl, task)
	}
	var ret liboct.HttpRet
	ret.Status = liboct.RetStatusOK
	ret.Message = fmt.Sprintf("%d tasks founded", len(tl))
	ret.Data = tl
	retInfo, _ := json.MarshalIndent(ret, "", "\t")
	w.Write([]byte(retInfo))
}
