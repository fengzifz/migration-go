# Migration Go

A very simple migration tool like Laravel migration.

But it only support 3 commands now:
- **make**: create a migration file
- **migrate**: migrate your database
- **rollback <step?>**: rollback your database, `step` default is `1`

### Directory Structure
```
├── config
│   └── db.go               # DB connection
├── database
│   ├── migrations          # Migration files
│   └── stubs               # Migration templates
|        ├── blank.stub
|        └── create.stub
├── migrate.go
└── .env                    # DB information
```

### Installation
#### 1. Add `.env` file.

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

#### 2. Copy `config/` and `migrate.go` to your project directory.
It contains `db.go` file and `stubs/` directory.
- `db.go`: connect to database and return `*sql.DB`
- `stubs`: contains two migration templates

#### 3. Dependence

```
go get github.com/go-sql-driver/mysql
go get github.com/joho/godotenv
```

or use `dep` to manage your packages.

#### 4. Usage

You can build `migrate.go`, and add to the $PATH, in order to use `migrate <command>` command directly. Or you should run `go run migrate.go <command>`

##### Create

Use `make` to create a migration file.

It will create a file in `database/migrations/`:

```
go run migrate.go make create_user_table

# It will create a migration file in database/migration/20180914180229_create_user_table.go
```

If the migration file's name has the prefix `create_` and the suffix `_table`, it will create a migration file copied from `create.stub`, otherwise from `blank.stub`.

##### Migrate

Use `migrate` to migrate your database.

```
go run migrate.go migrate
```

##### Rollback

Use `rollback step?` to rollback your database. `step?` means it is optional, default is `1` step.

```
# Rollback one step
go run migrate.go rollback
# or rollback two steps
go run migrate.go rollback 2
```