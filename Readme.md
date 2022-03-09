# Verve challenge: thoughts on solution

## Requirements 
- huge batch writes every 30 minutes (could be tens of millions row or even billions)
- up to few millions of requests per minute (~16k+ rps)
- new batch rewrites the old one
- data queried as single record by uuid

## The idea
Since data is queried by uuid we obviously need to index it.
A dense index would be wasteful, so we can use sparse index to make reads faster.

UUID can be stored as 16 bytes number, price as decimal number and date as unix timestamp.
Such huge amount of data should be stored by columns as it would allow us to save a lot of space with data compression and improve reads even further.

In order to rewrite data we can you 2 options:
- truncate table and write new data. Truncate shouldn't take a long time because the data is well compressed but reads may stall during this operation.
- first write data to a new table and drop the old one. This approach will use extra space as we will need to store 2 datasets at the same time.

To improve performance even further and reduce storage we can do data sharding across multiple machines which will allow us to speed up writes.
The sharding key will be uuid and we should use consistent hashing in order to minimize network activity in case of resharding. 

Considering all the requirements above I will choose Clickhouse database to store the data. It might be not the best tool for the task but:
- it meets the requirements
- it can exchange tables
- I have experience with it to build a working example fast enough
	
In this project I have implemented basic rest api for retrieving records. Rewriting data can be done like this:

`` clickhouse-client --query "insert into new_promotions select toUUID(col1), col2, parseDateTime32BestEffort(substr(col3, 1, 26)) from input('col1 String, col2 Float, col3 String') format CSV" < promotions.csv ``

Then we can replace old table with new one:

`` clickhouse-client --query "exchange table new_promotions, promotions" ``

## Other solutions considered

- Key-value storage: would take a lot of space

## Things not mentioned

Reliability: we can replicate shards

## QA
1. Billion of entries? Writes still fast because of data sharding, clickhouse handle writes decent for big batches of data. Reads complexity log(n), but still very fast because index is sparse and data is compressed.
2. Erase and write whole storage? Prepare second table and then exchange tables
3. 1M reads per minute? Data is sharded, requests are load balanced
4. deployment, scaling, monitoring? I would implement metrics within REST service, expose it with expvar package, monitor service and database with grafana. Scaling is horizontal, just add more machines.

## How to run

1. Install clickhouse-server
2. Create table: clickhouse-client --queries-file sql/promotions.sql
3. Insert data: see example above
4. Build service: go build cmd/rest/ -o rest
5. Run service: ./rest -c config.yaml
6. curl 'http://0.0.0.0:8080/promotions/d018ef0b-dbd9-48f1-ac1a-eb4d90e57118'
