# mboot PXE 自动安装环境说明

目前 `mboot` PXE 自动安装链路已跑通，可通过 iPXE 菜单启动 Ubuntu 22.04.5 自动安装。

## 当前目录

```text
data/
├── boot/
│   ├── http/
│   │   └── ubuntu2204/
│   │       ├── boot.ipxe
│   │       ├── casper/
│   │       │   ├── vmlinuz
│   │       │   └── initrd
│   │       ├── nocloud/
│   │       │   ├── user-data
│   │       │   └── meta-data
│   │       └── ubuntu-22.04.5-live-server-amd64.iso
│   ├── netboot/
│   │   ├── netboot.xyz.efi
│   │   ├── netboot.xyz.kpxe
│   │   ├── netboot.xyz-undionly.kpxe
│   │   └── netboot.xyz-arm64.efi
│   └── tftp/
│       └── local-vars.ipxe
├── logs/
│   └── pxe.log
├── pxe.db
└── pxe.toml
```

## 当前启动菜单

Web 管理界面已启用 `iPXE Menu`，当前菜单项：

```text
Ubuntu 22.04.5 Autoinstall
启动文件：ubuntu2204/boot.ipxe
```

菜单由 `dynamic.ipxe` 动态生成，可通过以下命令查看：

```bash
curl 'http://10.221.34.11:81/dynamic.ipxe?bootfile=ipxemenu'
```

## PXE 启动流程

```text
物理机 / 虚拟机
      ↓
DHCP 获取 IP 和 PXE 启动信息
      ↓
iPXE
      ↓
dynamic.ipxe 动态菜单
      ↓
ubuntu2204/boot.ipxe
      ↓
vmlinuz + initrd
      ↓
下载 Ubuntu 22.04.5 ISO
      ↓
读取 NoCloud user-data / meta-data
      ↓
Ubuntu Autoinstall
```

## BIOS / UEFI 区分

KVM 测试环境：

```text
SeaBIOS → Legacy BIOS
OVMF    → UEFI
```

实际物理服务器则直接由主板 Boot Mode 决定：

```text
Boot Mode = Legacy → Legacy BIOS PXE
Boot Mode = UEFI   → UEFI PXE
```

PXE 服务端可以通过 DHCP Option 93 判断客户端启动模式：

```text
ARCH = 0     → Legacy BIOS x86
ARCH = 7/9   → UEFI x86_64
ARCH = 11    → UEFI ARM64
```

Linux 系统安装完成后也可以查看：

```bash
[ -d /sys/firmware/efi ] && echo UEFI || echo "Legacy BIOS"
```

## 当前验证结果

目前已验证：

```text
✓ DHCP 分配地址正常
✓ PXE/iPXE 启动正常
✓ dynamic.ipxe 动态菜单正常
✓ Web 菜单配置生效
✓ boot.ipxe 加载正常
✓ vmlinuz/initrd 加载正常
✓ Ubuntu ISO HTTP 下载正常
✓ NoCloud Autoinstall 链路正常
```

测试过程中 4GB 内存虚拟机在下载 Live ISO 时出现：

```text
wget: short write: No space left on device
```

测试虚拟机建议配置：

```text
CPU：4 vCPU
内存：8 GB
磁盘：50 GB
网卡：virtio
网络：br0
启动顺序：PXE → 本地磁盘
```

