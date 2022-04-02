# idea-go


## 运行环境

- 开发语言: go1.18+
- 开发框架: gin
- RPC框架: grpc
- 数据库: mysql, redis
- 消息队列: 
- 代码管理工具：

## gitee

### 地址

[https://e.gitee.com/DXTeam/repos/DXTeam/idea-go](https://e.gitee.com/DXTeam/repos/DXTeam/idea-go)

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
├── app         
│      ├── grpc       # 提供grpc服务
│      ├── http       # 提供http服务
│      ├── models     # 数据模型层，数据表实体类
│      └── services   # 服务层
├── bootstrap     
│      ├── app.go
│      ├── config.go
│      ├── database.go
│      └── redis.go
├── config          # 各个环境配置文件夹
│      ├── bvt
│      ├── local
│      ├── online
│      └── test
├── go.mod
├── go.sum
├── helpers        # 常用类库封装
│      ├── config
│      ├── env
│      ├── errors
│      ├── logger
│      └── strings
├── main.go         # 入口文件
└── routes          # 路由文件夹
    ├── base.go
    └── other.go
    
```

## 本地环境配置

1. 拉取代码
    ```shell
    git clone git@gitee.com:DXTeam/idea-go.git
    ```

2. 安装vendor包里的git submodule
    ```shell
    # 进入目录
    cd idea-go
    # 初次下载
   
    ```

3. 启动
   ```shell
   go run server.go -e=local
   ```

## git-子模块

### gitlab地址

- [grpc-services-proto](https://gitlab.meiyou.com/meiyou-services/grpc-services-proto)
- [grpc-services-core](https://gitlab.meiyou.com/meiyou-services/grpc-services-core)

### 使用说明

> 新项目需要引用子模块

1. 添加子模块
    ```shell
    git submodule add  git@gitlab.meiyou.com:meiyou-services/grpc-services-proto.git vendor/gitlab.meiyou.com/meiyou-services/grpc-services-proto.git
    git submodule add  git@gitlab.meiyou.com:meiyou-services/grpc-services-core.git vendor/gitlab.meiyou.com/meiyou-services/grpc-services-core.git
    ```

2. 加载子模块, 第一次clone项目默认不会加载子模块, 需手动加载
    ```shell
    git submodule update --init --recursive
    ```
3. 后期子模块更新
    ```shell
    git submodule sync --recursive
    git submodule update --remote
    ```

## 相关文档

- [gin框架](https://github.com/gin-gonic/gin)
- [grpc文档](https://grpc.io/docs/)
- [git流程发布教程文档](http://wiki.meiyou.com/pages/viewpage.action?pageId=6848559)
- [git子模块文档](https://git-scm.com/book/zh/v1/Git-%E5%B7%A5%E5%85%B7-%E5%AD%90%E6%A8%A1%E5%9D%97)
