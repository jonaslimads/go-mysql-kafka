# Example

This example uses a testing database found in https://github.com/datacharmer/test_db

## Running the example

You must first run the MySQL container, load the database and run the Go container:

```
cd docker

docker-compose up -d go-mysql-kafka-db

# wait a few seconds, then
./load_database.sh

docker-compose up go-mysql-kafka
```

The Go docker image uses https://github.com/githubnemo/CompileDaemon to ensure hot-reload, so you can freely update `main.go` code.
Once the Go instance is running, connect as Kafka consumer in another tab:  

```
docker exec -it go-mysql-kafka-broker bash -c '
    kafka-console-consumer.sh --bootstrap-server go-mysql-kafka-broker:9092 \
    --topic employees --from-beginning --partition 0 '
```

Update an entry in the database via any client of your choice or shell i.e:

```
docker exec -it go-mysql-kafka-db bash -c '
    mysql -uroot -padmin --database=employees \
    --execute="UPDATE employees SET first_name=\"John\", last_name=\"Wick\" WHERE emp_no=10001" '
```

This will be listened by the Kafka consumer resulting in something like:
```json
{"Table":{"Schema":"employees","Name":"employees","Columns":[{"Name":"emp_no","Type":1,"Collation":"","RawType":"int(11)","IsAuto":false,"IsUnsigned":false,"EnumValues":null,"SetValues":null},{"Name":"birth_date","Type":8,"Collation":"","RawType":"date","IsAuto":false,"IsUnsigned":false,"EnumValues":null,"SetValues":null},{"Name":"first_name","Type":5,"Collation":"latin1_swedish_ci","RawType":"varchar(14)","IsAuto":false,"IsUnsigned":false,"EnumValues":null,"SetValues":null},{"Name":"last_name","Type":5,"Collation":"latin1_swedish_ci","RawType":"varchar(16)","IsAuto":false,"IsUnsigned":false,"EnumValues":null,"SetValues":null},{"Name":"gender","Type":3,"Collation":"latin1_swedish_ci","RawType":"enum('M','F')","IsAuto":false,"IsUnsigned":false,"EnumValues":["M","F"],"SetValues":null},{"Name":"hire_date","Type":8,"Collation":"","RawType":"date","IsAuto":false,"IsUnsigned":false,"EnumValues":null,"SetValues":null}],"Indexes":[{"Name":"PRIMARY","Columns":["emp_no"],"Cardinality":[283118]}],"PKColumns":[0],"UnsignedColumns":null},"Action":"update","Rows":[[10001,"1953-09-02","Test","Facello",1,"1986-06-26"],[10001,"1953-09-02","John","Wick",1,"1986-06-26"]],"Header":{"Timestamp":1584821499,"EventType":31,"ServerID":1,"EventSize":83,"LogPos":231689114,"Flags":0}}
```

An update entry to **departments** table won't be streamed to Kafka because no handler was added for this table.
You have to explicitly tell which tables will stream to which Kafka topics. See main.go for full implementation.  