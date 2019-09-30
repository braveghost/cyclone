package cyclone

import (
	"context"
	healthy "cyclone/healthy"
	"fmt"
	logging "github.com/braveghost/joker"
	"github.com/braveghost/rogue"
	"github.com/micro/go-grpc"
	"github.com/micro/go-micro"
	"github.com/micro/go-micro/client"
	"github.com/pkg/errors"
	"math/rand"
	"sync"
	"time"
)

var (
	defaultHealthInterval = int64(5)
	defaultMasterCount    = 1
	alarmFunc             func(string)
)

var (
	MicroServiceIsNullErr        = errors.New("Micro service is null")
	MicroServiceHealthHandlerErr = errors.New("Micro service health handler is error")
)

func SetAlarmFunc(fn func(string)) {
	alarmFunc = fn
}

type checker func(*ServiceBuilder)

type ServiceBuilder struct {
	name           string
	tags           map[string]string
	count          int
	heartBeat      *rogue.HeartBeat
	status         bool
	interval       int64
	healthFunc     checker
	service        micro.Service
	start          chan *struct{}
	alert          chan string
	config         chan *Setting // 备用, 更新配置使用, 免配置中心侵入
	error          chan error
	lock           sync.Mutex
	registerConf   *RegistryConf
	registerOption micro.Option
}

// 加载集群 tag
func (sb *ServiceBuilder) getTag(ops ...micro.Option) []micro.Option {
	sb.tags[clusterKey] = clusterMaster
	ops = append(ops, micro.Metadata(sb.tags))
	return ops
}

// 初始化注册中心
func (sb *ServiceBuilder) getRegister(ops ...micro.Option) ([]micro.Option, error) {
	reg, err := NewRegistry(sb.registerConf)

	if err != nil {
		return ops, err
	}

	_, err = reg.GetService(sb.name)
	if err != nil {
		return ops, err
	}
	sb.registerOption = micro.Registry(reg)
	ops = append(ops, sb.registerOption)
	return ops, err

}

func (sb *ServiceBuilder) extendOps(ops ...micro.Option) ([]micro.Option, error) {
	var err error
	ops, err = sb.getRegister(ops...)
	if err != nil {
		return ops, err
	}
	ops = sb.getTag(ops...)
	return ops, nil
}

func (sb *ServiceBuilder) discovery() ([]*SrvInfo, error) {
	m, err := NewMonitor("ServiceBuilderDiscovery", &MonitorConfig{
		Registry: sb.registerConf,
	})
	if err != nil {
		return nil, err
	}
	pp := m.HealthService(sb.name)
	return pp.Active, nil
}

var healthClient healthy.CycloneHealthyService

// todo 单例不要 error
func (sb *ServiceBuilder) GetHealthyClient() (healthy.CycloneHealthyService, error) {
	if healthClient == nil {
		srv := grpc.NewService(
			sb.registerOption,
		).Client()
		err := srv.Init()
		if err != nil {
			return nil, err
		}
		healthClient = healthy.NewCycloneHealthyService(sb.name, srv)
	}
	return healthClient, nil
}

// 监控提醒
func (sb *ServiceBuilder) alarm() {
	if alarmFunc == nil {
		logging.Warnw("Cyclone.ServiceBuilder.AlarmFuncIsNull.Warn", "status", "alarm not run")
		return
	}
	logging.Infow("Cyclone.ServiceBuilder.AlarmFunc.Info", "status", "alarm is run")

	for {
		select {
		case msg := <-sb.alert:
			alarmFunc(msg)
		}

	}
}
func (sb *ServiceBuilder) Run(ops ...micro.Option) error {
	go sb.alarm()
	var err error

	ops, err = sb.extendOps(ops...)
	if err != nil {
		return err
	}
	//
	go sb.healthFunc(sb)
	srv := sb.service

	if srv != nil {
		select {
		case <-sb.start:
			srv.Init(ops...)
			return srv.Run()
		case err = <-sb.error:
			return err
		}
	}
	return MicroServiceIsNullErr
}

type Setting struct {
	Threshold int64 // 计数器阈值, 溢出后表服务不可用
	Duration  int64 // 计数器统计时间周期, 距离当前多少秒内
	Masters   int
	Interval  int64
	Tags      map[string]string
	Registry  *RegistryConf
}

func NewServiceBuilder(srv micro.Service, fn checker, hdlr healthy.CycloneHealthyHandler, set *Setting) (*ServiceBuilder, error) {
	if fn == nil {
		fn = defaultCheckerHealth
	}

	if set.Masters <= 0 {
		set.Masters = defaultMasterCount
	}
	if set.Tags == nil {
		set.Tags = make(map[string]string)
	}
	if set.Interval <= 0 {
		set.Interval = defaultHealthInterval
	}

	err := healthy.RegisterCycloneHealthyHandler(srv.Server(), hdlr)
	if err != nil {
		return nil, MicroServiceHealthHandlerErr
	}
	return &ServiceBuilder{
		count:        set.Masters,
		tags:         set.Tags,
		service:      srv,
		name:         srv.Server().Options().Name,
		interval:     set.Interval,
		healthFunc:   fn,
		registerConf: set.Registry,
		start:        make(chan *struct{}, 1),
		error:        make(chan error),
		alert:        make(chan string),
		lock:         sync.Mutex{},
		heartBeat:    rogue.NewHeartBeat(set.Threshold, set.Duration),
	}, nil
}

// 数量检查, master 节点如果小于指定数量就启动, 否则等待并监听服务状态状态
func defaultCheckerHealth(sb *ServiceBuilder) {
	max := sb.interval * int64(time.Second)
	min := int64(0.8 * float64(max))

	for {
		// 服务发现
		act, err := sb.discovery()
		if err != nil {
			logging.Errorw("Cyclone.ServiceBuilder.HealthFunc.Discovery.Error",
				"err", err)
			sb.error <- err
			return
		}

		// 服务数量检查
		ct := len(act)
		if ct < sb.count {
			logging.Infow("Cyclone.ServiceBuilder.HealthFunc.Count.StartService.Info",
				"config_health_count", sb.count, "center_health_count", ct)
			sb.start <- &struct{}{}
			return
		}
		logging.Debugw("Cyclone.ServiceBuilder.HealthFunc.Verify.Debug",
			"config_health_count", sb.count, "center_health_count", ct)

		// 健康检查
		for _, s := range act {
			if val, ok := s.Tags[clusterKey]; ok && val == clusterMaster {
				cli, err := sb.GetHealthyClient()
				if err != nil {
					logging.Errorw("Cyclone.ServiceBuilder.HealthFunc.GetHealthyClient.Error",
						"err", err)
					sb.error <- err
					return
				}
				res, err := cli.Healthy(context.Background(), &healthy.CycloneRequest{},
					func(option *client.CallOptions) {
						option.Address = []string{s.Address}
					},
				)
				if err != nil {
					logging.Errorw("Cyclone.ServiceBuilder.HealthFunc.Healthy.Error",
						"err", err)
					sb.error <- err
					return
				}

				sb.heartBeat.AddSignal(&healthSignal{res, sb.alert})
				if sb.heartBeat.Status() {
					//	todo 关闭服务
					logging.Infow("Cyclone.ServiceBuilder.HealthFunc.Zombies.StartService.Info",
						"config_health_count", sb.count, "center_health_count", ct)
					sb.start <- &struct{}{}
					return
				}
			} else {
				logging.Warnw("Cyclone.ServiceBuilder.HealthFunc.Tag.Warn",
					"tags", s.Tags)
			}
		}

		time.Sleep(time.Duration(RandomInt64n(min, max)))

	}

}

type healthSignal struct {
	res *healthy.CycloneResponse
	ch  chan string
}

// 心跳状态, 每一次就计算一次
func (hc *healthSignal) Status() bool {
	res := hc.res
	switch res.Code {
	case healthy.CycloneResponse_Healthy:
		//	不做任何操作
		return false
	case healthy.CycloneResponse_Zombies:
		// 计数统计
		return true
	case healthy.CycloneResponse_Sick:
		//	告警
		hc.ch <- alertSickMsg(res.Response)
		return false
	default:
		//	告警
		hc.ch <- fmt.Sprintf(serviceStatusCodeErrorMsg, res.Response.Name, res.Code)
		return false
	}

}

var (
	serviceStatusHeadMsg = `
================================ [ServiceStatus] %v ================================
	`

	serviceStatusCodeErrorMsg = `
================================ [ServiceStatus] %v ================================
	[Code]: %v
    [Error]: response code is not support
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

func RandomInt64n(min, max int64) int64 {
	rand.Seed(time.Now().UnixNano())
	return rand.Int63n(max-min) + min
}
