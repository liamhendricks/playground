module gorm.io/playground

go 1.16

require (
	bitbucket.org/clearlinkit/goat v0.0.6
	github.com/google/uuid v0.0.0-20161128191214-064e2069ce9c
	github.com/icrowley/fake v0.0.0-20180203215853-4178557ae428
	github.com/jackc/pgx/v4 v4.14.1 // indirect
	github.com/mattn/go-sqlite3 v2.0.3+incompatible // indirect
	github.com/pressly/goose v2.7.0+incompatible
	github.com/spf13/viper v1.0.0
	github.com/stretchr/testify v1.7.0
	golang.org/x/crypto v0.0.0-20220112180741-5e0467b6c7ce // indirect
	gorm.io/driver/mysql v1.2.3
	gorm.io/driver/postgres v1.2.3
	gorm.io/driver/sqlite v1.2.6
	gorm.io/driver/sqlserver v1.2.1
	gorm.io/gorm v1.22.5
)

replace gorm.io/gorm => ./gorm
