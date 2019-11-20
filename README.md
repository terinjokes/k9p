# K9P

A virtual filesystem for Kubernetes cluster state.

## Install

K9P uses Go modules, and can be installed with a Go 1.13+. Goal is to allow operators to
use familiar unix tools to interact with their Kubernetes clusters, rather than learning kubectl.

```console
$ git clone https://git.terinstock.com/k9p && cd k9p
$ go get ./cmd/k9p/...
```

## Mounting

### Plan 9

K9P can be mounded like other remote 9P servers:

```console
$ import 'tcp!localhost!1564` /k8s
```

### Linux

Supported on systems compiled with `CONFIG_NET_9P`, otherwise use 9pfuse.

```console
# mount -t 9p -o trans=tcp,port=1564 127.0.0.1 /mnt/k8s
```

### FUSE

K9P can be mounted as a remote 9P server using 9pfuse:

```console
$ 9pfuse 'tcp!127.0.0.1!1564' $HOME/k8s
```

### Applications

So applications can directly interact with 9P servers, such as 9p from plan9ports:

```console
$ p9 -a 'tcp!9p.example.com!1564' ls -l /
d-r-xr-x-r-x I 0 terin terin 0 July 3  2019 cluster-scoped
d-r-xr-x-r-x I 0 terin terin 0 July 3  2019 namespaces
```

## Future

K9P is in a very early WIP state. There's lots of improvements, contributions welcome.

* Support passing authentication to the server for the attach mount.
* Load resources on-demand, rather than creating informers on attach.
* Add support for way more resource types, including CRDs.
* Mutate and remove resources.
* Port forwards modeled after Plan 9's netfs?
