# Run redpanda cluster

Create all the volumes

```bash
docker volume create redpanda1 && \
docker volume create redpanda2 && \
docker volume create redpanda3
```

run the cluster

```bash
docker-compose up -d
```

## Test everything is working

This will allow you to run a simple producer/consumer to make sure everything is working as expected, quickly:

### Create a topic

```bash
docker exec -it redpanda1 rpk topic create test --brokers=localhost:9092
```

### Run the producer

```bash
docker exec -it redpanda1 rpk topic produce test --brokers=localhost:9092
```

### Run the consumer

```bash
docker exec -it redpanda1 rpk topic consume test --brokers=localhost:9092
```

### clean up

```bash
docker-compose stop
```

```bash
docker-compose rm
```

```bash
docker volume rm redpanda1 && \
docker volume rm redpanda2 && \
docker volume rm redpanda3
```
