package main

import (
	"../../lib/libocit"
	"../../lib/routes"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
)

type CaseRepo struct {
	Name       string
	CaseFolder string
	//We can disable a repo
	Enable bool
	Groups []string
}

type TCServerConf struct {
	Repos    []CaseRepo
	CacheDir string
	Port     int
	Debug    bool
}

type MetaUnit struct {
	ID            string
	Name          string
	GroupDir      string
	LibFolderName string
	Status        string
	//0 means not tested
	TestedTime       int64
	LastModifiedTime int64
}

var store = map[string]*MetaUnit{}
var pub_config TCServerConf

func RefreshRepo(repo CaseRepo) {
	if repo.Enable == false {
		if repo.Debug == true {
			fmt.Println("The repo ", repo.Name, " is disabled")
		}
		return
	}

	var cmd string
	repoDir := libocit.PreparePath(pub_config.CacheDir, repo.Name)

	git_check_url := path.Join(repoDir, ".git/config")
	_, err := os.Stat(git_check_url)
	if err != nil {
		//Clean: remove the '/$' if there was one
		cmd = "cd " + path.Dir(path.Clean(repoDir)) + " ; git clone https://" + repo.Name + ".git"
	} else {
		cmd = "cd " + repoDir + " ; git pull"
	}

	if repo.Debug == true {
		fmt.Println("Refresh by using ", cmd)
	}
	c := exec.Command("/bin/sh", "-c", cmd)
	c.Run()
	if repo.Debug == true {
		fmt.Println("Refresh done")
	}
}

func LastModified(case_dir string) (last_modified int64) {
	last_modified = 0
	files, _ := ioutil.ReadDir(case_dir)
	for _, file := range files {
		if file.IsDir() {
			sub_lm := LastModified(path.Join(case_dir, file.Name()))
			if last_modified < sub_lm {
				last_modified = sub_lm
			}
		} else {
			if last_modified < file.ModTime().Unix() {
				last_modified = file.ModTime().Unix()
			}
		}
	}
	return last_modified
}

func LoadCase(groupDir string, caseName string, caseLibFolderName string) {
	caseDir := path.Join(groupDir, caseName)
	_, err_msgs := libocit.ValidateByDir(caseDir, "")
	if len(err_msgs) == 0 {
		last_modified := LastModified(caseDir)
		store_md := libocit.MD5(caseDir)
		if v, ok := store[store_md]; ok {
			//Happen when we refresh the repo
			(*v).LastModifiedTime = last_modified
			fi, err := os.Stat(path.Join(caseDir, "report.md"))
			if err != nil {
				(*v).TestedTime = 0
			} else {
				(*v).TestedTime = fi.ModTime().Unix()
			}
			if (*v).LastModifiedTime > (*v).TestedTime {
				(*v).Status = "idle"
			} else {
				(*v).Status = "tested"
			}
		} else {
			var meta MetaUnit
			meta.ID = store_md
			meta.Name = caseName
			meta.GroupDir = groupDir
			meta.LibFolderName = caseLibFolderName
			fi, err := os.Stat(path.Join(caseDir, "report.md"))
			if err != nil {
				meta.TestedTime = 0
			} else {
				meta.TestedTime = fi.ModTime().Unix()
			}
			meta.LastModifiedTime = last_modified
			if meta.LastModifiedTime > meta.TestedTime {
				meta.Status = "idle"
			} else {
				meta.Status = "tested"
			}
			store[store_md] = &meta
		}
	} else {
		fmt.Println("Error in loading case: ", caseDir, " . Skip it")
		return
	}
}

func LoadCaseGroup(groupDir string, libDir string) {
	files, _ := ioutil.ReadDir(groupDir)
	for _, file := range files {
		if file.IsDir() {
			if len(libDir) > 0 {
				if libDir == file.Name() {
					continue
				} else {
					LoadCase(groupDir, file.Name(), libDir)
				}
			} else {
				LoadCase(groupDir, file.Name(), "")
			}
		}
	}
}

func LoadDB() {
	repos := pub_config.Repos
	for index := 0; index < len(repos); index++ {
		RefreshRepo(repos[index])
	}

	//TODO: get all the case job with repo infos
	for g_index := 0; g_index < len(pub_config.Groups); g_index++ {
		repo_name := strings.Replace(path.Base(pub_config.GitRepo), ".git", "", 1)
		group_dir := path.Join(pub_config.CacheDir, repo_name, pub_config.CaseFolderName, pub_config.Groups[g_index].Name)
		LoadCaseGroup(group_dir, pub_config.Groups[g_index].LibFolderName)
	}
}

func ListCases(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("Status")
	page_string := r.URL.Query().Get("Page")
	page, err := strconv.Atoi(page_string)
	if err != nil {
		page = 0
	}
	page_size_string := r.URL.Query().Get("PageSize")
	page_size, err := strconv.Atoi(page_size_string)
	if err != nil {
		page_size = 10
	}

	var case_list []MetaUnit
	cur_num := 0
	for _, tc := range store {
		if status != "" {
			if status != tc.Status {
				continue
			}
		}
		cur_num += 1
		if (cur_num >= page*page_size) && (cur_num < (page+1)*page_size) {
			case_list = append(case_list, *tc)
		}

	}

	case_string, err := json.MarshalIndent(case_list, "", "\t")
	if err != nil {
		w.Write([]byte("[]"))
	} else {
		w.Write([]byte(case_string))
	}

}

func GetCase(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get(":ID")
	meta := store[id]
	files := libocit.GetDirFiles(meta.GroupDir, meta.Name)
	if len(meta.LibFolderName) > 0 {
		lib_files := libocit.GetDirFiles(meta.GroupDir, meta.LibFolderName)
		for index := 0; index < len(lib_files); index++ {
			files = append(files, lib_files[index])
		}
	}
	tar_url := libocit.TarFileList(files, meta.GroupDir, meta.Name)

	file, err := os.Open(tar_url)
	defer file.Close()
	if err != nil {
		//FIXME: add to head
		w.Write([]byte("Cannot open the file: " + tar_url))
		return
	}

	buf := bytes.NewBufferString("")
	buf.ReadFrom(file)
	//TODO: write head, filename and the etc
	w.Write([]byte(buf.String()))
}

func GetCaseReport(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get(":ID")
	meta := store[id]
	repo_name := strings.Replace(path.Base(pub_config.GitRepo), ".git", "", 1)
	report_url := path.Join(pub_config.CacheDir, repo_name, pub_config.CaseFolderName, meta.GroupDir, meta.Name, "report.md")

	_, err := os.Stat(report_url)
	if err != nil {
		//FIXME: 404 error head
		w.Write([]byte("Cannot find the report"))
		return
	}
	content := libocit.ReadFile(report_url)
	w.Write([]byte(content))
}

func RefreshCases(w http.ResponseWriter, r *http.Request) {
	RefreshRepo()
	var ret libocit.HttpRet
	ret.Status = "OK"
	ret_string, _ := json.MarshalIndent(ret, "", "\t")
	w.Write([]byte(ret_string))
}

func main() {
	content := libocit.ReadFile("./tcserver.conf")
	json.Unmarshal([]byte(content), &pub_config)
	LoadDB()

	port := fmt.Sprintf(":%d", pub_config.Port)
	fmt.Println("Listen to port ", port)
	mux := routes.New()
	//TODO: list repos; add repos; enable/disable repos
	mux.Get("/case", ListCases)
	mux.Post("/case", RefreshCases)
	mux.Get("/case/:ID", GetCase)
	mux.Get("/case/:ID/report", GetCaseReport)
	http.Handle("/", mux)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
