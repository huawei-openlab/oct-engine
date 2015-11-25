package liboct

import (
	//	"fmt"
	"testing"
)

func DemoResourceA() (res Resource) {
	res.ResourceCommon.Class = ResourceOS
	res.ResourceCommon.Distribution = "busybox"
	res.ResourceCommon.Version = "0.1"
	res.ResourceCommon.Arch = "x86"
	res.ResourceCommon.CPU = 2
	res.ResourceCommon.Memory = 10

	res.URL = "resourceA url"

	return res
}

func DemoResourceB() (res Resource) {
	res.ResourceCommon.Class = ResourceOS
	res.ResourceCommon.Distribution = "busybox"
	res.ResourceCommon.Version = "0.1"
	res.ResourceCommon.Arch = "x86"
	res.ResourceCommon.CPU = 1
	res.ResourceCommon.Memory = 3

	res.URL = "resourceB url"
	return res
}

func DemoUnitA() (unit TestUnit) {
	res := DemoResourceA()
	unit.ResourceCommon = res.ResourceCommon
	unit.Name = "Demo Unit A"
	unit.Status = TestStatusInit

	return unit
}

func TestUnitString(t *testing.T) {
	unit := DemoUnitA()
	val := unit.String()

	nUnit, _ := UnitFromString(val)
	if nUnit.Name == unit.Name && unit.ResourceCommon == nUnit.ResourceCommon {
		t.Log("Unit Unmarshal successful")
	} else {
		t.Error("Unit Unmarshal failed")
	}

	nVal := nUnit.String()
	if val == nVal {
		t.Log("Unit string successful")
	} else {
		t.Log("Unit string failed")
	}
}

func TestUnitApply(t *testing.T) {
	db := GetDefaultDB()
	db.RegistCollect(DBResource)
	resA := DemoResourceA()
	db.Add(DBResource, resA)
	resB := DemoResourceB()
	db.Add(DBResource, resB)

	unit := DemoUnitA()
	if unit.Apply() == nil {
		t.Log("Unit Apply OK successful")
	} else {
		t.Error("Unit Apply OK failed")
	}

	unit.ResourceCommon.CPU = 200
	if unit.Apply() != nil {
		t.Log("Unit Apply failed successful")
	} else {
		t.Error("Unit Apply failed failed")
	}
}

func TestActionCommandString(t *testing.T) {
	var cmd TestActionCommand
	cmd.Action = TestActionRun
	cmd.Command = "ls -al "
	val := cmd.String()

	nCmd, _ := ActionCommandFromString(val)
	if nCmd.Action == cmd.Action && nCmd.Command == cmd.Command {
		t.Log("Action Command Unmarshal successful")
	} else {
		t.Error("Action Command Unmarshal failed")
	}

	nVal := nCmd.String()
	if val == nVal {
		t.Log("Action Command string successful")
	} else {
		t.Log("Action Command string failed")
	}
}
