package liboct

import (
	"encoding/json"
)

const SchedulerCacheDir = "/tmp/.test_schedular_cache"

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
		for index := 0; index < len(s.Case.Units); index++ {
			s.Case.Units[index].SetSchedulerID(id)
		}
	}
}

func (s *Scheduler) GetID() string {
	return s.ID
}

func SchedulerNew(tc TestCase) (s Scheduler, err error) {
	s.Case = tc
	for index := 0; index < len(s.Case.Units); index++ {
		s.Case.Units[index].SetStatus(TestStatusInit)
	}
	return s, nil
}

func (s *Scheduler) Command(action TestAction) (err error) {
	for index := 0; index < len(s.Case.Units); index++ {
		switch action {
		case TestActionApply:
			err = s.Case.Units[index].Apply()
			if err == nil {
				//TODO: each unit should have its own files
				s.Case.Units[index].SetBundleURL(s.Case.BundleURL)
			}
		case TestActionDeploy:
			err = s.Case.Units[index].Deploy()
		case TestActionRun:
			err = s.Case.Units[index].Run()
		case TestActionCollect:
			err = s.Case.Units[index].Collect()
		case TestActionDestroy:
			err = s.Case.Units[index].Destroy()
		}

		if err != nil {
			break
		}
	}
	return err
}

func (s *Scheduler) GetStatus() TestStatus {
	for index := 0; index < len(s.Case.Units); index++ {
		//TODO: should we make the return value a list?
		return s.Case.Units[index].GetStatus()
	}
	return TestStatusInit
}
