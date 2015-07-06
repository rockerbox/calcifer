package main

import (
	"runtime"
	"fmt"
	"github.com/mailgun/oxy/forward"
	"github.com/mailgun/oxy/testutils"
	"log"
	"net/http"
	"net/url"
	"time"
	"encoding/json"
	"flag"
	"io/ioutil"
	"sync"
        "github.com/rockerbox/websocketproxy"
	"strings"
)

type HostMap struct {
	External string
	SRV string
}
type Config struct {
	mutex sync.RWMutex
	Hosts []HostMap
}

func (config *Config) UpdateHost(host HostMap) {
	config.mutex.Lock()
	shouldAppend := true
	for _, h := range config.Hosts {
		if (host.External == h.External) {
			h.SRV = host.SRV	
			shouldAppend = false
		}
	}

	if (shouldAppend) {
		config.Hosts = append(config.Hosts,host)
	}
	config.mutex.Unlock()
}

var config Config

func load(config Config, lookup map[string]string) {
	for _,elem := range config.Hosts {
		log.Println("Adding entry for:",elem.External, elem.SRV)
		lookup[elem.External] = elem.SRV
        }
}

func tokenListContainsValue(header http.Header, name string, value string) bool {
	for _, v := range header[name] {
		for _, s := range strings.Split(v, ",") {
			if strings.EqualFold(value, strings.TrimSpace(s)) {
				return true
			}
		}
	}
	return false
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU() * 2)
	configjson := flag.String("c", "config.json", "Path to config. (default config.json)")
        port := flag.String("p","8080","port to run on")

	flag.Parse()

	file, err := ioutil.ReadFile(*configjson)
	err = json.Unmarshal(file, &config)
	log.Println(config, err)
	
	fwd, _ := forward.New()
        c := NewCache(time.Duration(5)*time.Second) 

	lookup := make(map[string]string, len(config.Hosts) +30)
	load(config,lookup)
	

	redirect := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		srvq, exists := lookup[req.Host]
		if exists {
                	data, found := c.Get(srvq)
                	if !found {
				hosts, _, _ := DNSSRV(srvq)	
				log.Println("Looked up host:",srvq,hosts[0])
				c.Set(srvq,hosts[0])
				data = hosts[0]
			}
			if tokenListContainsValue(req.Header, "Connection", "upgrade") {
				swsurl := "ws://" + data + req.URL.Path + "?" + req.URL.RawQuery 
				wsurl, _ := url.Parse(swsurl)
				websocketproxy.NewProxy(wsurl).ServeHTTP(w,req)
			} else {
				req.URL = testutils.ParseURI("http://" + data)
				fwd.ServeHTTP(w, req)
			}
		} else {
			switch req.URL.Path {
			case "/hosts": 
				resp, _ := json.Marshal(config.Hosts)
				fmt.Fprintf(w,string(resp))
			case "/host":
				defer req.Body.Close()
				body, err := ioutil.ReadAll(req.Body)
				if err == nil && len(body) > 0 {
					var h HostMap
					json.Unmarshal(body,&h)
					config.UpdateHost(h)
					resp, _ := json.Marshal(config.Hosts)
					load(config, lookup)
					fmt.Fprintf(w,string(resp))
				}
			default:
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("HTTP status code returned!"))
			}
		}
	})

	s := &http.Server{
		Addr:    ":" + *port,
		Handler: redirect,
	}
	s.ListenAndServe()
}
