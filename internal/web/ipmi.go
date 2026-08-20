package web

import (
	"context"
	"database/sql"
	"errors"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"mboot/internal/ipmi"
	"mboot/internal/storage"
)

func ipmiID(c *gin.Context) (int64, bool) {
	id, e := strconv.ParseInt(c.Param("id"), 10, 64)
	if e != nil || id < 1 {
		Fail(c, 400, "VALIDATION_ERROR", "无效的节点 ID")
		return 0, false
	}
	return id, true
}

func (h *Handler) listIPMINodes(c *gin.Context) {
	rows, e := h.app.Storage().ListIPMINodes(c)
	if e != nil {
		Fail(c, 500, "IPMI_LIST_FAILED", e.Error())
		return
	}
	out := make([]storage.PublicIPMINode, 0, len(rows))
	for _, n := range rows {
		out = append(out, n.Public())
	}
	OK(c, out)
}

func validateIPMINode(n *storage.IPMINode) error {
	n.Name = strings.TrimSpace(n.Name)
	n.Address = strings.TrimSpace(n.Address)
	n.Username = strings.TrimSpace(n.Username)
	n.Interface = strings.TrimSpace(n.Interface)
	n.Vendor = strings.ToLower(strings.TrimSpace(n.Vendor))
	if n.Name == "" || n.Address == "" || n.Username == "" {
		return errors.New("名称、BMC 地址和用户名不能为空")
	}
	if strings.ContainsAny(n.Address, "/\\ ?#") {
		return errors.New("BMC 地址格式不正确")
	}
	if ip := net.ParseIP(strings.Trim(n.Address, "[]")); ip == nil {
		if u, e := url.Parse("http://" + n.Address); e != nil || u.Hostname() == "" {
			return errors.New("BMC 地址格式不正确")
		}
	}
	if n.Interface == "" {
		n.Interface = "lanplus"
	}
	if n.Interface != "lanplus" && n.Interface != "lan" {
		return errors.New("接口仅支持 lanplus 或 lan")
	}
	if n.Vendor == "" {
		n.Vendor = "generic"
	}
	return nil
}

func (h *Handler) saveIPMINode(c *gin.Context) {
	var n storage.IPMINode
	if e := c.ShouldBindJSON(&n); e != nil {
		Fail(c, 400, "VALIDATION_ERROR", "请求格式错误")
		return
	}
	if c.Param("id") != "" {
		id, ok := ipmiID(c)
		if !ok {
			return
		}
		n.ID = id
	}
	if e := validateIPMINode(&n); e != nil {
		Fail(c, 400, "VALIDATION_ERROR", e.Error())
		return
	}
	if n.ID == 0 && n.Password == "" {
		Fail(c, 400, "VALIDATION_ERROR", "新节点必须填写密码")
		return
	}
	saved, e := h.app.Storage().SaveIPMINode(c, n)
	if e != nil {
		Fail(c, 409, "IPMI_SAVE_FAILED", e.Error())
		return
	}
	OK(c, saved.Public())
}

func (h *Handler) deleteIPMINode(c *gin.Context) {
	id, ok := ipmiID(c)
	if !ok {
		return
	}
	if e := h.app.Storage().DeleteIPMINode(c, id); e != nil {
		code := 500
		if errors.Is(e, sql.ErrNoRows) {
			code = 404
		}
		Fail(c, code, "IPMI_DELETE_FAILED", e.Error())
		return
	}
	OK(c, gin.H{"deleted": id})
}

func (h *Handler) runIPMI(c *gin.Context, args []string) (storage.IPMINode, string, bool) {
	id, ok := ipmiID(c)
	if !ok {
		return storage.IPMINode{}, "", false
	}
	n, e := h.app.Storage().GetIPMINode(c, id)
	if e != nil {
		Fail(c, 404, "IPMI_NODE_NOT_FOUND", "IPMI 节点不存在")
		return n, "", false
	}
	ctx, cancel := timeContext(c, 20*time.Second)
	defer cancel()
	out, e := (ipmi.Runner{}).Run(ctx, n, args...)
	if e != nil {
		h.app.EventHub().Publish("error", "ipmi", n.Name+": "+e.Error())
		Fail(c, http.StatusBadGateway, "IPMI_COMMAND_FAILED", e.Error())
		return n, "", false
	}
	return n, out, true
}

func timeContext(c *gin.Context, d time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(c.Request.Context(), d)
}

func (h *Handler) probeIPMINode(c *gin.Context) {
	n, out, ok := h.runIPMI(c, ipmi.BMCInfoArgs())
	if !ok {
		return
	}
	power, e := (ipmi.Runner{}).Run(c.Request.Context(), n, "chassis", "power", "status")
	if e != nil {
		power = e.Error()
	}
	OK(c, gin.H{"bmc": out, "power": power, "capabilities": gin.H{"bmc": true, "power": true, "boot": true, "bios": n.Vendor != "generic", "bios_note": "BIOS 设置依赖厂商 OEM 支持；通用模式仅可读取启动参数。"}})
}
func (h *Handler) ipmiPower(c *gin.Context) {
	var req struct {
		Action string `json:"action"`
	}
	if c.ShouldBindJSON(&req) != nil {
		Fail(c, 400, "VALIDATION_ERROR", "请求格式错误")
		return
	}
	a, e := ipmi.PowerArgs(req.Action)
	if e != nil {
		Fail(c, 400, "VALIDATION_ERROR", e.Error())
		return
	}
	n, out, ok := h.runIPMI(c, a)
	if !ok {
		return
	}
	h.app.EventHub().Publish("info", "ipmi", n.Name+" 电源操作: "+req.Action)
	OK(c, gin.H{"action": req.Action, "output": out})
}
func (h *Handler) ipmiBoot(c *gin.Context) {
	var req struct {
		Device      string `json:"device"`
		Persistent  bool   `json:"persistent"`
		UEFI        bool   `json:"uefi"`
		PowerAction string `json:"power_action"`
	}
	if c.ShouldBindJSON(&req) != nil {
		Fail(c, 400, "VALIDATION_ERROR", "请求格式错误")
		return
	}
	a, e := ipmi.BootArgs(req.Device, req.Persistent, req.UEFI)
	if e != nil {
		Fail(c, 400, "VALIDATION_ERROR", e.Error())
		return
	}
	n, out, ok := h.runIPMI(c, a)
	if !ok {
		return
	}
	if req.PowerAction != "" {
		pa, e := ipmi.PowerArgs(req.PowerAction)
		if e != nil {
			Fail(c, 400, "VALIDATION_ERROR", e.Error())
			return
		}
		var ok2 bool
		_, _, ok2 = h.runIPMI(c, pa)
		if !ok2 {
			return
		}
	}
	h.app.EventHub().Publish("info", "ipmi", n.Name+" 设置启动目标: "+req.Device)
	OK(c, gin.H{"device": req.Device, "output": out, "power_action": req.PowerAction})
}
func (h *Handler) ipmiBIOS(c *gin.Context) {
	a, _ := ipmi.BIOSArgs("5")
	n, out, ok := h.runIPMI(c, a)
	if !ok {
		return
	}
	OK(c, gin.H{"vendor": n.Vendor, "standard_boot_options": out, "oem_supported": n.Vendor != "generic", "note": "IPMI 标准不定义完整 BIOS 属性模型，厂商扩展需单独适配。"})
}
