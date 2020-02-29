## Kafka utilities:

### Connect producer:
```
docker exec -it mysql-cdc-kafka bash -c \
    'kafka-console-producer.sh --broker-list mysql-cdc-kafka:9092 --topic employees'
```

### Connect consumer:
```
docker exec -it mysql-cdc-kafka bash -c \
    'kafka-console-consumer.sh --bootstrap-server mysql-cdc-kafka:9092 --topic employees --from-beginning --partition 0'
```