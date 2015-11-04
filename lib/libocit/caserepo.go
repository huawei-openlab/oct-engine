package libocit

import (
	//	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
)

const TestCaseCache string = "/tmp/.testcase_cache/"

type TestCaseRepo struct {
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

	//used to get the repo data, default to : /tmp/tcserver_cache/
	cacheDir string
	cases    []TestCase
	//The id is not public since it should be set in the implementation
	timestamp int64
}

func (repo *TestCaseRepo) IsValid() (msgs []string, valid bool) {
	valid = true
	//If the Name is empty, we can use MD(URL)
	if len(repo.Name) == 0 {
		msgs = append(msgs, "Name is empty")
		valid = false
	}
	if len(repo.URL) == 0 {
		msgs = append(msgs, "URL is empty")
		valid = false
	}
	if len(repo.CaseFolder) == 0 {
		msgs = append(msgs, "CaseFolder is empty")
		valid = false
	}
	if len(repo.Groups) == 0 {
		msgs = append(msgs, "Group is empty")
		valid = false
	}
	return msgs, valid
}

func (repo *TestCaseRepo) SetCacheDir(cacheDir string) {
	if repo.cacheDir != cacheDir {
		repo.cacheDir = cacheDir
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
	if len(repo.cacheDir) == 0 {
		repo.cacheDir = TestCaseCache
	}
	if len(repo.Type) == 0 || repo.Type == "git" {
		var cmd string
		repoDir := PreparePath(repo.cacheDir, repo.URL)
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

		_, err = os.Stat(path.Join(repo.cacheDir, repo.URL, repo.CaseFolder))
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
		repo.cases = nil
	}
}

//True means updated, false, means no changes
func (repo *TestCaseRepo) loadCases() bool {
	fileinfo, err := os.Stat(path.Join(repo.cacheDir, repo.URL, repo.CaseFolder))
	if err != nil {
		return false
	}
	timestamp := fileinfo.ModTime().Unix()
	if timestamp <= repo.timestamp {
		return false
	} else {
		repo.timestamp = timestamp
	}

	repo.cases = nil
	for index := 0; index < len(repo.Groups); index++ {
		groupDir := path.Join(repo.cacheDir, repo.URL, repo.CaseFolder, repo.Groups[index])
		files, _ := ioutil.ReadDir(groupDir)
		for _, file := range files {
			if !file.IsDir() {
				continue
			}
			if tc, err := CaseFromBundle(path.Join(groupDir, file.Name())); err == nil {
				if tc.IsValid() {
					tc.SetRepoID(repo.ID)
					repo.cases = append(repo.cases, tc)
				}
			}
		}
	}
	return true
}

func (repo *TestCaseRepo) GetCases() []TestCase {
	if repo.Enable {
		return repo.cases
	}
	return nil
}

func (repo *TestCaseRepo) GetCase(groupAndName string) (tc TestCase, err error) {
	caseDir := path.Join(repo.cacheDir, repo.URL, repo.CaseFolder, groupAndName)
	return CaseFromBundle(caseDir)
}

func (repo *TestCaseRepo) Purge() error {
	repoDir := path.Join(repo.cacheDir, repo.URL)
	return os.RemoveAll(repoDir)
}
