# TinyKV

TinyKV is a fast, simple, and lightweight key-value store implemented in Go. It is compatible with the **Redis CLI** and supports the following commands:

## Supported Commands
SET, GET, INCR, DECR, INCRBY, DECRBY, DEL

APPEND, LPUSH, LRANGE, LPOP, RPOP, RPUSH

HSET, HGET, HGETALL, HDEL

## Usage 

1. Clone the Repo
2. 
    ``` 
    go mod tidy 
    ```
3. ```
    go build 
    ```
4. ```
    ./tinykv.exe
    ```
5. A redis cli compatible server will open at port 6379

## Contributing

Contributions are welcome! If you find a bug or have a feature request, please open an issue on the GitHub repository. If you'd like to contribute code, please fork the repository and submit a pull reque
