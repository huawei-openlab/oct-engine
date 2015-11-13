//NOTE: this file is used for the 'Schedular'
//TODO: all 'sync' mode now
package libocit

import (
	"encoding/json"
	"fmt"
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
	switch val {
	case "apploy":
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
	schedulerID string
	//runtime ID, used to keep track of the relevant hostTest/container
	resourceID string
	//TODO: use the test bundle URL, but should put files into a smaller piece
	bundleURL string
}

type TestCommand struct {
	Deploy  string
	Run     string
	Collect string
}

//Used for tranfer between scheduler and ocitd/containerpool
type TestActionCommand struct {
	Action  TestAction
	Command string
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
	if id != t.schedulerID {
		t.schedulerID = id
	}
}

func (t *TestUnit) GetSchedulerID() string {
	return t.schedulerID
}

func (t *TestUnit) SetResourceID(id string) {
	if id != t.resourceID {
		t.resourceID = id
	}
}

func (t *TestUnit) GetResourceID() string {
	return t.resourceID
}

func (t *TestUnit) SetBundleURL(url string) {
	if url != t.bundleURL {
		t.bundleURL = url
	}
}

func (t *TestUnit) GetBundleURL() string {
	return t.bundleURL
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
	if t.Status != TestStatusInit {
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
			break
		}
	}
	if len(id) > 0 {
		t.resourceID = id
		t.Status = TestStatusAllocated
		ok = true
	} else {
		t.Status = TestStatusAllocateFailed
		ok = false
	}
	return ok
}

func (t *TestUnit) Deploy() bool {
	resInterface, ok := DBGet(DBResource, t.resourceID)
	if !ok {
		return false
	}
	res, _ := ResourceFromString(resInterface.String())
	deployURL := fmt.Sprintf("%s/task", res.URL)
	params := make(map[string]string)
	params["id"] = t.schedulerID
	ret := SendFile(deployURL, t.bundleURL, params)
	if ret.Status == RetStatusOK {
		return true
	}
	return false
}

func (t *TestUnit) Run() bool {
	if t.Status != TestStatusDeployed {
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
	if t.Status != TestStatusRun {
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
	if ok {
		t.Status = TestStatusFinish
	} else {
		t.Status = TestStatusDestroyFailed
	}
	return ok
}

func (t *TestUnit) command(action TestAction) bool {
	resInterface, ok := DBGet(DBResource, t.resourceID)
	if !ok {
		return false
	}

	var cmd TestActionCommand
	cmd.Action = action
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
	deployURL := fmt.Sprintf("%s/task/%s", res.URL, t.schedulerID)

	ret := SendCommand(deployURL, []byte(action))
	if ret.Status == RetStatusOK {
		return true
	}
	return false
}
