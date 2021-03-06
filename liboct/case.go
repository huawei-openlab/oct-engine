//This file is used for the 'Schedular'
package liboct

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/Sirupsen/logrus"
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

type TestCasePub struct {
	//set in runtime implementation
	ID          string
	RepoID      string
	Name        string
	Summary     string
	Version     string
	License     string
	Group       string
	Owner       string
	Description string
	Units       []TestUnit
}

type TestCase struct {
	TestCasePub
	//This is not necessary, but since the developer names the bundle,
	//just don't want to miss it
	BundleName string
	/* the dir */
	BundleURL string
}

func (tc TestCase) String() string {
	val, _ := json.Marshal(tc)
	return string(val)
}

func CaseFromString(val string) (tc TestCase, err error) {
	err = json.Unmarshal([]byte(val), &tc)
	return tc, err
}

func (tc *TestCase) IsValid() error {
	//TODO
	return nil
}

func (tc *TestCase) SetID(id string) {
	if tc.ID != id {
		tc.ID = id
	}
	for index := 0; index < len(tc.Units); index++ {
		unitID := MD5(fmt.Sprintf("%s%d", id, index))
		//Do we really need to setID?
		tc.Units[index].SetID(unitID)
		tc.Units[index].SetSchedulerID(id)
	}
}

func (tc *TestCase) GetID() string {
	return tc.ID
}

func (tc *TestCase) SetRepoID(id string) {
	if tc.RepoID != id {
		tc.RepoID = id
	}
}

func (tc *TestCase) GetRepoID() string {
	return tc.RepoID
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

func CaseFromBundle(BundleURL string) (tc TestCase, err error) {
	configURL := path.Join(BundleURL, TestCaseConfigFile)
	cf, err := os.Open(configURL)
	if err != nil {
		return tc, err
	}
	defer cf.Close()

	if err = json.NewDecoder(cf).Decode(&tc); err != nil {
		return tc, err
	}
	tc.SetBundleURL(BundleURL)
	return tc, nil
}

func (tc *TestCase) SetBundleURL(bundle string) {
	if bundle != tc.BundleURL {
		tc.BundleURL = bundle
		tc.BundleName = path.Base(tc.BundleURL)
		for index := 0; index < len(tc.Units); index++ {
			tc.Units[index].SetBundleURL(bundle)
		}
	}
}

func (tc *TestCase) GetBundleURL() string {
	return tc.BundleURL
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
	tarURL := tc.GetBundleTarURL()

	file, err := os.Open(tarURL)
	defer file.Close()
	if err != nil {
		logrus.Info(err)
		return ""
	}
	buf := bytes.NewBufferString("")
	buf.ReadFrom(file)

	return buf.String()
}

func (tc *TestCase) GetReportContent() string {
	reportURL := path.Join(tc.BundleURL, TestCaseReportFile)

	file, err := os.Open(reportURL)
	defer file.Close()
	if err != nil {
		logrus.Info(err)
		return ""
	}
	buf := bytes.NewBufferString("")
	buf.ReadFrom(file)

	return buf.String()
}

func (tc *TestCase) GetReportStatus() (hasReport bool, caseUpdated bool, err error) {
	if len(tc.BundleURL) == 0 {
		return false, true, errors.New("Please use a testcase with bundle.")
	}
	cfi, err := os.Stat(path.Join(tc.BundleURL, TestCaseConfigFile))
	if err != nil {
		return false, true, errors.New("Critial issue, invalid case bundle")
	}

	rfi, err := os.Stat(path.Join(tc.BundleURL, TestCaseReportFile))
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

	sfi, err := os.Stat(path.Join(tc.BundleURL, TestCaseSourceDir))
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

//If cacheURL is nil, tar the bundle in the same dir
func (tc *TestCase) GetBundleTarURL() string {
	needUpdate := false

	bfi, err := os.Stat(tc.BundleURL)
	if err != nil {
		logrus.Warn(err)
		return ""
	}
	tfi, err := os.Stat(fmt.Sprintf("%s.tar.gz", tc.BundleURL))
	if err != nil {
		needUpdate = true
	} else {
		btimestamp := bfi.ModTime().Unix()
		ttimestamp := tfi.ModTime().Unix()
		if btimestamp > ttimestamp {
			needUpdate = true
		}
	}

	if needUpdate {
		return tc.generateBundleTar()
	}
	return fmt.Sprintf("%s.tar.gz", tc.BundleURL)
}

func (tc *TestCase) generateBundleTar() string {
	files := GetDirFiles(tc.BundleURL, "")
	tarURL, _ := TarFileList(files, tc.BundleURL, "")
	return tarURL
}

//If bundleURL is nil, untar the bundle in the same dir
func CaseFromTar(tarURL string, bundleURL string) (TestCase, error) {
	bundleDir := ""
	if len(bundleURL) == 0 {
		bundleDir = strings.TrimSuffix(tarURL, ".tar.gz")
	} else {
		bundleDir = bundleURL
	}
	UntarFile(tarURL, bundleDir)
	return CaseFromBundle(bundleDir)
}
