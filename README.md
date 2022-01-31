# Open Telemetry demo with OPA and go
This projects demonstrates how applications and OPA can easily be instrumented with Open Telemtry (OTEL). Read on for a Quickstart and detailed information about the demo. 

## Quickstart

In the root folder (same as this readme), run 

``` bash
docker-compose up
```

After giving it some seconds to start-up and generate some data, navigate to [http://localhost:16686/search](http://localhost:16686/search) and browse the traces that have been generated for the services `go-client`, `go-server` and `opa`.

