/*
 * @Author: FunctionSir
 * @License: AGPLv3
 * @Date: 2025-04-15 20:23:21
 * @LastEditTime: 2025-04-16 20:34:49
 * @LastEditors: FunctionSir
 * @Description: -
 * @FilePath: /any-ecs-doh-proxy/handlers.go
 */

package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"os"
	"strings"
	"time"

	"github.com/miekg/dns"
)

const DEFAULT_HOME_PAGE = "<h1>This is an any-ecs-doh-proxy server!</h1>"

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	if !Config.HasKey("options", "HomePage") {
		w.Write([]byte(DEFAULT_HOME_PAGE))
		return
	}
	homePage, err := os.ReadFile(Config["options"]["HomePage"])
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 InternalServerError"))
		return
	}
	w.Write(homePage)
}

func dnsForward(query []byte, w http.ResponseWriter, r *http.Request) {
	countryCode := strings.TrimSpace(r.PathValue("CountryCode"))
	province := strings.TrimSpace(r.PathValue("Province"))
	city := strings.TrimSpace(r.PathValue("City"))
	if len(countryCode) <= 1 || len(province) <= 0 || len(city) <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("400 BadRequest"))
		return
	}
	var dnsMsg dns.Msg
	err := dnsMsg.Unpack(query)
	opt := &dns.OPT{Hdr: dns.RR_Header{Name: ".", Rrtype: dns.TypeOPT}}
	ecs := &dns.EDNS0_SUBNET{Code: dns.EDNS0SUBNET}
	ecs.SourceNetmask = uint8(24)
	ecs.Family = 1
	ip := getIp(countryCode, province, city)
	if ip == nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("400 BadRequest"))
		return
	}
	addr, ok := netip.AddrFromSlice(ip)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 InternalServerError"))
		return
	}
	p := netip.PrefixFrom(addr, 24)
	ecs.Address = net.IP(p.Masked().Addr().AsSlice())
	opt.Option = append(opt.Option, ecs)
	dnsMsg.Extra = []dns.RR{opt}
	packed, err := dnsMsg.Pack()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 InternalServerError"))
		return
	}
	resp, err := http.Post(Config["options"]["Upstream"], "application/dns-message", bytes.NewReader(packed))
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte("502 BadGateway"))
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte("502 BadGateway"))
		return
	}
	w.Header().Set("Content-Type", "application/dns-message")
	w.Write(body)
}

func getReqHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("dns")
	if len(q) <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("400 BadRequest"))
		return
	}
	query, err := base64.RawURLEncoding.DecodeString(q)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("400 BadRequest"))
		return
	}
	dnsForward(query, w, r)
}

func postReqHandler(w http.ResponseWriter, r *http.Request) {
	q, err := io.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 InternalServerError"))
		return
	}
	dnsForward(q, w, r)
}

func queryHandler(w http.ResponseWriter, r *http.Request) {
	Status.TotQueries.Add(1)
	switch r.Method {
	case http.MethodGet:
		getReqHandler(w, r)
	case http.MethodPost:
		postReqHandler(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("405 MethodNotAllowed"))
	}
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	str := fmt.Sprintf("<title>Server Status</title><h1>Server Status</h1><h2>Uptime: %ds</h2><h2>Tot Queries: %d</h2>",
		time.Now().Unix()-Status.StartedAt, Status.TotQueries.Load())
	w.Write([]byte(str))
}
