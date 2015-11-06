package libocit

import (
	//	"fmt"
	"testing"
)

func TestDBRegist(t *testing.T) {
	val := DBRegist(DBRepo)
	if val {
		t.Log("DB Regist OK successful!")
	} else {
		t.Error("DB Regist OK failed!")
	}

	val = DBRegist(DBRepo)
	if !val {
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

	_, ok := DBGet(DBRepo, id)
	if ok {
		t.Log("DBGet OK successful !")
	} else {
		t.Error("DBGet OK failed !")
	}

	_, ok = DBGet(DBRepo, "invalid")
	if !ok {
		t.Log("DBGet Failed successful !")
	} else {
		t.Error("DBGet Failed failed !")
	}
}

/*
func DBModify(collect DBCollectName, id string, val DBInterface) bool {
	if !DBCollectExist(collect) {
		return false
	}
	_, ok := OCTDB[collect][id]
	if ok {
		OCTDB[collect][id] = val
	}

	return ok
}

func DBRemove(collect DBCollectName, id string) bool {
	if !DBCollectExist(collect) {
		return false
	}
	delete(OCTDB[collect], id)
	return true
}

func DBLookup(collect DBCollectName, query DBQuery) (ids []string) {
	switch collect {
	case DBCase:
	case DBRepo:
	case DBTask:
	}
	return ids
}
*/
