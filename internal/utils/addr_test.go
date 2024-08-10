package utils

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGetHostFromRemoteAddr_ValidIPv4(t *testing.T) {
	host, err := GetHostFromRemoteAddr("192.168.1.1:8080")
	require.NoError(t, err)
	assert.Equal(t, "192.168.1.1", host)
}

func TestGetHostFromRemoteAddr_ValidIPv6(t *testing.T) {
	host, err := GetHostFromRemoteAddr("[2001:db8::1]:8080")
	require.NoError(t, err)
	assert.Equal(t, "2001:db8::1", host)
}

// Non-concurrent tests for invalid inputs
func TestGetHostFromRemoteAddr_InvalidAddrNoPort(t *testing.T) {
	_, err := GetHostFromRemoteAddr("192.168.1.1")
	assert.Error(t, err)
}

func TestGetHostFromRemoteAddr_InvalidAddrInvalidIP(t *testing.T) {
	_, err := GetHostFromRemoteAddr("999.999.999.999:8080")
	assert.Error(t, err)
}

func TestGetHostFromRemoteAddr_InvalidAddrGarbage(t *testing.T) {
	_, err := GetHostFromRemoteAddr("invalid-address")
	assert.Error(t, err)
}
