package testdb

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"cyber-ecosystem/shared-go/orm/ent/client"
)

type Config struct {
	DB            client.DBConfig
	MigrationsDir string
}

var dbName = regexp.MustCompile(`^[a-z_][a-z0-9_]{0,62}$`)

func Prepare(ctx context.Context, cfg Config) (*client.EntClient, func() error, error) {
	if !dbName.MatchString(cfg.DB.DBName) {
		return nil, nil, fmt.Errorf("testdb: database name %q must match %s", cfg.DB.DBName, dbName)
	}

	adminCfg := cfg.DB
	adminCfg.DBName = "postgres"
	admin, err := client.NewEntClient(adminCfg)
	if err != nil {
		return nil, nil, fmt.Errorf("testdb: admin connection: %w", err)
	}
	fail := func(err error) (*client.EntClient, func() error, error) {
		_ = admin.DB.Close()
		return nil, nil, err
	}

	if _, err := admin.DB.ExecContext(ctx,
		fmt.Sprintf("DROP DATABASE IF EXISTS %s WITH (FORCE);", cfg.DB.DBName)); err != nil {
		return fail(fmt.Errorf("testdb: drop database %s: %w", cfg.DB.DBName, err))
	}
	if _, err := admin.DB.ExecContext(ctx,
		fmt.Sprintf("CREATE DATABASE %s;", cfg.DB.DBName)); err != nil {
		return fail(fmt.Errorf("testdb: create database %s: %w", cfg.DB.DBName, err))
	}

	data, err := client.NewEntClient(cfg.DB)
	if err != nil {
		return fail(fmt.Errorf("testdb: data connection: %w", err))
	}
	failData := func(err error) (*client.EntClient, func() error, error) {
		_ = data.DB.Close()
		return fail(err)
	}

	entries, err := os.ReadDir(cfg.MigrationsDir)
	if err != nil {
		return failData(fmt.Errorf("testdb: read migrations: %w", err))
	}
	played := 0
	for _, e := range entries {
		if filepath.Ext(e.Name()) != ".sql" {
			continue
		}
		sql, err := os.ReadFile(filepath.Join(cfg.MigrationsDir, e.Name()))
		if err != nil {
			return failData(fmt.Errorf("testdb: read %s: %w", e.Name(), err))
		}
		if _, err := data.DB.ExecContext(ctx, string(sql)); err != nil {
			return failData(fmt.Errorf("testdb: apply %s: %w", e.Name(), err))
		}
		played++
	}
	if played == 0 {
		return failData(fmt.Errorf("testdb: no .sql files under %s", cfg.MigrationsDir))
	}

	cleanup := func() error {
		err := data.DB.Close()
		if _, derr := admin.DB.ExecContext(context.Background(),
			fmt.Sprintf("DROP DATABASE IF EXISTS %s WITH (FORCE);", cfg.DB.DBName)); derr != nil {
			err = errors.Join(err, derr)
		}
		err = errors.Join(err, admin.DB.Close())
		if err != nil {
			return fmt.Errorf("testdb: cleanup: %w", err)
		}
		return nil
	}
	return data, cleanup, nil
}
