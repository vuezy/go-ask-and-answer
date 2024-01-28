# Golang - Ask&Answer
A backend API server built with Go that is designed for use with [my kotlin-ask-and-answer repository](https://github.com/vuezy/kotlin-ask-and-answer).\
The project uses [goose](https://github.com/pressly/goose) as a database migration tool and [sqlc](https://github.com/sqlc-dev/sqlc) to generate Go code from SQL queries.

## Setup
Start the MySQL database server.\
Then, create a `.env` file in the root directory and put the following into it:
```
PORT=<port for your server>
DSN=<user>:<password>@tcp(<host>:<port>)/<dbname>?parseTime=true
JWT_SECRET=<secret key to sign and verify jwt>
```

Build and run the server with this command:
```
go build && go-ask-and-answer.exe
```