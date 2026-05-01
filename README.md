# NoteMG / 笔记管理系统

[English](#english) | [中文](#chinese)

---

<a name="english"></a>
## English

### Personal Note Management System

A modern, self-hosted note management application with Typora-like editing experience.

#### Features

- **Single Binary Deployment** - Zero dependencies, embedded frontend
- **Typora-like Editor** - Vditor IR (Instant Rendering) mode
- **Hybrid Storage** - SQLite metadata + Markdown files (Git-friendly)
- **Full-text Search** - Bleve index, fast fuzzy search
- **Modular Design** - Plugin hook system, easy to extend
- **Secure** - JWT auth, bcrypt hashing, CSRF protection
- **Import/Export** - Markdown, HTML, ZIP formats

#### Quick Start

```bash
# Build frontend
cd frontend && npm install && npm run build

# Copy frontend for embedding
cp -r frontend/dist cmd/notemg/frontend/dist

# Build binary
go build -ldflags="-s -w" -o notemg ./cmd/notemg/

# Run
./notemg init    # Initialize data directory
./notemg serve   # Start server (default: http://localhost:8080)
```

First visit will prompt to set a password.

#### Tech Stack

| Layer | Technology |
|---|---|
| Backend | Go 1.21+ |
| Database | SQLite (modernc.org/sqlite) |
| Markdown | goldmark + GFM + CJK |
| Search | Bleve |
| Auth | JWT |
| Frontend | Vite + TypeScript |
| Editor | Vditor |

#### API Endpoints

| Method | Path | Description |
|---|---|---|
| POST | /api/auth/init | Set initial password |
| POST | /api/auth/login | Login |
| GET | /api/auth/status | Check initialization |
| GET | /api/notes | List notes |
| POST | /api/notes | Create note |
| GET | /api/notes/:id | Get note |
| PUT | /api/notes/:id | Update note |
| DELETE | /api/notes/:id | Delete note |
| GET | /api/folders | List folders tree |
| GET | /api/tags | List tags |
| GET | /api/search?q= | Search notes |

#### Configuration

`configs/config.yaml`:

```yaml
server:
  port: 8080

data:
  dir: "./data"

auth:
  jwt_secret: "change-me"
```

---

<a name="chinese"></a>
## 中文

### 个人笔记管理系统

现代化、自托管的笔记管理应用，提供类 Typora 的编辑体验。

#### 特性

- **单一可执行文件** - Go + 嵌入式前端，零依赖部署
- **类 Typora 编辑器** - Vditor IR 即时渲染模式
- **混合存储架构** - SQLite 元数据 + Markdown 文件，支持 Git 版本管理
- **全文搜索** - Bleve 索引，高性能模糊搜索
- **模块化设计** - 插件 Hook 机制，易于扩展
- **安全可靠** - JWT 认证，bcrypt 密码哈希，CSRF 保护
- **导入导出** - 支持 Markdown、HTML、ZIP 格式

#### 快速开始

```bash
# 构建前端
cd frontend && npm install && npm run build

# 复制前端到嵌入目录
cp -r frontend/dist cmd/notemg/frontend/dist

# 构建二进制
go build -ldflags="-s -w" -o notemg ./cmd/notemg/

# 运行
./notemg init    # 初始化数据目录
./notemg serve   # 启动服务 (默认: http://localhost:8080)
```

首次访问会提示设置密码。

#### 技术栈

| 层级 | 技术 |
|---|---|
| 后端 | Go 1.21+ |
| 数据库 | SQLite (modernc.org/sqlite) |
| Markdown | goldmark + GFM + CJK |
| 搜索 | Bleve |
| 认证 | JWT |
| 前端 | Vite + TypeScript |
| 编辑器 | Vditor |

#### API 接口

| 方法 | 路径 | 说明 |
|---|---|---|
| POST | /api/auth/init | 设置初始密码 |
| POST | /api/auth/login | 登录 |
| GET | /api/auth/status | 检查初始化状态 |
| GET | /api/notes | 笔记列表 |
| POST | /api/notes | 创建笔记 |
| GET | /api/notes/:id | 获取笔记 |
| PUT | /api/notes/:id | 更新笔记 |
| DELETE | /api/notes/:id | 删除笔记 |
| GET | /api/folders | 文件夹树形列表 |
| GET | /api/tags | 标签列表 |
| GET | /api/search?q= | 搜索笔记 |

#### 配置

`configs/config.yaml`:

```yaml
server:
  port: 8080

data:
  dir: "./data"

auth:
  jwt_secret: "请修改为随机密钥"
```

#### 目录结构

```
notemg/
├── cmd/notemg/           # 入口程序
├── internal/             # 后端代码
│   ├── handler/          # HTTP 处理器
│   ├── store/            # 数据存储
│   ├── security/         # 认证安全
│   └── markdown/         # Markdown 渲染
├── frontend/             # 前端源码
├── configs/              # 配置文件
└── build/                # 构建输出
```

#### 数据存储

```
data/
├── notemg.db             # SQLite 元数据库
├── notes/                # Markdown 文件
├── attachments/          # 图片附件
└── index/                # 搜索索引
```

---

## License / 许可证

MIT