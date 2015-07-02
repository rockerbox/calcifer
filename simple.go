package main

import (
	"runtime"
	"fmt"
	"github.com/mailgun/oxy/forward"
	"github.com/mailgun/oxy/testutils"
	"log"
	"net/http"
	"time"
	"encoding/json"
	"flag"
	"io/ioutil"
)

type HostMap struct {
	External string
	SRV string
}
type Config struct {
	Hosts []HostMap
}
var config Config

func load(config Config, lookup map[string]string) {
	for _,elem := range config.Hosts {
		log.Println("Adding entry for:",elem.External, elem.SRV)
		lookup[elem.External] = elem.SRV
        }
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU() * 2)
	configjson := flag.String("c", "config.json", "Path to config. (default config.json)")
	file, err := ioutil.ReadFile(*configjson)
	err = json.Unmarshal(file, &config)
	log.Println(config, err)
	
	fwd, _ := forward.New()
        c := NewCache(time.Duration(5)*time.Second) 

	lookup := make(map[string]string, len(config.Hosts) -1)
	load(config,lookup)
	

	redirect := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		srvq, exists := lookup[req.Host]
		if exists {
                	data, found := c.Get(srvq)
                	if !found {
				hosts, _, _ := DNSSRV(srvq)	
				c.Set(srvq,hosts[0])
				data = hosts[0]
			}
			req.URL = testutils.ParseURI("http://" + data)
			fwd.ServeHTTP(w, req)
		} else {
			switch req.URL.Path {
			case "/hosts": 
				resp, _ := json.Marshal(config.Hosts)
				fmt.Fprintf(w,string(resp))
			case "/add_host":
				defer req.Body.Close()
				body, err := ioutil.ReadAll(req.Body)
				if err == nil && len(body) > 0 {
					var h HostMap
					json.Unmarshal(body,&h)
					config.Hosts = append(config.Hosts,h)
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
		Addr:    ":8080",
		Handler: redirect,
	}
	s.ListenAndServe()
}
