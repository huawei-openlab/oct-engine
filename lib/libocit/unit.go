//NOTE: this file is used for the 'Schedular'
//TODO: all 'sync' mode now
package libocit

import (
	"encoding/json"
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

type TUClass string

const (
	TUOS        TUClass = "os"
	TUContainer         = "container"
)

type Resource struct {
	Class        TUClass
	ID           string
	Distribution string
	Version      string
	Arch         string
	CPU          int64
	Memory       int64
	//Includiing port
	URL string
}

type TestUnit struct {
	Resource
	//Suggest to name the unit, easier to write/maintain, must be different
	Name string

	//deploy files: script/data
	Commands TestCommand
	//FIXME: I don't want to use Children..
	//	Children []TestUnit

	Status TestStatus

	//the id of the unit
	id string
	//the id of test task
	testID string
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

type testunit TestUnit

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

func (t *TestUnit) SetTestID(id string) {
	if id != t.testID {
		t.testID = id
	}
}

func (t *TestUnit) GetTestID() string {
	return t.testID
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

func (t *TestUnit) ChangeStatus(succ bool) bool {
	switch t.Status {
	case TestStatusInit:
		t.Status = TestStatusAllocating
	case TestStatusAllocating:
		if succ {
			t.Status = TestStatusAllocated

		} else {
			t.Status = TestStatusAllocateFailed
		}
	case TestStatusAllocated:
		t.Status = TestStatusDeploying
	case TestStatusDeploying:
		if succ {
			t.Status = TestStatusDeployed
		} else {
			t.Status = TestStatusDeployFailed
		}
	case TestStatusDeployed:
		t.Status = TestStatusRunning
	case TestStatusRunning:
		if succ {
			t.Status = TestStatusRun
		} else {
			t.Status = TestStatusRunFailed
		}
	case TestStatusRun:
		t.Status = TestStatusCollecting
	case TestStatusCollecting:
		if succ {
			t.Status = TestStatusCollected
		} else {
			t.Status = TestStatusCollectFailed
		}
	case TestStatusCollected:
		t.Status = TestStatusDestroying
	case TestStatusDestroying:
		if succ {
			t.Status = TestStatusFinish
		} else {
			t.Status = TestStatusDestroyFailed
		}
	default:
		return false

	}
	return true
}

func (t *TestUnit) SetStatus(s TestStatus) {
	if s != t.Status {
		t.Status = s
	}
}

func (t *TestUnit) GetStatus() TestStatus {
	return t.Status
}
