package main

import (
	"../../lib/libocit"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"
)

func GetResourceQuery(r *http.Request) libocit.DBQuery {
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

	//TODO: use a list , also add the list to the lib
	query.Params = make(map[string]string)
	query.Params["Class"] = r.URL.Query().Get("Class")
	query.Params["Distribution"] = r.URL.Query().Get("Distribution")
	query.Params["Version"] = r.URL.Query().Get("Version")
	query.Params["Arch"] = r.URL.Query().Get("Arch")
	query.Params["CPU"] = r.URL.Query().Get("CPU")
	query.Params["Memory"] = r.URL.Query().Get("Memory")

	return query
}

func GetResource(w http.ResponseWriter, r *http.Request) {
	query := GetResourceQuery(r)

	ids := libocit.DBLookup(libocit.DBResource, query)

	var ret libocit.HttpRet
	if len(ids) == 0 {
		ret.Status = libocit.RetStatusFailed
		ret.Message = "Cannot find the avaliable resource"
	} else {
		ret.Status = libocit.RetStatusOK
		ret.Message = "Find the avaliable resource"
		var rss []libocit.DBInterface
		for index := 0; index < len(ids); index++ {
			res, _ := libocit.DBGet(libocit.DBResource, ids[index])
			rss = append(rss, res)
		}

		ret.Data = rss
	}

	body, _ := json.MarshalIndent(ret, "", "\t")
	w.Write([]byte(body))
}

func PostResource(w http.ResponseWriter, r *http.Request) {
	var res SchedulerResource
	var ret libocit.HttpRet

	result, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	if pub_config.Debug {
		fmt.Println(string(result))
	}
	json.Unmarshal([]byte(result), &res)
	if err := res.IsValid(); err != nil {
		ret.Status = libocit.RetStatusFailed
		ret.Message = err.Error()
	} else {
		lock.Lock()
		if id, ok := libocit.DBAdd(libocit.DBResource, res); ok {
			ret.Status = "OK"
			ret.Message = fmt.Sprintf("Success in adding the resource: %s ", id)
		} else {
			ret.Status = libocit.RetStatusFailed
			ret.Message = "this resource is already exist"
		}
		lock.Unlock()
	}
	ret_body, _ := json.MarshalIndent(ret, "", "\t")
	w.Write([]byte(ret_body))
}

func DeleteResource(w http.ResponseWriter, r *http.Request) {
	var ret libocit.HttpRet
	id := r.URL.Query().Get("ID")
	lock.Lock()
	if libocit.DBRemove(libocit.DBResource, id) {
		ret.Status = libocit.RetStatusOK
		ret.Message = "Success in remove the resource"
	} else {
		ret.Status = libocit.RetStatusFailed
		ret.Message = "Cannot find the resource"
	}
	lock.Unlock()
	ret_body, _ := json.MarshalIndent(ret, "", "\t")
	w.Write([]byte(ret_body))
}

var lock = sync.RWMutex{}
