package server

import (
	"fmt"
	"log"
	"net"
	"net/http"
)

type IpFilterMiddleware struct {
	net *net.IPNet
}

func NewIpFilterMiddleware(cidr string) *IpFilterMiddleware {
	if cidr == "" {
		return &IpFilterMiddleware{}
	}
	_, n, err := net.ParseCIDR(cidr)
	if err != nil {
		panic("CIDR is not valid")
	}
	return &IpFilterMiddleware{net: n}
}

func (f *IpFilterMiddleware) Filter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if f.net == nil {
			return
		}

		ip := r.Header.Get("X-Real-IP")
		log.Println(ip)
		if ip != "" {
			i := net.ParseIP(ip)
			if i != nil && f.net.Contains(i) {
				next.ServeHTTP(w, r)
				return
			}

		}
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("X-Real-IP is not defined\n"))
		w.Write([]byte(fmt.Sprintf("ip = %v", ip)))
	})
}
