package cyclone

import (
	healthy "github.com/braveghost/cyclone/healthy"
	logging "github.com/braveghost/joker"
	"github.com/braveghost/meteor/errutil"
	"github.com/micro/go-micro"
	"github.com/micro/go-micro/registry"
	"github.com/pkg/errors"
	"math/rand"
	"sync"
	"time"
)

const (
	defaultHealthInterval = int64(5)
	defaultMasterCount    = 1
)

var (
	alarmFunc func(string)
	// Error
	MicroServiceIsNullErr        = errors.New("Micro service is null")
	MicroServiceHealthHandlerErr = errors.New("Micro service health handler is error")
	MicroServicesErr             = errors.New("Micro service is null")
)

func SetAlarmFunc(fn func(string)) {
	alarmFunc = fn
}

type checker func(*ServiceBuilder)

type ServiceBuilder struct {
	*Setting
	Name         string
	status       bool
	HealthFunc   checker
	Service      micro.Service
	start        chan *struct{}
	sameStartFns []func()
	alert        chan string
	error        chan error
	config       chan *Setting // 备用, 更新配置使用, 免配置中心侵入
	lock         sync.Mutex
	register     registry.Registry
}

// 加载集群 tag
func (sb *ServiceBuilder) getTag(ops ...micro.Option) []micro.Option {
	sb.Tags[clusterKey] = clusterMaster
	ops = append(ops, micro.Metadata(sb.Tags))
	return ops
}

// 初始化注册中心
func (sb *ServiceBuilder) getRegister(ops ...micro.Option) ([]micro.Option, error) {

	if sb.register == nil {
		register, err := NewRegistry(sb.Registry)
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

func (sb *ServiceBuilder) RegisterSameStart(fns ...func()) {
	sb.sameStartFns = fns

}
func (sb *ServiceBuilder) Run(ops ...micro.Option) error {
	go sb.alarm()
	var err error

	ops, err = sb.extendOps(ops...)
	if err != nil {
		return err
	}
	go sb.HealthFunc(sb)
	srv := sb.Service

	if srv != nil {
		select {
		case <-sb.start:
			srv.Init(ops...)
			for _, fn := range sb.sameStartFns {
				go fn()
			}
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

func NewServiceBuilder(srv micro.Service, fn checker, set *Setting) (*ServiceBuilder, error) {

	if srv == nil {
		return nil, MicroServicesErr
	}
	if set == nil {
		set = &Setting{}
	}

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

	healthy.RegistryHealthy(nil)

	err := healthy.RegisterCycloneHealthyHandler(srv.Server(), healthy.HealthyHandler{})
	if err != nil {
		return nil, MicroServiceHealthHandlerErr
	}

	return &ServiceBuilder{
		Setting:    set,
		Name:       srv.Server().Options().Name,
		Service:    srv,
		HealthFunc: fn,
		start:      make(chan *struct{}, 1),
		error:      make(chan error),
		alert:      make(chan string),
		lock:       sync.Mutex{},
	}, nil
}

// 数量检查, master 节点如果小于指定数量就启动, 否则等待并监听服务状态状态
func defaultCheckerHealth(sb *ServiceBuilder) {
	max := sb.Interval * int64(time.Second)
	min := int64(0.8 * float64(max))
	shc := &SrvHealthyConfig{
		Name:      sb.Name,
		Duration:  sb.Interval * 5,
		Threshold: 3,
	}

	for {
		m, err := NewMonitor(sb.register, sb.MonitorConfig)
		if err != nil {
			//	todo 告警
			sb.error <- err
			return
		}

		hi, err := m.GetHealth(shc)
		logging.Debugw("Cyclone.ServiceBuilder.CheckerHealth.GetHealth.Debug",
			"masters", sb.Masters, "count", hi.Count, "health_count", hi.HealthCount, "err", err)
		if errutil.Is(MonitorSrvNotFoundErr, err) {
			sb.start <- &struct{}{}
			logging.Info("Cyclone.ServiceBuilder.CheckerHealth.MonitorSrvNotFoundErr.Start.Info")
			return

		} else {
			if hi.Count < sb.Masters {
				sb.start <- &struct{}{}
				logging.Info("Cyclone.ServiceBuilder.CheckerHealth.HeartHealthCountBeat.Start.Info")
				return
			}
		}
		logging.Info("Cyclone.ServiceBuilder.CheckerHealth.Continue.Info")

		time.Sleep(time.Duration(RandomInt64n(min, max)))
	}

}

func RandomInt64n(min, max int64) int64 {
	rand.Seed(time.Now().UnixNano())
	return rand.Int63n(max-min) + min
}
