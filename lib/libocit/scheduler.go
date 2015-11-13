package libocit

import (
	"encoding/json"
)

type SchedulerUnit interface {
	Apply() string
	//	Deploy(id string, bundleURL string) bool
	Run(id string, action TestAction) bool
	GetStatus() TestStatus
}

type Scheduler struct {
	ID      string
	Case    TestCase
	UnitIDs []string
}

func (s Scheduler) String() string {
	val, _ := json.Marshal(s)
	return string(val)
}

func SchedulerFromString(val string) (s Scheduler, err error) {
	err = json.Unmarshal([]byte(val), &s)
	return s, err
}

func (s *Scheduler) SetID(id string) {
	if id != s.ID {
		s.ID = id
	}
}

func (s *Scheduler) GetID() string {
	return s.ID
}

func SchedulerNew(tc TestCase) (s Scheduler, ok bool) {
	s.Case = tc
	for index := 0; index < len(s.Case.Units); index++ {
		s.Case.Units[index].SetStatus(TestStatusInit)
	}
	return s, true
}

func (s *Scheduler) Command(action TestAction) (succ bool) {
	succ = true
	for index := 0; index < len(s.Case.Units); index++ {
		ok := true
		switch action {
		case TestActionApply:
			ok = s.Case.Units[index].Apply()
			if ok {
				//TODO: each unit should have its own files
				s.Case.Units[index].SetBundleURL(s.Case.BundleURL)
			}
		case TestActionDeploy:
			ok = s.Case.Units[index].Deploy()
		case TestActionRun:
			ok = s.Case.Units[index].Run()
		case TestActionCollect:
			ok = s.Case.Units[index].Collect()
		case TestActionDestroy:
			ok = s.Case.Units[index].Destroy()
		}
		if ok == false {
			succ = false
			break
		}
	}
	return succ
}

func (s *Scheduler) GetStatus() TestStatus {
	for index := 0; index < len(s.Case.Units); index++ {
		//TODO: should we make the return value a list?
		return s.Case.Units[index].GetStatus()
	}
	return TestStatusInit
}
