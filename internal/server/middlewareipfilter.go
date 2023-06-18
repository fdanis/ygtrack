package server

import (
	"fmt"
	"log"
	"net"
	"net/http"
)

type IPFilterMiddleware struct {
	net *net.IPNet
}

func NewIPFilterMiddleware(cidr string) *IPFilterMiddleware {
	if cidr == "" {
		return &IPFilterMiddleware{}
	}
	_, n, err := net.ParseCIDR(cidr)
	if err != nil {
		panic("CIDR is not valid")
	}
	return &IPFilterMiddleware{net: n}
}

func (f *IPFilterMiddleware) Filter(next http.Handler) http.Handler {
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
