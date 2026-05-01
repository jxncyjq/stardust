package databases

import (
	"bytes"
	"errors"
	"log"
	"text/template"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

// DBConn 数据库接口 - 直接使用gorm.DB
type DBConn *gorm.DB

func NewDBConn(dbConfig *Config) (DBConn, error) {
	if dbConfig.UseMasterSlave {
		return NewMSConn(dbConfig)
	}
	return NewSingleConn(dbConfig)
}

// NewSingleConn 初始化数据库连接
// mysql fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=%s",user,pwd,host,db,charset)
// postgres fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s sslcert=%s sslkey=%s sslrootcert=%s",
//
//	host, port, user, name, pass, sslMode, SslCert, SslKey, SslRootCert)
func NewSingleConn(c *Config) (DBConn, error) {
	if nil == c || "" == c.Master {
		return nil, errors.New("config or config.Url can not be null")
	}

	var conn *gorm.DB
	var err error

	switch c.DbType {
	case "mysql":
		conn, err = gorm.Open(mysql.Open(c.Master), &gorm.Config{})
	case "postgres":
		conn, err = gorm.Open(postgres.Open(c.Master), &gorm.Config{})
	default:
		conn, err = gorm.Open(mysql.Open(c.Master), &gorm.Config{})
	}

	if err != nil {
		log.Println("failed to initializing db connection:", err)
		return nil, err
	}

	// 获取底层sql.DB以设置连接池参数
	sqlDB, err := conn.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(c.MaxIdle)
	sqlDB.SetMaxOpenConns(c.MaxConn)

	return conn, nil
}

// NewMSConn 初始化主从数据库连接, master不能为空，slaves可以为空
func NewMSConn(c *Config) (DBConn, error) {
	if nil == c || "" == c.Master {
		return nil, errors.New("config or config.Url can not be null")
	}

	// 先创建主库连接
	conn, err := NewSingleConn(c)
	if err != nil {
		return nil, err
	}

	// 如果没有从库配置，直接返回主库连接
	if len(c.Slaves) == 0 {
		log.Println("No slave databases configured, using master only")
		return conn, nil
	}

	// 配置从库
	var replicas []gorm.Dialector
	for _, slave := range c.Slaves {
		var dialector gorm.Dialector
		switch c.DbType {
		case "postgres":
			dialector = postgres.Open(slave)
		default:
			dialector = mysql.Open(slave)
		}
		replicas = append(replicas, dialector)
	}

	// 使用 DBResolver 插件配置读写分离
	err = (*gorm.DB)(conn).Use(dbresolver.Register(dbresolver.Config{
		Replicas: replicas,
		Policy:   dbresolver.RandomPolicy{}, // 随机选择从库
	}).SetMaxIdleConns(c.MaxIdle).SetMaxOpenConns(c.MaxConn))

	if err != nil {
		return nil, err
	}

	log.Printf("Master-Slave configuration enabled: 1 master, %d slaves", len(c.Slaves))
	return conn, nil
}

// GetConnStr 检查传入字典和类型，返回数据库连接字符串，
// @params c map[string]interface{}{ "host":"127.0.0.1","port":3306,"user":"root","passwd":"123456","dbname":"csv"}
// @param t string "mysql" or "postgres"
// @return string, error
func GetConnStr(c map[string]interface{}, t string) (string, error) {
	var constr bytes.Buffer
	var tpr *template.Template
	var err error
	switch t {
	case "mysql":
		{
			tmpl := `{{.user}}:{{.pwd}}@tcp({{.host}}:{{.port}})/{{.db}}?charset=utf8mb4&parseTime=true&loc=Local`
			tpr, err = template.New("config").Parse(tmpl)
		}
	case "postgres":
		{
			tmpl := `host={{.host}} port={{.port}} user={{.user}} dbname={{.db}} password={{.pwd}} sslmode=disable`
			tpr, err = template.New("config").Parse(tmpl)
		}
	default:
		return "", errors.New("unsupported db type")
	}
	// 错误检查
	if err != nil {
		return "", err
	}
	if err = tpr.Execute(&constr, c); err != nil {
		return "", err
	}
	return constr.String(), nil
}
