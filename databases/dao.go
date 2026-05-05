package databases

import (
	"errors"
	"fmt"
	"reflect"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Dao interface {
	Count(bean interface{}) (int64, error)
	Exists(bean interface{}) (bool, error)

	InsertOne(entry interface{}) (int64, error)
	InsertMany(entries ...interface{}) (int64, error)

	Update(bean interface{}, where ...interface{}) (int64, error)
	UpdateById(id interface{}, bean interface{}) (int64, error)

	// 新增的 Upsert 方法
	Upsert(where interface{}, bean interface{}) (int64, error)
	UpsertById(id interface{}, bean interface{}) (int64, error)
	UpsertMany(wheres []interface{}, beans []interface{}) (int64, error)

	Delete(bean interface{}) (int64, error)
	DeleteById(id interface{}, bean interface{}) (int64, error)
	GetDBMetas() (map[string]interface{}, error)
	GetTableMetas(tableName string) ([]map[string]interface{}, error)
	FindById(id interface{}, bean interface{}) (bool, error)
	FindOne(bean interface{}) (bool, error)
	FindMany(rowsSlicePtr interface{}, orderBy string, condiBean ...interface{}) error
	FindAndCount(rowsSlicePtr interface{},
		pageable Pageable, condiBean ...interface{}) (int64, error)

	Query(rowsSlicePtr interface{}, sql string, Args ...interface{}) error
	CallProcedure(procName string, args ...interface{}) ([][]map[string]interface{}, error)
	Native() DBConn
	Migrations(tables []interface{}) error
}

type SessionDao interface {
	Dao
	DB() *gorm.DB
	Begin() error
	Commit() error
	Rollback() error
	Close()
}

type BaseDao interface {
	Dao
	NewSession() SessionDao
}

type OrmBaseDao struct {
	conn    DBConn
	tx      *gorm.DB
	inTrans bool
}

func NewBaseDao(conn DBConn) BaseDao {
	return &OrmBaseDao{
		conn:    conn,
		tx:      nil,
		inTrans: false,
	}
}

func (m *OrmBaseDao) DB() *gorm.DB {
	if m.inTrans && m.tx != nil {
		return m.tx
	}
	return (*gorm.DB)(m.conn)
}

// NewSession 创建一个session
func (m *OrmBaseDao) NewSession() SessionDao {
	return &OrmBaseDao{
		conn:    m.conn,
		tx:      nil,
		inTrans: false,
	}
}

// Begin 开启事务
func (m *OrmBaseDao) Begin() error {
	if m.inTrans {
		return errors.New("transaction already started")
	}
	m.tx = (*gorm.DB)(m.conn).Begin()
	if m.tx.Error != nil {
		return m.tx.Error
	}
	m.inTrans = true
	return nil
}

// Close 关闭事务
func (m *OrmBaseDao) Close() {
	if m.inTrans && m.tx != nil {
		// 如果事务还在进行中，进行回滚
		m.tx.Rollback()
	}
	m.inTrans = false
	m.tx = nil
}

func (m *OrmBaseDao) Commit() error {
	if !m.inTrans || m.tx == nil {
		return errors.New("no active transaction")
	}
	err := m.tx.Commit().Error
	m.inTrans = false
	m.tx = nil
	return err
}

func (m *OrmBaseDao) Rollback() error {
	if !m.inTrans || m.tx == nil {
		return errors.New("no active transaction")
	}
	err := m.tx.Rollback().Error
	m.inTrans = false
	m.tx = nil
	return err
}

func (m *OrmBaseDao) InsertOne(entry interface{}) (int64, error) {
	db := m.DB()
	result := db.Create(entry)
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}

func (m *OrmBaseDao) InsertMany(entries ...interface{}) (int64, error) {
	if len(entries) == 0 {
		return 0, nil
	}

	db := m.DB()
	result := db.Create(entries)
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}

func (m *OrmBaseDao) Update(bean interface{}, where ...interface{}) (int64, error) {
	db := m.DB()

	// 构建查询条件
	query := db.Model(bean)
	if len(where) > 0 {
		if len(where) == 1 {
			query = query.Where(where[0])
		} else if len(where) >= 2 {
			query = query.Where(where[0], where[1:]...)
		}
	}

	result := query.Updates(bean)
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}

func (m *OrmBaseDao) UpdateById(id interface{}, bean interface{}) (int64, error) {
	db := m.DB()
	result := db.Model(bean).Where("id = ?", id).Updates(bean)
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}

// Upsert 如果数据存在则更新，不存在则插入。where 用于 FirstOrCreate 查找条件。
func (m *OrmBaseDao) Upsert(where interface{}, bean interface{}) (int64, error) {
	db := m.DB()
	result := db.Where(where).Assign(bean).FirstOrCreate(bean)
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}

// UpsertById 根据ID进行Upsert操作
func (m *OrmBaseDao) UpsertById(id interface{}, bean interface{}) (int64, error) {
	db := m.DB()

	// 使用 GORM 的 Clauses 实现原子性 Upsert，指定冲突列为 id
	result := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}).Create(bean)

	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}

// UpsertMany 批量Upsert操作
func (m *OrmBaseDao) UpsertMany(wheres []interface{}, beans []interface{}) (int64, error) {
	if len(wheres) != len(beans) {
		return 0, fmt.Errorf("wheres 和 beans 长度不一致")
	}

	var totalAffected int64 = 0

	for i, bean := range beans {
		affected, err := m.Upsert(wheres[i], bean)
		if err != nil {
			return totalAffected, err
		}
		totalAffected += affected
	}

	return totalAffected, nil
}

func (m *OrmBaseDao) Delete(bean interface{}) (int64, error) {
	db := m.DB()
	result := db.Delete(bean)
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}

func (m *OrmBaseDao) DeleteById(id interface{}, bean interface{}) (int64, error) {
	db := m.DB()
	result := db.Delete(bean, id)
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}

func (m *OrmBaseDao) Query(rowsSlicePtr interface{}, sql string, Args ...interface{}) error {
	db := m.DB()
	return db.Raw(sql, Args...).Scan(rowsSlicePtr).Error
}

func (m *OrmBaseDao) FindById(id interface{}, bean interface{}) (bool, error) {
	db := m.DB()
	result := db.First(bean, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, result.Error
	}
	return true, nil
}

func (m *OrmBaseDao) FindOne(bean interface{}) (bool, error) {
	db := m.DB()
	result := db.First(bean, bean)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, result.Error
	}
	return true, nil
}

func (m *OrmBaseDao) Count(bean interface{}) (int64, error) {
	db := m.DB()
	var count int64
	err := db.Model(bean).Count(&count).Error
	return count, err
}

func (m *OrmBaseDao) Exists(bean interface{}) (bool, error) {
	count, err := m.Count(bean)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
func (m *OrmBaseDao) FindMany(rowsSlicePtr interface{}, sort string, condiBean ...interface{}) error {
	db := m.DB()
	query := db

	// 添加条件
	if len(condiBean) > 0 {
		query = query.Where(condiBean[0])
	}

	// 添加排序
	if sort != "" {
		query = query.Order(sort)
	}

	return query.Find(rowsSlicePtr).Error
}

func (m *OrmBaseDao) FindAndCount(rowsSlicePtr interface{},
	pageable Pageable, condiBean ...interface{}) (int64, error) {

	db := m.DB()

	// 获取反射值来确定模型类型
	sliceValue := reflect.ValueOf(rowsSlicePtr)
	if sliceValue.Kind() != reflect.Ptr || sliceValue.Elem().Kind() != reflect.Slice {
		return 0, errors.New("rowsSlicePtr must be a pointer to slice")
	}

	// 获取slice元素类型
	sliceType := sliceValue.Elem().Type()
	elementType := sliceType.Elem()

	// 创建一个新的实例来获取模型
	modelInstance := reflect.New(elementType).Interface()

	query := db.Model(modelInstance)

	// 添加条件
	if len(condiBean) > 0 {
		query = query.Where(condiBean[0])
	}

	// 计算总数
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}

	// 分页查询
	query = query.Offset(pageable.Skip()).Limit(pageable.Limit())

	// 添加排序
	if pageable.Sort() != "" {
		query = query.Order(pageable.Sort())
	}

	err := query.Find(rowsSlicePtr).Error
	return count, err
}

func (m *OrmBaseDao) Native() DBConn {
	return m.conn
}

func (m *OrmBaseDao) Migrations(tables []interface{}) error {
	db := (*gorm.DB)(m.conn)
	// 禁止 AutoMigrate 自动创建 FK 约束，避免微服务环境下的约束冲突
	noFKDB := db.Session(&gorm.Session{NewDB: true})
	noFKDB.DisableForeignKeyConstraintWhenMigrating = true
	return noFKDB.AutoMigrate(tables...)
}

func (m *OrmBaseDao) GetDBMetas() (map[string]interface{}, error) {
	// GORM doesn't have direct equivalent to xorm's DBMetas
	// This is a simplified implementation
	db := (*gorm.DB)(m.conn)

	var tables []string
	var query string

	// 根据数据库类型选择不同的查询
	switch db.Dialector.Name() {
	case "mysql":
		query = "SHOW TABLES"
	case "postgres":
		query = "SELECT tablename FROM pg_tables WHERE schemaname = 'public'"
	default:
		return nil, errors.New("unsupported database type")
	}

	if err := db.Raw(query).Scan(&tables).Error; err != nil {
		return nil, err
	}

	result := make(map[string]interface{})
	for _, table := range tables {
		result[table] = make(map[string]interface{})
	}

	return result, nil
}

func (m *OrmBaseDao) GetTableMetas(tableName string) ([]map[string]interface{}, error) {
	db := (*gorm.DB)(m.conn)

	var columns []map[string]interface{}

	// 根据数据库类型选择不同的查询
	switch db.Dialector.Name() {
	case "mysql":
		// MySQL 使用 information_schema 避免 SQL 注入
		query := `SELECT COLUMN_NAME, COLUMN_TYPE, IS_NULLABLE, COLUMN_KEY, COLUMN_DEFAULT, EXTRA
				FROM information_schema.COLUMNS
				WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?`
		if err := db.Raw(query, tableName).Scan(&columns).Error; err != nil {
			return nil, err
		}
	case "postgres":
		query := `SELECT column_name, data_type, is_nullable
				FROM information_schema.columns
				WHERE table_name = ? AND table_schema = 'public'`
		if err := db.Raw(query, tableName).Scan(&columns).Error; err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("unsupported database type")
	}

	return columns, nil
}

func (m *OrmBaseDao) CallProcedure(procName string, args ...interface{}) ([][]map[string]interface{}, error) {
	db := (*gorm.DB)(m.conn)

	// 验证存储过程名称，防止SQL注入
	for _, c := range procName {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '.') {
			return nil, fmt.Errorf("invalid procedure name: %s", procName)
		}
	}

	// 构建存储过程调用的 SQL 语句
	placeholders := ""
	for i := range args {
		if i > 0 {
			placeholders += ", "
		}
		placeholders += "?"
	}
	sql := fmt.Sprintf("CALL %s(%s)", procName, placeholders)

	// 执行存储过程
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	rows, err := sqlDB.Query(sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var allResults [][]map[string]interface{}

	for {
		var results []map[string]interface{}
		columns, err := rows.Columns()
		if err != nil {
			return nil, err
		}

		for rows.Next() {
			// 创建一个切片来保存每一行的列值
			values := make([]interface{}, len(columns))
			valuePtrs := make([]interface{}, len(columns))
			for i := range columns {
				valuePtrs[i] = &values[i]
			}

			// 扫描当前行的列值到切片中
			if err := rows.Scan(valuePtrs...); err != nil {
				return nil, err
			}

			// 创建一个映射来保存列名和对应的值
			rowMap := make(map[string]interface{})
			for i, col := range columns {
				var v interface{}
				val := values[i]
				b, ok := val.([]byte)
				if ok {
					v = string(b)
				} else {
					v = val
				}
				rowMap[col] = v
			}

			results = append(results, rowMap)
		}
		allResults = append(allResults, results)
		if !rows.NextResultSet() {
			break
		}
	}
	return allResults, nil
}
