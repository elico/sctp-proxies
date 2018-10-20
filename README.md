## SCTP Proxies
Since Squid-Cache doesn't have any support for SCTP socket, even a non stream based one I wrote these two proxy servers.
The original goal was to write a fully featured SCTP proxies.
These ideally should utilize the full capacity of SCTP which is to be able to communicate between two hosts (client and server) while shifting between src and destination IP addresses due to load balancing or failure of one or more routes on the path between the hosts.

Since the Squid-Cache project didn't got enough support these tools are a begining of something.

### TCP to SCTP
Listens on TCP port like on the loopback(127.0.0.1) of a client with the browser pointed to it and the remote SCTP server ipv4+port are configured as a peer.
Every new TCP connection is being proxied over SCTP to the remote ipv4+port.

### SCTP to TCP
Listens on a SCTP ipv4+port (only a single one) and proxies every incomming connection to a local or remote server TCP ipv4+port service.
The service is able to write a PROXY protocol header V1(only if is bound to one external IPv4 address and not loopback or all interfaces) which Squid-Cache support and there for will be able to enforce static or dynamic(external_acl, ICAP, other) source IP based acl's.

## building the deamons for all OS
```bash
./build.sh
```

## Refrences
- https://github.com/ishidawataru/sctp
- https://gist.github.com/legendtkl/c2483c73a3fdb01d36ed8f37d93d3b5c
- https://github.com/pires/go-proxyproto
- https://www.haproxy.org/download/1.8/doc/proxy-protocol.txt
- http://gogs.ngtech.co.il/NgTech-LTD/golang-build-software-binaries