# b618reboot go

## About
This is a version of [b618reboot](https://github.com/mkorz/b618reboot/) written in Go.

In addition to issuing simply reboot command, it can also pull stats.

Happy to accept improvements via pull requests!

## Install
### Build
```
go get github.com/mkorz/b618reboot-go
```
)
## Usage
### Command line
#### Reboot:
```
./b618reboot-go reboot -url http://192.168.1.1 -username admin -password ROUTER_ADMIN_PASSWORD
```
#### Signal stats:
```
./b618reboot-go signal-stats -url http://192.168.1.1 -username admin -password ROUTER_ADMIN_PASSWORD
```

Alternatively, instead of passing commandline parameters, you can provide the values via the following environment variables:
 * ROUTER_URL
 * ROUTER_USERNAME
 * ROUTER_PASSWORD

## Docker
### Reboot:
```
docker run --rm -it  mkorz/b618reboot-go reboot -url http://192.168.1.1 -username admin -password ROUTER_ADMIN_PASSWORD
```
or using environment variables:
```
docker run --rm -it -e ROUTER_URL=http://192.168.1.1 -e ROUTER_USERNAME=admin -e ROUTER_PASSWORD=ROUTER_ADMIN_PASSWORD mkorz/b618reboot-go reboot
```
### Signal stats:
```
docker run --rm -it  mkorz/b618reboot-go signal-stats -url http://192.168.1.1 -username admin -password ROUTER_ADMIN_PASSWORD
```
with environment variables:
```
docker run --rm -it -e ROUTER_URL=http://192.168.1.1 -e ROUTER_USERNAME=admin -e ROUTER_PASSWORD=ROUTER_ADMIN_PASSWORD  mkorz/b618reboot-go signal-stats
```
