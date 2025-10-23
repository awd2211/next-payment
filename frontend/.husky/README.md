# Husky Git Hooks

本项目使用 Husky 和 lint-staged 自动格式化代码。

## Pre-commit Hook

提交代码前会自动执行以下检查和修复：

1. **TypeScript/TSX 文件**：
   - 运行 ESLint 自动修复
   - 运行 Prettier 格式化

2. **其他文件** (JSON, CSS, SCSS, MD)：
   - 运行 Prettier 格式化

## 配置

配置文件在 `package.json` 中的 `lint-staged` 字段：

```json
{
  "lint-staged": {
    "*.{ts,tsx}": [
      "eslint --fix",
      "prettier --write"
    ],
    "*.{json,css,scss,md}": [
      "prettier --write"
    ]
  }
}
```

## 跳过检查

如果紧急情况需要跳过检查：

```bash
git commit --no-verify -m "urgent fix"
```

**注意**：不建议经常跳过检查，这会降低代码质量。
