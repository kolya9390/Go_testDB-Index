package postgres

import (
	"database/sql"
	"db-index/config"
	"db-index/internal/domain"
	"fmt"
	"path/filepath"

	"github.com/Masterminds/squirrel"
	migrate "github.com/rubenv/sql-migrate"
	"go.uber.org/zap"
)

type Soraged interface {
	Add(rates domain.Rates) error
	InitDB(config config.ConfigDB) error
	MigrateDB() error
}

type DB struct {
	db *sql.DB
	psql squirrel.StatementBuilderType
	logger *zap.Logger
}

func NewStorage(logger *zap.Logger) *DB {

	return &DB{logger: logger}

}

func (d *DB) InitDB(configDB config.ConfigDB) error {
    connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
	configDB.Host, configDB.Port, configDB.User, configDB.Password, configDB.Database)

    db, err := sql.Open("postgres", connStr)

    if err != nil {
		d.logger.Error("Conect to DB",zap.String("err",err.Error()))
        return err
    }

	d.db = db
    return nil
}

func (d *DB) MigrateDB() error {
	
	migrations := &migrate.FileMigrationSource{
		Dir: filepath.Join("migrations"),
	}
	n, err := migrate.Exec(d.db, "postgres", migrations, migrate.Up)

	if err != nil {
		d.logger.Error("migrate",zap.String("err",err.Error()))
		return err
	}

	d.logger.Info("Migrate",zap.String("Number of migrations",fmt.Sprintf("Successful %d migrations",n)))
	return nil

}

func (d *DB) Add(rates domain.Rates) error {
	query := d.psql.Insert("rates").
            Columns("timestamp", "ask_price", "bid_price").
            Values(rates.Timestamp.Abs(), rates.AskPrice,rates.BidPrice)

        sqlQuery, args, err := query.ToSql()

		_, err = d.db.Exec(sqlQuery, args...)
        if err != nil {
			d.logger.Error("Insert",zap.String("err",err.Error()))
            return err
        }
		d.logger.Info("DB",zap.String("RatesInsert","Successful"))
	
	return nil
}
