package gorm

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/pkg/errors"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
	"gorm.io/plugin/dbresolver"

	"github.com/kovercjm/tool-go/logger"
	"github.com/kovercjm/tool-go/repository"
)

type Repository struct {
	*gorm.DB
}

type gormKey struct{}

func (r *Repository) ToCtx(ctx context.Context, connection interface{}) context.Context {
	return context.WithValue(ctx, gormKey{}, connection)
}

func (r *Repository) Ctx(ctx context.Context) *Repository {
	if ctxDB, ok := ctx.Value(gormKey{}).(*gorm.DB); ok {
		return &Repository{ctxDB.WithContext(ctx)}
	}
	return &Repository{r.DB.WithContext(ctx)}
}

func New(config *repository.Config, dbLogger logger.Logger) (*Repository, error) {
	var (
		logLevel      gormLogger.LogLevel
		slowThreshold time.Duration
		gormDB        *gorm.DB
		err           error
	)
	switch strings.ToLower(config.DBLogLevel) {
	case "silent":
		logLevel = gormLogger.Silent
	case "error":
		logLevel = gormLogger.Error
	case "warn":
		logLevel = gormLogger.Warn
	default:
		logLevel = gormLogger.Info
	}
	if config.DBSlowThresholdMS > 0 {
		slowThreshold = time.Duration(config.DBSlowThresholdMS) * time.Millisecond
	}

	switch {
	case config.DBHost == "":
		gormDB, err = memorySQLiteInit(config, dbLogger, logLevel, slowThreshold)
	default:
		// TODO add support for other repository types
		gormDB, err = mySQLInit(config, dbLogger, logLevel, slowThreshold)
	}
	if err != nil {
		return nil, errors.Wrap(err, "gorm initialize failed")
	}
	if err = gormDB.Use(dbresolver.Register(dbresolver.Config{}).
		SetConnMaxLifetime(60 * time.Second).
		SetMaxIdleConns(config.DBConnPoolSize).
		SetMaxOpenConns(config.DBConnPoolSize),
	); err != nil {
		return nil, errors.Wrap(err, "gorm initialize dbresolver")
	}
	//if err = gormDB.Use(gormOpentracing.New()); err != nil {
	//	return nil, errors.Wrap(err, "gorm initialize tracing")
	//}
	return &Repository{gormDB}, nil
}

func mySQLInit(config *repository.Config, dbLogger logger.Logger, logLevel gormLogger.LogLevel, slowThreshold time.Duration) (*gorm.DB, error) {
	mysqlConfig := config.MySQL()
	dbName := mysqlConfig.DBName
	mysqlConfig.DBName = ""
	sqlDB, err := sql.Open("mysql", mysqlConfig.FormatDSN())
	if err != nil {
		return nil, errors.Wrap(err, "failed to establish database connection")
	}
	if _, err = sqlDB.Exec(fmt.Sprintf("create database if not exists `%s`;", config.DBName)); err != nil {
		return nil, errors.Wrap(err, "failed to init database")
	}
	mysqlConfig.DBName = dbName

	return gorm.Open(
		MySQLDialector{Dialector: mysql.Dialector{Config: &mysql.Config{DSN: config.MySQL().FormatDSN()}}},
		&gorm.Config{
			CreateBatchSize: 1000,
			Logger: Logger{
				l:                         dbLogger.NoCaller(),
				IgnoreRecordNotFoundError: config.DBIgnoreRecordNotFoundError,
				SlowThreshold:             slowThreshold,
			}.LogMode(logLevel),
			PrepareStmt:            true,
			SkipDefaultTransaction: true,
		},
	)
}

func memorySQLiteInit(config *repository.Config, dbLogger logger.Logger, logLevel gormLogger.LogLevel, slowThreshold time.Duration) (*gorm.DB, error) {
	return gorm.Open(
		sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", config.DBName)),
		&gorm.Config{
			CreateBatchSize: 1000,
			Logger: Logger{
				l:                         dbLogger.NoCaller(),
				IgnoreRecordNotFoundError: config.DBIgnoreRecordNotFoundError,
				SlowThreshold:             slowThreshold,
			}.LogMode(logLevel),
			PrepareStmt:            true,
			SkipDefaultTransaction: true,
		},
	)
}
