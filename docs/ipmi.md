# IPMI 集成说明

## 本次修正内容

- 将原“客户端”领域明确为“设备”，统一承载物理机和虚拟机的 PXE MAC、DHCP 地址、状态和启动记录。
- 将 IPMI 提升为设备的控制器适配器，通过 `device_id` 关联；保留旧 `/clients`、`/ipmi/nodes` 路由兼容。
- 完整 DHCP 模式按租约更新设备地址；ProxyDHCP 模式通过 iPXE MAC 回调更新实际来源 IP，并校验 MAC/IP 格式。
- 修正 `lan` 与 `lanplus` 的协议提示：`lanplus` 为推荐的 IPMI 2.0，避免把 IPMI 1.5 认证错误误判为密码错误。
- 电源操作改为中文说明和二次确认，服务端对改变状态的请求强制要求确认标志。
- Boot Override 仅允许设置下一次启动，不允许同一请求自动重启，也不允许修改持久启动顺序。
- 修正启动模式语义：必须明确选择 Legacy BIOS 或 UEFI；不再把缺少 `efiboot` 错误描述为“保持当前模式”。
- 控制器密码不通过 API 回显，也不放入命令行参数或事件日志；当前数据库仍为明文保存，生产环境必须限制数据目录权限。

## 设备与控制器

mboot 将资源拆成两个相互独立的层次：

- **设备（Device）**：物理机和虚拟机共用的数据模型，保存 PXE 网卡 MAC、DHCP 当前 IP、固件类型、状态和启动记录。设备不要求具备 BMC。
- **控制器（Controller）**：通过 `device_id` 关联设备，负责电源和启动控制。当前适配器为 `ipmi`；VMware、Proxmox、libvirt 和 Hyper-V 应作为并列适配器接入。

兼容路由 `/api/v1/clients` 与 `/api/v1/ipmi/nodes` 保留；新的领域路由为 `/api/v1/devices` 和 `/api/v1/controllers`。

## 能力边界

本项目把裸金属管理拆成 PXE 数据面和 IPMI 带外控制面。当前 IPMI 控制面负责：

- BMC：连通性探测和 `mc info` 信息读取。
- Power：状态查询、开机、软关机、关机、重启和电源循环。
- Boot：设置 PXE、本地磁盘、虚拟光驱或 BIOS Setup 启动目标，支持一次性/持久和 UEFI 标记。
- BIOS：通用模式只读取 IPMI System Boot Options 参数 5。完整 BIOS 属性不属于 IPMI 标准，后续应按 Dell、HPE、Lenovo、浪潮、Supermicro 等厂商增加 OEM 适配器。

## 运行依赖

mboot 服务主机必须安装 `ipmitool` 并加入 `PATH`。服务通过 `lanplus`（推荐）或 `lan` 访问 BMC，密码使用 `IPMI_PASSWORD` 进程环境变量传给 `ipmitool`，不会出现在命令行参数、API 响应或事件日志中。

节点凭据保存在 mboot 的 SQLite 数据库中，因此生产环境必须限制数据目录和数据库文件的操作系统访问权限。后续若接入集中密钥系统，应将 `ipmi_nodes.password` 替换为密钥引用。

## API

- `GET/POST /api/v1/ipmi/nodes`
- `PUT/DELETE /api/v1/ipmi/nodes/:id`
- `POST /api/v1/ipmi/nodes/:id/probe`
- `POST /api/v1/ipmi/nodes/:id/power`
- `POST /api/v1/ipmi/nodes/:id/boot`
- `GET /api/v1/ipmi/nodes/:id/bios`

所有接口沿用管理后台登录认证。电源和启动操作使用服务端白名单，不接受任意 IPMI 命令。
