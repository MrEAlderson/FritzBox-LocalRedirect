// https://github.com/cromefire/fritzbox-cloudflare-dyndns/blob/main/pkg/polling/pollserver.go
package polling

import (
	"marcel/localredirect/pkg/avm"
	"marcel/localredirect/pkg/util"
	"log/slog"
	"net"
	"net/url"
	"strings"
	"time"
)

func StartPollServer(out chan<- *net.IP, localIp *net.IP, logger *slog.Logger, interval string, useIpv4 bool, useIpv6 bool,
	endpointUrl string, endpointTimeout string) *util.PollStatus {

	const subsystem = "fritzbox_polling"
	logger = logger.With(util.SubsystemAttr(subsystem))
	fritzbox := newFritzBox(logger, endpointUrl, endpointTimeout)

	// Import endpoint polling interval duration
	var ticker *time.Ticker

	if interval != "" {
		v, err := time.ParseDuration(interval)

		if err != nil {
			logger.Warn("Failed to parse FRITZBOX_ENDPOINT_INTERVAL, using defaults", util.ErrorAttr(err))
			ticker = time.NewTicker(300 * time.Second)
		} else {
			ticker = time.NewTicker(v)
		}
	} else {
		logger.Info("FRITZBOX_ENDPOINT_INTERVAL not found, disabling polling")
		return nil
	}

	status := util.PollStatus{Succeeded: true}

	go func() {
		lastV4 := net.IP{}
		lastV6 := net.IP{}

		poll := func() {
			success := true
			logger.Debug("Polling WAN IPs from router")

			defer func() {
				status.Succeeded = success
				status.Last = time.Now()
			}()

			if useIpv4 {
				ipv4, err := fritzbox.GetWanIpv4()

				if err != nil {
					logger.Warn("Failed to poll WAN IPv4 from router", util.ErrorAttr(err))
					success = false
				} else {
					if !lastV4.Equal(ipv4) {
						logger.Info("New WAN IPv4 found", slog.Any("ipv4", ipv4))
						out <- &ipv4
						lastV4 = ipv4
					}
				}
			}

			if *localIp == nil && useIpv6 {
				ipv6, err := fritzbox.GetwanIpv6()

				if err != nil {
					logger.Warn("Failed to poll WAN IPv6 from router", util.ErrorAttr(err))
					success = false
				} else {
					if !lastV6.Equal(ipv6) {
						logger.Info("New WAN IPv6 found", slog.Any("ipv6", ipv6))
						out <- &ipv6
						lastV6 = ipv6
					}
				}
			} else if useIpv6 {
				prefix, err := fritzbox.GetIpv6Prefix()

				if err != nil {
					logger.Warn("Failed to poll IPv6 Prefix from router", util.ErrorAttr(err))
					success = false
				} else {
					constructedIp := make(net.IP, net.IPv6len)
					copy(constructedIp, prefix.IP)

					maskLen, _ := prefix.Mask.Size()

					for i := 0; i < net.IPv6len; i++ {
						b := constructedIp[i]
						lb := (*localIp)[i]
						var mask byte = 0b00000000
						for j := 0; j < 8; j++ {
							if (i*8 + j) >= maskLen {
								mask += 0b00000001 << (7 - j)
							}
						}
						b += lb & mask
						constructedIp[i] = b
					}

					if !lastV6.Equal(prefix.IP) {
						logger.Info("New IPv6 Prefix found", slog.Any("prefix", prefix), slog.Any("ipv6", constructedIp))
						out <- &constructedIp
						lastV6 = prefix.IP
					}
				}
			}
		}

		poll()

		for {
			select {
			case <-ticker.C:
				poll()
			}
		}
	}()

	return &status
}

func newFritzBox(logger *slog.Logger, endpointUrl string, endpointTimeout string) *avm.FritzBox {
	fb := avm.NewFritzBox(logger)

	// Import FritzBox endpoint url
	if endpointUrl != "" {
		v, err := url.ParseRequestURI(endpointUrl)

		if err != nil {
			logger.Error("Failed to parse FRITZBOX_ENDPOINT_URL", util.ErrorAttr(err))
			panic(err)
		}

		fb.Url = strings.TrimRight(v.String(), "/")
		fb.Url = strings.TrimRight(v.String(), "/")
	} else {
		logger.Info("FRITZBOX_ENDPOINT_URL not found, disabling FritzBox polling")
		return nil
	}

	// Import FritzBox endpoint timeout setting
	if endpointTimeout != "" {
		v, err := time.ParseDuration(endpointTimeout)

		if err != nil {
			logger.Warn("Failed to parse FRITZBOX_ENDPOINT_TIMEOUT, using defaults", util.ErrorAttr(err))
		} else {
			fb.Timeout = v
		}
	}

	return fb
}
