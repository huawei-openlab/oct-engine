package liboct

import (
	"testing"
)

func DemoCase() (tcr TestCase) {
	tcr.Name = "democase"
	tcr.Summary = "used test be demo"
	tcr.Version = "0.1"
	tcr.License = "ab"
	tcr.Group = "demo"
	tcr.Owner = "oo"
	tcr.Description = "desr"

	var unit TestUnit
	unit.Name = "unitA"

	tcr.Units = append(tcr.Units, unit)
	return tcr
}

func TestCaseValid(t *testing.T) {
}

func TestLoadCase(t *testing.T) {
	bundleURL := "./test/bundle-noreport"
	_, err := CaseFromBundle(bundleURL)
	if err == nil {
		t.Log("Load case failed!")
	} else {
		t.Error("Load case successful!")
	}

}

func TestGetReportStatus(t *testing.T) {
	type statusTest struct {
		url string
		hr  bool
		cu  bool
		err bool
	}
	var statusTests []statusTest = make([]statusTest, 3)
	statusTests[0] = statusTest{"./test/bundle-noreport", false, true, true}
	statusTests[1] = statusTest{"./test/bundle-report-updated", true, true, true}
	statusTests[2] = statusTest{"./test/bundle-report-noupdated", true, false, true}

	for index := 0; index < len(statusTests); index++ {
		st := statusTests[index]
		tc, _ := CaseFromBundle(st.url)
		hasReport, caseUpdated, err := tc.GetReportStatus()
		ok := true
		if err != nil {
			ok = false
		}
		if hasReport == st.hr && caseUpdated == st.cu && ok == st.err {
			t.Log("GetReportStatus success.")
		} else {
			t.Error("GetReportStatus failed.", hasReport, caseUpdated, ok, st)
		}
	}
	tc := DemoCase()
	_, _, err := tc.GetReportStatus()
	if err != nil {
		t.Log("GetReportStatus with naked spec success.")
	} else {
		t.Error("GetReportStatus with nakded spec failed.")
	}
}

func TestGetBundleTarURL(t *testing.T) {
	bundleURL := "./test/bundle-noreport"
	tc, _ := CaseFromBundle(bundleURL)
	tarURL := tc.GetBundleTarURL()
	if tarURL == "./test/bundle-noreport.tar.gz" {
		t.Log("GetBundleTarURL ok success.")
	} else {
		t.Error("GetBundleTarURL ok failed.")
	}

	_, err := CaseFromTar(tarURL, "./cache/bundle-noreport")
	if err == nil {
		t.Log("CaseFromTar ok success.")
	} else {
		t.Error("CaseFromTar ok failed.")
	}
}
