package ipmi

import "testing"

//func NewBasicClient(tool, iface, host, user, pass, sn string, port, expire int) *client {
func TestNewBasicClient(t *testing.T) {
    c := NewBasicClient("ipmitool", "lan", "127.0.0.1", "user", "pass", "x", 623, 1)
    m1 := map[string]string {
        "ipmitool": c.tool,
        "lan": c.iface,
        "127.0.0.1": c.host,
        "user": c.user,
        "pass": c.pass,
        "x": c.sn,
    }
    m2 := map[int]int {
        623: c.port,
        1: c.snVerifiedExpireMinute,
    }
    m3 := map[bool]bool {
        false: c.snVerified,
    }
    for k, v := range m1 {
        if k != v {
            t.Fatalf("test %s == %s failed!\n%#v\n", k, v, c)
        }
    }
    for k, v := range m2 {
        if k != v {
            t.Fatalf("test %d == %d failed!\n%#v\n", k, v, c)
        }
    }
    for k, v := range m3 {
        if k != v {
            t.Fatalf("test %v == %v failed!\n%#v\n", k, v, c)
        }
    }
    opts := []string{"-I", "lan", "-H", "127.0.0.1", "-U", "user", "-P", "pass"}
    for i, v := range opts {
        if c.options[i] != v {
            t.Fatalf("test %s == %s failed!\n%#v\n", c.options[i], v, opts)
        }
    }
}
func TestNewSimpleClient(t *testing.T) {
    t.Skip("xxxx")
}
//func (c *client)SnRecentlyVerified() bool {
//func (c *client)Cmd(cmd ...string) *exec.Cmd {
//func (c *client)Run(cmd ...string) (string, error) {
//func (c *client)SafeRun(cmd ...string) (string, error) {
//func (c *client)FruPrint0() (string, error) {
//func (c *client)GetSn() (string, error) {
//func (c *client)VerifySn() error {
//func (c *client)PowerStatus() (string, error) {
//func (c *client)PowerOn() (string, error) {
//func (c *client)PowerOff() (string, error) {
//func (c *client)PowerReset() (string, error) {
//func (c *client)PowerSoft() (string, error) {
//func (c *client)Power(action string) (string, error) {
//func (c *client)BootPxe() (string, error) {
//func (c *client)BootDisk() (string, error) {
//func (c *client)BootBios() (string, error) {
//func (c *client)BootNone() (string, error) {
//func (c *client)BootCdrom() (string, error) {
//func (c *client)Boot(dev string) (string, error) {
//func ToolVersion(tool string) (string, error) {
//func (c *client)ToolVersion() (string, error) {
