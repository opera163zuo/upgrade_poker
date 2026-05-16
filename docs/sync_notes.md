# 部署同步说明

> 项目：upgrade_poker · 编译产物发布指南
> 对应提交：`046b0c0`

---

## 1. 编译产物清单

| 文件 | 大小 | 说明 |
|------|------|------|
| `release/upgrade_046b0c0_win64.exe` | ~7.6 MB | Windows 可执行文件 |
| `release/upgrade_046b0c0_win64.zip` | ~2.9 MB | Windows zip 打包 |
| `release/upgrade_046b0c0_linux64` | ~8.0 MB | Linux 可执行文件 (AMD64) |

## 2. 运行环境要求

| 平台 | 最低版本 |
|------|----------|
| **Windows** | Windows 10+ (64-bit) |
| **Linux** | Linux kernel 4.15+ (AMD64) |

## 3. 部署方式

**直接部署：** 将所需平台的 release 产物复制到目标机器即可运行，无需安装额外依赖。

### Windows
```
rem 方案 A：直接使用 exe
copy release\upgrade_046b0c0_win64.exe D:\target\upgrade.exe

rem 方案 B：解压 zip 使用
解压 release\upgrade_046b0c0_win64.zip 到 D:\target\
```

### Linux
```bash
cp release/upgrade_046b0c0_linux64 /opt/upgrade/upgrade
chmod +x /opt/upgrade/upgrade
```

## 4. 运行检查清单

- [ ] **assets/ 资源目录** — 确保与可执行文件同目录，包含必要的游戏资源文件
- [ ] **首次运行** — 程序会自动创建配置目录（`~/.upgrade_poker/` 或同目录下 `config/`）
- [ ] **逻辑分辨率** — 默认使用 1280×960 逻辑分辨率
- [ ] **防火墙/杀软** — Windows 下如被拦截，请添加至白名单

## 5. 获取方式

### 方式 A：Git 仓库（推荐）
```bash
git clone <仓库地址>
cd upgrade_poker
# release/ 目录包含所有已编译产物
ls -la release/
```

### 方式 B：直接下载
从 CI/CD 产物页面或版本发布页直接下载对应平台的 zip/exe 文件。

### 方式 C：自编译
```bash
# 安装依赖后
go build -o upgrade_linux64 .
GOOS=windows GOARCH=amd64 go build -o upgrade_win64.exe .
```

---

## 6. 版本说明

- 产物文件名中的哈希（如 `046b0c0`）对应编译时的 Git commit
- 如需特定版本，请 check out 对应 commit 后自行编译
- 最新版本始终使用最新的 release 目录产物
