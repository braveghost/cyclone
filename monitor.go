package cyclone

import (
	"fmt"
	"github.com/pkg/errors"
	"strings"
)

type (
	MatchType int
	MonitorType int
	matchFunc func(*SrvConfigInfo) (string, error)
)

const (
	MatchTypeFull   MatchType = iota // 完全匹配, 用于主机名
	MatchTypePrefix                  // 左匹配
	MatchTypeIn                      // 包含
	MatchTypeEqual                   // 服务健康数量相等
	MatchTypeScope                   // 服务健康数量范围
)

const (
	MonitorTypeNode    MonitorType = iota // 主机名
	MonitorTypeAddress                    // ip地址
	MonitorTypeCount                      // 服务健康数量相等

)

var (
	defaultMonitorConf *MonitorConfig
	monitorClients     = make(map[string]*monitor)

	headMsg = `
================================ [%v] %v ================================
------BaseInfo:
	[Count]: %v		[Health]: %v		[Error]: %v
	`

	activeMsg = `
------Active - N%v:
	[Node]: %v		[Address]: %v
`

	deathMsg = `
------Death - N%v:
	[Host]: %v
`
)

var (
	MonitorConfIsNullErr      = errors.New("monitor config is null")
	MonitorAddrISNullErr      = errors.New("monitor registry client address is null")
	MonitorSrvCountEqualErr   = errors.New("monitor service health count error")
	MonitorSrvMatchErr        = errors.New("monitor service health match error")
	MonitorSrvCountPeakErr    = errors.New("monitor service count less than peak")
	MonitorSrvCountValleyErr  = errors.New("monitor service count more than valley")
	MonitorSrvCountScopeErr   = errors.New("monitor service config count scope error")
	MonitorMatchFuncChoiceErr = errors.New("monitor service match choice function error")
)

type MonitorConfig struct {
	Registry *RegistryConf
	Type     MonitorType      // 主机名Or地址Or数量, 节点状态
	Services []*SrvConfigInfo // 服务信息
	Match    MatchType
}

type SrvConfigInfo struct {
	Name   string
	Hosts  []string
	Peak   int // 服务最大量, count+=peak
	Valley int // 服务服务最小量, count+=valley，valley<count<peak，组成一个范围值
}

func (sci SrvConfigInfo) Count() int {
	return len(sci.Hosts)
}

func InitConfig(mc *MonitorConfig) {
	defaultMonitorConf = mc
}

type client interface {
	initClient(rc *RegistryConf) error
	HealthService(name string) *srvBaseInfo
}

type srvBaseInfo struct {
	Health int
	Err    error
	Active []*SrvInfo
	Death  []*SrvInfo
}
type SrvInfo struct {
	Node    string
	Address string
	Tags    map[string]string
}

type monitor struct {
	client
	conf *MonitorConfig
	matchFunc
}

func (m *monitor) getHost(si *SrvInfo) string {
	switch m.conf.Type {
	case MonitorTypeAddress:
		return si.Address
	case MonitorTypeNode:
		return si.Node

	}
	return ""
}

func (m *monitor) matchErr(tag string) error {
	switch m.conf.Type {
	case MonitorTypeAddress:
		return errors.Wrapf(MonitorSrvMatchErr, "Match %v address", tag)
	case MonitorTypeNode:
		return errors.Wrapf(MonitorSrvMatchErr, "Match %v node", tag)
	}
	return MonitorSrvMatchErr
}

type matchInnerFunc func(map[string]*SrvInfo, string) bool

func (m *monitor) match(tag, fnTag string, info *SrvConfigInfo, fn matchInnerFunc) (string, error) {

	hs := m.HealthService(info.Name)

	var msg string

	if hs.Err == nil {

		getHost := m.getHost

		var hostSrv = make(map[string]*SrvInfo)
		for i, l := range hs.Active {
			hostSrv[getHost(l)] = hs.Active[i]
			msg += getActiveMsg(i+1, l.Node, l.Address)
		}

		for ii, ll := range info.Hosts {

			if !fn(hostSrv, ll) {
				msg += getDeathMsg(ii+1, ll)
				hs.Err = m.matchErr(fnTag)
			}
		}
	}
	msg = getHeadMsg(tag, info.Name, info.Count(), hs.Health, hs.Err) + msg
	return msg, hs.Err
}

// 服务关键字完全匹配
func (m *monitor) matchFull(info *SrvConfigInfo) (string, error) {

	return m.match(
		"ServiceMatchFull",
		"full",
		info,
		func(hs map[string]*SrvInfo, host string) bool {
			_, ok := hs[host]
			return ok
		})
}

// 服务关键字前缀匹配
func (m *monitor) matchPrefix(info *SrvConfigInfo) (string, error) {
	return m.match(
		"ServiceMatchPrefix",
		"prefix",

		info,
		func(hs map[string]*SrvInfo, host string) bool {
			var status bool

			for k := range hs {
				if strings.HasPrefix(k, host) {
					status = true
					break
				}
			}

			return status
		})

}

// 服务关键字包含匹配
func (m *monitor) matchIn(info *SrvConfigInfo) (string, error) {

	return m.match(
		"ServiceMatchIn",
		"in",
		info,
		func(hs map[string]*SrvInfo, host string) bool {
			var status bool
			for k := range hs {
				if strings.Contains(k, host) {
					status = true
					break
				}
			}

			return status
		})

}

// 服务数量等值匹配
func (m *monitor) matchEqual(info *SrvConfigInfo) (string, error) {
	hs := m.HealthService(info.Name)
	if hs.Err == nil {
		if hs.Health != info.Count() {
			hs.Err = MonitorSrvCountEqualErr
		}

	}

	return getHeadMsg("ServiceCountEqual", info.Name, info.Count(), hs.Health, hs.Err), hs.Err
}

// 服务数量范围匹配
func (m *monitor) matchScope(info *SrvConfigInfo) (string, error) {
	hs := m.HealthService(info.Name)

	if hs.Err == nil {
		if hs.Health > info.Peak {
			hs.Err = MonitorSrvCountPeakErr
		} else if hs.Health < info.Valley {
			hs.Err = MonitorSrvCountValleyErr

		}
		if hs.Err != nil {
			hs.Err = errors.Wrapf(hs.Err, "peak=%v,valley=%v,health=%v", info.Peak, info.Valley, hs.Health)
		}
	}

	return getHeadMsg("ServiceCountScope", info.Name, info.Count(), hs.Health, hs.Err), hs.Err
}

// 生成消息头
func getHeadMsg(tag string, name string, count, health int, err error) string {
	return fmt.Sprintf(headMsg, tag, name, count, health, err)
}

// 生成活跃服务消息
func getActiveMsg(idx int, node, addr string) string {
	return fmt.Sprintf(activeMsg, idx, node, addr)
}

// 生成不存在的服务的消息
func getDeathMsg(idx int, host string) string {
	return fmt.Sprintf(deathMsg, idx, host)
}

// 规则校验
func (m *monitor) monitorService(conf *MonitorConfig) []string {
	var bd []string

	for _, info := range conf.Services {
		var (
			msg string
			err error
		)
		err = checkScopeConf(info)
		if err != nil {
			msg = getHeadMsg("ServiceConfigError", info.Name, info.Count(), -1, err)
		} else {
			msg, err = m.matchFunc(info)
		}

		if err != nil {
			bd = append(bd, msg)
		}
	}
	return bd
}

// 启动
func (m *monitor) Run() ([]string, error) {

	return m.monitorService(m.conf), nil

}

// 检查服务的数量范围配置
func checkScopeConf(sci *SrvConfigInfo) error {
	if sci.Peak < 0 {
		return errors.Wrapf(MonitorSrvCountScopeErr, "%v SrvConfigInfo.Peak < 0", sci.Name)
	}
	if sci.Valley < 0 {
		return errors.Wrapf(MonitorSrvCountScopeErr, "%v SrvConfigInfo.Valley < 0", sci.Name)

	}
	if sci.Peak < sci.Valley {
		return errors.Wrapf(MonitorSrvCountScopeErr, "%v SrvConfigInfo.Peak < SrvConfigInfo.Valley", sci.Name)

	}
	return nil
}

// 选择规则处理函数
func (m *monitor) matchChoice() matchFunc {

	tp := m.conf.Type
	var f matchFunc
	if tp != MonitorTypeCount && tp != MonitorTypeNode && tp != MonitorTypeAddress {
		return f
	}

	switch m.conf.Match {
	case MatchTypeFull:
		if tp != MonitorTypeCount {

			f = m.matchFull
		}
	case MatchTypePrefix:
		if tp != MonitorTypeCount {

			f = m.matchPrefix
		}
	case MatchTypeIn:
		if tp != MonitorTypeCount {

			f = m.matchIn
		}
	case MatchTypeEqual:
		if tp == MonitorTypeCount {

			f = m.matchEqual
		}
	case MatchTypeScope:
		if tp == MonitorTypeCount {

			f = m.matchScope
		}
	}
	return f
}

// 创建监控器, 单例
func NewMonitor(name string, mc *MonitorConfig) (*monitor, error) {
	// todo 地址不可用判断以及注册失败判断
	defaultMonitor, ok := monitorClients[name]
	if ok && defaultMonitor != nil {
		// todo 配置变更后的重新初始化
		defaultMonitor.conf = mc
		return defaultMonitor, nil
	}

	if mc == nil {
		mc = defaultMonitorConf
	}

	if mc == nil {
		return nil, MonitorConfIsNullErr
	}

	return newMonitor(name, mc)

}

// 创建监控器，
func newMonitor(name string, mc *MonitorConfig) (*monitor, error) {
	var err error
	var mt *monitor
	reg := mc.Registry
	switch reg.Registry {
	case "consul":
		mt = &monitor{
			client: &matchConsul{},
			conf:   mc}
		mt.matchFunc = mt.matchChoice()
	default:
		err = RegistryNameErr
	}

	if mt != nil {
		if mt.matchFunc == nil {
			return nil, MonitorMatchFuncChoiceErr
		}
	}
	if mt != nil {

		match := mt.client
		err := match.initClient(mc.Registry)
		if err != nil {
			return nil, err
		}

		monitorClients[name] = mt
	}
	return mt, err
}

// 关闭某一监控器
func Close(name string) {
	delete(monitorClients, name)
}

// 清空监控器
func Clear() {
	monitorClients = make(map[string]*monitor)
}
