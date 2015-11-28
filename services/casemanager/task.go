package main

import (
	"../../liboct"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/Sirupsen/logrus"
)

func AddTask(w http.ResponseWriter, r *http.Request) {
	result, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()

	db := liboct.GetDefaultDB()
	caseID := string(result)
	caseInterface, err := db.Get(liboct.DBCase, caseID)
	if err != nil {
		liboct.RenderError(w, err)
		return
	}
	tc, _ := liboct.CaseFromString(caseInterface.String())

	bundleURL := tc.GetBundleTarURL()
	postURL := pubConfig.SchedulerURL
	if task, err := liboct.TestTaskNew(postURL, bundleURL, liboct.SchedulerDefaultPrio); err != nil {
		liboct.RenderError(w, err)
	} else {
		id, _ := db.Add(liboct.DBTask, task)
		liboct.RenderOK(w, id, nil)
	}
}

func GetTaskStatus(w http.ResponseWriter, r *http.Request) {
	db := liboct.GetDefaultDB()
	id := r.URL.Query().Get(":TaskID")
	if taskInterface, err := db.Get(liboct.DBTask, id); err != nil {
		liboct.RenderError(w, err)
	} else {
		liboct.RenderOK(w, "", taskInterface)
	}
}

// redirect to local to test
func GetTaskReport(w http.ResponseWriter, r *http.Request) {
	db := liboct.GetDefaultDB()
	id := r.URL.Query().Get(":TaskID")
	taskInterface, err := db.Get(liboct.DBTask, id)
	if err != nil {
		liboct.RenderError(w, err)
		return
	}

	task, _ := liboct.TaskFromString(taskInterface.String())
	// Here the task.PostURL is: http://ip:port/task/id
	getURL := fmt.Sprintf("%s/report", task.PostURL)
	logrus.Debugf("Get url ", getURL)
	resp, err := http.Get(getURL)
	if err != nil {
		liboct.RenderError(w, err)
		return
	}
	defer resp.Body.Close()
	resp_body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		liboct.RenderError(w, err)
		return
	}

	w.Write([]byte(resp_body))
}

func PostTaskAction(w http.ResponseWriter, r *http.Request) {
	db := liboct.GetDefaultDB()
	result, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	action, err := liboct.TestActionFromString(string(result))
	if err != nil {
		liboct.RenderError(w, err)
		return
	}
	id := r.URL.Query().Get(":TaskID")
	taskInterface, err := db.Get(liboct.DBTask, id)
	if err != nil {
		liboct.RenderError(w, err)
		return
	}

	task, _ := liboct.TaskFromString(taskInterface.String())
	if e := task.Command(action); e != nil {
		liboct.RenderError(w, err)
	} else {
		liboct.RenderOK(w, "", nil)
	}
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
	liboct.RenderOK(w, fmt.Sprintf("%d tasks founded", len(tl)), tl)
}
