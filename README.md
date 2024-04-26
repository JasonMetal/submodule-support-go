# support-go




## 运行环境

- 开发语言: go1.18+
- 开发框架: gin
- RPC框架: grpc
- 数据库: mysql, redis
- 消息队列: 
- 代码管理工具：

## github

### 地址

[https://github.com/JasonMetal/submodule-support-go](https://github.com/JasonMetal/submodule-support-go)

### 分支说明

- master：正式环境
- test：测试环境, 主要提供给测试同学测试使用
- bvt：预发环境, 正式环境数据库一致, 预发环境测试通过后才可以发布线上环境

## 目录结构

``` 
.
├── Dockerfile          
├── LICENSE
├── README.md
├── app                 # 应用代码
│   ├── entity            # 各种结构体定义
│   ├── grpc              # 提供GRPC服务
│   ├── http              # 提供HTTP服务
│   ├── logic             # 逻辑层
│   ├── models            # 数据模型层，数据表实体类
│   └── services          # 服务层
├── bootstrap          
│   ├── app.go
│   ├── config.go
│   ├── database.go
│   └── redis.go 
├── config              # 各个环境配置
│   ├── bvt
│   ├── local
│   ├── online
│   └── test
├── go-build.sh         # 用于ci构建
├── go.mod
├── go.sum
├── helpers             # 常用类库封装
│   ├── config
│   ├── env
│   ├── errors
│   ├── logger
│   └── strings
├── http-server
├── main.go             # 入口文件
└── routes              # 路由文件夹
    ├── base.go
    ├── other.go
    └── prize.go    
    
```

## 本地环境配置

1. 拉取代码
    ```shell
    git clone git@github.com:JasonMetal/submodule-support-go.git
    ```


2. 启动

   ```shell
   go run main.go -e=local # 本地可以省略local
   ```

## git-子模块
 - 无

## gitlab地址
- 无


## 使用说明
###  顺序流程图

![](https://static.manyidea.cloud/dev/uploads/uploadfile/images/20220407/img20407112944001599126087.png)



### 新模块路由配置:

1. routes目录中新建: 新路由名(例如:brand.go)

```go
package router

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"support-go/app/http/controllers/brand"
)

func RegisterOther(router *gin.Engine) {
	fmt.Println("Registered brand router")
	v2 := router.Group("/v2")
	{
		// http://127.0.0.1:50069/v2/test1/detail
		v2.GET("/test1/detail", func(ctx *gin.Context) {
			testOne.NewTestOne(ctx).GetTest1()
		})
		
		// http://127.0.0.1:50069/v2/test1/update
		v2.POST("/test1/update", func(ctx *gin.Context) {
			testOne.NewTestOne(ctx).UpdateTest1()
		})
	}
}

```
2. 控制器(app/http/controllers/brand)
```go
package prize

import (
	"github.com/gin-gonic/gin"
	"support-go/app/http/controllers"
	"support-go/app/logic/prize"
	"strconv"
)

type PrizeController struct {
	controllers.BaseController
}

func NewPrizeController(ctx *gin.Context) *PrizeController {
	return &PrizeController{
		controllers.NewBaseBaseController(ctx),
	}
}

func (p *PrizeController) GetList() {
	prizeLogic := prize.NewPrizeLogic(p.GCtx)
	rid, err := strconv.ParseInt(p.GetQueryDefault("rid", "0").Val, 10, 32)
	if err != nil {
		rid = 0
	}

	ranking := prizeLogic.GetPrizeList(uint32(rid))
	p.Success(ranking)
}


```



## 相关文档

- [gin框架](https://github.com/gin-gonic/gin)
- [grpc文档](https://grpc.io/docs/)
- [git子模块文档](https://git-scm.com/book/zh/v2/Git-%E5%B7%A5%E5%85%B7-%E5%AD%90%E6%A8%A1%E5%9D%97)
