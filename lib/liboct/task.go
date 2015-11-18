package liboct

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

	//Return from the scheduler
	SchedulerID string
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
	task.PostURL = postURL
	task.BundleURL = bundleURL
	task.Status = TestStatusInit
	task.Priority = prio
	return task, true
}

func (task *TestTask) SetID(id string) {
	if id != task.ID {
		task.ID = id
	}
}

func (task *TestTask) GetID() string {
	return task.ID
}

func (task *TestTask) SetSchedulerID(id string) {
	if id != task.SchedulerID {
		task.SchedulerID = id
	}
}

func (task *TestTask) GetSchedulerID() string {
	return task.SchedulerID
}

func (task *TestTask) Apply() (ok bool) {
	if task.Status != TestStatusInit && task.Status != TestStatusAllocateFailed {
		return false
	}
	params := make(map[string]string)
	params[SchedulerPriority] = strconv.Itoa(task.Priority)

	postURL := fmt.Sprintf("%s/task", task.PostURL)

	//DEBUG
	fmt.Println("apply task: ", postURL, task.BundleURL)

	ret := SendFile(postURL, task.BundleURL, params)
	if ret.Status == RetStatusOK {
		task.SetSchedulerID(ret.Message)
		task.PostURL = fmt.Sprintf("%s/task/%s", task.PostURL, task.GetSchedulerID())
		task.Status = TestStatusAllocated
		fmt.Println("apply return : ", task, ret.Message)
		ok = true
	} else {
		task.Status = TestStatusAllocateFailed
		ok = false
	}
	return ok
}

//Donnot need to send files now, since it will be done by the Apply function
func (task *TestTask) Deploy() (ok bool) {
	if task.Status != TestStatusAllocated && task.Status != TestStatusDeployFailed {
		return false
	}
	task.Status = TestStatusDeploying
	ret := SendCommand(task.PostURL, []byte(TestActionDeploy))
	fmt.Println("Deploy : ", task.PostURL, ret)
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
	if task.Status != TestStatusDeployed && task.Status != TestStatusRunFailed {
		return false
	}
	task.Status = TestStatusRunning
	ret := SendCommand(task.PostURL, []byte(TestActionRun))
	fmt.Println("Run ", ret)
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
	if task.Status != TestStatusRun && task.Status != TestStatusCollectFailed {
		return false
	}
	task.Status = TestStatusCollecting
	ret := SendCommand(task.PostURL, []byte(TestActionCollect))
	fmt.Println("Collect ", ret)
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

func (task *TestTask) Command(action TestAction) (ok bool) {
	ok = false
	switch action {
	case TestActionApply:
		ok = task.Apply()
	case TestActionDeploy:
		ok = task.Deploy()
	case TestActionRun:
		ok = task.Run()
	case TestActionCollect:
		ok = task.Collect()
	case TestActionDestroy:
		ok = task.Destroy()
	default:
		fmt.Println("The action is not supported")
		ok = false
	}
	fmt.Println("Command ", action, "  Update  ", task)
	DBUpdate(DBTask, task.ID, task)
	return ok
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
	DBUpdate(DBTask, task.ID, task)
	return needContinue
}
