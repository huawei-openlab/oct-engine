//NOTE: this file is used for the 'Schedular'
//TODO: all 'sync' mode now
package libocit

import (
	"encoding/json"
	"errors"
	"os"
	"path"
)

const (
	TestCaseConfigFile = "case.json"
	TestCaseSourceDir  = "source"
	TestCaseCollectDir = "collect"
	TestCaseReportFile = "report.md"
)

/*
The case should be like this:
       caseA
           |___ case.json
	   |___ source/
	   |___ report.md
*/

type TestCase struct {
	//set in runtime implementation
	ID          string
	Name        string
	Summary     string
	Version     string
	License     string
	Group       string
	Owner       string
	Description string
	Units       []TestUnit
	//This is not necessary, but since the developer names the bundle,
	//just don't want to miss it
	BundleName string
	//donnot expose to the public
	bundleURL string
	repoID    string
}

//In this part, all the Unit (with Children) should get the relevant id
func (tc *TestCase) Apply() (msgs []string, succ bool) {
	succ = true
	for index := 0; index < len(tc.Units); index++ {
		if m, ok := tc.Units[index].ApplyResource(); !ok {
			msgs = append(msgs, m...)
			succ = false
			break
		}
	}
	return msgs, succ
}

func (tc *TestCase) Deploy(url string) (msgs []string, succ bool) {
	if !tc.IsStatus(TestStatusAllocated) {
		msgs = append(msgs, "The test case is not well allocated.")
		return msgs, false
	}
	succ = true
	for index := 0; index < len(tc.Units); index++ {
		if m, ok := tc.Units[index].Deploy(url); !ok {
			msgs = append(msgs, m...)
			succ = false
			break
		}
	}
	return msgs, succ
}

func (tc *TestCase) Run() (msgs []string, succ bool) {
	if !tc.IsStatus(TestStatusDeployed) {
		msgs = append(msgs, "The test case is not well deployed.")
		return msgs, false
	}
	succ = true
	for index := 0; index < len(tc.Units); index++ {
		if m, ok := tc.Units[index].Run(); !ok {
			msgs = append(msgs, m...)
			succ = false
			break
		}
	}
	return msgs, succ
}

func (tc *TestCase) Collect() (msgs []string, succ bool) {
	if !tc.IsStatus(TestStatusRun) {
		msgs = append(msgs, "The test case is not well ran.")
		return msgs, false
	}
	succ = true
	for index := 0; index < len(tc.Units); index++ {
		if m, ok := tc.Units[index].Collect(); !ok {
			msgs = append(msgs, m...)
			succ = false
			break
		}
	}
	return msgs, succ
}

func (tc *TestCase) Free() (msgs []string, succ bool) {
	for index := 0; index < len(tc.Units); index++ {
		if m, ok := tc.Units[index].ReleaseResource(); !ok {
			msgs = append(msgs, m...)
			succ = false
		}
	}
	return msgs, succ
}

func (tc *TestCase) IsValid() bool {
	//TODO
	return true
}

func (tc *TestCase) IsStatus(status TestStatus) (succ bool) {
	for index := 0; index < len(tc.Units); index++ {
		if tc.Units[index].Status() != status {
			return false
		}
	}
	return true
}

func (tc *TestCase) SetID(id string) {
	if tc.ID != id {
		tc.ID = id
	}
}

func (tc *TestCase) GetID() string {
	return tc.ID
}

func (tc *TestCase) SetRepoID(id string) {
	if tc.repoID != id {
		tc.repoID = id
	}
}

func (tc *TestCase) GetRepoID() string {
	return tc.repoID
}

func (tc *TestCase) MatchStatus(status string) bool {
	if len(status) == 0 {
		return true
	}
	hasReport, caseUpdated, err := tc.GetReportStatus()
	if err != nil {
		return false
	}
	if status == "hasReport" && hasReport {
		return true
	}
	if status == "caseUpdated" && caseUpdated {
		return true
	}
	return false
}

func CaseFromBundle(bundleURL string) (tc TestCase, err error) {
	configURL := path.Join(bundleURL, TestCaseConfigFile)
	cf, err := os.Open(configURL)
	if err != nil {
		return tc, err
	}
	defer cf.Close()

	if err = json.NewDecoder(cf).Decode(&tc); err != nil {
		return tc, err
	}
	tc.bundleURL = bundleURL
	return tc, nil
}

func (tc *TestCase) GetBundleURL() string {
	return tc.bundleURL
}

func (tc *TestCase) SetBundleName(bundle string) {
	if bundle != tc.BundleName {
		tc.BundleName = bundle
	}
}

func (tc *TestCase) GetBundleName() string {
	return tc.BundleName
}

func (tc *TestCase) GetBundleContent() string {
	files := GetDirFiles(tc.bundleURL, "")
	tmpTarURL := TarFileList(files, tc.bundleURL, "")

	return ReadFile(tmpTarURL)
}

func (tc *TestCase) GetReportContent() string {
	reportURL := path.Join(tc.bundleURL, TestCaseReportFile)
	return ReadFile(reportURL)
}

func (tc *TestCase) GetReportStatus() (hasReport bool, caseUpdated bool, err error) {
	if len(tc.bundleURL) == 0 {
		return false, true, errors.New("Please use a testcase with bundle.")
	}
	cfi, err := os.Stat(path.Join(tc.bundleURL, TestCaseConfigFile))
	if err != nil {
		return false, true, errors.New("Critial issue, invalid case bundle")
	}

	rfi, err := os.Stat(path.Join(tc.bundleURL, TestCaseReportFile))
	if err != nil {
		hasReport = false
		return hasReport, true, nil
	} else {
		hasReport = true
	}

	rtimestamp := rfi.ModTime().Unix()
	ctimestamp := cfi.ModTime().Unix()

	if ctimestamp > rtimestamp {
		caseUpdated = true
		return hasReport, caseUpdated, nil
	} else {
		caseUpdated = false
	}

	sfi, err := os.Stat(path.Join(tc.bundleURL, TestCaseSourceDir))
	if err != nil {
		return hasReport, caseUpdated, nil
	}

	stimestamp := sfi.ModTime().Unix()
	if stimestamp > rtimestamp {
		caseUpdated = true
	} else {
		caseUpdated = false
	}
	return hasReport, caseUpdated, nil
}

//TODO:
func GetTar() {
}
