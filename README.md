# mboot

> 一个基于 Go + Vue 3 的轻量级 PXE / iPXE 网络启动管理平台。

`mboot` 用于集中管理服务器、物理机和虚拟机的网络启动环境，集成 DHCP / ProxyDHCP、TFTP、HTTP Boot、iPXE 动态菜单、设备管理、IPMI/BMC 控制器、启动文件管理以及 netboot.xyz。

项目面向服务器批量装机、实验室网络启动、数据中心裸机部署和离线操作系统安装场景。

---

## 1. 功能概览

`mboot` 将传统 PXE 环境中分散的多个组件整合为一个统一服务：

```text
                    ┌────────────────────┐
                    │       mboot        │
                    │                    │
                    │  Web Management UI │
                    └─────────┬──────────┘
                              │
        ┌─────────────────────┼──────────────────────┐
        │                     │                      │
        ▼                     ▼                      ▼
  DHCP / ProxyDHCP           TFTP                HTTP Boot
        │                     │                      │
        └──────────────┬──────┴──────────────┬───────┘
                       │                     │
                       ▼                     ▼
                 PXE / UEFI              iPXE Menu
                       │                     │
                       └──────────┬──────────┘
                                  ▼
                           Operating System
```

主要功能：

* DHCP Server
* ProxyDHCP
* BIOS PXE 启动
* UEFI PXE 启动
* x86_64 / ARM64 架构识别
* TFTP Server
* HTTP Boot
* HTTP Range 请求
* 动态 iPXE 菜单
* netboot.xyz 集成
* 物理机 / 虚拟机统一设备管理
* DHCP / iPXE 动态地址关联
* IPMI / BMC 信息探测
* IPMI 电源状态与安全控制
* PXE / 本地硬盘一次性启动覆盖
* Legacy BIOS / UEFI 启动模式选择
* 启动菜单管理
* HTTP / TFTP 文件管理
* 实时日志
* 系统诊断
* Web 管理控制台
* 用户认证
* SQLite 配置存储
* SSE 实时事件
* 单二进制部署

### 设备与控制器模型

```text
设备（物理机或虚拟机）
  ├── PXE 网卡 MAC
  ├── DHCP 当前 IP
  ├── 固件与启动记录
  └── 控制器
       ├── IPMI / BMC（已支持）
       └── VMware / Proxmox / libvirt / Hyper-V（待适配）
```

设备使用稳定的设备 ID 和 PXE MAC 跟踪动态 DHCP 地址；控制器通过 `device_id` 与设备关联。BMC 地址和业务网 DHCP 地址相互独立，不能使用 IP 直接推断关联关系。

当前 IPMI 控制器支持 BMC 信息、Power、一次性 Boot Override，以及标准 Boot Parameter 5 读取。完整 BIOS 属性不属于 IPMI 标准，需要对应服务器厂商的 OEM 适配器。

运行 IPMI 功能前，需要在 mboot 主机安装 `ipmitool` 并加入 `PATH`。推荐使用 `lanplus`（IPMI 2.0）；`lan` 仅用于兼容 IPMI 1.5 旧设备。

详细说明参见 [docs/ipmi.md](docs/ipmi.md)。

---

# 2. 适用场景

`mboot` 主要适用于以下环境。

### 数据中心裸机部署

例如：

```text
100 台服务器
      │
      ├── DHCP / ProxyDHCP
      │
      ├── TFTP
      │
      ├── iPXE
      │
      └── HTTP
            │
            ▼
     Ubuntu / Debian / Rocky
```

用于服务器批量安装操作系统。

---

### 已有 DHCP 网络

如果现有网络已经存在 DHCP，例如：

```text
Router / Firewall
       │
       │ DHCP
       ▼
    Server
```

推荐使用：

```text
ProxyDHCP
```

此时：

```text
现有 DHCP
   │
   └── 分配 IP / Gateway / DNS

mboot
   │
   └── 提供 PXE Boot Server / Boot File
```

避免重复 DHCP Server 导致地址冲突。

---

### 独立装机网络

如果 PXE 网络完全隔离，可以由 `mboot` 提供完整 DHCP：

```text
                     mboot
                       │
      ┌────────────────┼────────────────┐
      │                │                │
     DHCP             TFTP             HTTP
      │                │                │
      └────────────────┴────────────────┘
                       │
                       ▼
                   PXE Client
```

---

# 3. 技术架构

## Backend

后端使用：

* Go 1.25+
* Gin
* SQLite
* `modernc.org/sqlite`
* `log/slog`
* TOML
* SSE
* Go Embed

SQLite 使用纯 Go 实现，因此可以关闭 CGO，方便：

```text
Linux
Windows
macOS
```

之间进行交叉编译。

---

## Frontend

Web UI 使用：

* Vue 3
* TypeScript
* Vite
* Tailwind CSS
* Vue Router
* Pinia
* lucide-vue-next

整体采用后台管理系统布局：

```text
┌──────────────────┬─────────────────────────────────┐
│                  │                                 │
│  Navigation      │           Header                │
│                  ├─────────────────────────────────┤
│  Dashboard       │                                 │
│  Services        │                                 │
│  Clients         │                                 │
│  Boot Menu       │          Workspace              │
│  Files           │                                 │
│  netboot.xyz     │                                 │
│  Logs            │                                 │
│  Diagnostics     │                                 │
│                  │                                 │
└──────────────────┴─────────────────────────────────┘
```

---

# 4. PXE 启动架构

推荐生产环境使用：

```text
Existing DHCP
      │
      │ 分配 IP / Gateway / DNS
      ▼
PXE Client
      │
      │ PXE Request
      ▼
mboot ProxyDHCP
      │
      │ next-server + bootfile
      ▼
TFTP
      │
      │ .kpxe / .efi
      ▼
iPXE
      │
      │ HTTP
      ▼
dynamic.ipxe
      │
      ├── Local Boot
      │
      ├── boot.ipxe
      │
      └── netboot.xyz
```

这种架构的核心思想是：

```text
TFTP
只负责第一阶段小文件

HTTP
负责第二阶段大文件
```

即：

```text
PXE
 ↓
TFTP
 ↓
iPXE
 ↓
HTTP
 ↓
Kernel / initrd / ISO resources
```

这样相比完全使用 TFTP 传输大文件，性能和稳定性更好。

---

# 5. DHCP 模式

`mboot` 支持两种 DHCP 工作模式。

## ProxyDHCP

推荐用于已有 DHCP 网络。

```text
PXE Client
   │
   ├───────────────► Existing DHCP
   │                    │
   │                    └── IP / Gateway / DNS
   │
   └───────────────► mboot
                        │
                        └── Boot Server / Boot File
```

ProxyDHCP：

* 不分配 IP
* 不提供默认网关
* 不提供 DNS
* 只提供 PXE 启动信息

因此不会替代现有 DHCP。

---

## Full DHCP

用于独立 PXE 网络：

```text
PXE Client
     │
     ▼
DHCPDISCOVER
     │
     ▼
mboot
     │
     ▼
DHCPOFFER
     │
     ▼
DHCPREQUEST
     │
     ▼
DHCPACK
```

完整 DHCP 可以提供：

* IP 地址
* Subnet Mask
* Gateway
* DNS
* Lease
* TFTP Server
* Boot File

> 不要在已经存在 DHCP Server 的生产 VLAN 中随意启用 Full DHCP。

否则可能出现：

```text
DHCP Server A
       │
       ├──── Offer A
Client │
       └──── Offer B
       │
DHCP Server B
```

客户端最终选择哪个 DHCP Server 具有不确定性。

---

# 6. ProxyDHCP

ProxyDHCP 同时监听：

```text
UDP 67
UDP 4011
```

这是为了兼容不同的 PXE 固件。

典型流程：

```text
PXE Client
   │
   ├── DHCPDISCOVER
   │
   ├── DHCPOFFER
   │
   ├── DHCPREQUEST
   │
   └── DHCPACK
```

当前实现：

```text
DISCOVER -> OFFER
REQUEST  -> ACK
```

响应可能发送至：

```text
255.255.255.255:68
```

以及：

```text
Directed Broadcast
Client Unicast
Request Source
```

以提高不同 PXE / UEFI 固件的兼容性。

---

# 7. PXE 架构识别

`mboot` 通过 DHCP Option 93 判断客户端架构。

|   Option 93 | Architecture |
| ----------: | ------------ |
| 0 / Missing | BIOS         |
|           6 | UEFI IA32    |
|           7 | UEFI x64     |
|           9 | UEFI x64     |
|          10 | UEFI ARM32   |
|          11 | UEFI ARM64   |

启动文件根据客户端架构自动选择。

---

# 8. 启动文件选择

## BIOS

优先级：

```text
netboot/netboot.xyz.kpxe
        ↓
netboot/netboot.xyz-undionly.kpxe
        ↓
Configured BIOS Boot File
```

---

## UEFI x64

优先：

```text
data/boot/netboot/netboot.xyz.efi
```

如果不存在，则使用：

```text
boot_files.uefi_x64
```

---

## UEFI ARM64

优先：

```text
data/boot/netboot/netboot.xyz-arm64.efi
```

如果不存在，则使用：

```text
boot_files.uefi_arm64
```

---

## IA32 / ARM32

支持架构识别，但是默认不会自动准备对应 EFI。

需要管理员自行提供：

```text
UEFI IA32 EFI
```

或：

```text
UEFI ARM32 EFI
```

文件。

---

# 9. iPXE 二阶段启动

`mboot` 会识别 iPXE Client。

主要通过：

```text
DHCP Option 77
DHCP Option 60
DHCP Option 175
```

判断。

识别成功后，如果客户端支持 HTTP，可以进入：

```text
http://MBOOT_SERVER/dynamic.ipxe?bootfile=ipxemenu
```

启动流程：

```text
PXE Firmware
     │
     ▼
TFTP
     │
     ▼
iPXE EFI / KPXE
     │
     ▼
HTTP
     │
     ▼
dynamic.ipxe
     │
     ▼
Boot Menu
```

---

# 10. Dynamic iPXE Menu

`mboot` 可以动态生成 iPXE 菜单。

例如：

```text
mboot Boot Menu
────────────────────────

Run boot.ipxe

Ubuntu 22.04

Ubuntu 24.04

Debian 12

netboot.xyz

Boot Local Disk

iPXE Shell
```

菜单通过：

```text
/dynamic.ipxe
```

实时生成。

菜单可以根据：

* 系统配置
* PXE Client
* Boot Menu
* Boot File

动态变化。

---

# 11. boot.ipxe

默认动态菜单的第一项是：

```text
Run boot.ipxe
```

对应：

```text
data/boot/http/boot.ipxe
```

因此可以直接维护自己的 iPXE 安装脚本。

例如：

```ipxe
#!ipxe

dhcp

kernel http://10.0.0.10/ubuntu/vmlinuz
initrd http://10.0.0.10/ubuntu/initrd

imgargs vmlinuz \
  initrd=initrd \
  ip=dhcp \
  url=http://10.0.0.10/ubuntu/ubuntu.iso

boot
```

---

# 12. TFTP

TFTP Server 默认使用：

```text
UDP 69
```

客户端请求：

```text
RRQ
```

后，后续传输使用临时 UDP Port。

当前支持：

```text
blksize
tsize
OACK
```

默认最大 `blksize`：

```text
1428
```

用于减少 MTU 分片风险。

如果老旧 PXE 固件不兼容 OACK，则会回退：

```text
512 Byte Block
```

---

# 13. HTTP Boot

HTTP Boot 用于传输：

* kernel
* initrd
* iPXE script
* 安装资源
* 大型启动文件

支持：

```text
HTTP Range
```

因此比 TFTP 更适合大文件。

主要路径：

```text
/dynamic.ipxe

/client/report

/netboot/...

/*
```

其中：

```text
/netboot/...
```

直接映射至 netboot.xyz 本地文件目录。

---

# 14. netboot.xyz

`mboot` 集成了 netboot.xyz 辅助能力。

可管理：

```text
netboot.xyz.kpxe

netboot.xyz-undionly.kpxe

netboot.xyz.efi

netboot.xyz-arm64.efi
```

文件默认存放：

```text
data/boot/netboot/
```

该目录同时可以通过：

```text
TFTP
```

和：

```text
HTTP
```

访问，不需要重复保存文件。

---

# 15. 文件目录

运行后默认数据目录：

```text
data/
├── mboot.toml
├── mboot.db
├── secret.key
│
├── logs/
│   └── mboot.log
│
├── boot/
│   ├── netboot/
│   │   ├── netboot.xyz.kpxe
│   │   ├── netboot.xyz.efi
│   │   └── netboot.xyz-arm64.efi
│   │
│   ├── tftp/
│   │
│   └── http/
│       └── boot.ipxe
│
├── smb/
│
└── exports/
```

设计原则：

```text
大文件
镜像
kernel
initrd
EFI
KPXE
        │
        ▼
File System
```

数据库只保存：

```text
Configuration
Boot Menu
Clients
Users
Download Records
Events
```

避免把大型二进制文件写入 SQLite。

---

# 16. 项目结构

```text
mboot/
│
├── embed.ipxe
│
├── .github/
│   └── workflows/
│       ├── release.yml
│       ├── docker.yml
│       └── build-boot.yml
│
├── cmd/
│   └── mboot/
│
├── internal/
│   ├── app/
│   ├── bootmenu/
│   ├── command/
│   ├── config/
│   ├── dhcp/
│   ├── httpboot/
│   ├── ipxe/
│   ├── netboot/
│   ├── netutil/
│   ├── observability/
│   ├── platform/
│   ├── pxeopt/
│   ├── smb/
│   ├── storage/
│   ├── tftp/
│   ├── torrent/
│   └── web/
│
├── web/
│
├── docs/
│
├── go.mod
│
└── README.md
```

模块职责：

```text
internal/app
    └── Application Lifecycle

internal/dhcp
    └── DHCP / ProxyDHCP

internal/tftp
    └── TFTP Server

internal/httpboot
    └── HTTP Boot

internal/ipxe
    └── iPXE Script Generator

internal/bootmenu
    └── Boot Menu

internal/netboot
    └── netboot.xyz

internal/storage
    └── SQLite

internal/observability
    └── Event / Log

internal/web
    └── REST API + Web UI
```

---

# 17. 配置文件

核心启动配置：

```text
data/mboot.toml
```

示例：

```toml
[data]
dir = "./data"

[admin]
admin_addr = "127.0.0.1:8088"

[database]
path = "./data/mboot.db"

[security]
secret_file = "./data/secret.key"

[logging]
level = "info"
format = "text"
```

大部分运行参数保存在 SQLite 中，并通过 Web UI 配置。

---

# 18. Listen IP 与 Advertise IP

这是 PXE 部署中非常重要的两个概念。

## Listen IP

例如：

```text
0.0.0.0
```

代表服务监听所有本地网卡。

适合：

```text
DHCP Broadcast
```

场景。

---

## Advertise IP

Advertise IP 是告诉 PXE 客户端：

```text
去哪个 IP 下载启动文件
```

例如服务器：

```text
eth0
10.0.1.10
```

则建议：

```text
Advertise IP
=
10.0.1.10
```

客户端必须能够访问这个 IP。

---

# 19. Web UI

管理界面包含：

* 仪表盘
* 服务配置
* 客户端
* 启动菜单
* 文件管理
* netboot.xyz
* 操作菜单
* 用户
* 日志
* 系统诊断

整体结构：

```text
┌───────────────────┬──────────────────────────────────┐
│                   │                                  │
│  mboot            │             Header               │
│                   │                                  │
├───────────────────┼──────────────────────────────────┤
│                   │                                  │
│  仪表盘           │                                  │
│                   │                                  │
│  服务配置         │                                  │
│                   │                                  │
│  客户端           │            Workspace             │
│                   │                                  │
│  启动菜单         │                                  │
│                   │                                  │
│  文件管理         │                                  │
│                   │                                  │
│  netboot.xyz      │                                  │
│                   │                                  │
│  日志             │                                  │
│                   │                                  │
│  系统诊断         │                                  │
│                   │                                  │
└───────────────────┴──────────────────────────────────┘
```

移动端使用 Drawer Navigation。

---

# 20. 登录与安全

首次启动时：

```text
mboot
   │
   ▼
检查管理员账号
   │
   ├── 不存在
   │      │
   │      ▼
   │   Setup
   │
   └── 已存在
          │
          ▼
        Login
```

管理员用户名：

```text
3 - 32 characters
```

允许：

```text
A-Z
a-z
0-9
.
_
-
@
```

登录失败具有限流机制：

```text
10 分钟内失败 10 次
       │
       ▼
锁定 10 分钟
```

同时：

* Cookie 使用 HttpOnly
* Cookie 使用 SameSite
* 日志禁止记录密码
* 日志禁止记录 Token
* 日志禁止记录 Session
* 文件路径必须限制在配置根目录
* TFTP Upload 默认关闭

---

# 21. 快速构建

## Linux

安装 Node.js、Go 后：

```bash
git clone <repository>

cd mboot
```

构建 Web UI：

```bash
cd web

npm ci

npm run build

cd ..
```

运行测试：

```bash
go test ./...
```

静态检查：

```bash
go vet ./...
```

构建：

```bash
mkdir -p dist

go build \
  -trimpath \
  -ldflags="-s -w" \
  -o dist/mboot \
  ./cmd/mboot
```

运行：

```bash
./dist/mboot
```

---

# 22. Windows 构建

PowerShell：

```powershell
cd mboot

npm ci --prefix web

npm run build --prefix web

go test ./...

go vet ./...

New-Item `
  -ItemType Directory `
  -Force `
  -Path dist |
  Out-Null

go build `
  -trimpath `
  -ldflags="-s -w" `
  -o dist\mboot.exe `
  .\cmd\mboot
```

---

# 23. 构建参数

生产版本建议始终保留：

```text
-trimpath
```

以及：

```text
-ldflags="-s -w"
```

完整：

```bash
go build \
  -trimpath \
  -ldflags="-s -w" \
  -o dist/mboot \
  ./cmd/mboot
```

主要作用：

```text
减小 Binary Size

+

避免暴露本地 Build Path
```

---

# 24. iPXE 固件

项目根目录：

```text
embed.ipxe
```

用于构建自定义 iPXE Firmware。

它和运行时：

```text
dynamic.ipxe
```

不是同一个东西。

需要明确区分：

```text
embed.ipxe
     │
     │ Compile
     ▼
iPXE Firmware
```

和：

```text
mboot
   │
   │ Runtime
   ▼
dynamic.ipxe
```

---

# 25. iPXE 构建流水线

GitHub Actions：

```text
.github/workflows/build-boot.yml
```

可以构建：

```text
undionly.kpxe

ipxe-x86_64.efi

ipxe-arm64.efi
```

当前构建：

```text
iPXE v2.0.0
```

并开启：

```text
DOWNLOAD_PROTO_HTTPS
```

以支持 HTTPS。

---

# 26. 三种 iPXE Script

项目中存在三种用途完全不同的 iPXE Script。

## embed.ipxe

```text
Repository
    │
    ▼
Compile into Firmware
```

修改后必须重新编译 iPXE Firmware。

---

## local-vars.ipxe

路径：

```text
data/boot/tftp/local-vars.ipxe
```

用于 netboot.xyz 本地 Hook。

---

## dynamic.ipxe

路径：

```text
/dynamic.ipxe
```

由 `mboot` 在运行时动态生成。

依赖：

```text
Boot Menu
Clients
Service Config
Database
```

---

# 27. 离线部署

完全离线环境中，不应该依赖公网 netboot.xyz 菜单。

推荐：

```text
             Internet
                X
                │
                │
           Offline LAN
                │
                ▼
              mboot
                │
        ┌───────┼────────┐
        │       │        │
       TFTP    HTTP     DHCP
        │       │        │
        └───────┴────────┘
                │
                ▼
             Servers
```

准备：

```text
data/boot/http/
```

例如：

```text
ubuntu2204/
├── vmlinuz
├── initrd
├── ubuntu.iso
└── nocloud/
    ├── meta-data
    └── user-data
```

然后：

```text
boot.ipxe
```

直接引用本地 HTTP。

---

# 28. Ubuntu 自动安装示例架构

例如：

```text
PXE Client
     │
     ▼
mboot
     │
     ▼
TFTP
     │
     ▼
iPXE
     │
     ▼
boot.ipxe
     │
     ├── vmlinuz
     │
     ├── initrd
     │
     └── autoinstall
             │
             ▼
          NoCloud
             │
             ├── user-data
             └── meta-data
```

这样可以实现：

```text
Power On
   ↓
PXE
   ↓
Ubuntu Installer
   ↓
Autoinstall
   ↓
Installed System
```

最终实现无人值守安装。

---

# 29. 跨 VLAN 部署

如果 PXE Client 和 `mboot` 不在同一个 VLAN：

```text
VLAN A
PXE Client
    │
    ▼
L3 Switch / Router
    │
    │ DHCP Relay
    ▼
VLAN B
mboot
```

必须检查：

```text
UDP 67
UDP 68
UDP 69
UDP 4011
HTTP Boot TCP Port
```

以及：

```text
DHCP Relay
IP Helper
Firewall
ACL
```

配置。

特别需要注意 ProxyDHCP：

```text
UDP 4011
```

某些网络设备默认只 Relay：

```text
UDP 67/68
```

而不会转发 UDP 4011。

---

# 30. 防火墙端口

典型需要允许：

| Protocol |           Port | Purpose          |
| -------- | -------------: | ---------------- |
| UDP      |             67 | DHCP / ProxyDHCP |
| UDP      |             68 | DHCP Client      |
| UDP      |             69 | TFTP             |
| UDP      |           4011 | ProxyDHCP        |
| TCP      | HTTP Boot Port | HTTP Boot        |
| TCP      |     Admin Port | Web Console      |

注意：

TFTP 不只是：

```text
UDP 69
```

初始化请求到达 69 后，后续数据传输可能使用临时 UDP Port。

---

# 31. 常见故障排查

## Client 获取不到 IP

检查：

```bash
tcpdump -ni any port 67 or port 68
```

确认是否存在：

```text
DHCPDISCOVER

DHCPOFFER

DHCPREQUEST

DHCPACK
```

---

## 能获取 IP，但不启动 PXE

重点检查：

```text
Option 66

Option 67

siaddr

filename
```

可以抓包：

```bash
tcpdump -ni any -vvv port 67 or port 68
```

---

## TFTP Timeout

检查：

```text
UDP 69

Firewall

TFTP Root

Boot File

Temporary UDP Port
```

并确认客户端是否发送：

```text
RRQ
```

---

## UEFI 找不到启动文件

确认：

```text
DHCP Option 93
```

架构是否正确。

例如：

```text
7 / 9
```

应该选择：

```text
UEFI x64
```

而：

```text
11
```

应该选择：

```text
UEFI ARM64
```

不要将 ARM64 客户端发送到 x86_64 EFI。

---

## Secure Boot

如果：

```text
DHCP
  OK

TFTP
  OK

EFI Download
  OK

Boot
  Failed
```

需要检查：

```text
UEFI Secure Boot
```

未签名的 iPXE EFI 可能被固件拒绝执行。

---

## dynamic.ipxe 无法访问

检查：

```text
Client
   │
   ▼
Advertise IP
   │
   ▼
HTTP Boot Port
```

确保 Advertise IP 是客户端可以访问的地址。

---

# 32. 常见固件兼容问题

以下环境需要特别注意。

### Secure Boot

未签名 EFI 可能无法启动。

### UEFI IA32

需要自行准备 IA32 EFI。

### UEFI ARM32

需要自行准备 ARM32 EFI。

### Wi-Fi PXE

大多数消费级无线网卡不支持 Firmware PXE。

### USB Ethernet PXE

取决于 BIOS / UEFI 是否内置对应 USB NIC Driver。

### ProxyDHCP

部分旧固件兼容性较差，可以尝试 Full DHCP 验证。

### VirtualBox Wi-Fi Bridge

广播报文可能受到 Host 网络栈限制。

---

# 33. 实时日志

Web UI 使用 SSE 获取实时事件。

架构：

```text
DHCP
TFTP
HTTP
System
   │
   ▼
Event Bus
   │
   ▼
SSE
   │
   ▼
Vue
```

日志按 Event ID 排序。

前端最多保存：

```text
1000
```

条近期事件，避免浏览器长时间运行后无限增长内存。

---

# 34. 数据存储设计

`mboot` 采用：

```text
SQLite
+
Filesystem
```

混合存储架构。

```text
               mboot
                 │
        ┌────────┴─────────┐
        │                  │
        ▼                  ▼
     SQLite            Filesystem
        │                  │
   Configuration          ISO
   User                   Kernel
   Client                 Initrd
   Menu                   EFI
   Event                  KPXE
                          Images
```

这种设计避免：

```text
Large Binary
      ↓
SQLite
```

导致数据库迅速膨胀。

---

# 35. GitHub Actions

当前流水线包括：

```text
release.yml
```

用于：

```text
Windows
Linux
macOS
```

多平台 Binary Release。

---

```text
docker.yml
```

用于构建：

```text
Multi-Architecture Docker Image
```

---

```text
build-boot.yml
```

用于构建：

```text
iPXE Firmware
```

---

# 36. 质量检查

提交或发布前建议执行：

```bash
go test ./...
```

```bash
go vet ./...
```

```bash
npm run typecheck --prefix web
```

```bash
npm run build --prefix web
```

完整：

```bash
go test ./... &&
go vet ./... &&
npm run typecheck --prefix web &&
npm run build --prefix web
```

---

# 37. 开发原则

新增 API：

```text
API Layer
   │
   ▼
Business Layer
```

API 不应承载复杂业务逻辑。

协议模块：

```text
Protocol Parsing
       │
       ▼
Business Decision
```

应该保持分离。

新增配置项时必须同时更新：

```text
Default
Validation
Backend
Frontend
Documentation
```

避免只修改某一层导致配置不一致。

---

# 38. PXE 协议开发注意事项

修改 DHCP 时必须特别注意以下 Options：

```text
53
54
60
66
67
77
93
97
175
```

其中：

```text
53 = DHCP Message Type

54 = DHCP Server Identifier

60 = Vendor Class

66 = TFTP Server

67 = Boot File

77 = User Class

93 = Client Architecture

97 = Client Machine Identifier

175 = iPXE Options
```

---

# 39. 测试建议

协议相关模块建议逐步增加：

```text
DHCP Golden Test

PXE Option 43 Test

iPXE Script Golden Test

TFTP RRQ Test

TFTP WRQ Test

HTTP Range Test

Path Traversal Test
```

PXE 属于典型的协议兼容型软件，仅依赖 UI 功能测试不足以保证不同硬件固件的兼容性。

---

# 40. 已知限制

目前需要注意：

* TFTP 基于 UDP，对丢包比较敏感。
* MTU 和 blksize 会影响 TFTP 稳定性。
* 不同 BIOS / UEFI 的 PXE 实现存在明显差异。
* Secure Boot 可能阻止未签名 iPXE EFI。
* 完整 DHCP 模式可能与现有 DHCP 冲突。
* ProxyDHCP 跨 VLAN 依赖 DHCP Relay / IP Helper。
* UDP 4011 不一定被所有 Relay 默认转发。
* UEFI IA32 / ARM32 默认需要自行准备 EFI。
* netboot.xyz 在线模式依赖互联网。
* 完全离线环境必须准备本地 Kernel / Initrd / Installation Resources。

---

# 41. 推荐生产部署架构

对于已有网络基础设施的数据中心，推荐：

```text
                     ┌─────────────────────┐
                     │ Existing DHCP       │
                     │ Router / Firewall   │
                     └─────────┬───────────┘
                               │
                           DHCP IP
                               │
                               ▼
                       ┌─────────────┐
                       │ PXE Client  │
                       └──────┬──────┘
                              │
                  PXE Boot Information
                              │
                              ▼
                     ┌────────────────┐
                     │     mboot      │
                     │                │
                     │ ProxyDHCP      │
                     │ TFTP           │
                     │ HTTP Boot      │
                     │ iPXE Menu      │
                     │ Web Console    │
                     └───────┬────────┘
                             │
                             ▼
                     Local HTTP Mirror
                             │
                ┌────────────┼────────────┐
                ▼            ▼            ▼
             Ubuntu        Debian       Rocky
```

这种模式的优点：

* 不修改现有 DHCP 地址分配体系
* PXE 服务和基础网络职责分离
* TFTP 只承担第一阶段启动
* 大文件通过 HTTP 传输
* 可以集中管理启动菜单
* 方便后续扩展无人值守安装
* 更适合数据中心服务器批量部署

---

# 42. 项目定位

`mboot` 不是单纯的：

```text
TFTP Server
```

也不是单纯的：

```text
DHCP Server
```

它更接近一个：

```text
Bare-Metal Network Boot Control Plane
```

即：

```text
                mboot
                  │
        ┌─────────┼──────────┐
        │         │          │
    Discovery    Boot      Management
        │         │          │
      DHCP       TFTP       Web UI
   ProxyDHCP      HTTP       API
        │         │          │
        └─────────┼──────────┘
                  │
                  ▼
            Bare Metal
```

长期可以继续向以下方向扩展：

```text
PXE
 │
 ├── Bare Metal Provisioning
 │
 ├── OS Autoinstall
 │
 ├── Hardware Inventory
 │
 ├── Redfish / 厂商 BIOS OEM 适配
 │
 ├── VMware / Proxmox / libvirt / Hyper-V 控制器
 │
 ├── Kickstart
 │
 ├── Ubuntu Autoinstall
 │
 ├── Cloud-init
 │
 └── Cluster Provisioning
```

最终形成完整的裸金属服务器自动化部署平台。

---

# License

This project is licensed under the MIT License. See [LICENSE](LICENSE) for details.
