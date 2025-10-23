package migration

import (
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// Migrator 数据库迁移管理器
type Migrator struct {
	m *migrate.Migrate
}

// NewMigrator 创建迁移管理器
// dbURL: 数据库连接字符串，例如 "postgres://user:pass@localhost:5432/db?sslmode=disable"
// migrationsPath: 迁移文件目录路径，例如 "file://./migrations"
func NewMigrator(dbURL, migrationsPath string) (*Migrator, error) {
	m, err := migrate.New(migrationsPath, dbURL)
	if err != nil {
		return nil, fmt.Errorf("创建迁移实例失败: %w", err)
	}

	return &Migrator{m: m}, nil
}

// Up 执行所有待执行的 up 迁移
func (mg *Migrator) Up() error {
	if err := mg.m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("执行 up 迁移失败: %w", err)
	}
	log.Println("数据库迁移完成")
	return nil
}

// Down 回滚所有迁移
func (mg *Migrator) Down() error {
	if err := mg.m.Down(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("执行 down 迁移失败: %w", err)
	}
	log.Println("数据库回滚完成")
	return nil
}

// Steps 执行指定数量的迁移步骤
// n > 0: 向上迁移 n 步
// n < 0: 向下回滚 n 步
func (mg *Migrator) Steps(n int) error {
	if err := mg.m.Steps(n); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("执行迁移步骤失败: %w", err)
	}
	log.Printf("迁移 %d 步完成", n)
	return nil
}

// Migrate 迁移到指定版本
func (mg *Migrator) Migrate(version uint) error {
	if err := mg.m.Migrate(version); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("迁移到版本 %d 失败: %w", version, err)
	}
	log.Printf("迁移到版本 %d 完成", version)
	return nil
}

// Version 获取当前数据库版本
func (mg *Migrator) Version() (uint, bool, error) {
	version, dirty, err := mg.m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return 0, false, fmt.Errorf("获取数据库版本失败: %w", err)
	}
	return version, dirty, nil
}

// Force 强制设置数据库版本（用于修复脏状态）
func (mg *Migrator) Force(version int) error {
	if err := mg.m.Force(version); err != nil {
		return fmt.Errorf("强制设置版本失败: %w", err)
	}
	log.Printf("强制设置版本为 %d", version)
	return nil
}

// Close 关闭迁移连接
func (mg *Migrator) Close() error {
	srcErr, dbErr := mg.m.Close()
	if srcErr != nil {
		return fmt.Errorf("关闭源失败: %w", srcErr)
	}
	if dbErr != nil {
		return fmt.Errorf("关闭数据库失败: %w", dbErr)
	}
	return nil
}

// Status 获取迁移状态
func (mg *Migrator) Status() (string, error) {
	version, dirty, err := mg.Version()
	if err != nil {
		return "", err
	}

	status := fmt.Sprintf("当前版本: %d", version)
	if dirty {
		status += " (脏状态，需要修复)"
	}
	return status, nil
}
