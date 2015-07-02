### Calcifer

A simple, lightweight reverse proxy written in go which maps external requests based on the Host header to internal DNS SRV records. The main use case is in conjunction with service discovery systems such as mesos-dns and consol-dns which publish an SRV record to a DNS for the purpose of providing host and port mapping for the service.


