package migration

import (
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"go.uber.org/zap"
)

type Config struct {
	MigrationsPath string
	DatabaseURL    string
	Logger         *zap.Logger
}

// RunMigrations 执行数据库迁移
func RunMigrations(cfg Config) error {
	if cfg.Logger == nil {
		logger, _ := zap.NewProduction()
		cfg.Logger = logger
	}

	cfg.Logger.Info("开始数据库迁移",
		zap.String("migrations_path", cfg.MigrationsPath),
	)

	m, err := migrate.New(
		fmt.Sprintf("file://%s", cfg.MigrationsPath),
		cfg.DatabaseURL,
	)
	if err != nil {
		return fmt.Errorf("创建迁移实例失败: %w", err)
	}
	defer m.Close()

	// 获取当前版本
	version, dirty, err := m.Version()
	if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
		return fmt.Errorf("获取数据库版本失败: %w", err)
	}

	if dirty {
		cfg.Logger.Warn("数据库处于dirty状态，尝试强制到当前版本",
			zap.Uint("version", version),
		)
		// 强制设置版本并清除dirty状态
		if err := m.Force(int(version)); err != nil {
			return fmt.Errorf("清除dirty状态失败: %w", err)
		}
	}

	cfg.Logger.Info("当前数据库版本", zap.Uint("version", version))

	// 执行迁移
	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("执行迁移失败: %w", err)
	}

	// 获取最新版本
	newVersion, _, err := m.Version()
	if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
		return fmt.Errorf("获取新版本失败: %w", err)
	}

	if errors.Is(err, migrate.ErrNoChange) {
		cfg.Logger.Info("数据库已是最新版本", zap.Uint("version", version))
	} else {
		cfg.Logger.Info("数据库迁移完成",
			zap.Uint("old_version", version),
			zap.Uint("new_version", newVersion),
		)
	}

	return nil
}

// MigrateDown 回滚指定步数的迁移
func MigrateDown(cfg Config, steps int) error {
	if cfg.Logger == nil {
		logger, _ := zap.NewProduction()
		cfg.Logger = logger
	}

	m, err := migrate.New(
		fmt.Sprintf("file://%s", cfg.MigrationsPath),
		cfg.DatabaseURL,
	)
	if err != nil {
		return fmt.Errorf("创建迁移实例失败: %w", err)
	}
	defer m.Close()

	version, _, _ := m.Version()
	cfg.Logger.Info("开始回滚迁移",
		zap.Uint("current_version", version),
		zap.Int("steps", steps),
	)

	if err := m.Steps(-steps); err != nil {
		return fmt.Errorf("回滚迁移失败: %w", err)
	}

	newVersion, _, _ := m.Version()
	cfg.Logger.Info("迁移回滚完成",
		zap.Uint("old_version", version),
		zap.Uint("new_version", newVersion),
	)

	return nil
}

// MigrateTo 迁移到指定版本
func MigrateTo(cfg Config, version uint) error {
	if cfg.Logger == nil {
		logger, _ := zap.NewProduction()
		cfg.Logger = logger
	}

	m, err := migrate.New(
		fmt.Sprintf("file://%s", cfg.MigrationsPath),
		cfg.DatabaseURL,
	)
	if err != nil {
		return fmt.Errorf("创建迁移实例失败: %w", err)
	}
	defer m.Close()

	currentVersion, _, _ := m.Version()
	cfg.Logger.Info("迁移到指定版本",
		zap.Uint("current_version", currentVersion),
		zap.Uint("target_version", version),
	)

	if err := m.Migrate(version); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("迁移到版本 %d 失败: %w", version, err)
	}

	cfg.Logger.Info("迁移完成", zap.Uint("version", version))
	return nil
}

// Reset 重置数据库（删除所有表）
func Reset(cfg Config) error {
	if cfg.Logger == nil {
		logger, _ := zap.NewProduction()
		cfg.Logger = logger
	}

	cfg.Logger.Warn("开始重置数据库（将删除所有数据）")

	m, err := migrate.New(
		fmt.Sprintf("file://%s", cfg.MigrationsPath),
		cfg.DatabaseURL,
	)
	if err != nil {
		return fmt.Errorf("创建迁移实例失败: %w", err)
	}
	defer m.Close()

	if err := m.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("重置数据库失败: %w", err)
	}

	cfg.Logger.Info("数据库重置完成")
	return nil
}
