# IPMI 集成说明

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
