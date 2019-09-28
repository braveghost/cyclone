package cyclone

import (
	"fmt"
	"github.com/micro/go-micro"
	"github.com/pkg/errors"
	"sync"
	"time"
)

var (
	MicroServiceIsNullErr = errors.New("Micro service is null")
	MicroMasterSetErr     = errors.New("Micro master service is 0")
)

type checker func(*ServiceBuilder)

type ServiceBuilder struct {
	name          string
	tags          map[string]string
	count         int
	status        bool
	countFn       checker
	interval      int
	healthFn      checker
	service       micro.Service
	startCh       chan *struct{}
	errorCountCh  chan error
	errorHealthCh chan error
	lock     sync.Mutex
	register      *RegistryConf
}

func (sb *ServiceBuilder) getTag(ops ...micro.Option) []micro.Option {
	sb.tags[clusterKey] = clusterMaster
	ops = append(ops, micro.Metadata(sb.tags))
	return ops
}

func (sb *ServiceBuilder) getRegister(ops ...micro.Option) []micro.Option {
	reg, err := NewRegistry(sb.register)
	if err != nil {
		sb.errorCountCh <- err
		return ops
	} else {
		ops = append(ops, micro.Registry(reg))
	}

	return ops
}

func (sb *ServiceBuilder) extendOps(ops ...micro.Option) []micro.Option {
	ops = sb.getRegister(ops...)
	ops = sb.getTag(ops...)
	fmt.Println(ops)
	return ops
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
func (sb *ServiceBuilder) Run(errFn func(err error), ops ...micro.Option) error {
	go sb.countFn(sb)
	go sb.healthFn(sb)
	srv := sb.service
	if srv != nil {
		for {

			select {
			case <-sb.startCh:
				fmt.Println("ccccccc", sb)

				srv.Init(sb.extendOps(ops...)...)
				fmt.Println("xxxxxxx")
				return srv.Run()
			case err := <-sb.errorHealthCh:
				fmt.Println("errorHealthCh")

				if errFn != nil {

				errFn(err)
				}
			case err := <-sb.errorCountCh:
				fmt.Println("errorCountCh")
				return err

			}
		}
	}
	return MicroServiceIsNullErr
}

type Setting struct {
	Service   micro.Service
	CountFn   checker
	HealthFn  checker
	Threshold int64 // 计数器阈值, 溢出后表服务不可用
	Duration  int64 // 计数器统计时间周期, 距离当前多少秒内
	Masters   int
	Interval  int
	Tags      map[string]string
	Registry  *RegistryConf
}

func NewSrvSignal(set *Setting) (*ServiceBuilder, error) {
	if set.CountFn == nil {
		set.CountFn = defaultCheckerCount
	}
	if set.HealthFn == nil {
		set.HealthFn = defaultCheckerHealth
	}

	if set.Masters <= 0 {
		return nil, MicroMasterSetErr
	}
	if set.Tags == nil {
		set.Tags = make(map[string]string)
	}

	return &ServiceBuilder{
		count:         set.Masters,
		tags:          set.Tags,
		service:       set.Service,
		name:          set.Service.Server().Options().Name,
		countFn:       set.CountFn,
		healthFn:      set.HealthFn,
		register:      set.Registry,
		startCh:       make(chan *struct{}),
		errorCountCh:  make(chan error),
		errorHealthCh: make(chan error),
		lock:     sync.Mutex{},
	}, nil
}

// 默认数量检查, master 节点如果小于指定数量就启动, 否则等待并监听服务状态状态
func defaultCheckerCount(sb *ServiceBuilder) {
	act, err := sb.discovery()
	if err != nil {
		sb.errorCountCh <- err
	} else {
		if len(act) < sb.count {
			sb.startCh <- &struct{}{}
		}
	}

}

// slave 节点, 如果
func defaultCheckerHealth(sb *ServiceBuilder) {
	for {
		act, err := sb.discovery()
		if err != nil {
			sb.errorHealthCh <- err
		} else {
			ct := 0
			for _, s := range act {
				for key, tag := range s.Tags {
					// todo 此处应该是健康检查+心跳计数, 待完善
					if key == clusterKey && tag == clusterMaster {
						ct += 1
					}
				}
			}
			if ct < sb.count {
				sb.startCh <- &struct{}{}
				break
			}
		}
		time.Sleep(time.Duration(sb.interval) * time.Second)
	}

}

//
//// 心跳状态, 每一次就计算一次
//func (sb *SignalBuilder) Status() bool {
//	if !sb.statusFlag {
//		var count int
//		for _, s := range sb.Services {
//			for key, tag := range s.Tags {
//				if tag == masterTag {
//					count++
//				}
//			}
//		}
//		if count < sb.MasterCount {
//			sb.status = true
//		}
//		sb.statusFlag = true
//	}
//	return sb.status
//
//}
//
//func example() {
//	hb := rogue.NewHeartBeat(5, 10)
//	hb.AddSignal(&SrvSignal{false})
//	hb.AddSignal(&SrvSignal{true})
//	hb.AddSignal(&SrvSignal{false})
//	hb.AddSignal(&SrvSignal{false})
//	hb.AddSignal(&SrvSignal{false})
//	hb.AddSignal(&SrvSignal{false})
//	hb.AddSignal(&SrvSignal{false})
//	time.Sleep(time.Second * 11)
//	//hb.AddBeat(&SrvSignal{false})
//	//hb.AddBeat(&SrvSignal{false})
//	fmt.Println(hb.Status())
//}
