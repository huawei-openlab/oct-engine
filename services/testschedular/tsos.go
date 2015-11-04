package main

import (
	"../../lib/libocit"
	"fmt"
)

type TSOSUnit struct {
	libocit.TestUnit
}

//TODO add lock
var OSStore map[string]TSResource

func ApplyOS(unit TSOSUnit) string {
	for _, os := range OSStore {
		if os.Match(unit) {
			if os.Allocate(unit) {
				return os.ID
			}
		}
	}
	return ""
}

func GetOS(id string) (TSResource, bool) {
	v, ok := OSStore[id]
	return v, ok
}

func (os *TSResource) Match(unit TSOSUnit) bool {
	if unit.Distribution != os.Distribution {
		return false
	}
	if unit.Version != os.Version {
		return false
	}
	if unit.Arch != os.Arch {
		return false
	}

	//TODO: better calculation
	if unit.CPU > os.CPU {
		return false
	}
	if unit.Memory > os.Memory {
		return false
	}
	return true
}

func (os *TSResource) Allocate(unit TSOSUnit) bool {
	if os.MaxJobs > len(os.TestUnitIDs) {
		os.TestUnitIDs = append(os.TestUnitIDs, unit.GetID())
		return true
	}
	return false
}

func (os *TSResource) Deploy(unit TSOSUnit) bool {
	postURL := fmt.Sprintf("%s/task", os.URL)
	params := make(map[string]string)
	params[libocit.TestActionID] = unit.GetID()

	//FIXME: it is better to send the related the file to the certain host OS
	ret := libocit.SendFile(postURL, unit.GetBundleURL(), params)
	if ret.Status == libocit.RetStatusOK {
		return true
	}

	return false
}

func (os *TSResource) Run(unit TSOSUnit) bool {
	postURL := fmt.Sprintf("%s/task/%s", os.URL, unit.GetID())
	ret := libocit.SendCommand(postURL, []byte(libocit.TestActionRun))
	if ret.Status == libocit.RetStatusOK {
		return true
	}
	return false
}

func (os *TSResource) Collect(unit TSOSUnit) bool {
	return true
}

func (os *TSResource) Destroy(unit TSOSUnit) bool {
	return true
}

func (unit TSOSUnit) Run(action libocit.TestAction) bool {
	switch action {
	case libocit.TestActionApply:
		unit.SetStatus(libocit.TestStatusAllocating)
		id := ApplyOS(unit)
		if len(id) > 0 {
			unit.SetResourceID(id)
			unit.SetStatus(libocit.TestStatusAllocated)
		} else {
			unit.SetStatus(libocit.TestStatusAllocateFailed)
		}
	case libocit.TestActionDeploy:
		unit.SetStatus(libocit.TestStatusDeploying)
		os, ok := GetOS(unit.GetResourceID())
		if !ok {
			return false
		}
		if os.Deploy(unit) {
			unit.SetStatus(libocit.TestStatusDeployed)
		} else {
			unit.SetStatus(libocit.TestStatusDeployFailed)
		}
	case libocit.TestActionRun:
		unit.SetStatus(libocit.TestStatusRunning)
		os, ok := GetOS(unit.GetResourceID())
		if !ok {
			return false
		}
		if os.Run(unit) {
			unit.SetStatus(libocit.TestStatusRun)
		} else {
			unit.SetStatus(libocit.TestStatusRunFailed)
		}
	case libocit.TestActionCollect:
		unit.SetStatus(libocit.TestStatusCollecting)
		os, ok := GetOS(unit.GetResourceID())
		if !ok {
			return false
		}
		if os.Collect(unit) {
			unit.SetStatus(libocit.TestStatusCollected)
		} else {
			unit.SetStatus(libocit.TestStatusCollectFailed)
		}
	case libocit.TestActionDestroy:
		unit.SetStatus(libocit.TestStatusDestroying)
		os, ok := GetOS(unit.GetResourceID())
		if !ok {
			return false
		}
		if os.Destroy(unit) {
			unit.SetStatus(libocit.TestStatusFinish)
		} else {
			unit.SetStatus(libocit.TestStatusDestroyFailed)
		}
	}
	return true
}

func (unit TSOSUnit) GetStatus() libocit.TestStatus {
	return unit.GetStatus()
}
