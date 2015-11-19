package liboct

import (
	//	"fmt"
	"testing"
)

func TestDBRegist(t *testing.T) {
	err := DBRegist(DBRepo)
	if err == nil {
		t.Log("DB Regist OK successful!")
	} else {
		t.Error("DB Regist OK failed!")
	}

	err = DBRegist(DBRepo)
	if err != nil {
		t.Log("DB Regist Failed successful!")
	} else {
		t.Error("DB Regist Failed failed!")
	}
}

func TestDBGet(t *testing.T) {
	var repo TestCaseRepo
	DBRegist(DBRepo)

	repo.Name = "repo name"
	id, _ := DBAdd(DBRepo, repo)

	_, err := DBGet(DBRepo, id)
	if err == nil {
		t.Log("DBGet OK successful !")
	} else {
		t.Error("DBGet OK failed !")
	}

	_, err = DBGet(DBRepo, "invalid")
	if err != nil {
		t.Log("DBGet Failed successful !")
	} else {
		t.Error("DBGet Failed failed !")
	}
}

func TestDBUpdate(t *testing.T) {
	var repo TestCaseRepo
	DBRegist(DBRepo)

	repo.Name = "repo name"
	id, _ := DBAdd(DBRepo, repo)

	repo.Name = "updated name"
	if err := DBUpdate(DBRepo, id, repo); err == nil {
		t.Log("DBUpdate OK successful !")
	} else {
		t.Error("DBUpdate Failed failed !")
	}

	nRepoI, _ := DBGet(DBRepo, id)
	nRepo, _ := RepoFromString(nRepoI.String())
	if nRepo.Name == repo.Name {
		t.Log("DBUpdate Name OK successful !")
	} else {
		t.Error("DBUpdate Name Failed failed !")
	}
}
