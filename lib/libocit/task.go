package libocit

import (
	"encoding/json"
	"fmt"
	"strconv"
)

const SchedulerPriority = "Priority"
const SchedulerDefaultPrio = 100

type TestTask struct {
	ID      string
	PostURL string
	//a tar.gz file
	BundleURL string
	Status    TestStatus
	Priority  int
}

func (task TestTask) String() string {
	val, _ := json.Marshal(task)
	return string(val)
}

func TaskFromString(val string) (task TestTask, err error) {
	err = json.Unmarshal([]byte(val), &task)
	return task, err
}

func TestTaskNew(postURL string, bundleURL string, prio int) (task TestTask, ok bool) {
	params := make(map[string]string)
	params[SchedulerPriority] = strconv.Itoa(prio)
	ret := SendFile(postURL, bundleURL, params)
	if ret.Status == RetStatusOK {
		task.ID = ret.Message
		task.PostURL = fmt.Sprintf("%s/%s", postURL, task.ID)
		task.BundleURL = bundleURL
		task.Status = TestStatusInit
		task.Priority = prio
		task.Status = TestStatusAllocated
		ok = true
	} else {
		task.Status = TestStatusAllocateFailed
		ok = false
	}
	return task, ok
}

func (task *TestTask) SetID(id string) {
	if id != task.ID {
		task.ID = id
	}
}

func (task *TestTask) GetID() string {
	return task.ID
}

func (task *TestTask) Deploy() (ok bool) {
	if task.Status != TestStatusAllocated {
		return false
	}
	task.Status = TestStatusDeploying
	ret := SendCommand(task.PostURL, []byte(TestActionDeploy))
	if ret.Status == RetStatusOK {
		task.Status = TestStatusDeployed
		ok = true
	} else {
		task.Status = TestStatusDeployFailed
		ok = false
	}
	return ok
}

func (task *TestTask) Run() (ok bool) {
	if task.Status != TestStatusDeployed {
		return false
	}
	task.Status = TestStatusRunning
	ret := SendCommand(task.PostURL, []byte(TestActionRun))
	if ret.Status == RetStatusOK {
		task.Status = TestStatusRun
		ok = true
	} else {
		task.Status = TestStatusRunFailed
		ok = false
	}
	return ok
}

func (task *TestTask) Collect() (ok bool) {
	if task.Status != TestStatusRun {
		return false
	}
	task.Status = TestStatusCollecting
	ret := SendCommand(task.PostURL, []byte(TestActionCollect))
	if ret.Status == RetStatusOK {
		task.Status = TestStatusCollected
		ok = true
	} else {
		task.Status = TestStatusCollectFailed
		ok = false
	}
	return ok
}

func (task *TestTask) Destroy() (ok bool) {
	task.Status = TestStatusDestroying
	ret := SendCommand(task.PostURL, []byte(TestActionDestroy))
	if ret.Status == RetStatusOK {
		task.Status = TestStatusFinish
		ok = true
	} else {
		task.Status = TestStatusDestroyFailed
		ok = false
	}
	return ok
}

func (task *TestTask) Command(action TestAction) bool {
	switch action {
	case TestActionDeploy:
		return task.Deploy()
	case TestActionRun:
		return task.Run()
	case TestActionCollect:
		return task.Collect()
	case TestActionDestroy:
		return task.Destroy()
	default:
		fmt.Println("The action is not supported")
	}
	return false
}

func (task *TestTask) Loop() (needContinue bool) {
	needContinue = false
	switch task.Status {
	case TestStatusInit:
		task.Status = TestStatusAllocating
		params := make(map[string]string)
		params[TestActionID] = task.ID
		ret := SendFile(task.PostURL, task.BundleURL, params)
		fmt.Println("Run send file : ", task.PostURL, task.BundleURL)
		if ret.Status == RetStatusOK {
			task.ID = ret.Message
			task.PostURL = fmt.Sprintf("%s/%s", task.PostURL, task.ID)
			task.Status = TestStatusAllocated
		} else {
			task.Status = TestStatusAllocateFailed
		}
	case TestStatusAllocated:
		needContinue = task.Deploy()
	case TestStatusDeployed:
		needContinue = task.Run()
	case TestStatusRun:
		needContinue = task.Collect()
	case TestStatusCollected:
		task.Destroy()
		needContinue = false
	}
	return needContinue
}
