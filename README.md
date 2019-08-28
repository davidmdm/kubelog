# kubelog

## Installing

To install kubelog run

```
go get github.com/davidmdm/kubelog
```

## Commands:

### Get Apps
```
kubelog get app(s)
```

options:
```
  -n : restrict to one namespace
```

This command will output all namespaces with their pod names joined by prefix

For example, if my namespace had pods example-pod-1566995400-gvjhr, example-pod-7646fb874c-k5nqb, and different-pod-dcb6d6d44-vg7s9
the output would be:

```
myNamespace
  example-pod
  different-pod
```

### Log Apps
```
kubelog pod
```

options:
```
  -n: (string) namespace (required)
  -t: (flag)   log timestamps
  -s: (string) start logs since ie: 5m
```

This command will output the logs for the pod (app) name you gave from the previous command.
It joins all pods output into one unified stream. 











