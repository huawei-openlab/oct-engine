package liboct

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/Sirupsen/logrus"
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

	//Used in octd, keep track of the unit name
	Name string
}

func (task TestTask) String() string {
	val, _ := json.Marshal(task)
	return string(val)
}

func TaskFromString(val string) (task TestTask, err error) {
	err = json.Unmarshal([]byte(val), &task)
	return task, err
}

func TestTaskNew(postURL string, bundleURL string, prio int) (task TestTask, err error) {
	if len(postURL) == 0 || len(bundleURL) == 0 {
		return task, errors.New("The post url or bundle url cannot be nil")
	}
	task.PostURL = postURL
	task.BundleURL = bundleURL
	task.Status = TestStatusInit
	task.Priority = prio
	return task, nil
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

func (task *TestTask) Apply() error {
	if task.Status != TestStatusInit && task.Status != TestStatusAllocateFailed {
		return errors.New(fmt.Sprintf("Cannot apply the resource in the status %s.", task.Status))
	}
	params := make(map[string]string)
	params[SchedulerPriority] = strconv.Itoa(task.Priority)

	postURL := fmt.Sprintf("%s/task", task.PostURL)

	logrus.Debugf("apply task: %v %v", postURL, task.BundleURL)

	ret := SendFile(postURL, task.BundleURL, params)
	if ret.Status == RetStatusOK {
		task.SetSchedulerID(ret.Message)
		task.PostURL = fmt.Sprintf("%s/task/%s", task.PostURL, task.GetSchedulerID())
		task.Status = TestStatusAllocated
		logrus.Debugf("apply return : %v %v", task, ret.Message)
		return nil
	} else {
		logrus.Warnf("apply return : %v %v ", task, ret.Message)
		task.Status = TestStatusAllocateFailed
	}
	return errors.New("Failed to apply.")
}

//Donnot need to send files now, since it will be done by the Apply function
func (task *TestTask) Deploy() error {
	if task.Status != TestStatusAllocated && task.Status != TestStatusDeployFailed {
		return errors.New(fmt.Sprintf("Cannot deploy the test in the status %s.", task.Status))
	}
	task.Status = TestStatusDeploying
	ret := SendCommand(task.PostURL, []byte(TestActionDeploy))
	logrus.Debugf("Deploy : %v %v", task.PostURL, ret)
	if ret.Status == RetStatusOK {
		task.Status = TestStatusDeployed
		return nil
	} else {
		task.Status = TestStatusDeployFailed
	}
	return errors.New("Failed to deploy.")
}

func (task *TestTask) Run() error {
	if task.Status != TestStatusDeployed && task.Status != TestStatusRunFailed {
		return errors.New(fmt.Sprintf("Cannot run the test in the status %s.", task.Status))
	}
	task.Status = TestStatusRunning
	ret := SendCommand(task.PostURL, []byte(TestActionRun))
	logrus.Debugf("Run %v", ret)
	if ret.Status == RetStatusOK {
		task.Status = TestStatusRun
		return nil
	} else {
		task.Status = TestStatusRunFailed
	}
	return errors.New("Failed to run.")
}

func (task *TestTask) Collect() error {
	if task.Status != TestStatusRun && task.Status != TestStatusCollectFailed {
		return errors.New(fmt.Sprintf("Cannot collect the report in the status %s.", task.Status))
	}
	task.Status = TestStatusCollecting
	ret := SendCommand(task.PostURL, []byte(TestActionCollect))
	logrus.Debugf("Collect %v", ret)
	if ret.Status == RetStatusOK {
		task.Status = TestStatusCollected
		return nil
	} else {
		task.Status = TestStatusCollectFailed
	}
	return errors.New("Failed to collect.")
}

func (task *TestTask) Destroy() error {
	task.Status = TestStatusDestroying
	ret := SendCommand(task.PostURL, []byte(TestActionDestroy))
	if ret.Status == RetStatusOK {
		task.Status = TestStatusFinish
		return nil
	} else {
		task.Status = TestStatusDestroyFailed
	}
	return errors.New("Failed to Destroy.")
}

func (task *TestTask) Command(action TestAction) (err error) {
	switch action {
	case TestActionApply:
		err = task.Apply()
	case TestActionDeploy:
		err = task.Deploy()
	case TestActionRun:
		err = task.Run()
	case TestActionCollect:
		err = task.Collect()
	case TestActionDestroy:
		err = task.Destroy()
	default:
		return errors.New(fmt.Sprintf("The action %s is not supported.", action))
	}
	logrus.Debugf("Command %v Update %v", action, task)
	db := GetDefaultDB()
	db.Update(DBTask, task.ID, task)
	return nil
}

func (task *TestTask) Loop() (needContinue error) {
	switch task.Status {
	case TestStatusInit:
		needContinue = task.Apply()
	case TestStatusAllocated:
		needContinue = task.Deploy()
	case TestStatusDeployed:
		needContinue = task.Run()
	case TestStatusRun:
		needContinue = task.Collect()
	case TestStatusCollected:
		needContinue = task.Destroy()
	default:
		needContinue = nil
	}
	db := GetDefaultDB()
	db.Update(DBTask, task.ID, task)
	return needContinue
}
