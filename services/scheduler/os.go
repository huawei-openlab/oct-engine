package main

import (
	"../../lib/libocit"
	"fmt"
)

type SchedulerOSUnit struct {
	libocit.TestUnit
}

func (unit *SchedulerOSUnit) Deploy(id string) bool {
	res, ok := SchedulerRQuery(id)
	if !ok {
		return false
	}
	postURL := fmt.Sprintf("%s/task", res.URL)
	params := make(map[string]string)
	params[libocit.TestActionID] = id

	fmt.Println("OS deploy ", id, postURL, unit.GetBundleURL())
	//FIXME: it is better to send the related the file to the certain host OS
	ret := libocit.SendFile(postURL, unit.GetBundleURL(), params)
	if ret.Status == libocit.RetStatusOK {
		return true
	}

	return false
}

func (unit *SchedulerOSUnit) RunCommand(id string, action libocit.TestAction) bool {
	res, ok := SchedulerRQuery(id)
	if !ok {
		return false
	}
	postURL := fmt.Sprintf("%s/task/%s", res.URL, id)

	fmt.Println("OS run command ", action, postURL)

	//TODO: check if action is valid, or check it in the ocitd
	ret := libocit.SendCommand(postURL, []byte(action))
	if ret.Status == libocit.RetStatusOK {
		return true
	}
	return false
}

//interface function
func (unit SchedulerOSUnit) GetStatus() libocit.TestStatus {
	return unit.TestUnit.GetStatus()
}

func (unit SchedulerOSUnit) Apply() string {
	fmt.Println("OS apply")
	return SchedulerRApply(unit.TestUnit)
}

//interface function
func (unit SchedulerOSUnit) Run(id string, action libocit.TestAction) bool {
	fmt.Println("OS run ", id, action)
	switch action {
	case libocit.TestActionDeploy:
		return unit.Deploy(id)
	default:
		return unit.RunCommand(id, action)
	}
	return true
}
