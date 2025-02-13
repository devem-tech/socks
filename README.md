# SOCKS5 Proxy Server

[![Go Version](https://img.shields.io/github/go-mod/go-version/devem-tech/socks)](https://go.dev/doc)
[![License](https://img.shields.io/github/license/devem-tech/socks)](LICENSE)

A high-performance, SOCKS5 proxy server written in Go. This server is designed for secure, low-latency data forwarding between clients and remote hosts, with robust error handling, connection monitoring, and customizable metrics tracking.

## Features

- **SOCKS5 Protocol Support**: Handles basic SOCKS5 commands for TCP connections.
- **Connection Management**: Manages concurrent client connections with monitoring for active and idle timeouts.
- **Error Handling**: Detailed error logging and filtering for non-critical errors.
- **Metrics Tracking**: Real-time monitoring for:
  - Active connections
  - Data sent and received
  - Connection duration and errors
  - Bandwidth usage
- **DNS Resolution**: Supports resolving domain names to IP addresses.
- **Customizable Settings**: Easily configure network, address, and monitoring intervals.

## Installation

1. **Clone the repository**:
    ```bash
    git clone https://github.com/devem-tech/socks.git
    cd socks
    ```

2. **Install dependencies**:
    Ensure you have Go installed (version 1.17 or higher).
    ```bash
    go mod download
    ```

3. **Build the server**:
    ```bash
    go build -o server cmd/main.go
    ```

## Usage

### Running the Server

Start the server with the following command:

```bash
./server
```

By default, the server listens on `127.0.0.1:1080`. This can be configured in the `New` function by passing in custom options.

### Configuration

You can configure the server using `options` struct in the code, which allows customizing:
- **Network and Address**: Specify the network protocol (`tcp`, `udp`) and address.
- **DNS Resolver**: Set up a custom DNS resolver.
- **Metrics**: Configure metrics tracking for connection monitoring.

### Metrics

The server tracks key metrics using a `metrics` interface:
- **Active Connections**: The current number of active connections.
- **Data Sent and Received**: Bytes sent to and received from clients.
- **Error Counts**: Counts for various error types, including connection and copy errors.
- **Bandwidth Usage**: Total data transferred per second.

## Example Code

```go
package main

import (
	"log"
	"github.com/devem-tech/socks/server"
)

func main() {
	// Initialize and configure server
	srv := server.New(
		server.Network("tcp"),
		server.Address("127.0.0.1:1080"),
	)

	// Start the server
	log.Println("Starting SOCKS5 server on 127.0.0.1:1080...")
	srv.Serve()
}
```

## Contributing

We welcome contributions! If you have suggestions for improvements or find any issues, please open an issue or submit a pull request.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
