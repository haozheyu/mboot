package netboot

import (
	"fmt"
	"os"
	"path/filepath"

	"mboot/internal/booturl"
	"mboot/internal/observability"
)

const LocalVarsFile = "local-vars.ipxe"

func EnsureLocalVars(tftpRoot, advertiseIP, httpAddr string, events *observability.Hub) (string, bool, error) {
	if err := os.MkdirAll(tftpRoot, 0755); err != nil {
		return "", false, err
	}

	target := filepath.Join(tftpRoot, LocalVarsFile)

	if info, err := os.Stat(target); err == nil && !info.IsDir() {
		if events != nil {
			events.Publish("info", "netboot.xyz", "local-vars.ipxe 已存在，跳过生成")
		}
		return target, false, nil
	}

	script := LocalVarsScript(advertiseIP, httpAddr)

	if err := os.WriteFile(target, []byte(script), 0644); err != nil {
		return "", false, err
	}

	if events != nil {
		events.Publish("info", "netboot.xyz", "已生成 local-vars.ipxe: "+target)
	}

	return target, true, nil
}

func LocalVarsScript(advertiseIP, httpAddr string) string {
	base := booturl.HTTPBase(advertiseIP, httpAddr)

	return fmt.Sprintf(`#!ipxe

# ============================================================
# mboot Local PXE Menu
# Generated automatically - do not edit manually
# ============================================================

# Ensure network configuration is available.
isset ${net0/ip} || dhcp || goto network_failed

# iPXE timeout is milliseconds.
# 60000 = 60 seconds.
set menu-timeout 60000

# mboot HTTP boot service.
set bootserver %s

# ============================================================
# Main Menu
# ============================================================

:main_menu
menu PXE Install Menu

item --gap -- ---------------- OS Installation ----------------
item ubuntu2204 Ubuntu 22.04.5 Autoinstall

item --gap -- ---------------- Boot Options -------------------
item localboot Boot From Local Disk
item netbootxyz Load netboot.xyz Menu

item --gap -- ---------------- Tools --------------------------
item show_info Show Boot Information
item shell iPXE Shell
item reboot Reboot
item exit Exit iPXE

choose --timeout ${menu-timeout} --default ubuntu2204 selected || goto localboot
goto ${selected}


# ============================================================
# Ubuntu 22.04.5 Autoinstall
# ============================================================

:ubuntu2204

echo
echo ============================================================
echo Starting Ubuntu 22.04.5 Autoinstall
echo ============================================================
echo

# Current Ubuntu installation image is amd64.
iseq ${buildarch} x86_64 && goto ubuntu2204_start ||
iseq ${buildarch} amd64 && goto ubuntu2204_start ||

echo ERROR: Unsupported architecture: ${buildarch}
echo Ubuntu 22.04.5 installation currently supports amd64 only.
sleep 5
goto main_menu


:ubuntu2204_start

echo Boot Server : ${bootserver}
echo Client IP   : ${net0/ip}
echo MAC Address : ${net0/mac}
echo

chain --autofree ${bootserver}/ubuntu2204/boot.ipxe || goto boot_failed


# ============================================================
# Boot Information
# ============================================================

:show_info

echo
echo ============================================================
echo PXE Boot Information
echo ============================================================
echo

echo Platform           : ${platform}
echo Architecture       : ${buildarch}
echo MAC Address        : ${net0/mac}
echo Client IP          : ${net0/ip}
echo Netmask            : ${net0/netmask}
echo Gateway            : ${net0/gateway}
echo DNS                : ${dns}

echo
echo DHCP Information
echo ------------------------------------------------------------
echo Next Server        : ${next-server}
echo Boot File          : ${filename}
echo ProxyDHCP Server   : ${proxydhcp/next-server}
echo ProxyDHCP File     : ${proxydhcp/filename}

echo
echo mboot Information
echo ------------------------------------------------------------
echo HTTP Boot Server   : ${bootserver}
echo Ubuntu Boot Script : ${bootserver}/ubuntu2204/boot.ipxe

echo
echo ============================================================

sleep 10
goto main_menu


# ============================================================
# Local Disk Boot
# ============================================================

:localboot

echo
echo Booting from local disk...
echo

sanboot --no-describe --drive 0x80 || goto localboot_failed


:localboot_failed

echo
echo Local disk boot failed.
echo Returning to PXE menu...
echo

sleep 3
goto main_menu


# ============================================================
# netboot.xyz
# ============================================================

:netbootxyz

echo
echo Loading netboot.xyz...
echo

chain --autofree https://boot.netboot.xyz || goto boot_failed


# ============================================================
# iPXE Shell
# ============================================================

:shell

echo
echo Entering iPXE shell.
echo Type "exit" to return to PXE menu.
echo

shell
goto main_menu


# ============================================================
# Reboot
# ============================================================

:reboot

echo
echo Rebooting system...
sleep 2
reboot


# ============================================================
# Exit
# ============================================================

:exit

echo
echo Exiting iPXE...
exit


# ============================================================
# Network Failure
# ============================================================

:network_failed

echo
echo ============================================================
echo PXE Network Initialization Failed
echo ============================================================
echo

echo DHCP failed or no network interface received an IP address.
echo

echo Check:
echo   1. Physical network link
echo   2. Switch port / VLAN
echo   3. DHCP server
echo   4. PXE DHCP options
echo   5. Network interface firmware

echo
sleep 5
shell


# ============================================================
# Ubuntu Boot Failure
# ============================================================

:boot_failed

echo
echo ============================================================
echo PXE Boot Failed
echo ============================================================
echo

echo Client Information:
echo ------------------------------------------------------------
echo IP          : ${net0/ip}
echo MAC         : ${net0/mac}
echo Architecture: ${buildarch}
echo Platform    : ${platform}

echo
echo Boot Server:
echo ------------------------------------------------------------
echo ${bootserver}

echo
echo Check:
echo ------------------------------------------------------------
echo 1. Client network connectivity
echo 2. mboot HTTP service
echo 3. ${bootserver}/ubuntu2204/boot.ipxe
echo 4. casper/vmlinuz
echo 5. casper/initrd
echo 6. ubuntu-22.04.5-live-server-amd64.iso
echo 7. nocloud/user-data
echo 8. nocloud/meta-data
echo

echo Entering iPXE shell...
echo

sleep 5
shell
goto main_menu

`, base)
}

//func LocalVarsScript(advertiseIP, httpAddr string) string {
//	base := booturl.HTTPBase(advertiseIP, httpAddr)
//
//	return fmt.Sprintf(`#!ipxe
//isset ${net0/ip} || dhcp || goto failed
//set menu-timeout 60000
//
//set public-mirror https://mirrors.cernet.edu.cn
//set local-mirror %s
//
//set debian-mirror-host mirrors.cernet.edu.cn
//set debian-mirror-dir /debian
//set debian-security-path /debian-security
//set debian-release trixie
//
//isset ${proxydhcp/next-server} && set use_proxydhcp_settings true ||
//
//cpuid --ext 29 && set debian_arch amd64 || set debian_arch arm64
//iseq ${debian_arch} amd64 && set alpine_arch x86_64 || set alpine_arch aarch64
//
//:main_menu
//menu PXE Install Menu
//item --gap -- OS Installation
//item public_debian Public Install Debian 13
//item public_alpine Public Install Alpine Linux
//item local_debian Local Install Debian 13
//item local_alpine Local Install Alpine Linux
//item --gap -- Tools
//item show_info Show Boot Information
//item shell iPXE Shell
//item exit Load netboot.xyz Menu
//choose --timeout ${menu-timeout} --default public_debian selected || goto exit
//goto ${selected}
//
//:public_debian
//imgfree
//
//set debian-base ${public-mirror}/debian/dists/${debian-release}/main/installer-${debian_arch}/current/images/netboot/debian-installer/${debian_arch}
//
//kernel ${debian-base}/linux \
//	initrd=initrd.gz \
//	ip=dhcp \
//	auto=true \
//	priority=critical \
//	mirror/country=manual \
//	mirror/http/hostname=${debian-mirror-host} \
//	mirror/http/directory=${debian-mirror-dir} \
//	mirror/http/proxy= \
//	apt-setup/services-select=security \
//	apt-setup/security_host=${debian-mirror-host} \
//	apt-setup/security_path=${debian-security-path} \
//	language=zh_CN \
//	country=CN \
//	locale=zh_CN.UTF-8 \
//	keymap=us \
//	hostname=debian \
//	domain= \
//	passwd/root-login=false \
//	passwd/make-user=true \
//	partman-auto/method=regular \
//	partman-auto/choose_recipe=atomic \
//	pkgsel/run_tasksel=false \
//	pkgsel/include=openssh-server,curl,wget,vim,sudo \
//	pkgsel/upgrade=none \
//	popularity-contest/participate=false \
//	openssh-server/password-auth=true
//
//initrd ${debian-base}/initrd.gz
//boot || goto failed
//
//:public_alpine
//imgfree
//set alpine-base ${public-mirror}/alpine/v3.23/releases/${alpine_arch}/netboot
//kernel ${alpine-base}/vmlinuz-lts initrd=initramfs-lts ip=dhcp alpine_repo=${public-mirror}/alpine/v3.23/main modloop=${alpine-base}/modloop-lts
//initrd ${alpine-base}/initramfs-lts
//boot || goto failed
//
//:local_debian
//imgfree
//
//set local-debian-base ${local-mirror}/debian/dists/${debian-release}/main/installer-${debian_arch}/current/images/netboot/debian-installer/${debian_arch}
//
//kernel ${local-debian-base}/linux \
//	initrd=initrd.gz \
//	ip=dhcp \
//	auto=true \
//	priority=critical \
//	mirror/country=manual \
//	mirror/http/hostname=${debian-mirror-host} \
//	mirror/http/directory=${debian-mirror-dir} \
//	mirror/http/proxy= \
//	apt-setup/services-select=security \
//	apt-setup/security_host=${debian-mirror-host} \
//	apt-setup/security_path=${debian-security-path} \
//	language=zh_CN \
//	country=CN \
//	locale=zh_CN.UTF-8 \
//	keymap=us \
//	hostname=debian \
//	domain= \
//	passwd/root-login=false \
//	passwd/make-user=true \
//	partman-auto/method=regular \
//	partman-auto/choose_recipe=atomic \
//	pkgsel/run_tasksel=false \
//	pkgsel/include=openssh-server,curl,wget,vim,sudo \
//	pkgsel/upgrade=none \
//	popularity-contest/participate=false \
//	openssh-server/password-auth=true
//
//initrd ${local-debian-base}/initrd.gz
//boot || goto failed
//
//:local_alpine
//imgfree
//set local-alpine-base ${local-mirror}/alpine/v3.23/releases/${alpine_arch}/netboot
//kernel ${local-alpine-base}/vmlinuz-lts initrd=initramfs-lts ip=dhcp alpine_repo=${local-mirror}/alpine/v3.23/main modloop=${local-alpine-base}/modloop-lts
//initrd ${local-alpine-base}/initramfs-lts
//boot || goto failed
//
//:show_info
//echo
//echo PXE boot information
//echo debian_arch: ${debian_arch}
//echo alpine_arch: ${alpine_arch}
//echo platform: ${platform}
//echo mac: ${net0/mac}
//echo ip: ${net0/ip}
//echo next-server: ${next-server}
//echo proxydhcp next-server: ${proxydhcp/next-server}
//echo filename: ${filename}
//echo proxydhcp filename: ${proxydhcp/filename}
//echo public mirror: ${public-mirror}
//echo local mirror: ${local-mirror}
//echo release: ${debian-release}
//sleep 8
//goto main_menu
//
//:shell
//shell
//goto main_menu
//
//:failed
//echo Boot failed. Check network, files and boot parameters.
//sleep 5
//shell
//
//:exit
//chain --autofree https://boot.netboot.xyz
//`, base)
//}
