package postgres

import (
	"context"
	"database/sql"
	"db-index/config"
	"db-index/internal/domain"
	"db-index/pkg/logster"
	"fmt"
	"path/filepath"
	"time"

	"github.com/Masterminds/squirrel"
	_ "github.com/lib/pq"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/uptrace/opentelemetry-go-extra/otelsql"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

var tracer = otel.Tracer("fiber-server")

type Storaged interface {
	AddRates(ctx context.Context,rates domain.Rates) error
	InitDB(ctx context.Context,configDB config.ConfigDB) error
	Stop()
}

type DB struct {
	db *sql.DB
	psql squirrel.StatementBuilderType
	tracer trace.Tracer
	logger logster.Factory
}

func NewStorage(logger logster.Factory,tp trace.Tracer,) *DB {
	return &DB{
		logger: logger,
		tracer: tp,
		psql: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}

}

func (d *DB) InitDB(ctx context.Context,configDB config.ConfigDB) error {
	d.logger.Bg().Info("Connecting to DB with config", 
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
		d.db, err = otelsql.Open("postgres", constr,
		otelsql.WithAttributes(semconv.DBSystemPostgreSQL),
		otelsql.WithDBName(configDB.Database))

		err = d.db.Ping()

		if err == nil {
			break
		}

		time.Sleep(time.Second*1)
	}

    if err != nil {
		d.logger.For(ctx).Error("Conect to DB",zap.String("err",err.Error()))
        return err
    }

	d.logger.For(ctx).Info("DB",zap.String("Connect to db","Successful"))

	err = migrateDB(configDB.MigrationsDir,d.db) 
	if err != nil {
		d.logger.For(ctx).Error("migrate",zap.String("err",err.Error()))
		return err
	}
	d.logger.For(ctx).Info("DB",zap.String("migrate","Successful"))
	
    return nil
}

func (d DB)Stop(){
	d.db.Close()
}

func migrateDB(path string, db *sql.DB) error {
	
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
	ctx, span := d.tracer.Start(ctx, "AddRates", trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(attribute.Key("rates.AskPrice").String(rates.AskPrice))
	defer span.End()

	query := d.psql.Insert("rates").
            Columns("timestamp", "ask_price", "bid_price").
            Values(rates.Timestamp, rates.AskPrice,rates.BidPrice)

        sqlQuery, args, err := query.ToSql()
		if err != nil {
			d.logger.For(ctx).Error("DB",zap.String("err ToSql",err.Error()))
            return err
        }

		_ = d.db.QueryRowContext(ctx,sqlQuery,args...)
        
		d.logger.Bg().Info("DB",zap.String("RatesInsert","Successful"))
	
	return nil
}
