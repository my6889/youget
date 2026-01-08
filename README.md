### 项目简介
为了更方便更高效的使用you-get，可以用该工具批量下载在线视频，并在下载完成后自动提取音频文件。  

主要功能：
- 批量下载在线视频
- 支持传入Cookies
- 自动提取音频（方便获取高品质歌曲）

---

### 环境要求
- 已安装并可在命令行中直接运行：
  - `you-get`
  - `ffmpeg`

---

### 安装you-get与ffmpeg
- you-get安装方式：
  - 使用 Python/pip：`pip install you-get`
  - 验证：执行 `you-get -V`
- ffmpeg安装方式：
  - 从官网或可信镜像下载可执行压缩包，解压后将 `bin` 目录加入 PATH。
  - 验证：执行 `ffmpeg -version`

参考：
- you-get文档：`https://you-get.org/`
- ffmpeg下载：`https://ffmpeg.org/download.html`

---

### 快速开始
1. 将本项目放到任意目录。
2. 在 `download-list.txt` 中，每行写入一个视频页面的 URL（空行会被忽略）。
3. 如需登录态下载或更高质量：
   - 获取 `cookies.sqlite` 或 `cookies.txt` 放在程序同级目录。
   - 程序会优先使用 `cookies.sqlite`，否则使用 `cookies.txt`，若均不存在则匿名下载。
4. 运行程序（会先检查依赖是否可用，缺少 you-get/ffmpeg 会直接退出提示）：
   - Windows：运行 `youget_windows_x64.exe` 或在源代码目录执行 `go run main.go`
   - Linux：运行 `youget_linux_amd64` 或在源代码目录执行 `go run main.go`
5. 对每个 URL，程序会：
   - 调用 `you-get` 下载视频；若失败会打印错误并继续下一条。
   - 自动解析下载结果：先尝试解析 `.mp4` 名称；解析到的文件名会去除路径前缀并优先尝试其 `sanitize` 版本（处理非法文件名字符）是否存在，若存在则使用该文件；否则回退到原始解析文件名并检查其存在；如果仍未找到，则退化为使用 `title` 生成的 `*.mp4` 文件名。
   - 调用 `ffmpeg` 提取音频，使用 `-q:a 0 -map a` 生成与视频同名的 `.mp3` 文件（视频文件保留）。如果目标 `.mp3` 已存在，程序将跳过转换以避免覆盖。

---

### Cookies 配置
- 将 `cookies.sqlite` 或 `cookies.txt` 置于程序同级目录。
- 优先级：`cookies.sqlite` > `cookies.txt` > 无Cookies。
- 未检测到Cookies文件时，将以匿名方式下载，可能会导致清晰度或可下载性受限。

#### 从 Firefox 获取 cookies.sqlite（推荐）
为确保可下载需登录或更高清晰度的视频，建议直接使用 Firefox 的 `cookies.sqlite`：

1) 退出（完全关闭）Firefox
- 请先关闭所有 Firefox 窗口，避免 `cookies.sqlite` 处于占用状态，无法复制。

2) 找到当前登录网站所使用的 Firefox 配置目录（Profile）
- 方法一（最简单）：
  - 在 Firefox 地址栏输入：`about:support`
  - 在“应用程序基本信息”中找到“配置文件文件夹”（或“Profile Directory/Folder”）
  - 点击“打开文件夹”（Open Folder），进入该 Profile 目录
- 方法二（按系统路径定位）：
  - Windows：`%APPDATA%\Mozilla\Firefox\Profiles\<profile>.default-release` 等
  - macOS：`~/Library/Application Support/Firefox/Profiles/<profile>.default-release`
  - Linux：`~/.mozilla/firefox/<profile>.default-release`

3) 复制 `cookies.sqlite`
- 在上述 Profile 目录中，找到 `cookies.sqlite`，复制到本项目的程序同级目录（与 `main.go/youget.exe` 同级）。
- 若存在多个 Profile，请选择你已登录目标网站的那个（可根据最近修改时间判断，或在 `about:profiles` 查看当前使用的 Profile）。

4) 再次启动程序
- 程序会优先使用当前目录下的 `cookies.sqlite` 进行下载与解析。

注意事项：
- `cookies.sqlite` 中包含登录态敏感信息，请勿泄露或分享。
- 若网站登录状态过期，请在 Firefox 中重新登录后，重复上述步骤复制最新的 `cookies.sqlite`。
- 若你使用 Firefox Multi-Account Containers/多容器，请确保复制的是对应容器所使用的 Profile 的 `cookies.sqlite`。

---

### 常见问题（FAQ）
- 运行即退出并提示you-get/ffmpeg未安装？
  - 请按前文“安装 you-get 与 ffmpeg”完成安装，并确认已加入 PATH。
- 报错 “Get title failed: parse json failed ...”？
  - 某些站点的 `you-get --json` 输出会混杂日志，程序会尝试自动提取 JSON；若仍失败，请手动执行 `you-get --json <url>` 查看输出并反馈。
- 已下载视频但未生成MP3？
  - 确认 `ffmpeg` 可用；或检查下载的视频扩展名是否非常规，必要时在源码中补充扩展名列表。
  - 如果目标 `.mp3` 已存在，程序会跳过转换并打印提示信息。

---

### 从源码构建（可选）
前提：已安装 Go（建议 1.20+）。

在项目目录执行：
```
go build -o youget.exe
```
随后运行：
```
./youget.exe
```

#### 交叉编译示例（生成指定二进制名）
- 在 Bash (Linux/macOS)：
```
GOOS=linux GOARCH=amd64 go build -o youget_linux_amd64 .
GOOS=windows GOARCH=amd64 go build -o youget_windows_x64.exe .
```
- 在 PowerShell：
```
$env:GOOS = 'linux'; $env:GOARCH = 'amd64'; go build -o youget_linux_amd64 .
$env:GOOS = 'windows'; $env:GOARCH = 'amd64'; go build -o youget_windows_x64.exe .
```
生成的二进制示例：`youget_linux_amd64`, `youget_windows_x64.exe`。

---

### 目录与文件说明
- `main.go`：主程序，下载、解析元数据、转换 MP3 的流程。
- `download-list.txt`：待下载的 URL 列表，一行一个。
- `cookies.sqlite` / `cookies.txt`：可选，用于需要登录态或更高清晰度的下载。

---

### 注意事项
- 仅用于抓取你有权下载的内容，请遵守网站的服务条款与当地法律。
- 某些网站可能随时间变更接口，导致下载或解析失败，建议更新 `you-get` 至最新版本。