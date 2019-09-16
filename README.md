# kubelog

## Installing

To install kubelog run

```
go get github.com/davidmdm/kubelog
```

## Commands:

### Get Services
```
kubelog get (services|svc)
```

options:
```
  -n : restrict to one namespace
```

This command will output all namespaces with their service names

```
myNamespace
  service_1
  service_2
```

### Tail Service Logs
```
kubelog -n namespace service [service2, [service3, [...]]]
```

options:
```
  -n: (string) namespace (required)
  -t: (flag)   log timestamps
  -s: (string) start logs since ie: 5m
```

If you want to tail all services in a namespace simply pass "*" as sole argument. 

```
kubelog -n namespace "*"
```
