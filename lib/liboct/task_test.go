package liboct

import (
	//	"fmt"
	"testing"
)

func DemoTask() (task TestTask) {
	task.ID = "abd"
	task.PostURL = "0.0.0.0"
	task.BundleURL = "/dev/nul"
	task.Status = TestStatusInit

	return task
}

func TestTaskString(t *testing.T) {
	task := DemoTask()
	val := task.String()

	nTask, _ := TaskFromString(val)
	if nTask.ID == task.ID && task.PostURL == nTask.PostURL && task.BundleURL == nTask.BundleURL && task.Status == nTask.Status {
		t.Log("Task Unmarshal successful")
	} else {
		t.Error("Task Unmarshal failed")
	}

	nVal := nTask.String()
	if val == nVal {
		t.Log("Task string successful")
	} else {
		t.Log("Task string failed")
	}
}

func TestTaskLoop(t *testing.T) {
	task := DemoTask()
	needContinue := task.Loop()
	if needContinue != true {
		t.Log("Task Loop break successful")
	} else {
		t.Error("Task Loop break failed")
	}

	task.Status = TestStatusAllocated
	ok := task.Deploy()
	if ok != true {
		t.Log("Task Deploy failed successful")
	} else {
		t.Error("Task Deploy failed failed")
	}

	task.Status = TestStatusDeployed
	ok = task.Run()
	if ok != true {
		t.Log("Task Run failed successful")
	} else {
		t.Error("Task Run failed failed")
	}

	task.Status = TestStatusRun
	ok = task.Collect()
	if ok != true {
		t.Log("Task Collect failed successful")
	} else {
		t.Error("Task Collect failed failed")
	}

	task.Status = TestStatusCollected
	ok = task.Destroy()
	if ok != true {
		t.Log("Task Destroy failed successful")
	} else {
		t.Error("Task Destroy failed failed")
	}
}
