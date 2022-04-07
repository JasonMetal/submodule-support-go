package common

import (
	"context"
	"fmt"
	"github.com/jinzhu/gorm"
	"idea-go/bootstrap"
)

// MysqlClient mysql连接对象
type MysqlClient struct {
	DB     *gorm.DB
	Master bool
}

func NewMysqlClient(ctx context.Context, name string) *MysqlClient {
	mc := new(MysqlClient)
	gdb, err := mc.WithDBContext(ctx, name) //gdb.GetMysqlInstance(name)

	mc.CheckMysqlError(err)
	mc.DB = gdb

	return mc
}

func (mc *MysqlClient) CheckMysqlError(err error) {
	if err == nil || err == gorm.ErrRecordNotFound {
		return
	}
	// TODO 日志收集
	//ErrorLog(err, "mysql")
	fmt.Println("CheckMysqlError")
	fmt.Println(err)
}
func (mc MysqlClient) WithDBContext(ctx context.Context, name string) (*gorm.DB, error) {
	instance, err := bootstrap.GetMysqlInstance(name)
	if err != nil {
		return nil, err
	}

	return instance.DB, err
}
