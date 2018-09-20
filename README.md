# Migration Go

[![Build Status](https://travis-ci.org/fengzifz/migration-go.svg?branch=master)](https://travis-ci.org/fengzifz/migration-go)

A very simple migration tool like Laravel migration.

Support commands:
- **make**: create a migration file
- **migrate**: migrate your database
- **rollback <step?>**: rollback your database, `step` default is `1`
- **refresh**: rollback all migrations and re-migrate

## Directory Structure
```
├── database
│   ├── migrations          # Migration files
│   └── stubs               # Migration templates
|        ├── blank.stub
|        └── create.stub
├── migrate.go
└── .env                    # DB information
```

## Installation
### 1. Add `.env` file.

Use `.env` file to save your database's information
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

### 2. Copy `migrate.go` to your project directory directly.

### 3. Dependence

```
go get github.com/go-sql-driver/mysql
go get github.com/joho/godotenv
```

or use `dep` to manage your packages.

### 4. Usage

You can build `migrate.go`, and add to the $PATH, in order to use `migrate <command>` command directly. Otherwise you should run `go run migrate.go <command>` instead.

#### Create

Use `make` to create a migration file.

It will create a directory in `database/migrations/`, and has two sql files `up.sql` and `down.sql`:

```
go run migrate.go make create_user_table

# database/migration/20180914180229_create_user_table/up.sql
# database/migration/20180914180229_create_user_table/down.sql
```

If the migration file's name has the prefix `create_` and the suffix `_table`, it will create a migration file copied from `create.stub`, otherwise from `blank.stub`.

#### Migrate

Use `migrate` to migrate your database.

```
go run migrate.go migrate
```

#### Rollback

Use `rollback step?` to rollback your database. `step?` means it is optional, default is `1` step.

```
# Rollback one step
go run migrate.go rollback
# or rollback two steps
go run migrate.go rollback 2
```

#### Refresh

Rollback all migrations and re-migrate

```
go run migrate.go refresh
```