package cyclone

import (
	logging "github.com/braveghost/joker"
	"github.com/micro/go-micro"
	"github.com/pkg/errors"
	"math/rand"
	"sync"
	"time"
)

var (
	defaultHealthInterval = int64(5)
	defaultMasterCount    = 1
)

var (
	MicroServiceIsNullErr = errors.New("Micro service is null")
)

type checker func(*ServiceBuilder)

type ServiceBuilder struct {
	name       string
	tags       map[string]string
	count      int
	status     bool
	interval   int64
	healthFunc checker
	service    micro.Service
	start      chan *struct{}
	config     chan *Setting // 备用, 更新配置使用, 免配置中心侵入
	error      chan error
	lock       sync.Mutex
	register   *RegistryConf
}

// 加载集群 tag
func (sb *ServiceBuilder) getTag(ops ...micro.Option) []micro.Option {
	sb.tags[clusterKey] = clusterMaster
	ops = append(ops, micro.Metadata(sb.tags))
	return ops
}

// 初始化注册中心
func (sb *ServiceBuilder) getRegister(ops ...micro.Option) ([]micro.Option, error) {
	reg, err := NewRegistry(sb.register)

	if err != nil {
		return ops, err
	}

	_, err = reg.GetService(sb.name)
	if err != nil {
		return ops, err
	}

	ops = append(ops, micro.Registry(reg))
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
	sb.lock.Lock()
	defer sb.lock.Unlock()

	m, err := NewMonitor("ServiceBuilderDiscovery", &MonitorConfig{
		Registry: sb.register,
	})
	if err != nil {
		return nil, err
	}
	pp := m.HealthService(sb.name)
	return pp.Active, nil
}
func (sb *ServiceBuilder) Run(ops ...micro.Option) error {
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

func NewServiceBuilder(srv micro.Service, fn checker, set *Setting) (*ServiceBuilder, error) {
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

	return &ServiceBuilder{
		count:      set.Masters,
		tags:       set.Tags,
		service:    srv,
		name:       srv.Server().Options().Name,
		interval:   set.Interval,
		healthFunc: fn,
		register:   set.Registry,
		start:      make(chan *struct{}, 1),
		error:      make(chan error),
		lock:       sync.Mutex{},
	}, nil
}

// 数量检查, master 节点如果小于指定数量就启动, 否则等待并监听服务状态状态
func defaultCheckerHealth(sb *ServiceBuilder) {
	max := sb.interval * int64(time.Second)
	min := int64(0.8 * float64(max))
	for {
		act, err := sb.discovery()
		if err != nil {
			logging.Infow("Cyclone.ServiceBuilder.HealthFunc.Discovery.Error",
				"err", err)
			sb.error <- err
		} else {
			if len(act) < sb.count {
				logging.Infow("Cyclone.ServiceBuilder.HealthFunc.Count.StartService.Info",
					"config_health_count", sb.count, "center_health_count", len(act))
				sb.start <- &struct{}{}
				break
			}

			ct := 0
			for _, s := range act {
				for key, tag := range s.Tags {
					// todo 此处应该是健康检查+心跳计数, 同主机检查，优先等待其他主机启动，防止所有节点启动在同一主机待完善
					if key == clusterKey && tag == clusterMaster {
						ct += 1
					}
				}
			}
			if ct < sb.count {
				logging.Infow("Cyclone.ServiceBuilder.HealthFunc.Health.StartService.Info",
					"config_health_count", sb.count, "center_health_count", ct)
				sb.start <- &struct{}{}

				break
			}

			logging.Debugw("Cyclone.ServiceBuilder.HealthFunc.Verify.Debug",
				"config_health_count", sb.count, "center_health_count", ct)
		}
		time.Sleep(time.Duration(RandomInt64n(min, max)))

	}

}

func RandomInt64n(min, max int64) int64 {
	rand.Seed(time.Now().UnixNano())
	return rand.Int63n(max-min) + min
}
