package dao

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/btcsuite/btclog"
	gormMysqlDriver "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"net"
	"os"
	"syscall"
	"time"

	inscLog "github.com/inscription-c/insc/inscription/log"
	_ "github.com/pingcap/tidb/pkg/extension/_import"
)

// DB is a struct that embeds gorm.DB to provide additional database functionality.
type DB struct {
	*gorm.DB
}

// DBOptions is a struct that holds the configuration options for the database.
type DBOptions struct {
	addr             string
	user             string
	password         string
	dbName           string
	noEmbedDB        bool
	dataDir          string
	serverPort       string
	serverStatusPort string
	startHeight      uint32

	log               btclog.Logger
	autoMigrateTables []interface{}
}

// DBOption is a function type that modifies DBOptions.
type DBOption func(*DBOptions)

// WithAddr returns a DBOption that sets the address of the database.
func WithAddr(addr string) DBOption {
	return func(o *DBOptions) {
		o.addr = addr
	}
}

// WithUser returns a DBOption that sets the user of the database.
func WithUser(user string) DBOption {
	return func(o *DBOptions) {
		o.user = user
	}
}

// WithPassword returns a DBOption that sets the password of the database.
func WithPassword(password string) DBOption {
	return func(o *DBOptions) {
		o.password = password
	}
}

// WithDBName returns a DBOption that sets the name of the database.
func WithDBName(dbName string) DBOption {
	return func(o *DBOptions) {
		o.dbName = dbName
	}
}

// WithLogger returns a DBOption that sets the logger of the database.
func WithLogger(log btclog.Logger) DBOption {
	return func(o *DBOptions) {
		o.log = log
	}
}

// WithAutoMigrateTables returns a DBOption that sets the tables to be auto migrated in the database.
func WithAutoMigrateTables(tables ...interface{}) DBOption {
	return func(o *DBOptions) {
		o.autoMigrateTables = tables
	}
}

// WithNoEmbedDB returns a DBOption that sets whether to use an embedded database or not.
func WithNoEmbedDB(noEmbed bool) DBOption {
	return func(options *DBOptions) {
		options.noEmbedDB = noEmbed
	}
}

// WithDataDir returns a DBOption that sets the data directory of the database.
func WithDataDir(dir string) DBOption {
	return func(o *DBOptions) {
		o.dataDir = dir
	}
}

// WithServerPort returns a DBOption that sets the server port of the database.
func WithServerPort(port string) DBOption {
	return func(o *DBOptions) {
		o.serverPort = port
	}
}

// WithStatusPort returns a DBOption that sets the status port of the database.
func WithStatusPort(port string) DBOption {
	return func(o *DBOptions) {
		o.serverStatusPort = port
	}
}

// Transaction is a method on DB that executes a function within a database transaction.
func (d *DB) Transaction(fn func(tx *DB) error) error {
	return d.DB.Transaction(func(tx *gorm.DB) error {
		d := &DB{DB: tx}
		return fn(d)
	})
}

// NewDB is a function that creates a new DB instance with the provided options.
func NewDB(opts ...DBOption) (*DB, error) {
	options := &DBOptions{}
	for _, opt := range opts {
		opt(options)
	}

	conn := "%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local"
	dsn := fmt.Sprintf(conn, options.user, options.password, options.addr, "")

	var err error
	var db *gorm.DB

	if !options.noEmbedDB {
		go TIDB(options)
		timeout := time.After(time.Second * 30)
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for range ticker.C {
			select {
			case <-timeout:
				return nil, fmt.Errorf("gorm open timeout")
			default:
				db, err = gorm.Open(gormMysqlDriver.Open(dsn), &gorm.Config{Logger: logger.Discard})
				if err != nil {
					var opErr *net.OpError
					if errors.As(err, &opErr) {
						var syscallErr *os.SyscallError
						if errors.As(opErr.Err, &syscallErr) &&
							errors.Is(syscallErr.Err, syscall.ECONNREFUSED) {
							continue
						}
					}
					return nil, err
				}
			}
			break
		}
	}

	createDb := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s;", options.dbName)
	if err = db.Exec(createDb).Error; err != nil {
		return nil, fmt.Errorf("gorm create database :%v", err)
	}

	dsn = fmt.Sprintf(conn, options.user, options.password, options.addr, options.dbName)
	db, err = gorm.Open(gormMysqlDriver.Open(dsn), &gorm.Config{Logger: &GormLogger{Logger: inscLog.Gorm}})
	if err != nil {
		return nil, fmt.Errorf("gorm open :%v", err)
	}
	db = db.Debug()
	if err := db.AutoMigrate(options.autoMigrateTables...); err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("gorm db :%v", err)
	}
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetMaxIdleConns(50)

	return &DB{
		DB: db,
	}, nil
}

// GormLogger is a struct that embeds btclog.Logger to provide additional logging functionality.
type GormLogger struct {
	btclog.Logger
}

// LogMode is a method on GormLogger that sets the log level.
func (g *GormLogger) LogMode(level logger.LogLevel) logger.Interface {
	switch level {
	case logger.Silent:
		g.Logger.SetLevel(btclog.LevelOff)
	case logger.Error:
		g.Logger.SetLevel(btclog.LevelError)
	case logger.Warn:
		g.Logger.SetLevel(btclog.LevelWarn)
	case logger.Info:
		g.Logger.SetLevel(btclog.LevelInfo)
	}
	return g
}

// Info is a method on GormLogger that logs an informational message.
func (g *GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	g.Logger.Info(append([]interface{}{msg}, data...))
}

// Warn is a method on GormLogger that logs a warning message.
func (g *GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	g.Logger.Warn(append([]interface{}{msg}, data...))
}

// Error is a method on GormLogger that logs an error message.
func (g *GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	g.Logger.Error(append([]interface{}{msg}, data...))
}

// Trace is a method on GormLogger that logs a trace message.
func (g *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin).Milliseconds()
	sql, rows := fc()
	sqlInfo := struct {
		Elapsed interface{}
		Rows    interface{}
		Err     error
		Sql     string
	}{
		Elapsed: elapsed,
		Rows:    rows,
		Sql:     sql,
	}
	if err != nil && err.Error() != "" {
		sqlInfo.Err = err
		sqlInfoByte, _ := json.Marshal(sqlInfo)
		g.Logger.Trace(string(sqlInfoByte))
	} else {
		sqlInfoByte, _ := json.Marshal(sqlInfo)
		g.Logger.Trace(string(sqlInfoByte))
	}
}
