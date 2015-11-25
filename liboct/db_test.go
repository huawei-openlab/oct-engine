package liboct

import (
	//	"fmt"
	"testing"
)

func TestDBRegist(t *testing.T) {
	db := GetDefaultDB()
	err := db.RegistCollect(DBRepo)
	if err == nil {
		t.Log("DB Regist OK successful!")
	} else {
		t.Error("DB Regist OK failed!")
	}

	err = db.RegistCollect(DBRepo)
	if err != nil {
		t.Log("DB Regist Failed successful!")
	} else {
		t.Error("DB Regist Failed failed!")
	}
}

func TestDBGet(t *testing.T) {
	var repo TestCaseRepo
	db := GetDefaultDB()
	db.RegistCollect(DBRepo)

	repo.Name = "repo name"
	id, _ := db.Add(DBRepo, repo)

	_, err := db.Get(DBRepo, id)
	if err == nil {
		t.Log("DBGet OK successful !")
	} else {
		t.Error("DBGet OK failed !")
	}

	_, err = db.Get(DBRepo, "invalid")
	if err != nil {
		t.Log("DBGet Failed successful !")
	} else {
		t.Error("DBGet Failed failed !")
	}
}

func TestDBUpdate(t *testing.T) {
	var repo TestCaseRepo
	db := GetDefaultDB()
	db.RegistCollect(DBRepo)

	repo.Name = "repo name"
	id, _ := db.Add(DBRepo, repo)

	repo.Name = "updated name"
	if err := db.Update(DBRepo, id, repo); err == nil {
		t.Log("DBUpdate OK successful !")
	} else {
		t.Error("DBUpdate Failed failed !")
	}

	nRepoI, _ := db.Get(DBRepo, id)
	nRepo, _ := RepoFromString(nRepoI.String())
	if nRepo.Name == repo.Name {
		t.Log("DBUpdate Name OK successful !")
	} else {
		t.Error("DBUpdate Name Failed failed !")
	}
}
