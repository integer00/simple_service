Simple go service
----

This simple service listens at 0.0.0.0:8080 and responding to requests

List of endpoints:
```shell script
/
/api
/healthcheck
```

## Usage:
Bring service and do a GET request to it!


`/api` endpoint will return simple json, ex.
```json
{"ip":"192.168.99.1:60179","status":"ok"}
```

`/healthcheck` endpoint return status of a service, ex.
```json
{"status":"ok"}
```


## Build

```shell script
go build -o simple_service
```

Or docker equivalent:

```shell script
docker build -f dist/docker/Dockerfile -t simple_service:latest .
```

```shell script
docker run -p 8080:8080 simple_service
```