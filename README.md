# TinyKV

TinyKV is a fast, simple, and lightweight key-value store implemented in Go. It is compatible with the **Redis CLI** and supports the following commands:

## Supported Commands

- **Key-Value Operations**
  - `SET`
  - `GET`
  - `INCR`
  - `DECR`
  - `INCRBY`
  - `DECRBY`
  - `DEL`

- **String Operations**
  - `APPEND`

- **List Operations**
  - `LPUSH`
  - `LRANGE`
  - `LPOP`
  - `RPOP`
  - `RPUSH`

- **Hash Operations**
  - `HSET`
  - `HGET`
  - `HGETALL`
  - `HDEL`

## Usage

### Local Setup

1. Clone the Repo:
    ```sh
    git clone https://github.com/Ayan-sh03/TinyKV.git
    ```
2. Run:
    ```sh
    go mod tidy
    ```
3. Build the project:
    ```sh
    go build
    ```
4. Start the server:
    ```sh
    ./tinykv.exe
    ```
5. A Redis CLI compatible server will open at port 6379.

### Getting Started with Docker

1. Build the Docker image:
    ```sh
    docker build -t tinykv .
    ```
2. Run the Docker container:
    ```sh
    docker run -d -p 6379:6379 tinykv
    ```

## Contributing

Contributions are welcome! If you find a bug or have a feature request, please open an issue on the GitHub repository. If you'd like to contribute code, please fork the repository and submit a pull request.
