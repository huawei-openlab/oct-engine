package main

import (
	"../../liboct"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"
)

func GetResourceQuery(r *http.Request) liboct.DBQuery {
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

func GetResourceStatus(w http.ResponseWriter, r *http.Request) {
}

func GetResource(w http.ResponseWriter, r *http.Request) {
	query := GetResourceQuery(r)
	db := liboct.GetDefaultDB()
	ids := db.Lookup(liboct.DBResource, query)

	var ret liboct.HttpRet
	if len(ids) == 0 {
		ret.Status = liboct.RetStatusFailed
		ret.Message = "Cannot find the avaliable resource"
	} else {
		ret.Status = liboct.RetStatusOK
		ret.Message = "Find the avaliable resource"
		var rss []liboct.DBInterface
		for index := 0; index < len(ids); index++ {
			res, _ := db.Get(liboct.DBResource, ids[index])
			rss = append(rss, res)
		}

		ret.Data = rss
	}

	body, _ := json.MarshalIndent(ret, "", "\t")
	w.Write([]byte(body))
}

func PostResource(w http.ResponseWriter, r *http.Request) {
	var res liboct.Resource
	var ret liboct.HttpRet
	db := liboct.GetDefaultDB()
	result, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	if pubConfig.Debug {
		fmt.Println(string(result))
	}
	json.Unmarshal([]byte(result), &res)
	if err := res.IsValid(); err != nil {
		ret.Status = liboct.RetStatusFailed
		ret.Message = err.Error()
	} else {
		lock.Lock()
		if id, e := db.Add(liboct.DBResource, res); e == nil {
			ret.Status = "OK"
			ret.Message = fmt.Sprintf("Success in adding the resource: %s ", id)
		} else {
			ret.Status = liboct.RetStatusFailed
			ret.Message = e.Error()
		}
		lock.Unlock()
	}
	ret_body, _ := json.MarshalIndent(ret, "", "\t")
	w.Write([]byte(ret_body))
}

func DeleteResource(w http.ResponseWriter, r *http.Request) {
	var ret liboct.HttpRet
	db := liboct.GetDefaultDB()
	id := r.URL.Query().Get("ID")
	lock.Lock()
	if err := db.Remove(liboct.DBResource, id); err == nil {
		ret.Status = liboct.RetStatusOK
		ret.Message = "Success in remove the resource"
	} else {
		ret.Status = liboct.RetStatusFailed
		ret.Message = err.Error()
	}
	lock.Unlock()
	ret_body, _ := json.MarshalIndent(ret, "", "\t")
	w.Write([]byte(ret_body))
}

var lock = sync.RWMutex{}
