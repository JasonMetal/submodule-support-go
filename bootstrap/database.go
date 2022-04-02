package bootstrap

import (
	"fmt"
	"github.com/jinzhu/gorm"
	yCfg "github.com/olebedev/config"
	"idea-go/helpers/config"
	"strings"
	"time"
)

type MysqlInstance struct {
	DSN string
	DB  *gorm.DB
}

var mysqlDbList = make(map[string]MysqlInstance)

func InitMysql() {
	path := fmt.Sprintf("./config/%s/mysql.yml", DevEnv)

	dbList := getDbNames(path)
	for _, dbname := range dbList {
		instance, err := initDbConn(dbname)
		if err == nil {
			mysqlDbList[dbname] = instance
		}
	}
}

func getDbNames(filename string) []string {
	DbNames := make([]string, 0)
	DBConfigs, err := config.GetConfig(filename)
	fmt.Println(filename)
	configList, err := DBConfigs.Map(filename)
	if err == nil {
		for DBName, _ := range configList {
			DbNames = append(DbNames, DBName)
		}
	}

	return DbNames
}

func initDbConn(dbName string) (MysqlInstance, error) {

	cfg, err := config.GetConfig("mysql")

	maxOpenConns, _ := cfg.Int("mysql." + dbName + ".maxOpenConns")
	maxIdleConns, _ := cfg.Int("mysql." + dbName + ".maxIdleConns")
	maxLifetime, _ := cfg.Int("mysql." + dbName + ".maxLifetime")
	tablePrefix, _ := cfg.String("mysql." + dbName + ".tablePrefix")
	debug, _ := cfg.Bool("mysql." + dbName + ".debug")
	charset, _ := cfg.String("mysql." + dbName + ".charset")

	servers, err := cfg.List("mysql." + dbName + ".servers")
	if err != nil || len(servers) < 1 {
		return MysqlInstance{}, err
	}

	if tablePrefix != "" {
		setTablePrefix(tablePrefix)
	}

	host, _ := yCfg.Get(servers[0], "host")
	port, _ := yCfg.Get(servers[0], "port")
	name, _ := yCfg.Get(servers[0], "db")
	user, _ := yCfg.Get(servers[0], "user")
	passwd, _ := yCfg.Get(servers[0], "passwd")

	addr := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		//addr := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True",
		user.(string), passwd.(string), host.(string), port.(int), name.(string), charset)

	if connTimeout, err := yCfg.Get(servers[0], "connTimeout"); err == nil {
		addr += fmt.Sprintf("&timeout=%dms", connTimeout.(int))
	}
	if readTimeout, err := yCfg.Get(servers[0], "readTimeout"); err == nil {
		addr += fmt.Sprintf("&readTimeout=%dms", readTimeout.(int))
	}
	if writeTimeout, err := yCfg.Get(servers[0], "writeTimeout"); err == nil {
		addr += fmt.Sprintf("&writeTimeout=%dms", writeTimeout.(int))
	}

	db, err := gorm.Open("mysql", addr)

	if err != nil {
		//return MysqlInstance{}, errors.New("connection is not exist")
		return MysqlInstance{}, err
	}

	if debug {
		db.LogMode(true)
	}

	err = db.DB().Ping()
	if err != nil {
		return MysqlInstance{}, err
	}

	if maxLifetime > 0 {
		db.DB().SetConnMaxLifetime(time.Duration(maxLifetime) * time.Second)
	}

	db.DB().SetMaxIdleConns(maxIdleConns)
	db.DB().SetMaxOpenConns(maxOpenConns)

	return MysqlInstance{fmt.Sprintf("%s:%d/%s", host.(string), port.(int), name.(string)), db}, nil
}

func setTablePrefix(TablePrefix string) {
	gorm.DefaultTableNameHandler = func(db *gorm.DB, defaultTableName string) string {
		if !strings.HasPrefix(defaultTableName, TablePrefix) {
			return TablePrefix + defaultTableName
		}
		return defaultTableName
	}
}
