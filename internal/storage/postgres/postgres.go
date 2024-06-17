package postgres

import (
	"context"
	"database/sql"
	"db-index/config"
	"db-index/internal/domain"
	"fmt"
	"path/filepath"
	"time"

	"github.com/Masterminds/squirrel"
	_ "github.com/lib/pq"
	migrate "github.com/rubenv/sql-migrate"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type Soraged interface {
	AddRates(ctx context.Context,rates domain.Rates) error
	InitDB(config config.ConfigDB) error
}

type DB struct {
	db *sql.DB
	tracer trace.Tracer
	psql squirrel.StatementBuilderType
	logger *zap.Logger
}

func NewStorage(logger *zap.Logger) *DB {
	return &DB{
		logger: logger,
		psql: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		tracer: otel.Tracer("fiber-server"),
	}

}

func (d *DB) InitDB(configDB config.ConfigDB) error {
	d.logger.Info("Connecting to DB with config", 
	zap.String("host", configDB.Host),
	zap.String("port", configDB.Port),
	zap.String("user", configDB.User),
	zap.String("dbname", configDB.Database),
	zap.String("migrate Dir", configDB.MigrationsDir))

	var err error
    constr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
	configDB.Host, configDB.Port, configDB.User, configDB.Password, configDB.Database)
	retrayConnect := 10

	for i:= 0 ; i < retrayConnect ; i++ {
		d.db, err = sql.Open("postgres", constr)

		err = d.db.Ping()

		if err == nil {
			break
		}
		d.logger.Error("DB",zap.String("Ping to db",err.Error()))
		time.Sleep(time.Second*1)
	} 

    if err != nil {
		d.logger.Error("Conect to DB",zap.String("err",err.Error()))
        return err
    }

	d.logger.Info("DB",zap.String("Connect to db","Successful"))

	err = migrateDB(configDB.MigrationsDir,d.db) 
	if err != nil {
		d.logger.Error("migrate",zap.String("err",err.Error()))
		return err
	}
	d.logger.Info("DB",zap.String("migrate","Successful"))
	
    return nil
}

func migrateDB(path string,db *sql.DB) error {
	
	migrations := &migrate.FileMigrationSource{
		Dir: filepath.Join(path),
	}
	_, err := migrate.Exec(db, "postgres", migrations, migrate.Up)

	if err != nil {
		return err
	}
	
	return nil

}

func (d *DB) AddRates(ctx context.Context,rates domain.Rates) error {
	tracer := otel.Tracer("fiber-server")
	_, span := tracer.Start(ctx, "AddRates", trace.WithAttributes())

	span.SetAttributes(attribute.Key("param.rates.AskPrice").String(rates.AskPrice))
	defer span.End()

	query := d.psql.Insert("rates").
            Columns("timestamp", "ask_price", "bid_price").
            Values(rates.Timestamp, rates.AskPrice,rates.BidPrice)

        sqlQuery, args, err := query.ToSql()
		if err != nil {
			d.logger.Error("DB",zap.String("err ToSql",err.Error()))
            return err
        }

		_, err = d.db.Exec(sqlQuery, args...)
        if err != nil {
			d.logger.Error("DB",zap.String("err Exec",err.Error()))
            return err
        }
		d.logger.Info("DB",zap.String("RatesInsert","Successful"))
	
	return nil
}
