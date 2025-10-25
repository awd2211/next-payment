# 🚀 快速启动指南

## 一键启动 (5分钟)

### 1. 启动基础设施
```bash
docker-compose up -d
```

### 2. 初始化数据库
```bash
cd backend && ./scripts/init-db.sh
```

### 3. 启动后端服务
```bash
./scripts/start-all-services.sh
```

### 4. 启动前端
```bash
# Admin Portal
cd frontend/admin-portal && npm install && npm run dev

# Merchant Portal (新终端)
cd frontend/merchant-portal && npm install && npm run dev
```

### 5. 访问系统
- **Admin Portal**: http://localhost:5173
- **Merchant Portal**: http://localhost:5174
- **Grafana**: http://localhost:40300 (admin/admin)
- **Prometheus**: http://localhost:40090
- **Jaeger**: http://localhost:40686

---

## 详细文档

- 📖 [完整快速启动指南](QUICK_START_GUIDE.md)
- 📊 [项目状态报告](PROJECT_STATUS_REPORT.md)
- 💻 [前端完成总结](FRONTEND_COMPLETE_SUMMARY.md)
- 🔧 [开发指南](CLAUDE.md)

---

**准备就绪后,开始测试支付流程! 🎉**
