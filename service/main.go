package main

import (
	"flag"
	"log"
	"net"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/denisbrodbeck/machineid"
	"github.com/kardianos/service"
	"github.com/valyala/fasthttp"
)

const (
	Duration = 15 * 60
	AppId    = "pc-monitoring"
	UrlSend  = "http://pc.watch:17518/pc/send"
)

var logger service.Logger

type program struct{}

func (p *program) Start(s service.Service) error {
	key, err := machineid.ProtectedID(AppId)
	if err != nil {
		return err
	}

	name, err := os.Hostname()
	if err != nil {
		return err
	}

	go p.run(key, name)
	return nil
}

func (p *program) Stop(s service.Service) error {
	return nil
}

func (p *program) run(key, username string) {
	for {
		_ = sendRequest(key, username, ipAddresses())

		runtime.GC()

		time.Sleep(time.Duration(Duration) * time.Second)
	}
}

func main() {
	isInstall := flag.Bool("install", false, "Install as a service")
	isUninstall := flag.Bool("uninstall", false, "Uninstall a service")
	flag.Parse()

	// Create Service
	prg := &program{}
	s, err := service.New(prg, &service.Config{
		Name:        "PCM",
		DisplayName: "PCM",
	})
	if err != nil {
		log.Fatal(err)
	}

	// Create Logger
	if logger, err = s.Logger(nil); err != nil {
		log.Fatal(err)
	}

	if *isInstall {
		err = s.Install()
	} else if *isUninstall {
		err = s.Uninstall()
	} else {
		err = s.Run()
	}

	if err != nil {
		logger.Error(err)
	}
}

func sendRequest(key, name, ip string) error {
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()

	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(UrlSend)
	req.URI().QueryArgs().Add("key", key)
	req.URI().QueryArgs().Add("name", name)
	req.URI().QueryArgs().Add("ip", ip)

	return fasthttp.Do(req, resp)
}

func ipAddresses() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return ""
	}

	var output strings.Builder

	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}

		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP

			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			if ip == nil || ip.IsLoopback() {
				continue
			}

			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}

			output.WriteString(ip.String() + ",")
		}
	}

	return strings.TrimRight(output.String(), ",")
}