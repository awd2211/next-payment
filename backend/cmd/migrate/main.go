package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/payment-platform/pkg/migration"
)

func main() {
	// 定义命令行参数
	var (
		dbURL          = flag.String("database", getEnv("DATABASE_URL", ""), "数据库连接字符串")
		migrationsPath = flag.String("path", getEnv("MIGRATIONS_PATH", "file://./migrations"), "迁移文件路径")
		command        = flag.String("command", "", "命令: up, down, steps, version, force, status")
		steps          = flag.Int("steps", 0, "迁移步骤数（用于 steps 命令）")
		version        = flag.Int("version", 0, "目标版本（用于 force 命令）")
	)
	flag.Parse()

	// 检查必需参数
	if *dbURL == "" {
		log.Fatal("错误: 必须提供数据库连接字符串（-database 或 DATABASE_URL 环境变量）")
	}

	if *command == "" {
		printUsage()
		os.Exit(1)
	}

	// 创建迁移管理器
	migrator, err := migration.NewMigrator(*dbURL, *migrationsPath)
	if err != nil {
		log.Fatalf("创建迁移管理器失败: %v", err)
	}
	defer migrator.Close()

	// 执行命令
	switch *command {
	case "up":
		log.Println("执行所有待执行的 up 迁移...")
		if err := migrator.Up(); err != nil {
			log.Fatalf("迁移失败: %v", err)
		}
		log.Println("✅ 迁移完成")

	case "down":
		log.Println("回滚所有迁移...")
		if err := migrator.Down(); err != nil {
			log.Fatalf("回滚失败: %v", err)
		}
		log.Println("✅ 回滚完成")

	case "steps":
		if *steps == 0 {
			log.Fatal("错误: steps 命令需要提供 -steps 参数")
		}
		direction := "向上"
		if *steps < 0 {
			direction = "向下"
		}
		log.Printf("执行 %s 迁移 %d 步...", direction, abs(*steps))
		if err := migrator.Steps(*steps); err != nil {
			log.Fatalf("迁移失败: %v", err)
		}
		log.Println("✅ 迁移完成")

	case "version":
		version, dirty, err := migrator.Version()
		if err != nil {
			log.Fatalf("获取版本失败: %v", err)
		}
		fmt.Printf("当前版本: %d\n", version)
		if dirty {
			fmt.Println("⚠️  数据库处于脏状态，需要修复")
		}

	case "force":
		if *version < 0 {
			log.Fatal("错误: force 命令需要提供有效的 -version 参数")
		}
		log.Printf("强制设置版本为 %d...", *version)
		if err := migrator.Force(*version); err != nil {
			log.Fatalf("强制设置版本失败: %v", err)
		}
		log.Println("✅ 版本设置完成")

	case "status":
		status, err := migrator.Status()
		if err != nil {
			log.Fatalf("获取状态失败: %v", err)
		}
		fmt.Println(status)

	default:
		log.Fatalf("未知命令: %s", *command)
	}
}

func printUsage() {
	fmt.Println(`
数据库迁移工具

用法:
  migrate -database <url> -path <path> -command <cmd> [选项]

命令:
  up              执行所有待执行的 up 迁移
  down            回滚所有迁移
  steps           执行指定数量的迁移步骤（需要 -steps 参数）
  version         显示当前数据库版本
  force           强制设置数据库版本（需要 -version 参数，用于修复脏状态）
  status          显示迁移状态

选项:
  -database       数据库连接字符串（或使用 DATABASE_URL 环境变量）
  -path           迁移文件路径（默认: file://./migrations）
  -steps          迁移步骤数（正数向上，负数向下）
  -version        目标版本号

示例:
  # 执行所有待执行的迁移
  migrate -database "postgres://user:pass@localhost/db?sslmode=disable" -command up

  # 回滚最后一个迁移
  migrate -database $DATABASE_URL -command steps -steps -1

  # 查看当前版本
  migrate -database $DATABASE_URL -command version

  # 修复脏状态（将版本设置为 1）
  migrate -database $DATABASE_URL -command force -version 1

环境变量:
  DATABASE_URL       数据库连接字符串
  MIGRATIONS_PATH    迁移文件路径
`)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
