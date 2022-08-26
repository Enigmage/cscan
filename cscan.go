package main

import (
	"flag"
	"fmt"
	"github.com/fatih/color"
	"net"
	"os"
	"sync"
	"time"
)

type Result struct {
	Port     int
	State    string
	Protocol string
	Service  string
}

type Scanner struct {
	Host     string
	Protocol string
}

func getCommonPortServices() map[int]string {
	ret := map[int]string{
		20:  "ftp",
		21:  "ftp",
		22:  "ssh",
		23:  "telnet",
		25:  "smtp",
		53:  "dns",
		110: "pop3",
		80:  "http",
		443: "https",
	}
	return ret
}

func scan(protocol, host string, openPorts chan<- Result) {
	var wg sync.WaitGroup
	portMap := getCommonPortServices()
	for i := 1; i <= 65_535; i++ {
		wg.Add(1)
		go func(port int) {
			defer wg.Done()
			address := fmt.Sprintf("%s:%d", host, port)
			timeout := time.Second * 60
			conn, err := net.DialTimeout(protocol, address, timeout)
			if err != nil {
				openPorts <- Result{port, "closed", protocol, "-"}
				return
			}
			conn.Close()
			service, ok := portMap[port]
			if !ok {
				service = "-"
			}
			openPorts <- Result{port, "open", protocol, service}
		}(i)
	}
	wg.Wait()
	defer close(openPorts)
}

func (p *Scanner) Start() {
	openPorts := make(chan Result)
	dt := time.Now()
	color.Set(color.FgHiGreen)
	fmt.Printf("Starting scan at %s\nHost: %s\n", dt.Format(time.UnixDate), p.Host)
	fmt.Println("Port\tState\tService\tProtocol")
	color.Unset()
	go scan(p.Protocol, p.Host, openPorts)
	for n := range openPorts {
		if n.State == "open" {
			fmt.Printf("%d\t%s\t%s\t%s\n", n.Port, n.State, n.Service, n.Protocol)
		}
	}
    color.Set(color.FgHiGreen)
    fmt.Println("\nTook", time.Since(dt), "to complete")
    color.Unset()
}

func displayAbout() {
	color.Set(color.FgHiYellow)
	fmt.Println("pscan: a simple and concurrent full connection type tcp/udp port scanner.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("\tpscan --host <hostname> --proto <protocol>")
	fmt.Println()
	fmt.Println("hostname: name of the target host. Default value: localhost")
	fmt.Println("protocol: can be tcp or udp. Default value: tcp")
	color.Unset()
}

func main() {
	var host, protocol string
	var about bool
    flag.StringVar(&protocol, "proto", "tcp", "Protcol to use, values: tcp/udp")
	flag.StringVar(&host, "host", "localhost", "Hostname")
	flag.BoolVar(&about, "about", false, "About pscan")
	flag.Parse()
	if (protocol != "tcp" && protocol != "udp") || flag.Arg(0) != "" {
        flag.Usage()
		os.Exit(1)
	}
	if about {
		displayAbout()
		os.Exit(0)
	}
	ps := &Scanner{
		Host:     host,
		Protocol: protocol,
	}
	ps.Start()
}
