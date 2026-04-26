package databases

import (
	"gorm.io/gorm"
)

type Table interface {
	TableName() string
	PrimaryKey() interface{}
}

type Entity interface {
	Id() interface{}
	Get() (bool, error)

	Count() (int64, error)
	Exists() (bool, error)

	Create() (int64, error)
	Update(where ...interface{}) (int64, error)
	UpdateByPk() (int64, error)

	Delete() (int64, error)
	DeleteByPk() (int64, error)
	Entity() interface{}
}

type SessionEntity interface {
	Entity
	DB() *gorm.DB
	Begin() error
	Commit() error
	Rollback() error
	Close()
}
type BaseEntity interface {
	Entity
	NewSession() SessionEntity
}

type entity struct {
	dao        BaseDao
	sessionDao SessionDao
	objectPtr  interface{}
}

func NewEntity(dao BaseDao, objectPtr interface{}) BaseEntity {
	return &entity{dao: dao, objectPtr: objectPtr}
}

func (e *entity) NewSession() SessionEntity {
	e.sessionDao = e.dao.NewSession()
	return e
}
func (e *entity) Entity() interface{} {
	return e.objectPtr
}

func (e *entity) Id() interface{} {
	return e.objectPtr.(Table).PrimaryKey()
}

func (e *entity) Exists() (bool, error) {
	if e.sessionDao != nil {
		return e.sessionDao.Exists(e.objectPtr)
	}
	return e.dao.Exists(e.objectPtr)
}
func (e *entity) Count() (int64, error) {
	if e.sessionDao != nil {
		return e.sessionDao.Count(e.objectPtr)
	}
	return e.dao.Count(e.objectPtr)
}

func (e *entity) Create() (int64, error) {
	if e.sessionDao != nil {
		return e.sessionDao.InsertOne(e.objectPtr)
	}
	return e.dao.InsertOne(e.objectPtr)
}

func (e *entity) Update(where ...interface{}) (int64, error) {
	if e.sessionDao != nil {
		return e.sessionDao.Update(e.objectPtr, where...)
	}
	return e.dao.Update(e.objectPtr, where...)
}

func (e *entity) UpdateByPk() (rows int64, err error) {
	if e.sessionDao != nil {
		return e.sessionDao.UpdateById(e.objectPtr.(Table).PrimaryKey(), e.objectPtr)
	}
	return e.dao.UpdateById(e.objectPtr.(Table).PrimaryKey(), e.objectPtr)
}

func (e *entity) Delete() (rows int64, err error) {
	if e.sessionDao != nil {
		return e.sessionDao.Delete(e.objectPtr)
	}
	return e.dao.Delete(e.objectPtr)
}

func (e *entity) DeleteByPk() (rows int64, err error) {
	if e.sessionDao != nil {
		return e.sessionDao.DeleteById(e.objectPtr.(Table).PrimaryKey(), e.objectPtr)
	}
	return e.dao.DeleteById(e.objectPtr.(Table).PrimaryKey(), e.objectPtr)
}

func (e *entity) Get() (has bool, err error) {
	if e.sessionDao != nil {
		return e.sessionDao.FindOne(e.objectPtr)
	}
	return e.dao.FindOne(e.objectPtr)
}

func (e *entity) DB() *gorm.DB {
	return e.sessionDao.DB()
}

func (e *entity) Begin() error {
	return e.sessionDao.Begin()
}

func (e *entity) Commit() error {
	return e.sessionDao.Commit()
}

func (e *entity) Rollback() error {
	return e.sessionDao.Rollback()
}

func (e *entity) Close() {
	e.sessionDao.Close()
}
