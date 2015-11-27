package liboct

import (
	"errors"
	"fmt"
	"time"
)

type DBCollectName string

const (
	DBCase     DBCollectName = "case"
	DBRepo                   = "repo"
	DBResource               = "resource"
	//testing task: for frontend users
	DBTask = "task"
	//scheduler: for backend servers
	DBScheduler = "scheduler"
	//scheduler unit
	DBUnit = "unit"
)

type DBInterface interface {
	String() string
}

type DBQuery struct {
	Page     int
	PageSize int
	Params   map[string]string
}

type DB struct {
	OCTDB map[DBCollectName](map[string]DBInterface)
}

var localDB DB

func DBNew() (db DB) {
	db.OCTDB = make(map[DBCollectName](map[string]DBInterface))
	return db
}

func GetDefaultDB() DB {
	if localDB.OCTDB == nil {
		localDB.OCTDB = make(map[DBCollectName](map[string]DBInterface))
	}
	return localDB
}

//The case, repo, resource should be consistent
//The task could always be different
func (db *DB) GenerateID(collect DBCollectName, val string) string {
	switch collect {
	case DBCase:
		tc, _ := CaseFromString(val)
		return MD5(tc.GetBundleContent())
	case DBRepo:
		repo, _ := RepoFromString(val)
		return MD5(fmt.Sprintf("%s%s", repo.URL, repo.CaseFolder))
	case DBTask:
	}
	return MD5(fmt.Sprintf("%d-%s", time.Now().Unix(), val))
}

func (db *DB) CollectExist(collect DBCollectName) error {
	if db.OCTDB == nil {
		return errors.New("The Database is not initialized.")
	}
	if _, ok := db.OCTDB[collect]; ok {
		return nil
	}

	return errors.New(fmt.Sprintf("The collect %s is not exist.", collect))
}

func (db *DB) RegistCollect(collect DBCollectName) error {
	if db.OCTDB == nil {
		db.OCTDB = make(map[DBCollectName](map[string]DBInterface))
	}
	_, ok := db.OCTDB[collect]
	if ok {
		return errors.New("The collect is already registed, donnot do it twice.")
	} else {
		db.OCTDB[collect] = make(map[string]DBInterface)
	}
	return nil
}

func (db *DB) Get(collect DBCollectName, id string) (DBInterface, error) {
	if err := db.CollectExist(collect); err != nil {
		return nil, err
	}

	if val, ok := db.OCTDB[collect][id]; ok {
		return val, nil
	}
	return nil, errors.New(fmt.Sprintf("Cannot find %s in the collect %s", id, collect))
}

func (db *DB) Add(collect DBCollectName, val DBInterface) (string, error) {
	if err := db.CollectExist(collect); err != nil {
		return "", err
	}

	id := db.GenerateID(collect, val.String())

	switch collect {
	case DBCase:
		tc, _ := CaseFromString(val.String())
		tc.SetID(id)
		db.OCTDB[collect][id] = tc
	case DBRepo:
		repo, _ := RepoFromString(val.String())
		repo.SetID(id)
		db.OCTDB[collect][id] = repo
	case DBTask:
		task, _ := TaskFromString(val.String())
		task.SetID(id)
		db.OCTDB[collect][id] = task
	case DBUnit:
		unit, _ := UnitFromString(val.String())
		unit.SetID(id)
		db.OCTDB[collect][id] = unit
	case DBResource:
		res, _ := ResourceFromString(val.String())
		res.SetID(id)
		db.OCTDB[collect][id] = res
	case DBScheduler:
		s, _ := SchedulerFromString(val.String())
		s.SetID(id)
		db.OCTDB[collect][id] = s
	}

	return id, nil
}

//If id exist,s modify it; if id does not exist, add it.
func (db *DB) Update(collect DBCollectName, id string, val DBInterface) error {
	if err := db.CollectExist(collect); err != nil {
		return err
	}

	if len(id) == 0 {
		return errors.New("Cannot update an empty id.")
	}

	db.OCTDB[collect][id] = val

	return nil
}

func (db *DB) Remove(collect DBCollectName, id string) error {
	if err := db.CollectExist(collect); err != nil {
		return err
	}
	delete(db.OCTDB[collect], id)
	return nil
}

func (db *DB) Lookup(collect DBCollectName, query DBQuery) (ids []string) {
	noLimit := false
	if query.Page < 0 {
		query.Page = 0
	}
	if query.PageSize < 0 {
		query.PageSize = 30
	} else if query.PageSize == 0 {
		noLimit = true
	}
	i := 0
	for id, _ := range db.OCTDB[collect] {
		if noLimit == true || (i >= query.Page*query.PageSize && i < (query.Page+1)*query.PageSize) {
			//TODO: check by 'reflect'
			switch collect {
			case DBCase:
			case DBRepo:
			case DBTask:
			}
			ids = append(ids, id)
		}
		i++
	}
	return ids
}
