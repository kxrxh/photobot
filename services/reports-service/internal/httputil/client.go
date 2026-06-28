package httputil

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const maxRedirects = 5

// PublicClient returns an HTTP client that blocks connections to private/loopback IPs.
func PublicClient(timeout time.Duration, allowPrivateIPs bool) *http.Client {
	dialer := &net.Dialer{Timeout: timeout}
	return &http.Client{
		Timeout: timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= maxRedirects {
				return errors.New("too many redirects")
			}
			return ValidateHTTPURL(req.URL.String())
		},
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				if !allowPrivateIPs {
					if err := assertPublicDialAddr(addr); err != nil {
						return nil, err
					}
				}
				return dialer.DialContext(ctx, network, addr)
			},
		},
	}
}

func ValidateHTTPURL(raw string) error {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return err
	}
	switch strings.ToLower(u.Scheme) {
	case "http", "https":
	default:
		return fmt.Errorf("unsupported url scheme %q", u.Scheme)
	}
	if strings.TrimSpace(u.Hostname()) == "" {
		return errors.New("url host is empty")
	}
	return nil
}

func assertPublicDialAddr(addr string) error {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		host = addr
	}
	ip := net.ParseIP(host)
	if ip == nil || isPublicIP(ip) {
		return nil
	}
	return fmt.Errorf("blocked IP %s", ip)
}

func isPublicIP(ip net.IP) bool {
	return !ip.IsLoopback() && !ip.IsPrivate() && !ip.IsLinkLocalUnicast() && !ip.IsUnspecified()
}
