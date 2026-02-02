<p align="center">
  <img src="https://github.com/Lvshujun0918/PaletteFlow/blob/main/frontend/src/assets/logo.png?raw=true" alt="PaletteFlow Logo" width="80" height="80" style="border-radius: 16px;"/>
  <h1 align="center">PaletteFlow</h1>
  <p align="center" style="color: #666; font-size: 16px;">配色，易如反掌</p>
</p>

PaletteFlow利用人工智能技术，快速生成美观、和谐的配色方案。输入简单描述，即可获得专业级调色板及色彩代码，支持导出多种格式。适合设计师、开发者和创意工作者使用。

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Go](https://img.shields.io/badge/go-1.25+-blue.svg)
![Vue](https://img.shields.io/badge/Vue-3.3+-green.svg)
![Docker](https://img.shields.io/badge/docker-ready-blue.svg)

## ✨ 核心功能

### 🤖 AI配色生成
- **广泛适配大预言模型**：集成 OpenAI/Claude/通义千问/DeepSeek 等多家AI服务
- **自然语言输入**：输入配色需求描述，AI自动生成5个协调配色
- **智能降级**：AI失败时自动降级到随机生成，确保服务可用性
- **快速模板**：8个预设配色主题，秒速生成

### 🎨 配色工具
- **对比度检查**：WCAG 2.0标准检查（本地计算）
  - 实时对比度比率计算
  - AA/AAA等级评分
  - 0-100可访问性评分
- **色盲模拟**：4种色盲模拟（本地矩阵变换）
  - Deuteranopia（红绿色盲）
  - Protanopia（红绿色弱）
  - Tritanopia（蓝黄色盲）
  - Achromatopsia（完全色盲）
- **颜色导出**：CSS变量 / JSON / PNG 三种格式

### 💾 历史管理
- **本地保存**：自动本地保存历史记录
- **自动备份**：生成新配色自动加入历史
- **快速恢复**：一键加载历史配色方案
- **最多20条**：自动删除超期记录

### 📱 用户体验
- **响应式设计**：桌面优先，适配平板和手机
- **实时通知**：操作反馈通过 Toast 通知
- **光滑动画**：卡片悬停和过渡效果

## 🏗️ 技术栈

### 后端
```
Go 1.25                    # 编程语言
├── Gin                    # Web框架
├── godotenv              # 环境变量管理
└── OpenAI兼容API         # 真实AI调用
```

### 前端
```
Vue 3.3                   # UI框架
├── Vite 4.5             # 构建工具
├── Axios                # HTTP客户端
└── CSS3                 # 响应式设计 + 动画
```

### 基础设施
```
Nginx 1.29               # 反向代理 + 静态服务
Supervisor              # 进程管理
Docker 20.10+          # 容器化部署
```

## 🚀 快速开始

### 方案1：Docker（推荐）

#### 要求
- Docker 20.10+

#### 配置

编辑 `backend/.env`：
```env
AI_API_KEY=sk-xxxxxxxxxxxx          # OpenAI或兼容API的密钥
AI_API_BASE_URL=https://api.openai.com/v1
AI_MODEL=gpt-3.5-turbo              # 模型名称
AI_TIMEOUT=30                       # 请求超时（秒）
```

#### 运行

```bash
# 构建镜像（前后端一体，单容器）
docker build -t ai-color-palette .

# 运行容器（同端口访问前后端）
docker run -p 5173:80 --env-file backend/.env ai-color-palette

# 访问应用
open http://localhost:5173
```

### 方案2：本地开发

#### 要求
- Go 1.25+
- Node.js 18+
- npm/yarn

#### 后端启动

```bash
cd backend

# 复制环境变量
cp .env.example .env
# 编辑 .env，填入真实 AI_API_KEY

# 安装依赖
go mod download

# 启动服务
go run main.go
# 后端运行在 http://localhost:8080
```

#### 前端启动

```bash
cd frontend

# 安装依赖
npm install

# 开发模式
npm run dev
# 应用运行在 http://localhost:5173
```

## 📚 API 文档

### 健康检查
**GET** `/api/health`

```bash
curl http://localhost:8080/api/health
```

### 生成配色
**POST** `/api/generate-palette`

请求：
```bash
curl -X POST http://localhost:8080/api/generate-palette \
  -H "Content-Type: application/json" \
  -d '{"prompt": "温暖的秋日色调"}'
```

响应：
```json
{
  "colors": ["#D97706", "#F59E0B", "#FBBF24", "#FCD34D", "#FEF3C7"],
  "timestamp": 1704067200,
  "description": "根据提示词生成的配色方案"
}
```

## 🔧 环境变量配置

支持多家AI服务提供商，通过 `AI_API_BASE_URL` 和 `AI_MODEL` 切换。

### OpenAI
```env
AI_API_KEY=sk-xxxxx
AI_API_BASE_URL=https://api.openai.com/v1
AI_MODEL=gpt-3.5-turbo
```

### Anthropic Claude
```env
AI_API_KEY=sk-ant-xxxxx
AI_API_BASE_URL=https://api.anthropic.com
AI_MODEL=claude-3-sonnet
```

### 通义千问
```env
AI_API_KEY=sk-xxxxx
AI_API_BASE_URL=https://dashscope.aliyuncs.com/compatible-mode/v1
AI_MODEL=qwen-turbo
```

### DeepSeek
```env
AI_API_KEY=sk-xxxxx
AI_API_BASE_URL=https://api.deepseek.com/v1
AI_MODEL=deepseek-chat
```

### 智谱 AI
```env
AI_API_KEY=xxxxx
AI_API_BASE_URL=https://open.bigmodel.cn/api/paas/v4
AI_MODEL=glm-4
```


## 📦 项目结构

```
ai-color-palette/
├── Dockerfile                    # 单容器前后端一体构建
├── supervisor.conf               # Supervisor 进程管理配置
├── .env                         # 环境变量（Git忽略）
├── .gitignore
├── backend/
│   ├── main.go                  # 入口点
│   ├── go.mod / go.sum         # Go依赖
│   ├── .env.example             # 环境变量模板
│   ├── config/
│   │   └── config.go            # 配置加载器
│   ├── ai/
│   │   └── client.go            # AI API客户端
│   ├── handler/
│   │   └── palette.go           # 配色生成处理器
│   └── .dockerignore
├── frontend/
│   ├── package.json             # npm依赖
│   ├── vite.config.js          # Vite配置
│   ├── nginx.conf               # Nginx反向代理配置
│   ├── .dockerignore
│   ├── dist/                    # 生产构建输出
│   └── src/
│       ├── App.vue              # 主组件
│       ├── components/          # 组件库
│       │   ├── ColorDisplay.vue
│       │   ├── ColorExport.vue
│       │   ├── ContrastCheck.vue
│       │   ├── ColorBlindCheck.vue
│       │   └── History.vue
│       └── utils/
│           ├── api.js           # Axios配置
│           └── colorUtils.js    # 颜色算法库
└── AI_INTEGRATION.md             # AI集成详解文档
```

## 🎯 常见问题

### ❓ AI生成失败怎么办？

系统会自动降级到随机生成。检查以下项：
- `backend/.env` 中 `AI_API_KEY` 是否配置
- 网络连接是否正常
- API额度是否充足
- `AI_API_BASE_URL` 是否正确

查看后端日志获取详细错误信息。

### ❓ 对比度检查在前端还是后端？

✅ **完全在前端本地计算**，无需网络请求，响应快，隐私更好。

使用 WCAG 2.0 标准算法：
- 先将 RGB 转换为 sRGB 线性值
- 计算相对亮度
- 计算对比度比率

### ❓ 历史记录在哪里？

保存在浏览器 **localStorage** 中，键值为 `ai_color_palette_history`，最多20条记录。

- 刷新页面自动恢复
- 清除浏览器缓存会丢失
- 支持跨标签页共享（同源）

### ❓ 如何部署到生产？

Docker 方式（推荐）：
```bash
# 构建镜像
docker build -t ai-color-palette:v1.0 .

# 运行容器，暴露到80端口
docker run -d -p 80:80 \
  --env-file backend/.env \
  --name ai-palette \
  ai-color-palette:v1.0

# 访问应用
curl http://localhost
```

K8s 部署：
```bash
# 将 docker run 转换为 Pod 资源
docker to kubernetes using kompose
```

### ❓ 如何限制历史记录大小？

修改 `frontend/src/App.vue`：
```javascript
// 第138行，修改最大记录数
const MAX_HISTORY = 30  // 改为你需要的数量
```

### ❓ 支持离线使用吗？

可以，使用本地AI模型：
1. 安装 [Ollama](https://ollama.ai)
2. 拉取模型：`ollama pull mistral`
3. 配置 `.env`：
   ```env
   AI_API_BASE_URL=http://localhost:11434/v1
   AI_MODEL=mistral
   ```

## 📄 许可证

MIT License - 详见 [LICENSE](./LICENSE) 文件

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！
