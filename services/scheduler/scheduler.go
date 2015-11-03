package main

import (
	"../../lib/libocit"
	"../../lib/routes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"time"
)

type ServerConfig struct {
	TSurl string
	CPurl string
	TCurl string
	Port  int
	Debug bool
}

const SchedularDefaultPrio = 100
const SchedularCacheDir = "/tmp/.schedular_cache"

//public variable
var pubConfig ServerConfig

var taskStore map[string]libocit.TestTask

func AddCase(id string, bundleName string) {
	url := fmt.Sprintf("%s/case/%s", pubConfig.TCurl, id)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		return
	}
	resp_body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		fmt.Println(err)
		return
	}

	taskID := fmt.Sprintf("%d", time.Now().Unix())
	taskDir := libocit.PreparePath(path.Join(SchedularCacheDir, taskID), "")
	//FIXME: I think I need to add a state.json to keep track of all the testing status ...
	bundleURL := fmt.Sprintf("%s/%s.tar.gz", taskDir, bundleName)
	f, err := os.Create(bundleURL)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()
	f.Write(resp_body)
	f.Sync()

	postURL := pubConfig.TSurl + "/task"
	task := libocit.TestTaskNew(taskID, postURL, bundleURL, SchedularDefaultPrio)
	taskStore[taskID] = task
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

	var tl []libocit.TestTask
	cur_num := 0
	for _, taskInStore := range taskStore {
		cur_num += 1
		if (cur_num >= page*pageSize) && (cur_num < (page+1)*pageSize) {
			tl = append(tl, taskInStore)
		}
	}

	var ret libocit.HttpRet
	ret.Status = "OK"
	ret.Message = fmt.Sprintf("%d tasks founded", len(tl))
	ret.Data = tl
	retInfo, _ := json.MarshalIndent(ret, "", "\t")
	w.Write([]byte(retInfo))

}

func GetTaskInfo(w http.ResponseWriter, r *http.Request) {
	var ret libocit.HttpRet
	taskID := r.URL.Query().Get(":ID")

	for _, taskInStore := range taskStore {
		if taskID == taskInStore.ID {
			ret.Status = "OK"
			ret.Data = taskInStore
			retInfo, _ := json.MarshalIndent(ret, "", "\t")
			w.Write([]byte(retInfo))
			return
		}
	}

	ret.Status = "Failed"
	ret.Message = "Cannot find the task, wrong ID provided"
	retInfo, _ := json.MarshalIndent(ret, "", "\t")
	w.Write([]byte(retInfo))
}

func LoadCases() {
	pageSize := 30
	for page := 0; ; page++ {
		url := fmt.Sprintf("%s/case?Page=%d&PageSize=%d", pubConfig.TCurl, page, pageSize)
		resp, err := http.Get(url)
		if err != nil {
			fmt.Println(err)
			break
		}
		resp_body, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			fmt.Println(err)
			break
		} else {
			//FIXME: how to convert interface to struct ?
			//Donnot need to add a new struct
			type tmpRet struct {
				Status  string
				Message string
				Data    []libocit.TestCase
			}

			var ret tmpRet
			json.Unmarshal([]byte(resp_body), &ret)
			for index := 0; index < len(ret.Data); index++ {
				AddCase(ret.Data[index].ID, ret.Data[index].GetBundleName())
			}
			if len(ret.Data) <= pageSize {
				break
			}
		}
	}
}

func init() {
	config_content := libocit.ReadFile("./scheduler.conf")
	json.Unmarshal([]byte(config_content), &pubConfig)

	taskStore = make(map[string]libocit.TestTask)
	//	postURL := pubConfig.TSurl + "/task"
	LoadCases()
}

func main() {
	mux := routes.New()
	mux.Get("/task", ListTasks)
	mux.Get("/task/:ID", GetTaskInfo)
	http.Handle("/", mux)

	port := fmt.Sprintf(":%d", pubConfig.Port)
	fmt.Println("Listen to ", port)

	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
	return
}
