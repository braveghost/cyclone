package cyclone

import (
	healthy "github.com/braveghost/cyclone/healthy"
	logging "github.com/braveghost/joker"
	"github.com/micro/go-micro"
	"github.com/micro/go-micro/registry"
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
	name         string
	tags         map[string]string
	count        int
	status       bool
	interval     int64
	healthFunc   checker
	service      micro.Service
	start        chan *struct{}
	alert        chan string
	error        chan error
	config       chan *Setting // 备用, 更新配置使用, 免配置中心侵入
	lock         sync.Mutex
	registerConf *RegistryConf
	register     registry.Registry
	monitorConf  *MonitorConfig
}

// 加载集群 tag
func (sb *ServiceBuilder) getTag(ops ...micro.Option) []micro.Option {
	sb.tags[clusterKey] = clusterMaster
	ops = append(ops, micro.Metadata(sb.tags))
	return ops
}

// 初始化注册中心
func (sb *ServiceBuilder) getRegister(ops ...micro.Option) ([]micro.Option, error) {

	if sb.register == nil {
		register, err := NewRegistry(sb.registerConf)
		if err != nil {
			return ops, err
		}
		_, err = register.GetService(register.String())
		if err != nil {
			return ops, err
		}
		sb.register = register
	}

	ops = append(ops, micro.Registry(sb.register))
	return ops, nil

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
	//m, err := NewMonitor("ServiceBuilderDiscovery", &MonitorConfig{
	//	Registry: sb.registerConf,
	//})
	//if err != nil {
	//	return nil, err
	//}
	//
	//
	//pp := sb.register.GetService(sb.name)
	return nil, nil
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
	Masters       int
	Interval      int64
	Tags          map[string]string
	Registry      *RegistryConf
	MonitorConfig *MonitorConfig
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
	if hdlr != nil {
		err := healthy.RegisterCycloneHealthyHandler(srv.Server(), healthy.HealthyHandler{})
		if err != nil {
			return nil, MicroServiceHealthHandlerErr
		}
	}

	return &ServiceBuilder{
		count:        set.Masters,
		tags:         set.Tags,
		service:      srv,
		interval:     set.Interval,
		healthFunc:   fn,
		registerConf: set.Registry,
		monitorConf:  set.MonitorConfig,

		name:  srv.Server().Options().Name,
		start: make(chan *struct{}, 1),
		error: make(chan error),
		alert: make(chan string),
		lock:  sync.Mutex{},
	}, nil
}

// 数量检查, master 节点如果小于指定数量就启动, 否则等待并监听服务状态状态
func defaultCheckerHealth(sb *ServiceBuilder) {
	max := sb.interval * int64(time.Second)
	min := int64(0.8 * float64(max))

	for {
		m, err := NewMonitor(sb.register, sb.monitorConf)
		if err != nil {
			//	todo 告警
			sb.error <- err
		}

		_, err = m.Run()
		if err == nil {
			sb.start <- &struct{}{}
			return
		}
		time.Sleep(time.Duration(RandomInt64n(min, max)))
	}

}

func RandomInt64n(min, max int64) int64 {
	rand.Seed(time.Now().UnixNano())
	return rand.Int63n(max-min) + min
}
