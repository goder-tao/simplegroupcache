package util

import (
	"net/http"
	"strings"
)

func RemoteIp(req *http.Request) string {
	var remoteAddr string
	// RemoteAddr
	remoteAddr = req.RemoteAddr
	if remoteAddr != "" {
		return remoteAddr
	}
	// ipv4
	remoteAddr = req.Header.Get("ipv4")
	if remoteAddr != "" {
		return strings.Split(remoteAddr, ":")[0]
	}
	//
	remoteAddr = req.Header.Get("XForwardedFor")
	if remoteAddr != "" {
		return strings.Split(remoteAddr, ":")[0]
	}
	// X-Forwarded-For
	remoteAddr = req.Header.Get("X-Forwarded-For")
	if remoteAddr != "" {
		return strings.Split(remoteAddr, ":")[0]
	}
	// X-Real-Ip
	remoteAddr = req.Header.Get("X-Real-Ip")
	if remoteAddr != "" {
		return strings.Split(remoteAddr, ":")[0]
	} else {
		remoteAddr = "127.0.0.1"
	}
	return remoteAddr
}
