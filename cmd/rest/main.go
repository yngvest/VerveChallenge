package main

import (
	"context"
	"database/sql"
	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/yngvest/VerveChallenge/server"
	"log"
	"os"
	"time"
)

type Config struct {
	Listen   string
	Database []string
}

func main() {
	cmd := &cobra.Command{
		Use:   "rest",
		Short: "Verve rest service",
		RunE:  run,
	}
	cmd.PersistentFlags().StringP("config", "c", "", "configuration `file` to load")
	cmd.Execute()
}

func run(cmd *cobra.Command, args []string) error {
	cfg, err := getConfig(cmd)
	if err != nil {
		return err
	}
	logger := log.New(os.Stdout, "rest", log.LstdFlags)
	db := clickhouseDb(cfg.Database)
	srv, err := server.NewServer(cfg.Listen, db, logger)
	if err != nil {
		return err
	}
	ctx := context.Background()
	go srv.Serve(ctx)
	<-srv.Done()
	return nil
}

func getConfig(cmd *cobra.Command) (*Config, error) {
	configFn, err := cmd.Flags().GetString("config")
	if err != nil {
		return nil, err
	}
	viper.SetConfigFile(configFn)
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}
	var cfg Config
	err = viper.Unmarshal(&cfg)
	return &cfg, err
}

func clickhouseDb(addr []string) *sql.DB {
	options := clickhouse.Options{
		Addr: addr,
		Auth: clickhouse.Auth{
			Database: "default",
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 15,
		},
		DialTimeout: 5 * time.Second,
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
	}
	db := clickhouse.OpenDB(&options)
	db.SetConnMaxLifetime(30 * time.Second)
	db.SetMaxIdleConns(64)
	db.SetMaxOpenConns(64)
	return db
}
