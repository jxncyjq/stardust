package databases

import (
	"gorm.io/gorm"
)

func DoGet(call func() (bool, error)) error {
	has, err := call()
	if nil != err {
		return err
	}
	if !has {
		return ErrGetEmpty
	}
	return nil
}
func DoUpdate(call func() (int64, error)) error {
	ret, err := call()
	if nil != err {
		return err
	}
	if 0 == ret {
		return ErrUpdatedEmpty
	}
	return nil
}

func DoInsert(call func() (int64, error)) error {
	ret, err := call()
	if nil != err {
		return err
	}
	if 0 == ret {
		return ErrInsertedEmpty
	}
	return nil
}

func DoDelete(call func() (int64, error)) error {
	ret, err := call()
	if nil != err {
		return err
	}
	if 0 == ret {
		return ErrDeletedEmpty
	}
	return nil
}

type SessionDoctor int

type SessionHandler func(tx *gorm.DB) (interface{}, SessionDoctor, error)

// WarpSession 事务装饰器
// error 为真事务回滚
// SessionDoctor 决定回滚还是提交
func WarpSession(db *gorm.DB, h SessionHandler) (interface{}, int, error) {
	tx := db.Begin()
	if tx.Error != nil {
		return nil, 0, tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	value, doctor, err := h(tx)
	if err != nil {
		tx.Rollback()
		return nil, 0, err
	}

	if doctor == SessionDoctorCommit {
		if err := tx.Commit().Error; err != nil {
			return nil, 0, err
		}
	} else {
		if err := tx.Rollback().Error; err != nil {
			return nil, 0, err
		}
	}

	return value, 0, nil
}

type SessionWrapper struct {
	dao     BaseDao
	session SessionDao
	err     error
}

type ExecuteFunc func(session SessionDao) error

func NewSessionWrapper(s BaseDao) SessionWrapper {
	sw := SessionWrapper{
		dao:     s,
		session: s.NewSession(),
		err:     nil,
	}
	sw.err = sw.session.Begin()
	return sw
}

func (sw SessionWrapper) Execute(call ExecuteFunc) SessionWrapper {
	// 有错误直接返回
	if sw.err != nil {
		return sw
	}
	if sw.err = call(sw.session); sw.err != nil {
		_ = sw.session.Rollback()
	}
	return sw
}

func (sw SessionWrapper) CommitAndClose() error {
	defer sw.session.Close()
	if sw.err == nil {
		sw.err = sw.Commit()
	}
	return sw.err
}

func (sw SessionWrapper) Commit() error {
	// 没有错误，提交commit
	if sw.err == nil {
		sw.err = sw.session.Commit()
	} else {
		// 有错误，回滚
		_ = sw.session.Rollback()
	}

	return sw.err
}

func (sw SessionWrapper) Close() error {
	// 有错误则回滚后更新
	if sw.err != nil {
		_ = sw.session.Rollback()
	}
	sw.session.Close()
	return sw.err
}

func (sw SessionWrapper) GetErr() error {
	return sw.err
}
