# Migration written in Go

[![Build Status](https://travis-ci.org/fengzifz/migration-go.svg?branch=master)](https://travis-ci.org/fengzifz/migration-go)

像 Laravel migration 那样管理数据库

## 目前支持的命令：
- **make:migration**: 创建 migration
- **make:seeder**: 创建 seeder
- **up**: 升级数据库
- **down <step?>**: 回滚，默认回滚 1 步，`<step?>` 是可选参数，直接填数字
- **refresh**: 重新运行所有的 migration
- **db:seed**: 运行 seeder

## 支持的数据库：
- MySQL

**日后会支持更多的类型的**

## 第三方依赖：
- github.com/joho/godotenv
- github.com/go-sql-driver/mysql
- github.com/fatih/color

## 目录结构
```
├── database
│   ├── migrations          # Migration files
├── migrate.go
└── .env                    # DB information
```

## 安装
1. 直接把 `migrate.go` 拷贝到项目的根目录，并根据你自己的项目所使用的包管理工具，下载欠缺的依赖；
2. 在根目录添加 `.env` 文件，并按照下面模板填入你的数据库信息：
```
APP_NAME=MigrationExample

DB_CONNECTION=mysql
DB_HOST=127.0.0.1
DB_PORT=3306
DB_DATABASE=
DB_USERNAME=
DB_PASSWORD=
DB_CHARSET=utf8
DB_PARSETIME=True
DB_LOC=Local
```

## 使用
1. 创建 migration
```
go run migrate.go make:migration create_user_table

# 然后会在 database/migration 下面创建对应的 migration，并生成一个 up.sql 和 一个 down.sql
# database/migration/20180914180229_create_user_table/up.sql
# database/migration/20180914180229_create_user_table/down.sql
```

2. 升级 migration
```
go run migrate.go up
```

3. 回滚
```
go run migrate.go down # 默认回滚 1 步

go run migrate.go down 2 # 回滚 2 步

```

4. 重新运行所有 migration
```
go run migrate.go refresh
```

5. 创建 seeder
```
go run migrate.go make:seeder user_seeder
```

6. 运行 seeder
```
go run migrate.go db:seed user_seeder
```