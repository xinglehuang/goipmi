package ipmi

import (
    "fmt"
    "os/exec"
    "time"
    "bufio"
    "strings"
    "regexp"
)

// default ipmitool settting
var (
    Ipmitool string = "ipmitool"
    Iface string = "lanplus"
    Port int = 623
    SnVerifiedExpireMinute int = 0
    ReSnInfo string = `^\s*Product Serial\s+:\s*(\w+)\s*$`
    ReToolVersionInfo string = `^ipmitool\s+version\s+(\S+)\s*$`
    ReSnInfoCompiled = regexp.MustCompile(ReSnInfo)
    ReToolVersionInfoCompiled = regexp.MustCompile(ReToolVersionInfo)
)

// client 私有，必须通过new方法创建
type client struct {
    tool string // ipmitool
    iface string // -I lanplus
    host string // -H host
    user string // -U user
    pass string // -P password
    sn string// expect device serial number
    port int // -p port

    // ipmitool options参数
    options []string

    // 序列号校验cache
    snVerified bool
    snVerifiedExpireMinute int
    lastSnVerifyTime time.Time
    
}

// getter
func (c *client)Tool() string { return c.tool }
func (c *client)Host() string { return c.host }
func (c *client)Iface() string { return c.iface }
func (c *client)User() string { return c.user }
func (c *client)Pass() string { return c.pass }
func (c *client)Sn() string { return c.sn }
func (c *client)Port() int { return c.port }
func (c *client)Options() []string { return c.options }
func (c *client)SnVerified() bool { return c.snVerified }
func (c *client)SnVerifiedExpireMinute() int { return c.snVerifiedExpireMinute }
func (c *client)LastSnVerifyTime() time.Time { return c.lastSnVerifyTime }

// setter
func (c *client)SetSnVerifiedExpireMinute(expire int) { c.snVerifiedExpireMinute = expire }


// 基础client
func NewBasicClient(tool, iface, host, user, pass, sn string, port, expire int) *client {
    options := []string {
        "-I", iface,
        "-H", host,
        "-U", user,
        "-P", pass,
    }
    if port != Port {
        options = append(options, "-p", fmt.Sprintf("%d", port))
    }
    return &client {tool, iface, host, user, pass, sn, port, options, false, expire, time.Time{}}
}

// return client with default config
func NewSimpleClient(host, user, pass, sn string) *client {
    return NewBasicClient(Ipmitool, Iface, host, user, pass, sn, Port, SnVerifiedExpireMinute)
}

// 判断客户端sn是否最近校验通过
func (c *client)SnRecentlyVerified() bool {
    return c.snVerified && time.Now().Before(
            c.lastSnVerifyTime.Add(time.Minute * time.Duration(c.snVerifiedExpireMinute)))
}

// 生成cmd对象
func (c *client)Cmd(cmd ...string) *exec.Cmd {
    var opts = make([]string, len(c.options) + len(cmd))
    copy(opts, c.options)
    opts = append(opts, cmd...)
    return exec.Command(c.tool, opts...)
}

// 执行cmd, 返回string输出
func (c *client)Run(cmd ...string) (string, error) {
    co := c.Cmd(cmd...)
    o, e := co.Output()
    return string(o), e
}

// 安全执行命令, 执行之前确保序列号已校验
func (c *client)SafeRun(cmd ...string) (string, error) {
    if ! c.SnRecentlyVerified() {
        // try verify
        e := c.VerifySn()
        if e != nil {
            return "", e
        }
    }
    // now sn verified, safe to run
    return c.Run(cmd...)
}


// run fru print 0
func (c *client)FruPrint0() (string, error) {
    return c.Run(CmdFruPrint0)
}

// 获取序列号
func (c *client)GetSn() (string, error) {
    var sn string

    // 获取fru print 0信息
    o, e := c.FruPrint0()
    if e != nil {
        return "", e
    }

    // 从fru print 0输出中寻找sn信息
    scanner := bufio.NewScanner(strings.NewReader(o))
    for scanner.Scan() {
         m := ReSnInfoCompiled.FindStringSubmatch(scanner.Text())
        if m != nil && len(m) > 1 {
            sn = m[1]
            break
        }
    }

    // 没有匹配到sn
    if sn == "" {
        err := fmt.Errorf("Host(%s) serial number info missing!", c.host)
        return "", err
    }

    return sn, nil
}
    

// 校验sn同时更新client校验信息
func (c *client)VerifySn() error {
    now := time.Now()
    verified := false
    
    // 获取成功但是匹配不上sn
    sn, e := c.GetSn()
    if e == nil && c.sn != sn {
        e = fmt.Errorf(
            "Verify host(%s) serial number (want: %s, real: %s) failed!",
            c.host, c.sn, sn,
        )
    }
    
    // 校验成功
    if e == nil {
        verified = true
    }

    // 更新cient校验信息
    c.lastSnVerifyTime = now
    c.snVerified = verified

    return e
}

// 电源控制
func (c *client)PowerStatus() (string, error) {
    return c.SafeRun(CmdPowerStatus)
}

func (c *client)PowerOn() (string, error) {
    return c.SafeRun(CmdPowerOn)
}

func (c *client)PowerOff() (string, error) {
    return c.SafeRun(CmdPowerOff)
}

func (c *client)PowerReset() (string, error) {
    return c.SafeRun(CmdPowerReset)
}

func (c *client)PowerSoft() (string, error) {
    return c.SafeRun(CmdPowerSoft)
}

func (c *client)Power(action string) (string, error) {
    switch action {
	    case "status":
	        return c.PowerStatus()
	    case "on":
	        return c.PowerOn()
	    case "off":
	        return c.PowerOff()
	    case "reset":
	        return c.PowerReset()
	    case "soft":
	        return c.PowerSoft()
	    default:
	        return "", fmt.Errorf("Power action(%s) not supported!", action)
    }
}

// 引导顺序
func (c *client)BootPxe() (string, error) {
    return c.SafeRun(CmdBootPxe)
}
func (c *client)BootDisk() (string, error) {
    return c.SafeRun(CmdBootDisk)
}
func (c *client)BootBios() (string, error) {
    return c.SafeRun(CmdBootBios)
}
func (c *client)BootNone() (string, error) {
    return c.SafeRun(CmdBootNone)
}
func (c *client)BootCdrom() (string, error) {
    return c.SafeRun(CmdBootCdrom)
}

func (c *client)Boot(dev string) (string, error) {
    switch dev {
	    case "pxe":
	        return c.BootPxe()
	    case "disk":
	        return c.BootDisk()
	    case "bios":
	        return c.BootBios()
	    case "cdrom":
	        return c.BootCdrom()
	    case "none":
	        return c.BootNone()
	    default:
	        return "", fmt.Errorf("Boot dev(%s) not supported!", dev)
    }
}

// tool version, default unknown
func ToolVersion(tool string) (string, error) {
    var version string = "unknown"

    co := exec.Command(tool, "-V")
    o, e := co.Output()
    if e != nil {
        return version, e
    }

    scanner := bufio.NewScanner(strings.NewReader(string(o)))
    for scanner.Scan() {
        m := ReToolVersionInfoCompiled.FindStringSubmatch(scanner.Text())
        if m != nil && len(m) > 1 {
            version = m[1]
            break
        }
    }

    if version == "unknown" {
        return version, fmt.Errorf("Detect tool(%s) version info failed!", tool)
    }

    return version, nil
}
func (c *client)ToolVersion() (string, error) {
    return ToolVersion(c.tool)
}
