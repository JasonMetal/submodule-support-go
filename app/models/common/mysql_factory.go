package common

import (
	"context"
)

// Manyideacloud 获取manyideacloud库客户端连接对象
func Manyideacloud(ctx context.Context) *MysqlClient {
	return NewMysqlClient(ctx, "manyideacloud")
}
