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

## Run sample producer with schema

### Create topic
If you haven't created the test topic, create the topic from the previous step

```bash
docker exec -it redpanda1 rpk topic create test --brokers=localhost:9092
```

### Register the schema
The producer will do it automatically if this step is not executed, but we want to make sure that we can register schemas separetely for the approach we are going to take in the project further on:

1. Make sure you have python running and you have the correct version (recommend pyenv)
2. Create a virtual env for installing the dependencies
```bash
python -m venv venv
```
3. Activate the virtual environment
```bash
source venv/bin/activate
```
4. Install all the dependencies
```bash
pip install -r requirements.txt
```
5. Register the sample schema
```bash
python register_schema.py
```

Also you can check the information about the schema
```bash
python get_schema_version_details.py
```

### Run sample
Make sure you have installed go correctly on your machine. For the sample we won't need to create an executable.

```bash
cd producer
```

Make sure you have all dependencies correctly before running
```bash
go mod tidy
```

Run the code
```bash
go run .
```

## clean up

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
