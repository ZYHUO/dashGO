# 项目重命名总结

## 更改日期
2024年12月13日

## 更改内容
将项目从 `xboard-go` 重命名为 `dashGO`

## 更改详情

### 1. 核心配置文件
- ✅ `go.mod` - 模块名从 `xboard` 改为 `dashgo`
- ✅ `web/package.json` - 包名从 `xboard-web` 改为 `dashgo-web`
- ✅ `web/index.html` - 标题从 `XBoard` 改为 `dashGO`
- ✅ `docker-compose.yaml` - 服务名和容器名更新
- ✅ `Makefile` - Docker 镜像名更新

### 2. Go 源代码
- ✅ 所有 Go 文件的 import 路径自动从 `xboard/` 更新为 `dashgo/`
- ✅ 包括 `cmd/`, `internal/`, `pkg/` 下的所有文件

### 3. 安装脚本
- ✅ `install.sh` - GitHub 仓库路径、安装目录、二进制文件名
- ✅ `setup.sh` - 所有配置和路径
- ✅ `troubleshoot.sh` - 容器名和日志路径

### 4. 文档文件
- ✅ `README.md` - 项目名称、仓库链接、命令示例
- ✅ `README_SETUP.md` - 安装指南
- ✅ `CHANGELOG.md` - 版本日志
- ✅ `.kiro/specs/traffic-stats-completion/COMPLETION_SUMMARY.md`

### 5. 前端代码
- ✅ `web/src/views/Login.vue` - 默认站点名
- ✅ `web/src/views/Register.vue` - 默认站点名
- ✅ `web/src/layouts/MainLayout.vue` - 默认站点名
- ✅ `web/src/views/admin/Settings.vue` - 占位符文本
- ✅ `web/src/views/admin/Hosts.vue` - Agent 安装命令

### 6. 数据库和配置
- ✅ 数据库名从 `xboard` 改为 `dashgo`
- ✅ SQLite 文件从 `xboard.db` 改为 `dashgo.db`
- ✅ Docker 密码从 `xboard_password` 改为 `dashgo_password`

### 7. 二进制文件名
- `xboard-server` → `dashgo-server`
- `xboard-agent` → `dashgo-agent`
- `xboard-mysql` → `dashgo-mysql`
- `xboard-redis` → `dashgo-redis`

### 8. GitHub 仓库
- 旧地址: `https://github.com/ZYHUO/xboard-go`
- 新地址: `https://github.com/ZYHUO/dashGO`

## 命名规则

| 类型 | 旧名称 | 新名称 |
|------|--------|--------|
| 仓库名 | xboard-go | dashGO |
| Go 模块 | xboard | dashgo |
| 显示名称 | XBoard | dashGO |
| 包名 | xboard-web | dashgo-web |
| 二进制 | xboard-* | dashgo-* |
| 容器 | xboard-* | dashgo-* |
| 数据库 | xboard | dashgo |

## 需要手动处理的事项

### 1. GitHub 仓库重命名
在 GitHub 上将仓库从 `xboard-go` 重命名为 `dashGO`：
1. 进入仓库设置 (Settings)
2. 在 "Repository name" 中输入 `dashGO`
3. 点击 "Rename"

### 2. 本地 Git 远程地址更新
```bash
git remote set-url origin https://github.com/ZYHUO/dashGO.git
```

### 3. 重新编译
```bash
# Windows
.\build-all.ps1 -Target all -Version 1.0.0

# Linux/macOS
./build-all.sh all 1.0.0
```

### 4. 更新已部署的实例
如果有已经部署的实例，需要：
1. 备份数据
2. 停止服务
3. 更新二进制文件
4. 更新配置文件中的应用名称
5. 重启服务

### 5. 文档中的下载链接
部分文档中包含具体的下载链接（如 `docs/prebuilt-binaries.md`），这些链接指向 `download.sharon.wiki`，需要根据实际情况更新。

## 验证清单

- [ ] Go 代码编译通过: `go build ./...`
- [ ] 前端构建成功: `cd web && npm run build`
- [ ] Docker 镜像构建成功: `docker build -t dashgo .`
- [ ] 安装脚本测试通过
- [ ] 文档链接检查完成
- [ ] GitHub 仓库已重命名
- [ ] Git 远程地址已更新

## 注意事项

1. **向后兼容性**: 旧的 `xboard` 数据库可以继续使用，只需更新配置文件中的应用名称
2. **迁移路径**: 如果从旧版本升级，数据库迁移工具仍然兼容
3. **API 端点**: 所有 API 端点保持不变，只是应用名称改变
4. **配置文件**: 旧的配置文件需要更新 `app.name` 字段为 `dashGO`

## 完成状态

✅ 所有代码文件已更新
✅ 所有配置文件已更新
✅ 所有文档已更新
✅ 所有脚本已更新

项目重命名完成！
