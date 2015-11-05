package main

import (
	"../../lib/libocit"
)

type TSUnit interface {
	Apply() string
	//	Deploy(id string, bundleURL string) bool
	Run(id string, action libocit.TestAction) bool
	GetStatus() libocit.TestStatus
}

type TSTask struct {
	//Test Task ID, from the schedular
	ID    string
	Case  libocit.TestCase
	Units []TSUnit
}

//TODO need a task store
func TSTaskNew(id string, tc libocit.TestCase) (task TSTask) {
	task.ID = id
	task.Case = tc

	for index := 0; index < len(tc.Units); index++ {
		unit := tc.Units[index]
		if unit.Class == libocit.TUOS {
			tsunit := TSOSUnit{unit}
			tsunit.SetBundleURL(tc.GetBundleURL())
			task.Units = append(task.Units, tsunit)
		} else {
			//			tsunit := TSContainerNew(unit)
		}
	}
	return task
}

func (task *TSTask) Run(action libocit.TestAction) bool {
	for index := 0; index < len(task.Units); index++ {
		//TODO async in the future
		succ := true
		task.Case.Units[index].ChangeStatus(succ)
		switch action {
		case libocit.TestActionApply:
			id := task.Units[index].Apply()
			task.Case.Units[index].SetResourceID(id)
			if len(id) > 0 {
				succ = true
			} else {
				succ = false
			}
			/*
				case libocit.TestActionDeploy:
					id := task.Case.Units[index].GetResourceID(id)
					bundleURL := task.Case.Units[index].GetBundleURL()
					succ = task.Units[index].Deploy(id, bundleURL) */
		default:
			id := task.Case.Units[index].GetResourceID()
			succ = task.Units[index].Run(id, action)
		}
		task.Case.Units[index].ChangeStatus(succ)
		if !succ {
			return false
		}
	}
	return true
}

func (task *TSTask) GetStatus() libocit.TestStatus {
	for index := 0; index < len(task.Units); index++ {
		//TODO: should we make the return value a list?
		return task.Units[index].GetStatus()
	}
	return libocit.TestStatusInit
}
