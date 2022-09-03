package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
)

type Result struct {
	Port     int64
	State    string
	Protocol string
	Service  string
}

type Scanner struct {
	Host      string
	Protocol  string
	portRange Range
	Timeout   time.Duration
}

type Range struct {
	start, end int64
}

func getCommonPortServices() map[int64]string {
	ret := map[int64]string{
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

func scan(protocol, host string, openPorts chan<- Result, r Range, timeout time.Duration) {
	wg := sync.WaitGroup{}
	portMap := getCommonPortServices()
	for i := r.start; i <= r.end; i++ {
		wg.Add(1)
		go func(port int64) {
			defer wg.Done()
			address := fmt.Sprintf("%s:%d", host, port)
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
	fmt.Printf("Scanning from port %v to %v\n", p.portRange.start, p.portRange.end)
	fmt.Println("Port\tState\tService\tProtocol")
	color.Unset()
	go scan(p.Protocol, p.Host, openPorts, p.portRange, p.Timeout)
	for n := range openPorts {
		if n.State == "open" {
			fmt.Printf("%d\t%s\t%s\t%s\n", n.Port, n.State, n.Service, n.Protocol)
		}
	}
	color.Set(color.FgHiGreen)
	fmt.Println("\nTook", time.Since(dt), "to complete")
	color.Unset()
}

func main() {
	var host, protocol, inputRange string
	var timeout time.Duration
	flag.StringVar(&protocol, "proto", "tcp", "Protcol to use, values: tcp/udp")
	flag.StringVar(&host, "host", "localhost", "Hostname")
	flag.StringVar(&inputRange, "port", "1-65535", "Provide the port or range of ports to scan")
	flag.DurationVar(&timeout, "timeout", time.Second*60, "Request timeout duration in seconds")
	flag.Parse()
	if (protocol != "tcp" && protocol != "udp") || flag.Arg(0) != "" {
		flag.Usage()
		os.Exit(1)
	}

	r := Range{}
	sp := strings.Split(inputRange, "-")
	f, err := strconv.ParseInt(sp[0], 10, 64)
	if err != nil {
		color.Red("Error: bad range provided!")
		os.Exit(1)
	}
	s, err := f, nil
	if len(sp) == 2 {
		s, err = strconv.ParseInt(sp[1], 10, 64)
		if err != nil {
			color.Red("Error: bad range provided!")
			os.Exit(1)
		}
	} else if len(sp) > 2 {
		color.Red("Error: bad range provided!")
		os.Exit(1)
	}

	r.start = func(a, b int64) int64 {
		if a <= b {
			return a
		}
		return b
	}(f, s)

	r.end = func(a, b int64) int64 {
		if a >= b {
			return a
		}
		return b
	}(f, s)

	ps := &Scanner{
		Host:      host,
		Protocol:  protocol,
		portRange: r,
		Timeout:   timeout,
	}
	ps.Start()
}
