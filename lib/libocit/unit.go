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

type TestUnit struct {
	//Suggest to name the unit, easier to write/maintain
	Name string
	//os or containers
	Class        TUClass
	Distribution string
	Version      string

	//deploy files: script/data
	Commands TestCommand
	//FIXME: I don't want to use Children..
	Children []TestUnit

	//runtime ID, used to keep track of the relevant hostTest/container
	id     string
	status TestStatus
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
	v.status = TestStatusInit
	*t = TestUnit(v)
	return nil

}

func (tu *TestUnit) ApplyResource() (msgs []string, succ bool) {
	succ = true
	if tu.Class == TUOS {
		//TODO : get from testserver
		// tu.id =
	} else if tu.Class == TUContainer {
		//TODO : get from container pool
	}
	for index := 0; index < len(tu.Children); index++ {
		if m, ok := tu.Children[index].ApplyResource(); !ok {
			msgs = append(msgs, m...)
			succ = false
			break
		}
	}
	if len(tu.id) != 0 {
		tu.status = TestStatusAllocated
	}
	return msgs, succ
}

func (tu *TestUnit) ReleaseResource() (msgs []string, valid bool) {
	valid = true
	if len(tu.id) == 0 {
		return msgs, true
	}
	if tu.Class == TUOS {
		//TODO : get from testserver
		// tu.id =
	} else if tu.Class == TUContainer {
		//TODO : get from container pool
	}
	return msgs, valid
}

func (tu *TestUnit) Deploy(url string) (msgs []string, succ bool) {
	for index := 0; index < len(tu.Children); index++ {
		if m, ok := tu.Children[index].ApplyResource(); !ok {
			msgs = append(msgs, m...)
			succ = false
			break
		}
	}
	return msgs, succ
}

func (tu *TestUnit) Run() (msgs []string, succ bool) {
	for index := 0; index < len(tu.Children); index++ {
		if m, ok := tu.Children[index].Run(); !ok {
			msgs = append(msgs, m...)
			succ = false
			break
		}
	}
	return msgs, succ
}

func (tu *TestUnit) Collect() (msgs []string, succ bool) {
	for index := 0; index < len(tu.Children); index++ {
		if m, ok := tu.Children[index].Collect(); !ok {
			msgs = append(msgs, m...)
			succ = false
			break
		}
	}
	return msgs, succ
}

func (tu *TestUnit) Status() TestStatus {
	return tu.status
}
