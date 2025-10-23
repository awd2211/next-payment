# 数据库迁移指南

## 📋 目录

- [工具介绍](#工具介绍)
- [安装](#安装)
- [迁移文件结构](#迁移文件结构)
- [使用方法](#使用方法)
- [常见场景](#常见场景)
- [最佳实践](#最佳实践)
- [故障排除](#故障排除)

---

## 工具介绍

我们使用 **golang-migrate** 作为数据库迁移工具，它提供：

- ✅ **版本控制** - 清晰的版本号管理
- ✅ **Up/Down 迁移** - 支持向上迁移和回滚
- ✅ **脏状态检测** - 自动检测未完成的迁移
- ✅ **多数据库支持** - PostgreSQL, MySQL, SQLite 等
- ✅ **CLI 和 Go 代码集成** - 灵活使用

## 安装

### 方法 1：使用包管理器

**macOS:**
```bash
brew install golang-migrate
```

**Linux:**
```bash
curl -L https://github.com/golang-migrate/migrate/releases/latest/download/migrate.linux-amd64.tar.gz | tar xvz
sudo mv migrate /usr/local/bin/
```

**Windows:**
```bash
scoop install migrate
```

### 方法 2：使用 Go

```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

### 验证安装

```bash
migrate -version
```

---

## 迁移文件结构

迁移文件采用 **顺序编号 + 描述** 的命名格式：

```
migrations/
├── 000001_create_notifications.up.sql
├── 000001_create_notifications.down.sql
├── 000002_create_templates.up.sql
├── 000002_create_templates.down.sql
├── 000003_create_webhooks.up.sql
├── 000003_create_webhooks.down.sql
├── 000004_insert_system_templates.up.sql
└── 000004_insert_system_templates.down.sql
```

### 文件命名规则

- **版本号**: 6位数字，例如 `000001`, `000002`
- **描述**: 简短的英文描述，使用下划线分隔
- **方向**: `.up.sql` 或 `.down.sql`

### 示例

**Up 迁移** (`000001_create_users.up.sql`):
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(100),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
```

**Down 迁移** (`000001_create_users.down.sql`):
```sql
DROP TABLE IF EXISTS users;
```

---

## 使用方法

### 方法 1：使用脚本（推荐）

我们提供了便捷的脚本 `scripts/migrate.sh`：

```bash
# 执行所有待执行的迁移
./scripts/migrate.sh up

# 回滚最后一个迁移
./scripts/migrate.sh steps -1

# 查看当前版本
./scripts/migrate.sh version

# 创建新的迁移文件
./scripts/migrate.sh create add_users_table
```

### 方法 2：直接使用 CLI

```bash
# 设置环境变量
export DATABASE_URL="postgres://user:pass@localhost:5432/db?sslmode=disable"

# 执行迁移
migrate -database $DATABASE_URL -path file://./migrations up

# 回滚
migrate -database $DATABASE_URL -path file://./migrations down
```

### 方法 3：在 Go 代码中使用

```go
import "github.com/payment-platform/pkg/migration"

func main() {
    dbURL := "postgres://user:pass@localhost:5432/db?sslmode=disable"
    migrationsPath := "file://./migrations"

    migrator, err := migration.NewMigrator(dbURL, migrationsPath)
    if err != nil {
        log.Fatal(err)
    }
    defer migrator.Close()

    // 执行迁移
    if err := migrator.Up(); err != nil {
        log.Fatal(err)
    }
}
```

---

## 常见场景

### 1. 初次部署（执行所有迁移）

```bash
./scripts/migrate.sh up
```

**输出示例:**
```
[INFO] 执行所有待执行的 up 迁移...
000001/u create_notifications (123.45ms)
000002/u create_templates (89.12ms)
000003/u create_webhooks (156.78ms)
000004/u insert_system_templates (45.23ms)
[INFO] ✅ 迁移完成
```

### 2. 回滚最后一个迁移

```bash
./scripts/migrate.sh steps -1
```

**使用场景**: 刚执行的迁移有问题，需要立即回滚

### 3. 查看当前数据库版本

```bash
./scripts/migrate.sh version
```

**输出示例:**
```
4
```

### 4. 创建新的迁移文件

```bash
./scripts/migrate.sh create add_user_roles
```

**生成的文件:**
```
migrations/000005_add_user_roles.up.sql
migrations/000005_add_user_roles.down.sql
```

### 5. 迁移到特定版本

```bash
# 迁移到版本 3
./scripts/migrate.sh goto 3
```

**使用场景**: 需要精确控制数据库版本

### 6. 修复脏状态

如果迁移过程中断（如数据库连接中断），可能会处于"脏状态"：

```bash
# 查看状态
./scripts/migrate.sh version
# 输出: 3 (dirty)

# 修复：强制设置为版本 3
./scripts/migrate.sh force 3

# 然后重新执行迁移
./scripts/migrate.sh up
```

---

## 最佳实践

### 1. 迁移文件应该是幂等的

**错误示例:**
```sql
-- ❌ 不幂等
CREATE TABLE users (...);
```

**正确示例:**
```sql
-- ✅ 幂等
CREATE TABLE IF NOT EXISTS users (...);
```

### 2. 总是提供 Down 迁移

即使不打算回滚，也应该编写 down 迁移：

```sql
-- up
ALTER TABLE users ADD COLUMN phone VARCHAR(20);

-- down
ALTER TABLE users DROP COLUMN phone;
```

### 3. 一个迁移文件只做一件事

**错误示例:**
```sql
-- ❌ 在一个文件中创建多个不相关的表
CREATE TABLE users (...);
CREATE TABLE products (...);
CREATE TABLE orders (...);
```

**正确示例:**
```sql
-- ✅ 每个表单独一个迁移文件
-- 000001_create_users.up.sql
CREATE TABLE users (...);

-- 000002_create_products.up.sql
CREATE TABLE products (...);
```

### 4. 使用事务（PostgreSQL）

```sql
BEGIN;

CREATE TABLE users (...);
CREATE INDEX idx_users_email ON users(email);

COMMIT;
```

### 5. 数据迁移要谨慎

对于大表的数据迁移，应该：
- 分批处理
- 添加超时控制
- 考虑停机窗口

```sql
-- ✅ 分批更新
UPDATE users SET status = 'active' WHERE status IS NULL LIMIT 10000;
-- 重复执行直到所有数据迁移完成
```

### 6. 在开发环境先测试

```bash
# 开发环境测试
export DATABASE_URL="postgres://localhost:5432/dev_db"
./scripts/migrate.sh up

# 验证无误后再部署到生产
```

### 7. 备份生产数据库

```bash
# 执行迁移前先备份
pg_dump -Fc payment_platform > backup_$(date +%Y%m%d_%H%M%S).dump

# 执行迁移
./scripts/migrate.sh up
```

---

## 故障排除

### 问题 1: 脏状态（Dirty State）

**症状:**
```bash
$ ./scripts/migrate.sh version
3 (dirty)
```

**原因**: 迁移过程中断（数据库连接断开、SQL 错误等）

**解决方案:**
```bash
# 1. 查看是哪个版本处于脏状态
./scripts/migrate.sh version

# 2. 检查数据库，确认迁移是否部分完成
psql -d payment_platform -c "SELECT * FROM schema_migrations;"

# 3. 根据情况选择：
# 选项 A: 如果迁移已完成，强制标记为完成
./scripts/migrate.sh force 3

# 选项 B: 如果迁移未完成，回滚到上一个版本
./scripts/migrate.sh force 2

# 4. 重新执行迁移
./scripts/migrate.sh up
```

### 问题 2: no change 错误

**症状:**
```bash
error: no change
```

**原因**: 没有待执行的迁移

**解决方案**: 这是正常的，表示数据库已经是最新版本

### 问题 3: schema_migrations 表不存在

**症状:**
```bash
error: relation "schema_migrations" does not exist
```

**原因**: 首次运行迁移工具

**解决方案**: migrate 会自动创建这个表，无需手动创建

### 问题 4: 版本冲突

**症状:**
```bash
error: Dirty database version 3. Fix and force version.
```

**原因**: 两个开发者同时创建了相同版本号的迁移

**解决方案**:
```bash
# 1. 重命名较新的迁移文件
mv 000003_feature_b.up.sql 000005_feature_b.up.sql
mv 000003_feature_b.down.sql 000005_feature_b.down.sql

# 2. 修复脏状态
./scripts/migrate.sh force 2

# 3. 重新执行迁移
./scripts/migrate.sh up
```

### 问题 5: SQL 语法错误

**症状:**
```bash
error: migration failed: syntax error at or near "CREAT"
```

**解决方案**:
```bash
# 1. 修复 SQL 文件中的语法错误

# 2. 如果迁移已部分执行，回滚
./scripts/migrate.sh steps -1

# 3. 重新执行
./scripts/migrate.sh up
```

---

## CI/CD 集成

### GitHub Actions 示例

```yaml
name: Database Migration

on:
  push:
    branches: [ main ]

jobs:
  migrate:
    runs-on: ubuntu-latest

    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: postgres
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

    steps:
      - uses: actions/checkout@v3

      - name: Install migrate
        run: |
          curl -L https://github.com/golang-migrate/migrate/releases/latest/download/migrate.linux-amd64.tar.gz | tar xvz
          sudo mv migrate /usr/local/bin/

      - name: Run migrations
        env:
          DATABASE_URL: postgres://postgres:postgres@localhost:5432/test?sslmode=disable
        run: |
          cd backend/services/notification-service
          migrate -database $DATABASE_URL -path file://./migrations up
```

---

## 参考资料

- [golang-migrate 官方文档](https://github.com/golang-migrate/migrate)
- [PostgreSQL 文档](https://www.postgresql.org/docs/)
- [数据库迁移最佳实践](https://www.prisma.io/dataguide/types/relational/what-are-database-migrations)

---

## 总结

使用 golang-migrate 工具可以：
- ✅ 版本化管理数据库结构
- ✅ 安全地执行迁移和回滚
- ✅ 团队协作时避免冲突
- ✅ 自动化部署流程

记住：**始终在开发环境测试迁移，备份生产数据，编写幂等的 SQL 语句**。
