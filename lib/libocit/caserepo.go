package libocit

import (
	//	"fmt"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
)

const TestCaseCache string = "/tmp/.testcase_cache/"

type TestCaseRepoPub struct {
	//set in runtime implementation
	ID string
	//Name is the short name of URL, make the repo management easier
	Name       string
	URL        string
	CaseFolder string
	//Type is the method to discover and get the repos. 'git' as a default
	Type string
	//We can disable a repo
	Enable bool
	Groups []string
}

type TestCaseRepo struct {
	TestCaseRepoPub
	//used to get the repo data, default to : /tmp/tcserver_cache/
	CacheDir string
	Cases    []TestCase
	//The id is not public since it should be set in the implementation
	Timestamp int64
}

func (repo TestCaseRepo) String() string {
	val, _ := json.Marshal(repo)
	return string(val)
}

func RepoFromString(val string) (tcr TestCaseRepo, err error) {
	err = json.Unmarshal([]byte(val), &tcr)
	return tcr, err
}

func (repo *TestCaseRepo) IsValid() error {
	//If the Name is empty, we can use MD(URL)
	if len(repo.Name) == 0 {
		return errors.New("Name is empty")
	}
	if len(repo.URL) == 0 {
		return errors.New("URL is empty")
	}
	if len(repo.CaseFolder) == 0 {
		return errors.New("CaseFolder is empty")
	}
	if len(repo.Groups) == 0 {
		return errors.New("Group is empty")
	}
	return nil
}

func (repo *TestCaseRepo) SetCacheDir(CacheDir string) {
	if repo.CacheDir != CacheDir {
		repo.CacheDir = CacheDir
	}
}

func (repo *TestCaseRepo) SetID(id string) {
	if id != repo.ID {
		repo.ID = id
	}
}

func (repo *TestCaseRepo) GetID() string {
	return repo.ID
}

func (repo *TestCaseRepo) Refresh() bool {
	if repo.Enable == false {
		return false
	}
	if len(repo.CacheDir) == 0 {
		repo.CacheDir = TestCaseCache
	}
	if len(repo.Type) == 0 || repo.Type == "git" {
		var cmd string
		repoDir := PreparePath(repo.CacheDir, repo.URL)
		git_check_url := path.Join(repoDir, ".git/config")
		_, err := os.Stat(git_check_url)
		if err != nil {
			//Clean: remove the '/$' if there was one
			cmd = "cd " + path.Dir(path.Clean(repoDir)) + " ; git clone https://" + repo.URL + ".git"
		} else {
			cmd = "cd " + repoDir + " ; git pull"
		}

		c := exec.Command("/bin/sh", "-c", cmd)
		c.Run()

		_, err = os.Stat(path.Join(repo.CacheDir, repo.URL, repo.CaseFolder))
		if err != nil {
			return false
		}
		return repo.loadCases()
	}
	return true
}

func (repo *TestCaseRepo) Modify(newRepo TestCaseRepo) {
	changed := false
	if len(newRepo.Name) > 0 && newRepo.Name != repo.Name {
		repo.Name = newRepo.Name
		changed = true
	}
	if len(newRepo.URL) > 0 && newRepo.Name != repo.Name {
		repo.URL = newRepo.URL
		changed = true
	}
	if len(newRepo.CaseFolder) > 0 && newRepo.CaseFolder != repo.CaseFolder {
		repo.CaseFolder = newRepo.CaseFolder
		changed = true
	}
	if len(newRepo.Type) > 0 && newRepo.Type != newRepo.Type {
		repo.Type = newRepo.Type
		changed = true
	}
	if newRepo.Enable != repo.Enable {
		repo.Enable = newRepo.Enable
		changed = true
	}
	//TODO: better comparing
	if len(newRepo.Groups) > 0 {
		repo.Groups = newRepo.Groups
		changed = true
	}

	if changed {
		repo.Cases = nil
	}
}

//True means updated, false, means no changes
func (repo *TestCaseRepo) loadCases() bool {
	fileinfo, err := os.Stat(path.Join(repo.CacheDir, repo.URL, repo.CaseFolder))
	if err != nil {
		return false
	}
	Timestamp := fileinfo.ModTime().Unix()
	if Timestamp <= repo.Timestamp {
		return false
	} else {
		repo.Timestamp = Timestamp
	}

	repo.Cases = nil
	for index := 0; index < len(repo.Groups); index++ {
		groupDir := path.Join(repo.CacheDir, repo.URL, repo.CaseFolder, repo.Groups[index])
		files, _ := ioutil.ReadDir(groupDir)
		for _, file := range files {
			if !file.IsDir() {
				continue
			}
			if tc, err := CaseFromBundle(path.Join(groupDir, file.Name())); err == nil {
				if tc.IsValid() == nil {
					tc.SetRepoID(repo.ID)
					repo.Cases = append(repo.Cases, tc)
				}
			}
		}
	}
	return true
}

func (repo *TestCaseRepo) GetCases() []TestCase {
	if repo.Enable {
		return repo.Cases
	}
	return nil
}

func (repo *TestCaseRepo) GetCase(groupAndName string) (tc TestCase, err error) {
	caseDir := path.Join(repo.CacheDir, repo.URL, repo.CaseFolder, groupAndName)
	return CaseFromBundle(caseDir)
}

func (repo *TestCaseRepo) Purge() error {
	repoDir := path.Join(repo.CacheDir, repo.URL)
	return os.RemoveAll(repoDir)
}
