package libocit

import (
	"testing"
)

func DemoRepo() (tcr TestCaseRepo) {
	tcr.Name = "liangchenye"
	tcr.URL = "github.com/liangchenye/oct-casedemo"
	tcr.Enable = true
	tcr.CaseFolder = "casedb"
	tcr.Groups = append(tcr.Groups, "oci-spec")

	tcr.SetCacheDir("/tmp/.temp_cache")
	return tcr
}

func TestCaseRepoValid(t *testing.T) {
	var tcr TestCaseRepo

	tcr.Name = "name"
	_, valid := tcr.IsValid()
	if valid {
		t.Error("Valid with wrong repo failed!")
	} else {
		t.Log("Valid with wrong repo successful!")
	}

	tcrok := DemoRepo()

	_, valid = tcrok.IsValid()
	if valid {
		t.Log("Valid with right repo successful!")
	} else {
		t.Error("Valid with right repo failed!")
	}
}

func TestCaseRepoRefresh(t *testing.T) {
	tcr := DemoRepo()
	if tcr.Refresh() {
		t.Log("Refresh with right repo successful!")
	} else {
		t.Error("Refresh with right repo failed!")
	}

	tcr.CaseFolder = "neverexistdb"
	if tcr.Refresh() {
		t.Error("Refresh with wrong repo failed!")
	} else {
		t.Log("Refresh with wrong repo successful!")
	}
}

func TestCaseRepoLoadCases(t *testing.T) {
	tcr := DemoRepo()
	tcr.Refresh()
	cases := tcr.LoadCases()
	if len(cases) == 2 {
		t.Log("Load cases successful!")
	} else {
		t.Error("Load cases failed!")
	}
}

func TestCaseRepoLoadCase(t *testing.T) {
	tcr := DemoRepo()
	tcr.Refresh()
	_, err := tcr.LoadCase("oci-spec/bundle")
	if err == nil {
		t.Log("Load right case successful!")
	} else {
		t.Error("Load right case failed!")
	}
	_, err = tcr.LoadCase("oci-spec/bundle-not-exist")
	if err != nil {
		t.Log("Load wrong case successful!")
	} else {
		t.Error("Load wrong case failed!")
	}
}

func TestCaseRepoPurge(t *testing.T) {
	//TO make it easier fo debug, donnot open it
	return

	tcr := DemoRepo()
	tcr.Purge()
	cases := tcr.LoadCases()
	if len(cases) == 0 {
		t.Log("Repo purge successful!")
	} else {
		t.Log("Repo purge failed!")
	}
}
