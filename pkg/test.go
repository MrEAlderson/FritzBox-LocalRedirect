package main

import (
	//"context"
	"fmt"
	"log/slog"
	"net"
	"strings"
	"time"

	"github.com/MrEAlderson/FritzBox-LocalRedirect/pkg/avm"
)

/*
	type IP struct {
		value net.IP
		ver   byte
	}
*/
type FritzIps struct {
	v4          net.IP
	v6          net.IP
	v6Prefix    *net.IPNet
	refreshTime time.Time
}

func (ips FritzIps) any_match(ip net.IP) bool {
	if ips.v4 != nil && ips.v4.Equal(ip) {
		return true
	}
	if ips.v6 != nil && ips.v6.Equal(ip) {
		return true
	}

	if ips.v6Prefix != nil {
		ipMasked := ip.Mask(ips.v6Prefix.Mask)

		return ips.v6Prefix.IP.Equal(ipMasked)
	}

	return false
}

func (ips FritzIps) all_nil() bool {
	return ips.v4 == nil && ips.v6 == nil && ips.v6Prefix == nil
}

/*
func get_ip(ip net.IP) IP {
	if ipv4 := ip.To4(); ipv4 != nil {
		return IP{
			value: ipv4,
			ver:   4,
		}
	} else {
		ipv6 := ip.To16()

		return IP{
			value: ipv6,
			ver:   6,
		}
	}
}

func (ip1 IP) match(ip2 IP, prefixLen int) bool {
	if ip1.ver != ip2.ver {
		return false
	}

	// v4
	if ip1.ver == 4 {
		return ip1.value.Equal(ip2.value)
	}

	// v6
	mask := net.CIDRMask(prefixLen, 128)
	na := ip1.value.Mask(mask)
	nb := ip2.value.Mask(mask)

	fmt.Println(na, nb)

	return na.Equal(nb)
}*/

func main() {
	ip := "192.168.178.64:5135"
	var ips []string
	pieces := strings.Split(ip, ":")
	ips = pieces[0 : len(pieces)-1]
	ip = strings.Join(ips, ":")
	fmt.Println(ip)
	/*r := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: time.Millisecond * time.Duration(10000),
			}
			return d.DialContext(ctx, network, "8.8.8.8:53")
		},
	}

	ips, _ := r.LookupIP(context.Background(), "ip", "l5gfd163xvqc.myfritz.net")
	ipObjs := make([]IP, len(ips))
	var index int = 0

	for _, ip := range ips {
		ipObjs[index] = get_ip(ip)
		index++
	}*/

	rootLogger := slog.Default()
	timeoutDuration, _ := time.ParseDuration("30s")
	fritzbox := &avm.FritzBox{
		Url:     "http://192.168.178.1:49000",
		Timeout: timeoutDuration,
		Logger:  rootLogger,
	}

	v4, _ := fritzbox.GetWanIpv4()
	v6, _ := fritzbox.GetwanIpv6()
	v6Prefix, _ := fritzbox.GetIpv6Prefix()
	fritzIps := &FritzIps{
		v4:          v4,
		v6:          v6,
		v6Prefix:    v6Prefix,
		refreshTime: time.Now(),
	}

	//ourIP := net.ParseIP("2j12:8202:98df:7000:f860:239c:9dac:dee2")
	fmt.Println(fritzIps)
	fmt.Println(fritzIps.all_nil())
	/*for _, ip := range ipObjs {
		fmt.Println(ip.match(ourIP, 48))
	}*/
}
