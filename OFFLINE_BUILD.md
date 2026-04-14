# 内网离线构建指南

## 概述

本指南说明如何在内网隔离环境中使用 vendor 方式构建项目。

## 两种构建方式对比

| 特性 | 标准构建 (go mod) | 内网构建 (vendor) |
|------|-------------------|-------------------|
| 网络要求 | 需要互联网 | 完全离线 |
| 依赖管理 | 自动下载 | 使用 vendor 目录 |
| 适用场景 | GitHub Actions、外网开发 | 内网隔离环境 |
| Makefile 命令 | `make build` | `make vendor-build` |

## 准备工作（在外网环境执行）

### 步骤 1：克隆项目

```bash
git clone http://159.226.186.140:7028/oliver/ping-x.git
cd ping-x
```

### 步骤 2：准备 vendor 目录

```bash
# 下载所有依赖到 vendor 目录
go mod vendor

# 验证 vendor 目录
ls -la vendor/
```

### 步骤 3：打包项目

```bash
# 方式1：使用 tar
cd ..
tar czf ping-x-offline.tar.gz ping-x/

# 方式2：使用 zip
zip -r ping-x-offline.zip ping-x/
```

### 步骤 4：传输到内网

通过以下方式将压缩包传输到内网环境：
- U盘拷贝
- 内网文件传输工具
- 其他安全传输方式

## 内网构建步骤

### 步骤 1：解压项目

```bash
# 解压 tar.gz
tar xzf ping-x-offline.tar.gz

# 或解压 zip
unzip ping-x-offline.zip

cd ping-x
```

### 步骤 2：验证 vendor 目录

```bash
# 确认 vendor 目录存在
ls -la vendor/

# 应该看到：
# vendor/github.com/
# vendor/golang.org/x/
# vendor/modules.txt
```

### 步骤 3：使用 vendor 方式构建

```bash
# 构建当前平台（Linux）
make vendor-build

# 或构建特定平台
make vendor-build-linux    # Linux amd64
make vendor-build-windows  # Windows amd64（需要 MinGW-w64）

# 或构建并打包所有平台
make vendor-release
```

### 步骤 4：验证构建结果

```bash
# 查看生成的文件
ls -lh dist/

# 应该看到：
# ping-x_v1.0.0_linux_amd64
# ping-x_v1.0.0_linux_amd64.tar.gz
# ping-x_v1.0.0_windows_amd64.exe
# ping-x_v1.0.0_windows_amd64.zip
```

## Makefile 命令参考

### 标准构建（go mod 方式）- 需要网络

```bash
make build              # 构建当前平台
make build-linux        # 构建 Linux 版本
make build-windows      # 构建 Windows 版本
make release            # 构建并打包所有平台
```

### 内网构建（vendor 方式）- 完全离线

```bash
make vendor-build              # 构建当前平台
make vendor-build-linux        # 构建 Linux 版本
make vendor-build-windows      # 构建 Windows 版本
make vendor-release            # 构建并打包所有平台
```

## Windows 交叉编译

如果需要在 Linux 上编译 Windows 版本，需要安装 MinGW-w64：

```bash
# Ubuntu/Debian
sudo apt-get install -y gcc-mingw-w64-x86-64

# CentOS/RHEL
sudo yum install -y mingw64-gcc
```

然后执行：

```bash
make vendor-build-windows
```

## 常见问题

### Q1: vendor 目录缺失

**错误信息**：
```
cannot find module for path ...
```

**解决方法**：
在外网环境重新执行：
```bash
go mod vendor
```

### Q2: vendor 目录不完整

**错误信息**：
```
missing go.sum entry
```

**解决方法**：
```bash
# 清理并重新生成
rm -rf vendor/
go mod vendor
```

### Q3: 版本信息不正确

**现象**：运行 `./ping-x --version` 显示 `dev` 或不正确的版本

**解决方法**：
使用 tag 来设置版本号：
```bash
# 创建 tag
git tag v1.0.0

# 使用 vendor 构建
make vendor-build

# 验证版本
./ping-x --version
# 输出：ping-x version v1.0.0
```

### Q4: 需要更新依赖版本

**场景**：需要更新某个依赖的版本

**解决方法**：
1. 在外网环境修改 `go.mod`
2. 执行 `go mod tidy`
3. 执行 `go mod vendor`
4. 重新打包传输到内网

```bash
# 外网环境
go get github.com/example/package@v1.2.3
go mod tidy
go mod vendor

# 重新打包
cd ..
tar czf ping-x-offline.tar.gz ping-x/
```

## 项目结构

```
ping-x/
├── cmd/                    # 命令行代码
├── internal/               # 内部实现
│   ├── config/            # 配置管理
│   ├── packet/            # 数据包处理
│   ├── proto/             # 协议实现
│   │   ├── ssm.go         # SSM 通用实现
│   │   ├── ssm_linux.go   # Linux SSM 实现
│   │   └── ssm_windows.go # Windows SSM 实现
│   └── stats/             # 统计分析
├── vendor/                 # 离线依赖（内网必需）
├── go.mod                  # Go 模块定义
├── go.sum                  # 依赖校验和
├── Makefile                # 构建脚本
├── main.go                 # 入口文件
├── RELEASE_GUIDE.md        # GitHub Release 指南
└── OFFLINE_BUILD.md        # 本文档
```

## 注意事项

1. **vendor 目录不要提交到 GitHub**：已在 `.gitignore` 中排除
2. **保持 go.mod 和 vendor 同步**：修改依赖后必须重新执行 `go mod vendor`
3. **版本号管理**：使用 git tag 来管理版本
4. **交叉编译**：Windows 版本需要 MinGW-w64 工具链
5. **测试验证**：构建后务必测试可执行文件是否正常工作

## 快速参考卡片

```bash
# === 外网准备 ===
git clone <repo-url>
cd ping-x
go mod vendor
cd ..
tar czf ping-x-offline.tar.gz ping-x/

# === 内网构建 ===
tar xzf ping-x-offline.tar.gz
cd ping-x
make vendor-build-linux
make vendor-build-windows
make vendor-release
ls -lh dist/
```

## 技术支持

如遇到问题，请检查：
1. vendor 目录是否完整
2. Go 版本是否为 1.17.5
3. Makefile 命令是否正确（使用 `vendor-*` 前缀）
4. 网络隔离环境是否完全离线
