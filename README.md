# tickets-api
Tickets API Simulation

## Seed the Database

* Change the DB url in db-seed/seed.py
  * `engine = create_engine('cockroachdb://root@192.168.86.74:26257/tickets')`
* Execute the Python script
  * `python db-seed/seed.py --num_users 500 --num_purchases 1000 --num_cancellations 200 --num_payments 4000`

## Tickets API Endpoints

* GET: `/user/:uuid/purchases` - get user purchases
* GET: `/user/:uuid/purchases/cancellations` - get user cancellations
* GET: `/search/users` - search users `http://localhost:3001/search/users?name=Abigail` for go-pg
* GET: `/search/users` - search users `http://localhost:3001/search/users?name=Abigail` for pgx

## Start GO-PG Tickets API Endpoint

```shell
cd gopg-api

go mod init gopg-api
go mod tidy
go run *.go
```

### Run K6 Stress Test

```
k6 run k6-gopg.js
```

## Start PGX Tickets API Endpoint

```shell
cd pgx-api

go mod init pgx-api
go mod tidy
go run *.go
```
### Run K6 Stress Test

```
k6 run k6-pgx.js
```

## Implicit/Explicit Transaction Example

* This example is built to test Read Committed transactions in CockroachDB 23.2.x (beta at the moment of this writing)

```
cd read-commit

go mod init implicit-explicit
go mod tidy
go run *.go
```

### Test Implicit Transactions

* Slightly modify your K6 script to hard code some user UUIDs in an array
  * `const uuids = [...uuids...]`
* http://localhost:8080/implicit/users/033560d1-32f0-4d7d-a745-232c053b93bf
* Run the stress test
  * `k6 run --vus 500 --duration 5m k6-implicit.js`
  
### Test Explicit Transactions (v23.2.x+)

* Slightly modify your K6 script to hard code some user UUIDs in an array
  * `const uuids = [...uuids...]`
* http://localhost:8080/explicit/users/033560d1-32f0-4d7d-a745-232c053b93bf
* Run the stress test
  * `k6 run --vus 500 --duration 5m k6-explicit.js`

