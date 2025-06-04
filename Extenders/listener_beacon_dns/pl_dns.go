package main

import (
	"fmt"
	"github.com/miekg/dns"
	"time"
)

type DNSConfig struct {
	HostBind   string `json:"host_bind"`
	PortBind   int    `json:"port_bind"`
	Domain     string `json:"domain"`
	Protocol   string `json:"protocol"`
	EncryptKey []byte `json:"encrypt_key"`
}

type DNS struct {
	Server *dns.Server
	Config DNSConfig
	Name   string
	Active bool
}

func (handler *DNS) Start(ts Teamserver) error {
	address := fmt.Sprintf("%s:%d", handler.Config.HostBind, handler.Config.PortBind)

	mux := dns.NewServeMux()
	mux.HandleFunc(handler.Config.Domain+".", handler.handleRequest)

	handler.Server = &dns.Server{Addr: address, Net: "udp", Handler: mux}

	go func() {
		if err := handler.Server.ListenAndServe(); err != nil {
			fmt.Println("Error starting DNS server:", err)
		} else {
			handler.Active = true
		}
	}()

	time.Sleep(500 * time.Millisecond)
	return nil
}

func (handler *DNS) Stop() error {
	handler.Active = false
	if handler.Server != nil {
		return handler.Server.Shutdown()
	}
	return nil
}

func (handler *DNS) handleRequest(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.Authoritative = true

	if len(r.Question) > 0 && r.Question[0].Qtype == dns.TypeTXT {
		m.Answer = append(m.Answer, &dns.TXT{
			Hdr: dns.RR_Header{Name: r.Question[0].Name, Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 0},
			Txt: []string{"OK"},
		})
	}
	_ = w.WriteMsg(m)
}
