
All the examples references the services `echoheaders-x` and `echoheaders-y`

```
kubectl run echoheaders --image=gcr.io/google_containers/echoserver:1.3 --replicas=1 --port=8080
kubectl expose deployment echoheaders --port=80 --target-port=8080 --name=echoheaders-x
kubectl expose deployment echoheaders --port=80 --target-port=8080 --name=echoheaders-x
```
