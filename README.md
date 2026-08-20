# mboot

> 面向物理机与虚拟机的 PXE/iPXE 网络启动和带外控制平台。

mboot 将 DHCP、ProxyDHCP、TFTP、HTTP Boot、iPXE 动态菜单、设备发现和 IPMI/BMC 控制整合为一个 Go 服务，并提供 Vue 3 管理界面。

它既可以管理裸金属服务器批量装机，也可以发现虚拟机 PXE 客户端。设备身份和控制能力相互独立：物理机可以关联 IPMI/BMC，虚拟机后续可以关联 VMware、Proxmox、libvirt 或 Hyper-V 控制器。

## 核心模型

```text
                         mboot
                           │
              ┌────────────┴────────────┐
              │                         │
         网络启动平面                 管理平面
              │                         │
   DHCP / ProxyDHCP / TFTP       Web UI / API / Events
       HTTP Boot / iPXE                 │
              │               ┌─────────┴─────────┐
              │               │                   │
            设备层          控制器层              用户与审计
              │               │
      物理机 / 虚拟机      IPMI / BMC（已支持）
      MAC / IP / 状态      VMware / Proxmox（规划）
```

### 设备

设备是物理机和虚拟机共用的资源模型，负责记录：

- PXE 网卡 MAC
- DHCP 当前地址
- BIOS/UEFI 固件类型
- 在线状态和最后启动记录
- 磁盘、网络等客户端上报信息

设备通过稳定的设备 ID 和 PXE MAC 跟踪动态 IP，不使用易变化的 DHCP 地址作为关联键。

### 控制器

控制器通过 `device_id` 关联设备，并提供设备类型对应的控制能力：

| 控制器 | 使用对象 | 状态 | 能力 |
| --- | --- | --- | --- |
| IPMI / BMC | 物理服务器 | 已支持 | BMC、Power、Boot、标准启动参数 |
| VMware | VMware 虚拟机 | 规划中 | 电源、虚拟启动设备 |
| Proxmox | QEMU/LXC | 规划中 | 电源、启动顺序 |
| libvirt | KVM/QEMU | 规划中 | 电源、启动设备 |
| Hyper-V | Hyper-V 虚拟机 | 规划中 | 电源、固件启动设置 |

BMC 管理地址与设备业务网地址是两套独立地址。mboot 不通过 IP 猜测两者关系，而是在控制器中显式选择设备，后续由 DHCP/iPXE 自动更新设备地址。

## 已实现能力

### 网络启动

- DHCP Server 和 ProxyDHCP
- BIOS、UEFI x64、UEFI ARM64 客户端识别
- TFTP 和 HTTP Boot
- HTTP Range 请求
- 动态 iPXE 菜单
- 本地启动、外部脚本和 netboot.xyz
- 完整 DHCP 与已有 DHCP 网络两种部署模式

### 设备管理

- 自动发现物理机和虚拟机
- 按 PXE MAC 更新动态 DHCP 地址
- ProxyDHCP 环境通过 iPXE 回调记录实际客户端地址
- 静态设备、批量设备和待认领设备
- Wake-on-LAN
- 设备状态和健康信息

### IPMI / BMC

- BMC 连通性及控制器信息探测
- 查询电源状态
- 开机、软关机、强制重启、电源循环和强制断电
- 设置下一次 PXE、本地硬盘、虚拟光驱或 BIOS Setup 启动
- 明确选择 Legacy BIOS 或 UEFI 启动模式
- 读取 IPMI System Boot Options 参数 5
- 操作事件记录

### 管理平台

- Vue 3 Web 控制台
- 用户登录和会话认证
- SQLite 配置及状态存储
- 文件管理和启动菜单管理
- SSE 实时事件
- 日志和系统诊断
- Go 单二进制部署

## IPMI 安全边界

IPMI 操作会直接影响物理服务器，mboot 对此采用以下限制：

- 查询操作不需要确认，所有状态变更都需要明确确认。
- 前端展示目标服务器、BMC 地址、动作含义及风险。
- 后端再次验证确认标志，不能只靠绕过页面直接误操作。
- Boot Override 仅允许下一次启动，不允许通过该接口修改永久启动顺序。
- 设置启动设备不会自动重启，电源操作必须单独确认。
- 启动模式必须明确选择 Legacy BIOS 或 UEFI。
- 密码通过 `IPMI_PASSWORD` 环境变量传递给 `ipmitool`，不会出现在进程参数、API 响应或事件日志中。

当前控制器密码保存在 SQLite 数据库中。生产环境必须限制数据目录权限；后续可替换为集中密钥系统或加密凭据存储。

完整 BIOS 属性并不属于 IPMI 标准。当前通用适配器只读取标准启动参数，BIOS 属性修改需要 Dell、HPE、Lenovo、浪潮、Supermicro 等厂商 OEM 适配器。

## 启动覆盖说明

页面中的 PXE 和硬盘启动均表示“仅下一次启动”：

```text
选择下一次启动设备
        │
        ├── PXE 网络启动
        ├── 默认本地硬盘
        ├── 虚拟光驱
        └── BIOS Setup
        │
        ├── Legacy BIOS
        └── UEFI
        │
        └── 写入 BMC，不自动重启
```

等价的常用命令如下：

```bash
# 下一次从 PXE 启动，Legacy BIOS
ipmitool -I lanplus -H BMC_IP -U USER chassis bootdev pxe

# 下一次从 PXE 启动，UEFI
ipmitool -I lanplus -H BMC_IP -U USER chassis bootdev pxe options=efiboot

# 下一次从默认硬盘启动，Legacy BIOS
ipmitool -I lanplus -H BMC_IP -U USER chassis bootdev disk

# 下一次从默认硬盘启动，UEFI
ipmitool -I lanplus -H BMC_IP -U USER chassis bootdev disk options=efiboot
```

`chassis bootparam get 5` 读取的是 BMC 保存的启动覆盖，不是当前操作系统的实际启动状态。

## 快速开始

### 环境要求

- Go 1.25 或兼容版本
- Node.js 和 npm（构建 Web 前端）
- `ipmitool`（使用 IPMI 控制器时）
- Linux 或 Windows

推荐使用 `lanplus`，它对应 IPMI 2.0。`lan` 使用 IPMI 1.5/RMCP，仅用于旧设备兼容。

### 构建前端

```bash
cd web
npm install
npm run build
cd ..
```

前端产物生成到 `internal/web/dist`，并嵌入 Go 二进制。

### 构建服务

Linux：

```bash
go build -o dist/mboot ./cmd/mboot
./dist/mboot
```

Windows PowerShell：

```powershell
go build -o dist\mboot.exe .\cmd\mboot
.\dist\mboot.exe
```

常用启动参数：

```text
-config      指定启动配置文件
-data-dir    覆盖数据目录
-host        覆盖管理界面监听地址
-port        覆盖管理界面端口
-no-browser  启动时不自动打开浏览器
```

首次打开管理界面时创建管理员账号。服务设置、用户、设备、菜单、事件和控制器保存在数据目录中。

### Docker

```bash
docker compose up --build
```

DHCP、ProxyDHCP 和 TFTP 使用 UDP，容器网络模式、防火墙和跨 VLAN DHCP Relay 需要根据实际网络单独配置。生产部署前请先在隔离网络验证。

## DHCP 部署模式

### 完整 DHCP

mboot 负责地址分配和 PXE 参数：

```text
客户端 ──DHCP──> mboot ──> 地址租约 + PXE 启动文件
```

设备地址会在 DHCP 分配和续租时按 MAC 自动更新。

### ProxyDHCP

已有 DHCP 服务器继续分配地址，mboot 只返回启动信息：

```text
客户端 ──DHCP──────> 现有 DHCP Server ──> IP / 网关 / DNS
   │
   └────ProxyDHCP──> mboot ─────────────> PXE 启动信息
```

由于 mboot 不拥有外部 DHCP 租约，iPXE 加载动态菜单时会携带 PXE MAC；mboot 使用回调来源地址更新设备当前 IP。

## API 概览

所有管理接口位于 `/api/v1`，除初始化和登录外均需要认证。

| 领域 | 路径 | 说明 |
| --- | --- | --- |
| 设备 | `/devices` | 设备查询、保存、删除和批量创建 |
| 控制器 | `/controllers` | 控制器查询、保存和删除 |
| BMC 探测 | `/controllers/:id/probe` | BMC 信息和电源状态 |
| 电源 | `/controllers/:id/power` | IPMI 电源操作 |
| 启动覆盖 | `/controllers/:id/boot` | 设置下一次启动设备 |
| BIOS 启动参数 | `/controllers/:id/bios` | 读取标准 Boot Parameter 5 |
| 配置 | `/config` | 服务配置 |
| 菜单 | `/menus` | PXE/iPXE 菜单 |
| 文件 | `/files` | 启动文件管理 |
| 事件 | `/events/stream` | SSE 实时事件 |

旧 `/clients` 和 `/ipmi/nodes` 路由暂时保留兼容。

## 项目结构

```text
mboot/
├── cmd/mboot/              程序入口
├── internal/app/           生命周期和服务编排
├── internal/config/        启动配置
├── internal/storage/       SQLite 数据模型
├── internal/dhcp/          DHCP / ProxyDHCP
├── internal/tftp/          TFTP
├── internal/httpboot/      HTTP Boot
├── internal/ipxe/          动态 iPXE 脚本
├── internal/ipmi/          IPMI 命令适配器
├── internal/web/           API 和嵌入式前端
├── web/                    Vue 3 管理界面
└── docs/                   专题文档
```

## 文档

- [IPMI、设备和控制器设计](docs/ipmi.md)
- [PXE 自动安装环境说明](docs/docs.md)

## 当前限制

- VMware、Proxmox、libvirt 和 Hyper-V 控制器尚未实现。
- 厂商 BIOS OEM 属性尚未实现。
- IPMI 凭据尚未接入加密存储或外部密钥系统。
- ProxyDHCP 跨 VLAN 场景依赖网络设备正确配置 DHCP Relay/IP Helper。
- 不同服务器 BIOS/UEFI 的 PXE 和 Boot Override 行为可能存在差异。

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE) for details.
