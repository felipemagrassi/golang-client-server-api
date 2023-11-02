Simple Golang project that creates a client-server to consume an external API using contexts, http, files and dbs.

## How to run

## Requirements

- Server should consume the external request within 200ms
- Server should register the response in a SQlite database within 10ms 
- Server should have an '/cotacao' endpoint using 8080 port
- Client should expect a response within 300ms
- Every timeout should log the error 
- Client should save the response in a file

## Special thanks

[Introductory tutorial to SQLite in Go](https://gosamples.dev/sqlite-intro/)
