//NOTE: this file is used for the 'Schedular'
//TODO: all 'sync' mode now
package liboct

import (
	"encoding/json"
	"fmt"
	"os"
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

func TestActionFromString(val string) (TestAction, bool) {
	fmt.Println("test action get ", val)
	switch val {
	case "apply":
		return TestActionApply, true
	case "deploy":
		return TestActionDeploy, true
	case "run":
		return TestActionRun, true
	case "collect":
		return TestActionCollect, true
	case "destroy":
		return TestActionDestroy, true
	}
	return TestActionAction, false
}

type TestUnit struct {
	ResourceCommon
	//Suggest to name the unit, easier to write/maintain, must be different
	Name string

	//deploy files: script/data
	Commands TestCommand
	//FIXME: I don't want to use Children..
	//	Children []TestUnit

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
	Deploy  string
	Run     string
	Collect string
}

//Used for tranfer between scheduler and octd/containerpool
type TestActionCommand struct {
	Action  string
	Command string
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
	v.Status = TestStatusInit
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

func (t *TestUnit) Apply() (ok bool) {
	ok = true
	if t.Status != TestStatusInit && t.Status != TestStatusAllocateFailed {
		return false
	}
	t.Status = TestStatusAllocating

	ids := DBLookup(DBResource, DBQuery{})
	var id string
	for index := 0; index < len(ids); index++ {
		resInterface, _ := DBGet(DBResource, ids[index])
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
		ok = true
	} else {
		t.Status = TestStatusAllocateFailed
		ok = false
	}
	fmt.Println("OS Apply", t, "----", ok)
	return ok
}

func (t *TestUnit) Deploy() bool {
	resInterface, ok := DBGet(DBResource, t.ResourceID)
	if !ok {
		fmt.Println("Cannot find the resource ", t.ResourceID)
		return false
	}
	res, _ := ResourceFromString(resInterface.String())
	deployURL := fmt.Sprintf("%s/task", res.URL)
	params := make(map[string]string)
	params["id"] = t.SchedulerID
	fmt.Println("Test Unit deploy ", deployURL, t.BundleURL, t.SchedulerID)

	tarURL := fmt.Sprintf("%s.tar.gz", t.BundleURL)
	_, err := os.Stat(tarURL)
	if err != nil {
		files := GetDirFiles(t.BundleURL, "")
		tarURL = TarFileList(files, t.BundleURL, "")
	}
	ret := SendFile(deployURL, tarURL, params)
	fmt.Println("Deploy result ", ret)
	if ret.Status == RetStatusOK {
		if t.command(TestActionDeploy) {
			t.Status = TestStatusDeployed
			return true
		}
	}
	t.Status = TestStatusDeployFailed
	return false
}

func (t *TestUnit) Run() bool {
	if t.Status != TestStatusDeployed && t.Status != TestStatusRunFailed {
		return false
	}
	t.Status = TestStatusRunning
	ok := t.command(TestActionRun)
	if ok {
		t.Status = TestStatusRun
	} else {
		t.Status = TestStatusRunFailed
	}
	return ok
}

func (t *TestUnit) Collect() bool {
	if t.Status != TestStatusRun && t.Status != TestStatusCollectFailed {
		return false
	}
	t.Status = TestStatusCollecting
	ok := t.command(TestActionCollect)
	if ok {
		t.Status = TestStatusCollected
	} else {
		t.Status = TestStatusCollectFailed
	}
	return ok
}

func (t *TestUnit) Destroy() bool {
	t.Status = TestStatusDestroying
	ok := t.command(TestActionDestroy)
	//TODO After destroy, should release the resource!
	if ok {
		t.Status = TestStatusFinish
	} else {
		t.Status = TestStatusDestroyFailed
	}
	return ok
}

func (t *TestUnit) command(action TestAction) bool {
	resInterface, ok := DBGet(DBResource, t.ResourceID)
	if !ok {
		return false
	}

	var cmd TestActionCommand
	cmd.Action = string(action)
	switch action {
	case TestActionDeploy:
		cmd.Command = t.Commands.Deploy
	case TestActionRun:
		cmd.Command = t.Commands.Run
	case TestActionCollect:
		//TODO: the user should just define the log url, and the oct compose a new command
		cmd.Command = t.Commands.Collect
	default:
		return false
	}

	res, _ := ResourceFromString(resInterface.String())
	deployURL := fmt.Sprintf("%s/task/%s", res.URL, t.SchedulerID)

	ret := SendCommand(deployURL, []byte(cmd.String()))
	fmt.Println("Send Command ", deployURL, cmd.String())
	if ret.Status == RetStatusOK {
		return true
	}
	return false
}
