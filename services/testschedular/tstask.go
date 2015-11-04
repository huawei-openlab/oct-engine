package main

import (
	"../../lib/libocit"
)

type TSUnit interface {
	Run(libocit.TestAction) bool
	GetStatus() libocit.TestStatus
}

type TSTask struct {
	//Test Task ID, from the schedular
	ID    string
	Case  libocit.TestCase
	Units []TSUnit
}

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

func (task *TSTask) Run(action libocit.TestAction) (succ bool) {
	succ = true
	for index := 0; index < len(task.Units); index++ {
		//TODO async in the future
		if !task.Units[index].Run(action) {
			succ = false
			break
		}
	}
	return succ
}

func (task *TSTask) GetStatus() {
	for index := 0; index < len(task.Units); index++ {
	}
}
