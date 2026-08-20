package ipmi

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"mboot/internal/storage"
)

var ErrToolMissing = errors.New("未找到 ipmitool，请先在 mboot 主机安装并加入 PATH")

type Runner struct{ Tool string }

func (r Runner) Run(ctx context.Context, n storage.IPMINode, args ...string) (string, error) {
	tool := r.Tool
	if tool == "" {
		tool = "ipmitool"
	}
	if _, err := exec.LookPath(tool); err != nil {
		return "", ErrToolMissing
	}
	base := []string{"-I", n.Interface, "-H", n.Address, "-U", n.Username, "-E"}
	cmd := exec.CommandContext(ctx, tool, append(base, args...)...)
	cmd.Env = append(cmd.Environ(), "IPMI_PASSWORD="+n.Password)
	b, err := cmd.CombinedOutput()
	out := strings.TrimSpace(string(b))
	if err != nil {
		if out == "" {
			out = err.Error()
		}
		return "", fmt.Errorf("IPMI 操作失败: %s", out)
	}
	return out, nil
}

func PowerArgs(action string) ([]string, error) {
	switch action {
	case "status", "on", "off", "cycle", "reset", "soft":
		return []string{"chassis", "power", action}, nil
	}
	return nil, errors.New("不支持的电源操作")
}

func BootArgs(device string, persistent bool, uefi bool) ([]string, error) {
	switch device {
	case "pxe", "disk", "cdrom", "bios", "none":
	default:
		return nil, errors.New("不支持的启动目标")
	}
	a := []string{"chassis", "bootdev", device}
	var options []string
	if persistent {
		options = append(options, "persistent")
	}
	if uefi {
		options = append(options, "efiboot")
	}
	if len(options) > 0 {
		a = append(a, "options="+strings.Join(options, ","))
	}
	return a, nil
}

func BMCInfoArgs() []string { return []string{"mc", "info"} }

// BIOSArgs deliberately exposes only the standardized IPMI System Boot Options
// parameter. Rich BIOS attributes are OEM-specific and belong in a vendor adapter.
func BIOSArgs(parameter string) ([]string, error) {
	if parameter != "5" {
		return nil, errors.New("通用 IPMI 仅开放 System Boot Options 参数 5；其他 BIOS 能力需要厂商适配器")
	}
	return []string{"chassis", "bootparam", "get", "5"}, nil
}
