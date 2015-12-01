package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/huawei-openlab/oct-engine/liboct"
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

	if len(ids) == 0 {
		liboct.RenderErrorf(w, "Cannot find the avaliable resource")
	} else {
		var rss []liboct.DBInterface
		for index := 0; index < len(ids); index++ {
			res, _ := db.Get(liboct.DBResource, ids[index])
			rss = append(rss, res)
		}
		liboct.RenderOK(w, "Find the avaliable resource", rss)
	}
}

func PostResource(w http.ResponseWriter, r *http.Request) {
	var res liboct.Resource
	db := liboct.GetDefaultDB()
	result, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	logrus.Debugf(string(result))
	json.Unmarshal([]byte(result), &res)
	if err := res.IsValid(); err != nil {
		liboct.RenderError(w, err)
	} else {
		lock.Lock()
		if id, err := db.Add(liboct.DBResource, res); err != nil {
			liboct.RenderError(w, err)
		} else {
			liboct.RenderOK(w, fmt.Sprintf("Success in adding the resource: %s ", id), nil)
		}
		lock.Unlock()
	}
}

func DeleteResource(w http.ResponseWriter, r *http.Request) {
	db := liboct.GetDefaultDB()
	id := r.URL.Query().Get("ID")
	lock.Lock()
	if err := db.Remove(liboct.DBResource, id); err != nil {
		liboct.RenderError(w, err)
	} else {
		liboct.RenderOK(w, "Success in remove the resource", nil)
	}
	lock.Unlock()
}

var lock = sync.RWMutex{}
