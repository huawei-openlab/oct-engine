package main

import (
	"../../lib/libocit"
	"../../lib/routes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

const SchedularCacheDir = "/tmp/.test_schedular_cache"

type SchedulerConfig struct {
	Port           int
	ServerListFile string
	CacheDir       string
	Debug          bool
}

func GetResult(w http.ResponseWriter, r *http.Request) {
	//TODO
	id := r.URL.Query().Get(":ID")
	fmt.Println(id)
}

//List all the hostOS status
func GetStatus(w http.ResponseWriter, r *http.Request) {
}

func ReceiveTask(w http.ResponseWriter, r *http.Request) {
	realURL, params := libocit.ReceiveFile(w, r, SchedularCacheDir)
	taskID := params["id"]

	//TODO
	//Untar the file and load the case.json
	content := ""
	if pub_config.Debug {
		fmt.Println(content)
	}
	var ret libocit.HttpRet
	var tc libocit.TestCase
	if err := json.Unmarshal([]byte(content), &tc); err != nil {
		ret.Status = "Failed"
		ret.Message = "The testcase is not standard. (.tar.gz or .json)"
		ret_string, _ := json.MarshalIndent(ret, "", "\t")
		w.Write([]byte(ret_string))
		return
	} else {
		ret.Status = "OK"
		ret.Message = "success in receiving task files"
		ret_string, _ := json.MarshalIndent(ret, "", "\t")
		w.Write([]byte(ret_string))
	}

	tc.SetBundleURL(realURL)
	task := SchedulerTaskNew(taskID, tc)

	if !task.Run(libocit.TestActionApply) {
		return
	}
	if !task.Run(libocit.TestActionDeploy) {
		return
	}
	if !task.Run(libocit.TestActionRun) {
		return
	}
	if !task.Run(libocit.TestActionCollect) {
		return
	}
	if !task.Run(libocit.TestActionDestroy) {
		return
	}
}

// Will use DB in the future, (mongodb for example)
func init() {
}

var pub_config SchedulerConfig

func main() {
	test()
	return

	config_content := libocit.ReadFile("./testserver.conf")
	json.Unmarshal([]byte(config_content), &pub_config)

	mux := routes.New()

	mux.Get("/resource", GetResource)
	mux.Post("/resource", PostResource)
	mux.Get("/:ID/status", GetStatus)
	mux.Post("/task", ReceiveTask)
	mux.Get("/:ID/result", GetResult)

	http.Handle("/", mux)
	port := fmt.Sprintf(":%d", pub_config.Port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func test() {
	err := SchedulerRInitFromFile("./servers.conf")
	if err != nil {
		fmt.Println(err)
		return
	}
	realURL := "/tmp/.schedular_cache/1446707259/bundle.tar.gz"
	taskID := "1446707259"

	//TODO
	//Untar the file and load the case.json
	content := libocit.ReadCaseFromTar(realURL)
	fmt.Println(content)

	var tc libocit.TestCase
	if err = json.Unmarshal([]byte(content), &tc); err != nil {
		fmt.Println("The testcase is not standard. (.tar.gz or .json)")
		return
	}

	tc.SetBundleURL(realURL)
	task := SchedulerTaskNew(taskID, tc)

	if !task.Run(libocit.TestActionApply) {
		fmt.Println("Failed in main apply ")
		return
	}
	fmt.Println(task.GetStatus())

	if !task.Run(libocit.TestActionDeploy) {
		fmt.Println("Failed in main deploy ")
		return
	}
	fmt.Println(task.GetStatus())

	if !task.Run(libocit.TestActionRun) {
		fmt.Println("Failed in main run ")
		return
	}
	fmt.Println(task.GetStatus())
	if !task.Run(libocit.TestActionCollect) {
		fmt.Println("Failed in main collect ")
		return
	}
	fmt.Println(task.GetStatus())
	if !task.Run(libocit.TestActionDestroy) {
		fmt.Println("Failed in main destroy ")
		return
	}
	fmt.Println(task.GetStatus())
}
