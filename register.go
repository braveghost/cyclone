package cyclone

import (
	"github.com/micro/go-micro/registry"
	"github.com/micro/go-micro/registry/mdns"
	"regexp"

	"github.com/micro/go-micro/registry/consul"

	"github.com/pkg/errors"
)


var (
	clusterKey    = "cluster-srv-key"
	clusterMaster = "master-tag"
	clusterSlave  = "slave-tag"
)

func SetClusterKey(key string) {
	clusterKey = key
}

func SetMasterTag(tag string) {
	clusterMaster = tag
}

func SetSlaveTag(tag string) {
	clusterSlave = tag
}



var (
	RegistryConfErr = errors.New("micro registry config error")
	RegistryNameErr = errors.New("micro registry name error")
	RegistryAddrErr = errors.New("micro registry address error")
)

var IpPortRegex = `^(?:(?:[0,1]?\d?\d|2[0-4]\d|25[0-5])\.){3}(?:[0,1]?\d?\d|2[0-4]\d|25[0-5])` +
	`:` + `([0-9]|[1-9]\d{1,3}|[1-5]\d{4}|6[0-5]{2}[0-3][0-5])$`

type RegistryConf struct {
	Registry        string   `json:"registry"`
	RegistryAddress []string `json:"registry_address" mapstructure:"registry_address"`
}

func checkRc(rc *RegistryConf) error {
	if len(rc.RegistryAddress) == 0 {
		return RegistryAddrErr
	}

	err := GetAddressSlice(rc.RegistryAddress)
	if err != nil {
		return err
	}

	return nil
}

// 免侵入micro
func NewRegistry(rc *RegistryConf) (reg registry.Registry, err error) {
	err = checkRc(rc)
	if err != nil {
		return reg, err
	}
	switch rc.Registry {
	case "consul":
		reg = consul.NewRegistry(
			func(op *registry.Options) {
				op.Addrs = rc.RegistryAddress
			})
	case "gossip":
		reg = consul.NewRegistry(
			func(op *registry.Options) {
				op.Addrs = rc.RegistryAddress
			})
	case "mdns":
		reg = mdns.NewRegistry(
			func(op *registry.Options) {
				op.Addrs = rc.RegistryAddress
			})
	case "etcdv3":

		//reg = etcdv3.NewRegistry(
		//	func(op *registry.Options) {
		//		op.Addrs = addrs
		//	})

	case "etcd":
		//reg = etcd.NewRegistry(func(op *registry.Options) {
		//	op.Addrs = addrs
		//})
	case "zookeeper":
		//
		//
		//
		//reg := etcdv3.NewRegistry(
		//	func(op *registry.Options) {
		//		op.Addrs = []string{"127.0.0.1:8500"}
		//	})
	default:
		err = RegistryNameErr
	}
	return reg, err
}

func VerifyIpAndPort(addr string) bool {
	match, _ := regexp.MatchString(IpPortRegex, addr)
	return match
}

// 检查配置中心的注册中心配置地址
func GetAddressSlice(addrs []string) error {
	for _, addr := range addrs {
		if ! VerifyIpAndPort(addr) {
			return RegistryAddrErr
		}
	}
	return nil
}


// 删除服务, 并替换注册
func ReplaceRegister()  {
	
}