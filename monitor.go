package cyclone

import (
	"context"
	"fmt"
	healthy "github.com/braveghost/cyclone/healthy"
	logging "github.com/braveghost/joker"
	"github.com/braveghost/rogue"
	"github.com/micro/go-micro"
	"github.com/micro/go-micro/client"
	"github.com/micro/go-micro/registry"
	"github.com/micro/go-micro/service/grpc"
	"github.com/pkg/errors"
	"strings"
	"sync"
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

	MatchTypeEqual  // 服务健康数量相等
	MatchTypeScope  // 服务健康数量范围
)

const (
	MonitorTypeAddress MonitorType = iota // ip地址
	MonitorTypeCount                      // 服务健康数量相等

)

var (
	defaultMonitorConf *MonitorConfig
	defaultMonitor     *Monitor
	heartBeats         = make(map[string]*rogue.HeartBeat)

	headMsg = `
================================ [%v] %v ================================
------BaseInfo:
	[Count]: %v		[Health]: %v		[Error]: %v
	`

	healthyMsg = `
------Healthy - N%v:
	[Id]: %v
	[Address]: %v
	[Error]: %v
`

	sickMsg = `
------Sick - N%v:
	[MatchField]: %v
`
)

var (
	MonitorSrvCountEqualErr   = errors.New("monitor service health count error")
	MonitorSrvMatchErr        = errors.New("monitor service health match error")
	MonitorSrvNotFoundErr     = errors.New("monitor service not found")
	MonitorSrvCountPeakErr    = errors.New("monitor service count less than peak")
	MonitorSrvCountValleyErr  = errors.New("monitor service count more than valley")
	MonitorSrvCountScopeErr   = errors.New("monitor service config count scope error")
	MonitorMatchFuncChoiceErr = errors.New("monitor service match choice function error")
)

func init() {
	defaultMonitorConf = &MonitorConfig{Name: "default"}
}

func HasDefaultMonitor() bool {
	if defaultMonitor == nil {
		return false
	}
	return true
}

func McCountEqual(name string, services ...*SrvConfigInfo) *MonitorConfig {
	return &MonitorConfig{
		Name:     name,
		Type:     MonitorTypeCount,
		Services: services,
		Match:    MatchTypeEqual,
	}
}

func McCountScope(name string, services ...*SrvConfigInfo) *MonitorConfig {
	return &MonitorConfig{
		Name:     name,
		Type:     MonitorTypeCount,
		Services: services,
		Match:    MatchTypeScope,
	}
}

type MonitorConfig struct {
	Name     string           // 配置名称, 根据配置名称单例客户端
	Type     MonitorType      // 主机名Or地址Or数量, 节点状态
	Services []*SrvConfigInfo // 服务信息
	Match    MatchType        // 匹配类型
}

type SrvConfigInfo struct {
	Name      string   // 服务名称
	Hosts     []string // 精确匹配主机
	Peak      int      // 服务最大量, count+=peak
	Valley    int      // 服务服务最小量, count+=valley，valley<count<peak，组成一个范围值
	Duration  int64    // 健康检查心跳持续时间
	Threshold int64    // 健康检查删除节点阈值, 即持续时间内到达阈值删除节点
}

type SrvHealthyConfig struct {
	Name      string // 服务名称
	Duration  int64  // 健康检查心跳持续时间
	Threshold int64  // 健康检查删除节点阈值, 即持续时间内到达阈值删除节点
}

func (sci SrvConfigInfo) Count() int {
	if sci.Peak > 0 {
		return sci.Peak
	}
	return len(sci.Hosts)
}

func InitConfig(mc *MonitorConfig) {
	defaultMonitorConf = mc
}

type healthInfo struct {
	Count       int // 注册中心的对应服务总数量, 暂未使用
	HealthCount int
	Healthy     []*SrvInfo
	Sick        []*SrvInfo
}

type SrvInfo struct {
	//Tag    string   // 通过tag 匹配
	Id      string
	Address string
	Tags    map[string]string
	ApiInfo []*healthy.ApiInfo
	Error   error
}

type Monitor struct {
	registry.Registry
	conf *MonitorConfig
	matchFunc
	alert chan string
	error chan error
	lock  sync.Mutex
}

func (m *Monitor) getHost(si *SrvInfo) string {
	switch m.conf.Type {
	case MonitorTypeAddress:
		return si.Address
	}
	return ""
}

func (m *Monitor) matchErr(tag string) error {
	switch m.conf.Type {
	case MonitorTypeAddress:
		return errors.Wrapf(MonitorSrvMatchErr, "Match %v address", tag)
	}
	return MonitorSrvMatchErr
}

type matchInnerFunc func(map[string]*SrvInfo, string) bool

func (m *Monitor) matchString(tag, fnTag string, info *SrvConfigInfo, fn matchInnerFunc) (string, error) {

	return m.match(
		tag,
		fnTag,
		info,
		func(hi *healthInfo) (string, error) {
			var (
				msg string
				err error
			)
			getHost := m.getHost

			var hostSrv = make(map[string]*SrvInfo)

			for i, l := range hi.Healthy {
				hostSrv[getHost(l)] = hi.Healthy[i]
				msg += getActiveMsg(i+1, l.Id, l.Address, "null")
			}

			for ii, ll := range info.Hosts {

				if !fn(hostSrv, ll) {
					msg += getDeathMsg(ii+1, ll)
					err = m.matchErr(fnTag)
				} else {

					delete(hostSrv, ll)
				}
			}
			return msg, err
		},

	)
}

func (m *Monitor) match(tag, fnTag string, info *SrvConfigInfo, fn func(*healthInfo) (string, error)) (string, error) {
	healthCount := -1
	var msg string

	hi, err := m.GetHealth(&SrvHealthyConfig{
		info.Name,
		info.Duration,
		info.Threshold})

	if err == nil {
		healthCount = hi.HealthCount
		msg, err = fn(hi)
	}
	return getHeadMsg(tag, info.Name, info.Count(), healthCount, err) + msg, err
}

// 健康状态检查
func (m *Monitor) GetHealth(info *SrvHealthyConfig) (*healthInfo, error) {
	hi := &healthInfo{}
	hs, err := m.GetService(info.Name)
	if err == nil {
		if len(hs) == 0 {
			err = errors.Wrapf(MonitorSrvNotFoundErr, "Service '%s'", info.Name)
		} else {
			hi, err = m.health(hs, info)
		}
	}
	return hi, err
}

// 服务关键字完全匹配
func (m *Monitor) matchFull(info *SrvConfigInfo) (string, error) {

	return m.matchString(
		"ServiceMatchFull",
		"full",
		info,
		func(hs map[string]*SrvInfo, host string) bool {
			_, ok := hs[host]
			return ok
		})
}

// 服务关键字前缀匹配
func (m *Monitor) matchPrefix(info *SrvConfigInfo) (string, error) {
	return m.matchString(
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
func (m *Monitor) matchIn(info *SrvConfigInfo) (string, error) {

	return m.matchString(
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

func (m *Monitor) MatchScope(info *SrvConfigInfo) (*healthInfo, error) {
	hi, err := m.GetHealth(&SrvHealthyConfig{
		info.Name,
		info.Duration,
		info.Threshold})
	if err == nil {
		if hi.HealthCount > info.Peak {
			err = MonitorSrvCountPeakErr
		} else if hi.HealthCount < info.Valley {
			err = MonitorSrvCountValleyErr

		}

		if err != nil {
			err = errors.Wrapf(err, "peak=%v,valley=%v,health=%v", info.Peak, info.Valley, hi.HealthCount)
		}
	}
	return hi, err
}

func (m *Monitor) matchScope(info *SrvConfigInfo) (string, error) {

	var msg string
	hi, err := m.MatchScope(info)

	healthCount := -1
	if hi != nil {
		healthCount = hi.HealthCount
		for idx, l := range hi.Healthy {
			msg += getActiveMsg(idx+1, l.Id, l.Address, "null")
		}
	}
	return getHeadMsg("ServiceCountScope", info.Name, info.Count(), healthCount, err) + msg, err
}

// 服务数量等值匹配
func (m *Monitor) MatchEqual(info *SrvConfigInfo) (*healthInfo, error) {
	hi, err := m.GetHealth(&SrvHealthyConfig{
		info.Name,
		info.Duration,
		info.Threshold})
	if err == nil {
		if hi.HealthCount != info.Count() {
			err = MonitorSrvCountEqualErr
		}
	}
	return hi, err
}

// 服务数量等值匹配
func (m *Monitor) matchEqual(info *SrvConfigInfo) (string, error) {

	var msg string
	hi, err := m.MatchEqual(info)
	healthCount := -1
	if hi != nil {

		healthCount = hi.HealthCount
		for idx, l := range hi.Healthy {
			msg += getActiveMsg(idx+1, l.Id, l.Address, "null")
		}
	}
	return getHeadMsg("ServiceCountEqual", info.Name, info.Count(), healthCount, err) + msg, err
}

var healthClients = make(map[string]healthy.CycloneHealthyService)

// todo 单例不要 error
func (m *Monitor) GetHealthyClient(name string) (healthy.CycloneHealthyService, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	hc, ok := healthClients[name]
	if !ok {

		srv := grpc.NewService(
			micro.Registry(m.Registry),
		).Client()
		err := srv.Init()
		if err != nil {
			return nil, err
		}
		hc = healthy.NewCycloneHealthyService(name, srv)
		healthClients[name] = hc
	}
	return hc, nil
}

// 删除节点, 程序非假死状态, 服务节点会重连注册中心
func (m *Monitor) removeNode(name string, node *registry.Node) error {

	return m.Deregister(&registry.Service{
		Name: name,
		Nodes: []*registry.Node{
			node,
		}},
	)
}

// 健康状态检查, 不是监控当前服务和配置是否一致
func (m *Monitor) health(srv []*registry.Service, info *SrvHealthyConfig) (*healthInfo, error) {
	var hi = &healthInfo{}
	var (
		healthCount int
		count       int
	)

	for _, service := range srv {
		for _, node := range service.Nodes {
			count += 1

			tags := node.Metadata
			val, ok := tags[clusterKey]
			if !(ok && val == clusterMaster) {
				// 不是依赖 cyclone 启动的服务
				logging.Warnw("Cyclone.ServiceBuilder.HealthFunc.Tag.Warn",
					"tags", tags, "node", node)
				continue
			}

			res, err := m.getHealthyRes(info.Name, node.Address)

			if err != nil {
				return hi, err
			}
			// 计数
			hb := m.getHeartBeat(info)
			st := hb.AddSignal(&healthSignal{res})
			// 状态计数
			ok, err = hb.Status()
			if err != nil {
				return nil, err
			}
			if ok {
				if res.Code != healthy.CycloneResponse_Healthy {
					_ = m.closeService(info, node)
				}
			}
			// 健康计数
			si := &SrvInfo{
				Id:      node.Id,
				Address: node.Address,
				Tags:    node.Metadata,
				Error:   st,
			}

			if res.Response != nil {
				si.ApiInfo = res.Response.ApiInfo
			}
			if st == nil {
				healthCount += 1
				hi.Healthy = append(hi.Healthy, si)
			} else {
				hi.Sick = append(hi.Sick, si)
			}
		}
	}

	hi.HealthCount = healthCount
	hi.Count = count
	return hi, nil
}

// 获取健康信息
func (m *Monitor) getHealthyRes(name, addr string) (*healthy.CycloneResponse, error) {
	cli, err := m.GetHealthyClient(name)
	if err != nil {
		logging.Errorw("Cyclone.Monitor.Health.GetHealthyClient.Error",
			"err", err)
		return nil, err
	}
	res, err := cli.Healthy(context.Background(), &healthy.CycloneRequest{},
		func(option *client.CallOptions) {
			option.Address = []string{addr}
		},
	)

	if err != nil {
		logging.Errorw("Cyclone.Monitor.Health.ClientQuery.Healthy.Error",
			"err", err)
		if res == nil {
			res = &healthy.CycloneResponse{}
		}
	}
	return res, err
}

// 关闭对应服务
func (m *Monitor) closeService(info *SrvHealthyConfig, node *registry.Node) error {

	cli, err := m.GetHealthyClient(info.Name)
	if err != nil {
		logging.Errorw("Cyclone.Monitor.Health.GetHealthyClient.Error",
			"err", err)
		return err
	}

	// 关闭服务, 删除注册中心的内容, 通过 Close 接口关闭服务
	err = m.removeNode(info.Name, node)
	if err != nil {
		logging.Errorw("Cyclone.Monitor.Health.RemoveNode.Error",
			"node_name", info.Name, "node", node, "err", err)
	} else {
		logging.Info("Cyclone.Monitor.Health.RemoveNode.Info")
	}

	_, err = cli.Close(context.Background(), &healthy.CycloneRequest{},
		func(option *client.CallOptions) {
			option.Address = []string{node.Address}
		},
	)
	if err != nil {
		logging.Errorw("Cyclone.Monitor.Health.CloseNode.Error",
			"node_name", info.Name, "node", node, "err", err)
	} else {
		logging.Infow("Cyclone.Monitor.Health.CloseNode.Info",
			"node_name", info.Name)
	}
	return err
}

func (m *Monitor) getHeartBeat(info *SrvHealthyConfig) *rogue.HeartBeat {
	var hb *rogue.HeartBeat
	if tmpHb, ok := heartBeats[info.Name]; ok {
		hb = tmpHb
	} else {
		hb = rogue.NewHeartBeat(info.Threshold, info.Duration, rogue.ECounterTypeRedis, rogue.UnitSecond)
		heartBeats[info.Name] = hb
	}
	return hb
}

// 生成消息头
func getHeadMsg(tag string, name string, count, health int, err error) string {
	return fmt.Sprintf(headMsg, tag, name, count, health, err)
}

// 生成活跃服务消息
func getActiveMsg(idx int, node, addr string, err string) string {
	return fmt.Sprintf(healthyMsg, idx, node, addr, err)
}

// 生成不存在的服务的消息
func getDeathMsg(idx int, host string) string {
	return fmt.Sprintf(sickMsg, idx, host)
}

// 规则校验
func (m *Monitor) monitorService(conf *MonitorConfig) []string {
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
func (m *Monitor) Run() ([]string, error) {
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
func (m *Monitor) matchChoice() matchFunc {

	tp := m.conf.Type
	var f matchFunc
	if tp != MonitorTypeCount && tp != MonitorTypeAddress {
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

// 创建监控器
func NewMonitor(reg registry.Registry, mc *MonitorConfig) (*Monitor, error) {
	return newMonitor(reg, mc)
}

func (m *Monitor) UpdateRegistry(reg registry.Registry) error {

	m.lock.Lock()
	defer m.lock.Unlock()

	m.Registry = reg
	healthClients = make(map[string]healthy.CycloneHealthyService)
	return nil
}

func (m *Monitor) UpdateConfig(mc *MonitorConfig) error {

	m.lock.Lock()
	defer m.lock.Unlock()

	m.conf = mc

	fn := m.matchChoice()
	if fn == nil {
		return errors.Wrap(MonitorMatchFuncChoiceErr, "monitor update config")
	}
	m.matchFunc = fn

	return nil
}

// default monitor
func GetMonitor(reg registry.Registry, mc *MonitorConfig) (*Monitor, error) {
	if defaultMonitor != nil {
		return defaultMonitor, nil
	}
	var err error
	defaultMonitor, err = newMonitor(reg, mc)
	return defaultMonitor, err

}

// 创建监控器，
func newMonitor(reg registry.Registry, mc *MonitorConfig) (*Monitor, error) {
	if mc == nil {
		mc = defaultMonitorConf
	}
	var err error
	var mt *Monitor
	mt = &Monitor{
		Registry: reg,
		conf:     mc,
		alert:    make(chan string),
		error:    make(chan error),
		lock:     sync.Mutex{},
	}
	fn := mt.matchChoice()
	if fn == nil {
		return nil, MonitorMatchFuncChoiceErr
	}
	mt.matchFunc = fn
	return mt, err
}

type healthSignal struct {
	res *healthy.CycloneResponse
}

var (
	MonitorServiceSickErr    = errors.New("monitor service sick")
	MonitorServiceZombiesErr = errors.New("monitor service zombies")
	MonitorServiceCodeErr    = errors.New("monitor service response code is not support")
)
// 心跳状态, 每一次就计算一次
func (hc *healthSignal) Status() error {
	res := hc.res
	switch res.Code {
	case healthy.CycloneResponse_Healthy:
		//	不做任何操作
		return nil
	case healthy.CycloneResponse_Zombies:
		// 计数统计
		return MonitorServiceZombiesErr
	case healthy.CycloneResponse_Sick:
		//	告警
		//hc.ch <- alertSickMsg(res.Response)
		return MonitorServiceSickErr
	default:
		//	告警
		//hc.ch <- fmt.Sprintf(serviceStatusCodeErrorMsg, res.Response.Name, res.Code)
		return MonitorServiceCodeErr
	}

}

var (
	serviceStatusHeadMsg = `
================================ [ServiceStatus] %v ================================
	`

	serviceStatusApiMsg = `
------N%v:
	[Api]: %v		[Error]: %v
`
)

func alertSickMsg(ss *healthy.ServiceStatus) string {
	msg := fmt.Sprintf(serviceStatusHeadMsg, ss.Name)

	for i, l := range ss.ApiInfo {
		msg = msg + fmt.Sprintf(serviceStatusApiMsg, i+1, l.Api, l.Error)
	}
	return msg
}
