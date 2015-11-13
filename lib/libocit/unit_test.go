package libocit

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
	DBRegist(DBResource)
	resA := DemoResourceA()
	DBAdd(DBResource, resA)
	resB := DemoResourceB()
	DBAdd(DBResource, resB)

	unit := DemoUnitA()
	if unit.Apply() == true {
		t.Log("Unit Apply OK successful")
	} else {
		t.Error("Unit Apply OK failed")
	}

	unit.ResourceCommon.CPU = 200
	if unit.Apply() == false {
		t.Log("Unit Apply failed successful")
	} else {
		t.Error("Unit Apply failed failed")
	}
}
