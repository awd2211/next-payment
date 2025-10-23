# 数据库迁移指南

本项目使用 [golang-migrate](https://github.com/golang-migrate/migrate) 进行数据库迁移管理。

## 📁 目录结构

```
backend/
├── services/
│   ├── admin-service/
│   │   └── migrations/
│   │       ├── 000001_init_schema.up.sql
│   │       ├── 000001_init_schema.down.sql
│   │       ├── 000002_seed_data.up.sql
│   │       └── 000002_seed_data.down.sql
│   ├── merchant-service/
│   │   └── migrations/
│   │       ├── 000001_init_schema.up.sql
│   │       └── 000001_init_schema.down.sql
│   └── ... (其他8个服务)
├── pkg/
│   └── migration/
│       └── migrate.go        # 迁移helper包
└── scripts/
    └── migrate.sh            # 迁移管理脚本
```

## 🚀 快速开始

### 1. 安装 golang-migrate CLI

```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

### 2. 执行迁移

```bash
# 迁移所有服务
./scripts/migrate.sh up all

# 迁移单个服务
./scripts/migrate.sh up admin-service

# 查看迁移状态
./scripts/migrate.sh status
```

## 📝 迁移脚本使用

### 基本命令

```bash
# 显示帮助
./scripts/migrate.sh help

# 执行迁移
./scripts/migrate.sh up all              # 所有服务
./scripts/migrate.sh up admin-service    # 单个服务

# 回滚迁移
./scripts/migrate.sh down admin-service 1   # 回滚1步
./scripts/migrate.sh down admin-service 2   # 回滚2步

# 重置数据库（危险操作！）
./scripts/migrate.sh reset admin-service    # 删除所有表

# 查看版本
./scripts/migrate.sh version all
./scripts/migrate.sh version admin-service

# 查看状态
./scripts/migrate.sh status

# 强制设置版本（修复dirty状态）
./scripts/migrate.sh force admin-service 1
```

### 环境变量

```bash
export DB_HOST=localhost
export DB_PORT=40432
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_SSL_MODE=disable
```

## 📋 服务和数据库映射

| 服务名称 | 数据库 | 表数量 |
|---------|-------|--------|
| admin-service | payment_admin | 47 |
| merchant-service | payment_merchant | 9 |
| payment-gateway | payment_gateway | 4 |
| order-service | payment_order | 4 |
| channel-adapter | payment_channel | 3 |
| risk-service | payment_risk | 3 |
| accounting-service | payment_accounting | 4 |
| notification-service | payment_notification | 4 |
| analytics-service | payment_analytics | 4 |
| config-service | payment_config | 4 |

## 🔧 在代码中集成迁移

### 使用 migration helper 包

```go
import (
    "payment-platform/pkg/migration"
    "go.uber.org/zap"
)

func main() {
    logger, _ := zap.NewProduction()
    
    // 执行迁移
    err := migration.RunMigrations(migration.Config{
        MigrationsPath: "./migrations",
        DatabaseURL:    "postgres://user:pass@localhost:5432/dbname?sslmode=disable",
        Logger:         logger,
    })
    if err != nil {
        logger.Fatal("迁移失败", zap.Error(err))
    }
    
    // 继续启动服务...
}
```

### 在服务启动时自动迁移

```go
func main() {
    // 加载配置
    dbURL := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
        dbUser, dbPassword, dbHost, dbPort, dbName, sslMode)
    
    // 执行迁移
    if err := migration.RunMigrations(migration.Config{
        MigrationsPath: "./migrations",
        DatabaseURL:    dbURL,
        Logger:         logger,
    }); err != nil {
        logger.Fatal("数据库迁移失败", zap.Error(err))
    }
    
    // 连接数据库
    db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
    // ...
}
```

## 📝 创建新的迁移

### 迁移文件命名规范

```
{version}_{description}.up.sql
{version}_{description}.down.sql
```

示例：
```
000003_add_user_roles.up.sql
000003_add_user_roles.down.sql
```

### UP 迁移示例

```sql
-- 000003_add_user_roles.up.sql
CREATE TABLE IF NOT EXISTS user_roles (
    user_id UUID NOT NULL,
    role_id UUID NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (user_id, role_id)
);

CREATE INDEX idx_user_roles_user_id ON user_roles(user_id);
```

### DOWN 迁移示例

```sql
-- 000003_add_user_roles.down.sql
DROP TABLE IF EXISTS user_roles;
```

## 🎯 最佳实践

### 1. 迁移文件原则

- ✅ 每个迁移应该是原子性的
- ✅ 总是提供 up 和 down 文件
- ✅ 使用 `IF EXISTS` 和 `IF NOT EXISTS`
- ✅ 先删除依赖（外键、索引），再删除表
- ✅ 使用事务（在需要的时候）

### 2. 版本号管理

- 使用递增的数字：`000001`, `000002`, `000003`
- 或使用时间戳：`20250123120000`

### 3. 数据迁移

对于包含数据的迁移：

```sql
-- UP
ALTER TABLE users ADD COLUMN new_field VARCHAR(50);
UPDATE users SET new_field = 'default' WHERE new_field IS NULL;

-- DOWN  
ALTER TABLE users DROP COLUMN new_field;
```

### 4. 安全检查

生产环境部署前：

```bash
# 1. 在开发环境测试
./scripts/migrate.sh up admin-service

# 2. 验证数据
# 连接数据库检查表结构和数据

# 3. 测试回滚
./scripts/migrate.sh down admin-service 1

# 4. 再次向上迁移
./scripts/migrate.sh up admin-service
```

## ⚠️ 常见问题

### Dirty 状态

如果迁移失败，数据库可能处于 "dirty" 状态：

```bash
# 查看当前版本
./scripts/migrate.sh version admin-service

# 如果显示 dirty，手动修复
./scripts/migrate.sh force admin-service <version>
```

### 回滚失败

如果回滚失败：

1. 检查 .down.sql 文件是否正确
2. 手动检查数据库状态
3. 必要时手动执行 SQL 清理

### 迁移版本冲突

多人协作时：

1. 在拉取代码后检查迁移版本
2. 如有冲突，重新编号或合并迁移
3. 与团队沟通迁移计划

## 🔍 调试

### 启用详细日志

```bash
export MIGRATE_VERBOSE=true
./scripts/migrate.sh up admin-service
```

### 检查迁移历史

```sql
SELECT * FROM schema_migrations ORDER BY version DESC;
```

### 手动执行迁移

```bash
migrate -path ./services/admin-service/migrations \
        -database "postgres://postgres:postgres@localhost:40432/payment_admin?sslmode=disable" \
        up
```

## 📚 参考资料

- [golang-migrate 官方文档](https://github.com/golang-migrate/migrate)
- [PostgreSQL 迁移最佳实践](https://www.postgresql.org/docs/current/)
- [数据库版本控制](https://martinfowler.com/articles/evodb.html)

## ✅ 检查清单

部署前确认：

- [ ] 所有迁移文件都有对应的 .up.sql 和 .down.sql
- [ ] 在开发环境测试过完整的 up/down 流程
- [ ] 备份生产数据库
- [ ] 准备好回滚方案
- [ ] 与团队沟通迁移时间窗口
- [ ] 设置数据库维护模式（如果需要）
- [ ] 执行迁移后验证数据完整性
