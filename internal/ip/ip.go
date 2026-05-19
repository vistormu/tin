package ip

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"sort"
	"strings"
	"time"
)

const publicIPEndpoint = "https://api.ipify.org"

type Interface struct {
	Name string
	V4   []string
	V6   []string
}

type Info struct {
	Interfaces []Interface
	Public     string
	PublicErr  error
}

func Gather(ctx context.Context) (Info, error) {
	ifaces, err := localInterfaces()
	if err != nil {
		return Info{}, err
	}

	info := Info{Interfaces: ifaces}
	info.Public, info.PublicErr = fetchPublicIP(ctx)
	return info, nil
}

func localInterfaces() ([]Interface, error) {
	raw, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("list interfaces: %w", err)
	}

	result := make([]Interface, 0, len(raw))
	for _, iface := range raw {
		if iface.Flags&net.FlagUp == 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		entry := Interface{Name: iface.Name}
		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}
			if ipNet.IP.IsLinkLocalUnicast() || ipNet.IP.IsLinkLocalMulticast() {
				continue
			}
			if v4 := ipNet.IP.To4(); v4 != nil {
				entry.V4 = append(entry.V4, v4.String())
			} else {
				entry.V6 = append(entry.V6, ipNet.IP.String())
			}
		}
		if len(entry.V4) == 0 && len(entry.V6) == 0 {
			continue
		}
		sort.Strings(entry.V4)
		sort.Strings(entry.V6)
		result = append(result, entry)
	}
	return result, nil
}

func fetchPublicIP(ctx context.Context) (string, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(timeoutCtx, http.MethodGet, publicIPEndpoint, nil)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("public ip lookup: status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 64))
	if err != nil {
		return "", err
	}

	addr := strings.TrimSpace(string(body))
	if net.ParseIP(addr) == nil {
		return "", fmt.Errorf("public ip lookup: invalid response %q", addr)
	}
	return addr, nil
}
