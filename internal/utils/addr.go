package utils

import (
	"fmt"
	"net"
)

func GetHostFromRemoteAddr(remoteAddr string) (string, error) {
	ip, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return "", err
	}

	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return "", fmt.Errorf("invalid IP address %q", ip)
	}

	return parsedIP.String(), nil
}
