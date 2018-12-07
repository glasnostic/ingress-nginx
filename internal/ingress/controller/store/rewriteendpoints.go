package store

import (
	"fmt"
	"os"

	core "k8s.io/api/core/v1"

	"github.com/golang/glog"
	"github.com/miekg/dns"
)

func updateEndpoints(endpoints *core.Endpoints) {
	router := os.Getenv("GLASNOSTIC_ROUTER_ADDRESS")

	domain := fmt.Sprintf("%s.%s.default.svc.cluster.local.", endpoints.Name, endpoints.Namespace)
	glog.Infof("DNS question domain: %s\n", domain)
	m1 := new(dns.Msg)
	m1.Id = dns.Id()
	m1.RecursionDesired = true
	m1.Question = make([]dns.Question, 1)
	m1.Question[0] = dns.Question{
		Name:   domain,
		Qtype:  dns.TypeA,
		Qclass: dns.ClassINET,
	}
	in, err := dns.Exchange(m1, router+":53")
	if err != nil {
		glog.Errorf("DNS exchange error: %s\n", err)
	}
	if len(in.Answer) == 0 {
		glog.Errorln("no answer")
	}

	for i, answer := range in.Answer {
		// FIXME: this only work for the endpoints that create by service automatically
		// because this kind of endpoints would have 1 subset as we expected
		ip := answer.(*dns.A).A
		endpoints.Subsets[0].Addresses[i].IP =
			ip.String()
	}
}
