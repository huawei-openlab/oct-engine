package libocit

import (
	"fmt"
)

type TestTask struct {
	ID      string
	PostURL string
	//a tar.gz file
	BundleURL string
	Status    TestStatus
	Priority  int
}

func TestTaskNew(postURL string, bundleURL string, prio int) (task TestTask) {
	task.PostURL = postURL
	task.BundleURL = bundleURL
	task.Status = TestStatusInit
	task.Priority = prio

	return task
}

func (task *TestTask) Init() (id string, ret HttpRet) {
	if task.Status == TestStatusInit {
		task.Status = TestStatusAllocating
		ret = SendFile(task.PostURL, task.BundleURL, nil)
		if ret.Status == RetStatusOK {
			//FIXME: use message to mean id is not a good idea
			task.ID = ret.Message
			task.PostURL = fmt.Sprintf("%s/%s", task.PostURL, task.ID)
			task.Status = TestStatusAllocated
		} else {
			task.Status = TestStatusAllocateFailed
		}
	}
	return task.ID, ret
}

func (task *TestTask) Run() (needContinue bool) {
	needContinue = false
	switch task.Status {
	case TestStatusAllocated:
		task.Status = TestStatusDeploying
		ret := SendCommand(task.PostURL, []byte(TestActionDeploy))
		if ret.Status == RetStatusOK {
			task.Status = TestStatusDeployed
			needContinue = true
		} else {
			task.Status = TestStatusDeployFailed
		}
	case TestStatusDeployed:
		task.Status = TestStatusRunning
		ret := SendCommand(task.PostURL, []byte(TestActionRun))
		if ret.Status == RetStatusOK {
			task.Status = TestStatusRun
			needContinue = true
		} else {
			task.Status = TestStatusRunFailed
		}
	case TestStatusRun:
		task.Status = TestStatusCollecting
		ret := SendCommand(task.PostURL, []byte(TestActionCollect))
		if ret.Status == RetStatusOK {
			task.Status = TestStatusCollected
			needContinue = true
		} else {
			task.Status = TestStatusCollectFailed
		}
	case TestStatusCollected:
		task.Status = TestStatusDestroying
		ret := SendCommand(task.PostURL, []byte(TestActionDestroy))
		if ret.Status == RetStatusOK {
			task.Status = TestStatusFinish
		} else {
			task.Status = TestStatusDestroyFailed
		}
	}
	return needContinue
}
