# CquptFunAnnihilator

CquptFunAnnihilator 是一个针对 EduForge 在线学习平台的自动化答题工具，能够自动获取题目列表、调用 AI 生成答案并提交。

## 功能特性

- 自动从学习看板获取题目列表
- 调用 AI（通义千问）生成编程题答案
- 自动提交代码到平台
- 支持批量答题模式
- 详细的日志记录

## 快速开始

### 方式一：直接使用 EXE 文件（推荐）

如果你不想编译代码，可以直接从项目的 Release 页面下载已经编译好的 EXE 文件：

1. **访问项目 Release 页面**
   - 打开项目主页
   - 点击右侧的 **"Releases"** 标签
   - 选择最新的版本

2. **下载 EXE 文件**
   - 在最新版本的 Assets 列表中，找到 `CquptFunAnnihilator.exe` 文件
   - 点击下载

3. **准备配置文件**
   - 在 EXE 文件所在目录创建 `config.yaml` 文件
   - 复制下面的配置内容并填入你的 API Key 和 Cookie

4. **运行程序**
   - 双击 `CquptFunAnnihilator.exe` 或在命令行中运行
   - 使用命令行参数指定操作（见下方使用说明）

### 方式二：从源码编译

如果你有 Go 开发环境，可以从源码编译：

**安装要求：**
- Go 1.21+

**编译步骤：**
```bash
# 克隆项目
git clone <项目地址>
cd CquptFunAnnihilator

# 编译
go build -o CquptFunAnnihilator.exe .

# 运行
.\CquptFunAnnihilator.exe -list
```

## 1. 获取阿里云 API Key

本工具使用阿里云通义千问 API 来生成答案，你需要先获取 API Key：

### 详细步骤：

#### 第一步：注册/登录阿里云账号
- 访问 [阿里云官网](https://www.aliyun.com/) 注册账号（如果已有账号可直接登录）
- 完成实名认证（首次使用需要）

#### 第二步：开通 DashScope 服务
- 访问 [DashScope 控制台](https://dashscope.console.aliyun.com/)
- 如果是首次使用，需要开通 DashScope 服务
- **好消息**：目前有免费额度，足够日常使用

#### 第三步：创建 API Key
1. 在控制台左侧菜单中找到 **"API-KEY 管理"**
2. 点击 **"创建 API-KEY"** 按钮
3. 输入一个便于识别的名称（如 "CquptFunAnnihilator"）
4. 点击确定后，系统会生成一个 API Key
5. **⚠️ 重要**：请立即复制并保存这个 API Key，它只会显示一次！

#### 第四步：查看 API Key
- 如果忘记保存，可以在 API-KEY 管理页面查看已创建的 Key
- 点击 Key 右侧的眼睛图标可以查看完整内容

## 2. 获取平台 Cookie（详细图文教程）

你需要从浏览器获取登录凭证 Cookie。以下是详细的步骤说明：

### 第一步：登录平台
1. 打开浏览器（推荐使用 Chrome 或 Edge）
2. 访问 `https://cqupt.fun`
3. 使用你的学号和密码登录

### 第二步：打开开发者工具
有两种方法可以打开开发者工具：

**方法一：快捷键**
- 按 `F12` 键

**方法二：右键菜单**
- 在页面空白处右键点击
- 选择 **"检查"** 或 **"Inspect"**

### 第三步：获取 Cookie（Chrome/Edge 浏览器）

1. **切换到 Application 标签**
   - 在开发者工具顶部，找到并点击 **"Application"** 标签

2. **展开 Cookies**
   - 在左侧菜单中，找到 **"Cookies"** 并展开它
   - 点击展开后的 `https://cqupt.fun`

3. **找到 ef_session**
   - 在右侧的 Cookie 列表中，找到名为 `ef_session` 的那一行
   - 这个 Cookie 就是你的登录凭证

4. **复制 Cookie 值**
   - 在 `ef_session` 那一行中，找到 **"Value"** 列
   - 双击 Value 列的内容，它会变成可编辑状态
   - 复制整个 Value 值（通常是一长串字符，如：`a1b2c3d4e5f6g7h8i9j0...`）
   - **注意**：只需要复制 Value 值，不要包含 `ef_session=` 这个名称

### 第三步：获取 Cookie（Firefox 浏览器）

1. **切换到存储标签**
   - 在开发者工具顶部，找到并点击 **"存储"** 或 **"Storage"** 标签

2. **展开 Cookies**
   - 在左侧菜单中，找到 **"Cookies"** 并展开它
   - 点击展开后的 `https://cqupt.fun`

3. **找到 ef_session**
   - 在右侧的 Cookie 列表中，找到名为 `ef_session` 的那一行

4. **复制 Cookie 值**
   - 点击 `ef_session` 那一行
   - 在下方详情中找到 **"值"** 或 **"Value"** 字段
   - 复制整个值

### 第四步：验证 Cookie

确保你复制的 Cookie 值：
- ✅ 是完整的（通常是一长串字符）
- ✅ 不包含空格或换行符
- ✅ 不包含 Cookie 名称（只需要 Value 值）
- ✅ 不是空的

### Cookie 格式示例：

**❌ 错误的格式：**
```
ef_session=a1b2c3d4e5f6g7h8i9j0
```
（包含了 Cookie 名称）

**✅ 正确的格式：**
```
a1b2c3d4e5f6g7h8i9j0
```
（只有 Value 值）

### 常见问题：

**Q: 找不到 ef_session 怎么办？**
A: 请确保：
- 已经成功登录平台
- 刷新一下页面后再查看
- 确认访问的是 `https://cqupt.fun` 而不是其他域名

**Q: Cookie 复制后无法使用？**
A: 请检查：
- 是否复制了完整的 Value 值
- 是否包含了多余的空格或换行符
- Cookie 是否已过期（重新登录即可获取新的）

**Q: Cookie 有有效期吗？**
A: 是的，Cookie 通常会在一段时间后失效。如果程序提示 Cookie 无效，按照上述步骤重新获取即可。

## 3. 配置配置文件

在 EXE 文件所在目录创建 `config.yaml` 文件，填入以下内容：

```yaml
platform:
  base_url: "https://cqupt.fun"
  timeout: 30

auth:
  cookie: "ef_session=你的_cookie值"

llm:
  api_key: "你的_dashscope_api_key"
  base_url: "https://dashscope.aliyuncs.com/compatible-mode/v1"
  model: "qwen-plus"
  max_tokens: 2000

solver:
  delay_ms: 300000
  max_retries: 3
  user_agent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"

logging:
  level: "info"
  file: "cqupt_fun.log"
```

### 配置说明：

| 配置项 | 说明 |
|--------|------|
| `platform.base_url` | 平台基础地址，通常不需要修改 |
| `platform.timeout` | HTTP 请求超时时间（秒） |
| `auth.cookie` | 从浏览器获取的完整 Cookie，格式为 `ef_session=你的值` |
| `llm.api_key` | 从阿里云 DashScope 获取的 API Key |
| `llm.base_url` | 通义千问 API 地址，通常不需要修改 |
| `llm.model` | 使用的模型，可选：`qwen-plus`、`qwen-turbo`、`qwen-max` |
| `llm.max_tokens` | 生成答案的最大 token 数 |
| `solver.delay_ms` | 提交间隔时间（毫秒），建议设置为较大值避免频繁请求 |
| `solver.max_retries` | 失败重试次数 |
| `logging.level` | 日志级别：`debug`、`info`、`warn`、`error` |
| `logging.file` | 日志文件路径 |

## 4. 运行程序

### 基本用法：

```bash
# 列出所有题目
CquptFunAnnihilator.exe -list

# 解决单个题目
CquptFunAnnihilator.exe -problem-id 1

# 解决所有题目
CquptFunAnnihilator.exe -solve-all
```

### 使用命令行参数（可选）：

你也可以直接通过命令行参数指定 Cookie，而不需要修改配置文件：

```bash
# 使用命令行参数指定 Cookie
CquptFunAnnihilator.exe -cookie "ef_session=你的cookie值" -list

# 同时指定配置文件和 Cookie（Cookie 参数优先级更高）
CquptFunAnnihilator.exe -config "my_config.yaml" -cookie "ef_session=你的cookie值" -problem-id 1
```

## 使用说明

### 命令行参数

| 参数 | 说明 |
|------|------|
| `-config <path>` | 指定配置文件路径（默认：config.yaml） |
| `-cookie <cookie>` | 直接指定认证 Cookie（优先级高于配置文件） |
| `-list` | 列出所有可用题目 |
| `-problem-id <id>` | 解决指定 ID 的题目 |
| `-solve-all` | 解决所有题目 |
| `-probe` | 探测 API 端点 |
| `-scrape` | 爬取前端 API 端点 |
| `-analyze` | 分析首页 API 线索 |
| `-dashboard` | 分析仪表盘数据 |
| `-parse-dashboard` | 解析仪表盘任务 |
| `-classroom-id <id>` | 指定教室 ID |
| `-task-id <id>` | 指定任务 ID |

## 项目结构

```
CquptFunAnnihilator/
├── analyzer/          # HTML 解析模块
│   ├── analyzer.go
│   ├── dashboard_parser.go
│   ├── html_parser.go
│   └── submission.go
├── auth/              # 认证模块
│   └── auth.go
├── client/            # HTTP 客户端
│   └── http.go
├── fetcher/           # 题目获取模块
│   └── fetcher.go
├── llm/               # LLM 客户端
│   └── client.go
├── logger/            # 日志模块
│   └── logger.go
├── models/            # 数据模型
│   └── models.go
├── probe/             # API 探测模块
│   └── probe.go
├── scraper/           # 前端爬取模块
│   └── scraper.go
├── solver/            # 解题模块
│   └── solver.go
├── tools/             # 工具脚本
├── config.yaml        # 配置文件（不要提交到仓库）
├── config.yaml.example # 配置示例
├── go.mod
├── go.sum
├── main.go
└── README.md
```

## 工作原理

1. **获取题目列表**：从 `/student/dashboard` 页面解析内嵌的 JSON 数据获取任务列表
2. **获取题目详情**：访问每个任务的详情页面获取完整题目描述
3. **生成答案**：调用通义千问 API 根据题目描述生成代码
4. **提交答案**：通过表单提交方式将代码提交到平台

## 常见问题

### Q: API Key 无效怎么办？
A: 请检查：
- API Key 是否正确复制（不要包含空格或换行）
- 是否已开通 DashScope 服务
- 账户是否有足够的免费额度
- 在 DashScope 控制台查看 API Key 是否正常

### Q: Cookie 失效怎么办？
A: Cookie 有有效期，失效后需要重新获取：
- 重新登录平台
- 按照上述详细步骤获取新的 Cookie
- 更新配置文件中的 `auth.cookie` 字段，或使用 `-cookie` 参数

### Q: 提交失败怎么办？
A: 可能的原因：
- Cookie 已失效，重新获取
- API Key 无效，检查配置
- 网络连接问题
- 平台服务器繁忙，可以增加 `solver.delay_ms` 延迟时间
- 查看日志文件 `cqupt_fun.log` 了解详细错误信息

### Q: 如何调整 AI 模型？
A: 在配置文件中修改 `llm.model` 字段：
- `qwen-turbo`：速度快，成本低
- `qwen-plus`：平衡性能和成本（推荐）
- `qwen-max`：性能最强，成本较高

### Q: EXE 文件无法运行？
A: 请检查：
- 是否被杀毒软件拦截（添加信任即可）
- 确保下载的是最新版本
- 检查 config.yaml 文件是否在正确位置

### Q: 可以在没有配置文件的情况下使用吗？
A: 可以！使用 `-cookie` 和 `-api-key` 命令行参数（如果支持），或者先创建配置文件。

## 注意事项

1. **请合理使用本工具**，遵守平台使用规则
2. **API Key 和 Cookie 属于敏感信息**，请妥善保管，不要泄露给他人
3. **config.yaml 文件不要提交到远程仓库**（已在 .gitignore 中配置）
4. 建议设置适当的延迟时间，避免对服务器造成压力
5. 本工具完全免费，请勿相信任何收费行为
6. 本工具仅用于学习和研究目的

## License

MIT License

## 作者

皖月清风

作者博客：https://pooiwoop-github-io.pages.dev/
