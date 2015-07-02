### Calcifer

A simple, lightweight reverse proxy written in go which maps external requests based on the Host header to internal DNS SRV records. The main use case is in conjunction with service discovery systems such as mesos-dns and consol-dns which publish an SRV record to a DNS for the purpose of providing host and port mapping for the service.

#### Examples

Start the server:
```
cat > config.json <<EOF 
{
  "Hosts":[
    {"external":"portal","SRV":"_portal.apps._tcp.marathon.mesos"}
  ]
} 
EOF
go build
./calcifer
```


Show the mapped host:
```
curl localhost:8080/hosts
> [{"External":"portal","SRV":"_portal.apps._tcp.marathon.mesos"}]
```

Add / Update a host:
```
curl -d '{"External":"portal2","SRV":"_portal.apps._tcp.marathon.mesos"}' localhost:8080/host
> [{"External":"portal","SRV":"_portal.apps._tcp.marathon.mesos"},{"External":"portal2","SRV":"_portal.apps._tcp.marathon.mesos"}]
```

Route to a host:
```
curl --header "Host: portal" localhost:8080
curl --header "Host: portal2" localhost:8080
```



#### Todos
- add load balancing

