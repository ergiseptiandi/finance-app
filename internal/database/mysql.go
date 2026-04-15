package database

import (
	"context"
	"database/sql"
	"net"

	"finance-backend/internal/config"

	"github.com/go-sql-driver/mysql"
)

func OpenMySQL(ctx context.Context, cfg config.MySQLConfig) (*sql.DB, error) {
	return openMySQL(ctx, cfg, false)
}

func OpenMySQLForMigration(ctx context.Context, cfg config.MySQLConfig) (*sql.DB, error) {
	return openMySQL(ctx, cfg, true)
}

func openMySQL(ctx context.Context, cfg config.MySQLConfig, multiStatements bool) (*sql.DB, error) {
	db, err := sql.Open("mysql", newDriverConfig(cfg, multiStatements).FormatDSN())
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}

func newDriverConfig(cfg config.MySQLConfig, multiStatements bool) *mysql.Config {
	driverConfig := mysql.NewConfig()
	driverConfig.Net = "tcp"
	driverConfig.Addr = net.JoinHostPort(cfg.Host, cfg.Port)
	driverConfig.User = cfg.User
	driverConfig.Passwd = cfg.Password
	driverConfig.DBName = cfg.Name
	driverConfig.ParseTime = true
	driverConfig.MultiStatements = multiStatements
	return driverConfig
}
