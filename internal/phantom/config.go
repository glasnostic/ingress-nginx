package phantom

import (
	"context"
	"fmt"
	"net"

	"k8s.io/ingress-nginx/internal/phantom/protobuf/router"

	"github.com/golang/glog"

	"google.golang.org/grpc"
)

const (
	GRPCPort = ":50051"
)

type Config struct {
	server string
}

func NewConfig(addr string) *Config {
	return &Config{
		server: addr,
	}
}

func GetPhantomFrom(addr string) (*net.IPNet, *net.IPNet, bool) {
	c := NewConfig(addr)
	nets, err := c.getNetworks()
	if err != nil {
		glog.Infof("failed at getting glasnostic router phantom network config, error: %s", err)
		return nil, nil, false
	}
	for _, n := range nets.Networks {
		if !n.Owned {
			continue
		}
		_, network, err := net.ParseCIDR(fmt.Sprintf("%s/%d", n.Real, n.Mask))
		if err != nil {
			glog.Infof("failed at parse real network CIDR, error: %s", err)
			return nil, nil, false
		}
		_, phantom, err := net.ParseCIDR(fmt.Sprintf("%s/%d", n.Virtual, n.Mask))
		if err != nil {
			glog.Infof("failed at parse phantom network CIDR, error: %s", err)
			return nil, nil, false
		}
		return network, phantom, true
	}
	return nil, nil, false
}

func (c *Config) getNetworks() (*router.Networks, error) {
	address := c.server + GRPCPort
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("failed to connect GRPC server: %v", err)
	}
	defer conn.Close()
	client := router.NewConfigServiceClient(conn)
	networks, err := client.GetNetworks(context.Background(), &router.Request{})
	if err != nil {
		return nil, err
	}
	return networks, nil
}
