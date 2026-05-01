# NoteMG

个人笔记管理系统 - B/S架构，Markdown编辑，类Typora体验

## 特性

- **单一可执行文件** - Go + 嵌入式前端，零依赖部署
- **Typora-like编辑体验** - Vditor IR即时渲染模式
- **混合存储架构** - SQLite元数据 + Markdown文件存储，支持Git版本管理
- **全文搜索** - Bleve索引，高性能模糊搜索
- **模块化设计** - 插件Hook机制，易于扩展
- **安全** - JWT认证，bcrypt密码哈希，CSRF保护
- **导入导出** - Markdown/HTML/ZIP格式支持

## 技术栈

| 层级 | 技术 |
|---|---|
| 后端 | Go 1.21+ |
| Web框架 | 自定义路由器（轻量） |
| 数据库 | SQLite (modernc.org/sqlite) |
| Markdown | goldmark + GFM + CJK |
| 搜索 | Bleve |
| 认证 | JWT |
| 前端 | Vite + TypeScript |
| 编辑器 | Vditor |

## 快速开始

### 构建

```bash
# 构建前端
cd frontend && npm install && npm run build

# 复制前端到嵌入目录
cp -r frontend/dist cmd/notemg/frontend/dist

# 构建二进制
go build -ldflags="-s -w" -o notemg ./cmd/notemg/
```

### 运行

```bash
# 初始化数据目录
./notemg init

# 启动服务
./notemg serve

# 指定端口
./notemg serve --port 9090

# 指定数据目录
./notemg serve --data ./mydata
```

首次访问 `http://localhost:8080` 会提示设置密码。

### 开发模式

```bash
# 后端热重载
go run ./cmd/notemg/ serve

# 前端开发服务器（端口5173）
cd frontend && npm run dev
```

## 目录结构

```
notemg/
├── cmd/notemg/main.go        # 入口
├── internal/
│   ├── config/               # 配置管理
│   ├── handler/              # HTTP处理器
│   ├── httputil/             # HTTP工具
│   ├── markdown/             # Markdown渲染
│   ├── model/                # 数据模型
│   ├── plugin/               # 插件系统
│   ├── search/               # 全文搜索
│   ├── security/             # 认证/安全
│   ├── server/               # 服务器/路由
│   └── store/                # 数据存储
├── frontend/                 # 前端源码
│   ├── src/
│   │   ├── api/              # API客户端
│   │   ├── components/       # UI组件
│   │   ├── editor/           # 编辑器配置
│   │   ├── views/            # 页面视图
│   │   └── styles/           # CSS样式
│   └── dist/                 # 构建产物
├── migrations/               # 数据库迁移
├── configs/                  # 配置文件
└── build/                    # 构建输出
```

## 数据存储

```
data/
├── notemg.db                 # SQLite元数据库
├── notes/                    # Markdown文件
│   ├── abc123.md
│   └── def456.md
├── attachments/              # 图片/附件
└── index/                    # 搜索索引
```

## API接口

### 认证

```
POST /api/auth/init          # 首次设置密码
POST /api/auth/login         # 登录
POST /api/auth/refresh       # 刷新Token
PUT  /api/auth/password      # 修改密码
GET  /api/auth/status        # 检查初始化状态
```

### 笔记

```
GET    /api/notes            # 列表
POST   /api/notes            # 创建
GET    /api/notes/:id        # 获取
PUT    /api/notes/:id        # 更新
DELETE /api/notes/:id        # 删除（软删除）
POST   /api/notes/:id/move   # 移动到文件夹
POST   /api/notes/:id/duplicate # 复制
```

### 文件夹

```
GET    /api/folders          # 树形列表
POST   /api/folders          # 创建
PUT    /api/folders/:id      # 更新
DELETE /api/folders/:id      # 删除
```

### 标签

```
GET    /api/tags             # 列表
POST   /api/tags             # 创建
DELETE /api/tags/:id         # 删除
```

### 搜索

```
GET /api/search?q=keyword    # 全文搜索
```

### 导入导出

```
POST /api/import/markdown    # 导入Markdown
POST /api/import/zip         # 导入ZIP
GET  /api/export/notes/:id   # 导出单个笔记
POST /api/export/batch       # 批量导出ZIP
```

## 配置

`configs/config.yaml`:

```yaml
server:
  host: "0.0.0.0"
  port: 8080

data:
  dir: "./data"

auth:
  jwt_secret: "your-secret-key"
  token_expire: "72h"

editor:
  autosave_interval: 1000
  default_mode: "ir"    # ir/wysiwyg/sv
```

## 插件系统

```go
type Plugin interface {
    Info() PluginInfo
    Init(app App) error
    Hooks() map[Hook]HookFunc
    Destroy() error
}

// 可用Hook
HookBeforeSave
HookAfterSave
HookBeforeDelete
HookAfterDelete
HookRender
```

## 安全

- JWT Token认证（HttpOnly Cookie + Header双模式）
- bcrypt密码哈希
- 登录失败锁定（5次/15分钟）
- CSRF保护
- 路径遍历防护

## 许可证

MIT