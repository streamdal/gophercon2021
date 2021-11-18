# Reliability Nirvana
## GopherCon 2021

This repo contains the various materials used as part of the ["Reliability Nirvana"](https://www.gophercon.com/agenda/speakers/1221929) 
presentation.

Contents:

* `./order` - idempotent sample order processing service
* `./notify` - idempotent sample notification service
* `./example1` - bare-bones sample producer & consumer
* `./docker-compose.yaml` - rabbitmq, etcd and statsd + graphite (for example visualization)
  * `docker-compose.yaml up -d` - to bring up all
  * or 
  * `docker-compose up -d $specific-dependency`

## Demo

### 1. Basic consume/produce

1. Bring up dependencies
   1. `docker-compose up -d`
2. Verify dependencies are up
3. Bing up services
   1. `cd order && go run *.go`
   2. In another terminal: `cd notify && go run *.go`
4. Publish an event
```bash
curl -i -u guest:guest -H "content-type:application/json" \
  -X POST http://localhost:15672/api/exchanges/%2f/events/publish \
  -d'{"properties":{},"routing_key":"foo","payload":"{\"id\":\"123\",\"type\":\"new_order\"}","payload_encoding":"string"}'
```
5. Observe services consume the event
   1. `orders_processed` metric will increase in Graphite UI (http://localhost:80)
   2. `notifications_sent` metric will increase

### 2. Recovery - one of the consumers is down (for a while)

1. Bring down notification service (via `ctrl-c`)
2. Publish an event
3. Observe service `order` consume the event
   1. `orders_processed` metric will increase
   2. `notifications_sent` metric will _not_ increase
4. Bring up "notify" service
5. See it consume the event
   1. `notifications_sent` metric will increase

### 3. Idempotency

1. Publish the event several times
2. "order" and "notify" services will both consume and process the messages each time
3. `orders_processed` and `notifications_sent` metrics will increase...
4. **^ This is incorrect** - it's the same event - we shouldn't be processing duplicate events
5. Enable etcd usage by restarting service and setting an ENV var:
   1. `cd order && ENABLE_ETCD=true go run *.go`
   2. `cd notify && ENABLE_ETCD=true go run *.go`
6. Emit message multiple times
7. Watch services consume first event and skip duplicate events
8. `orders_processed` & `notifications_sent` will increase only once
9. Restart the services and re-emit events
10. Services will instantly skip the message (due to initial cache import)
