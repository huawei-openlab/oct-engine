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
	result, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()

	caseID := string(result)
	tc, ok := caseStore[caseID]

	fmt.Println(caseStore)
	fmt.Println(caseID)
	if ok {
		bundleURL := tc.GetBundleURL()
		fmt.Println(bundleURL)
		//		task, ok := libocit.TestTaskNew(postURL, bundleURL, SchedularDefaultPrio)
		/*
			var ret libocit.HttpRet
			if ok {
				ret.Status = libocit.RetStatusOK
				ret.Data = task.ID
			} else {
				ret.Status = libocit.RetStatusFailed
			}
			retInfo, _ := json.MarshalIndent(ret, "", "\t")
			w.Write([]byte(retInfo))
		*/
	}
}

func GetTaskStatus(w http.ResponseWriter, r *http.Request) {
	var ret libocit.HttpRet
	//taskID := r.URL.Query().Get(":ID")
	/*
		for _, taskInStore := range taskStore {
			if taskID == taskInStore.ID {
				ret.Status = "OK"
				ret.Data = taskInStore
				retInfo, _ := json.MarshalIndent(ret, "", "\t")
				w.Write([]byte(retInfo))
				return
			}
		}
	*/
	ret.Status = "Failed"
	ret.Message = "Cannot find the task, wrong ID provided"
	retInfo, _ := json.MarshalIndent(ret, "", "\t")
	w.Write([]byte(retInfo))

}

func PostTaskAction(w http.ResponseWriter, r *http.Request) {

	var ret libocit.HttpRet
	//taskID := r.URL.Query().Get(":ID")
	/*
		for _, taskInStore := range taskStore {
			if taskID == taskInStore.ID {
				if taskInStore.Run() {
					ret.Status = "OK"
				} else {
					ret.Status = "Failed"
				}
				ret.Data = taskInStore
				retInfo, _ := json.MarshalIndent(ret, "", "\t")
				w.Write([]byte(retInfo))
				return
			}
		}
	*/
	ret.Status = "Failed"
	ret.Message = "Cannot find the task, wrong ID provided"
	retInfo, _ := json.MarshalIndent(ret, "", "\t")
	w.Write([]byte(retInfo))

}

func ListTasks(w http.ResponseWriter, r *http.Request) {
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
	fmt.Println(page, pageSize)
	var tl []libocit.TestTask
	/*
		for _, taskInStore := range taskStore {
			cur_num += 1
			if (cur_num >= page*pageSize) && (cur_num < (page+1)*pageSize) {
				tl = append(tl, taskInStore)
			}
		}
	*/
	var ret libocit.HttpRet
	ret.Status = "OK"
	ret.Message = fmt.Sprintf("%d tasks founded", len(tl))
	ret.Data = tl
	retInfo, _ := json.MarshalIndent(ret, "", "\t")
	w.Write([]byte(retInfo))

}
