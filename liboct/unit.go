//NOTE: this file is used for the 'Schedular'
//TODO: all 'sync' mode now
package liboct

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"

	"github.com/Sirupsen/logrus"
)

type TestStatus string

//Warning: this is not the test case status,  this is the test status, which is runtime
const (
	TestStatusInit           TestStatus = "init"
	TestStatusAllocating                = "allocating"
	TestStatusAllocated                 = "allocated"
	TestStatusAllocateFailed            = "allocate failed"
	TestStatusDeploying                 = "deploying"
	TestStatusDeployed                  = "deployed"
	TestStatusDeployFailed              = "deploy failed"
	TestStatusRunning                   = "running"
	TestStatusRun                       = "run"
	TestStatusRunFailed                 = "run failed"
	TestStatusCollecting                = "collecting"
	TestStatusCollected                 = "collect"
	TestStatusCollectFailed             = "collect failed"
	TestStatusDestroying                = "destroying"
	TestStatusFinish                    = "finish"
	TestStatusDestroyFailed             = "destroy failed"
)

type TestAction string

const (
	TestActionAction  TestAction = "action"
	TestActionID                 = "id"
	TestActionApply              = "apply"
	TestActionDeploy             = "deploy"
	TestActionRun                = "run"
	TestActionCollect            = "collect"
	TestActionDestroy            = "destroy"
)

func TestActionFromString(val string) (TestAction, error) {
	logrus.Debugf("test action get ", val)
	switch val {
	case "apply":
		return TestActionApply, nil
	case "deploy":
		return TestActionDeploy, nil
	case "run":
		return TestActionRun, nil
	case "collect":
		return TestActionCollect, nil
	case "destroy":
		return TestActionDestroy, nil
	}
	return TestActionAction, errors.New(fmt.Sprintf("Invalid action %s.", val))
}

type TestUnit struct {
	ResourceCommon
	//Suggest to name the unit, easier to write/maintain, must be different
	Name string

	//deploy files: script/data
	Commands TestCommand
	//FIXME: I don't want to use Children..
	//	Children []TestUnit
	ReportURL string

	Status TestStatus

	//the id of the unit: use less for now, FIXME
	id string
	//the id of the scheduler
	SchedulerID string
	//runtime ID, used to keep track of the relevant hostTest/container
	ResourceID string
	//TODO: use the test bundle URL, but should put files into a smaller piece
	BundleURL string
}

type TestCommand struct {
	Deploy string
	Run    string
}

//Used for tranfer between scheduler and octd/containerpool
type TestActionCommand struct {
	Action  string
	Command string
	//Used for container
	ResName string
}

func (t TestActionCommand) String() string {

	val, _ := json.Marshal(t)
	return string(val)
}

func ActionCommandFromString(val string) (t TestActionCommand, err error) {
	err = json.Unmarshal([]byte(val), &t)
	return t, err
}

type testunit TestUnit

func (t TestUnit) String() string {

	val, _ := json.Marshal(t)
	return string(val)
}

func UnitFromString(val string) (t TestUnit, err error) {
	err = json.Unmarshal([]byte(val), &t)
	return t, err
}

func (t *TestUnit) UnmarshalJSON(data []byte) error {
	var v testunit
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	if len(v.Status) == 0 {
		v.Status = TestStatusInit
	}
	*t = TestUnit(v)
	return nil

}

func (t *TestUnit) SetID(id string) {
	if id != t.id {
		t.id = id
	}
}

func (t *TestUnit) GetID() string {
	return t.id
}

func (t *TestUnit) SetSchedulerID(id string) {
	if id != t.SchedulerID {
		t.SchedulerID = id
	}
}

func (t *TestUnit) GetSchedulerID() string {
	return t.SchedulerID
}

func (t *TestUnit) SetResourceID(id string) {
	if id != t.ResourceID {
		t.ResourceID = id
	}
}

func (t *TestUnit) GetResourceID() string {
	return t.ResourceID
}

func (t *TestUnit) SetBundleURL(url string) {
	if url != t.BundleURL {
		t.BundleURL = url
	}
}

func (t *TestUnit) GetBundleURL() string {
	return t.BundleURL
}

func (t *TestUnit) SetStatus(s TestStatus) {
	if s != t.Status {
		t.Status = s
	}
}

func (t *TestUnit) GetStatus() TestStatus {
	return t.Status
}

func (t *TestUnit) Apply() error {
	if t.Status != TestStatusInit && t.Status != TestStatusAllocateFailed {
		logrus.Debugf("Cannot apply the test resource when the current status is :%s.", t.Status)
		return errors.New(fmt.Sprintf("Cannot apply the test resource when the current status is :%s.", t.Status))
	}
	t.Status = TestStatusAllocating

	db := GetDefaultDB()
	ids := db.Lookup(DBResource, DBQuery{})
	var id string
	for index := 0; index < len(ids); index++ {
		resInterface, _ := db.Get(DBResource, ids[index])
		res, _ := ResourceFromString(resInterface.String())
		if res.IsQualify(t.ResourceCommon) {
			id = ids[index]
			res.Allocate(*t)
			break
		}
	}
	if len(id) > 0 {
		t.ResourceID = id
		t.Status = TestStatusAllocated
		return nil
	} else {
		t.Status = TestStatusAllocateFailed
	}
	return errors.New("Cannot apply the resource")
}

func (t *TestUnit) Deploy() error {
	db := GetDefaultDB()
	resInterface, err := db.Get(DBResource, t.ResourceID)
	if err != nil {
		return err
	}
	res, _ := ResourceFromString(resInterface.String())
	deployURL := fmt.Sprintf("%s/task", res.URL)
	params := make(map[string]string)
	params["id"] = t.SchedulerID
	logrus.Debugf("Test Unit deploy ", deployURL, t.BundleURL, t.SchedulerID)

	tarURL := fmt.Sprintf("%s.tar.gz", t.BundleURL)
	_, err = os.Stat(tarURL)
	if err != nil {
		files := GetDirFiles(t.BundleURL, "")
		tarURL = TarFileList(files, t.BundleURL, "")
	}
	ret := SendFile(deployURL, tarURL, params)
	logrus.Debugf("Deploy result ", ret)
	if ret.Status == RetStatusOK {
		if err := t.command(TestActionDeploy); err == nil {
			t.Status = TestStatusDeployed
			return nil
		}
	}
	t.Status = TestStatusDeployFailed
	return errors.New(ret.Message)
}

func (t *TestUnit) Run() error {
	logrus.Debugf("TestUnit Run ", t)
	if t.Status != TestStatusDeployed && t.Status != TestStatusRunFailed {
		return errors.New(fmt.Sprintf("Cannot run the test when the current status is :%s.", t.Status))
	}
	t.Status = TestStatusRunning
	err := t.command(TestActionRun)
	if err == nil {
		t.Status = TestStatusRun
	} else {
		t.Status = TestStatusRunFailed
	}
	return err
}

func (t *TestUnit) Collect() error {
	if t.Status != TestStatusRun && t.Status != TestStatusCollectFailed && t.Status != TestStatusCollected {
		return errors.New(fmt.Sprintf("Cannot collect the report when the current status is :%s.", t.Status))
	}
	db := GetDefaultDB()
	resInterface, err := db.Get(DBResource, t.ResourceID)
	if err != nil {
		return err
	}
	res, _ := ResourceFromString(resInterface.String())
	t.Status = TestStatusCollecting
	reportURL := fmt.Sprintf("%s/task/%s/report?File=", res.URL, t.SchedulerID, t.ReportURL)

	resp, err := http.Get(reportURL)
	if err != nil {
		t.Status = TestStatusCollectFailed
		return err
	}
	defer resp.Body.Close()
	resp_body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Status = TestStatusCollectFailed
		return err
	}
	filename := path.Join(SchedulerCacheDir, t.SchedulerID, t.ReportURL)
	ioutil.WriteFile(filename, resp_body, os.ModePerm)

	t.Status = TestStatusCollected
	return nil
}

func (t *TestUnit) Destroy() error {
	t.Status = TestStatusDestroying
	err := t.command(TestActionDestroy)
	//TODO After destroy, should release the resource!
	if err == nil {
		t.Status = TestStatusFinish
	} else {
		t.Status = TestStatusDestroyFailed
	}
	return err
}

func (t *TestUnit) command(action TestAction) error {
	logrus.Debugf("Test Unit command ", action)
	db := GetDefaultDB()
	resInterface, err := db.Get(DBResource, t.ResourceID)
	if err != nil {
		return err
	}

	var cmd TestActionCommand
	cmd.ResName = t.ResName
	cmd.Action = string(action)
	switch action {
	case TestActionDeploy:
		cmd.Command = t.Commands.Deploy
	case TestActionRun:
		cmd.Command = t.Commands.Run
	default:
		return errors.New(fmt.Sprintf("The action '%s' is not support.", action))
	}

	res, _ := ResourceFromString(resInterface.String())
	deployURL := fmt.Sprintf("%s/task/%s", res.URL, t.SchedulerID)

	ret := SendCommand(deployURL, []byte(cmd.String()))
	logrus.Debugf("Send Command ", deployURL, cmd.String())
	if ret.Status == RetStatusOK {
		return nil
	}
	return errors.New(fmt.Sprintf("Failed to send the '%s' : %s.", action, ret.Message))
}
