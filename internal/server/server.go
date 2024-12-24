package server

import (
	"encoding/binary"
	"errors"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/valyala/fasthttp"
)

const (
	connectionTimeout = 10 * time.Second // Connection timeout for target connections
	idleTimeout       = 10 * time.Minute // Idle timeout for connections
)

const socksVersion = 5 // SOCKS protocol version

// SOCKS5 response codes.
//
//nolint:gochecknoglobals
var (
	success                 = []byte{socksVersion, 0x00, 0x00, 0x01, 0, 0, 0, 0, 0, 0}
	failure                 = []byte{socksVersion, 0x01, 0x00, 0x01, 0, 0, 0, 0, 0, 0}
	connectionRefused       = []byte{socksVersion, 0x05, 0x00, 0x01, 0, 0, 0, 0, 0, 0}
	commandNotSupported     = []byte{socksVersion, 0x07, 0x00, 0x01, 0, 0, 0, 0, 0, 0}
	addressTypeNotSupported = []byte{socksVersion, 0x08, 0x00, 0x01, 0, 0, 0, 0, 0, 0}
)

// Interface for closeWriter to close the write side of a connection.
type closeWriter interface {
	CloseWrite() error
}

// options defines configurable settings for the Server.
type options struct {
	network     string
	address     string
	dnsResolver dnsResolver
	metrics     metrics
}

// Server represents a SOCKS5 proxy server.
type Server struct {
	network           string
	address           string
	dnsResolver       dnsResolver
	metrics           metrics
	activeConnections atomic.Int64
}

// New returns a new Server instance with provided options.
func New(opts ...Option) *Server {
	o := &options{
		network:     "tcp",
		address:     "127.0.0.1:1080",
		dnsResolver: &defaultDNSResolver{},
		metrics:     &defaultMetrics{},
	}

	for _, opt := range opts {
		opt(o)
	}

	return &Server{
		network:     o.network,
		address:     o.address,
		dnsResolver: o.dnsResolver,
		metrics:     o.metrics,
	}
}

// Serve starts the server and listens for incoming client connections.
func (s *Server) Serve() {
	go s.startMonitoring()

	listener, err := net.Listen(s.network, s.address)
	if err != nil {
		log.Fatalf("Failed to start listener: %v", err)
	}
	defer listener.Close()

	log.Printf("Server is running on port %s...", s.address)

	allowedIP := os.Getenv("ALLOWED_IP")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			s.metrics.Increment(mErrorsAccept)

			continue
		}

		if allowedIP != "" {
			if tcpAddr, ok := conn.RemoteAddr().(*net.TCPAddr); ok && allowedIP != tcpAddr.IP.String() {
				log.Printf("Connection from %s is not allowed", conn.RemoteAddr().String())
				s.metrics.Increment(mErrorsUnauthorizedIP)

				continue
			}
		}

		go func() {
			s.activeConnections.Add(1)
			s.handle(conn)
			s.activeConnections.Add(-1)
		}()
	}
}

// handle manages the SOCKS5 client connection and traffic forwarding.
//
//nolint:cyclop,gocognit,funlen
func (s *Server) handle(client net.Conn) {
	defer client.Close()
	defer s.metrics.Timer(mConnectionDuration)()

	// Read authentication methods
	buf := make([]byte, 2)
	if _, err := io.ReadFull(client, buf); err != nil {
		s.reply(client, failure, mErrorsGreeting, "Failed to read greeting", err)
		return
	}

	// Check SOCKS version
	if buf[0] != socksVersion {
		s.reply(client, failure, mErrorsVersion, "Unsupported SOCKS version", nil)
		return
	}

	// Read and handle authentication methods
	methods := make([]byte, int(buf[1]))
	if _, err := io.ReadFull(client, methods); err != nil {
		s.reply(client, failure, mErrorsAuthMethods, "Failed to read auth methods", err)
		return
	}

	// Send "no authentication" response
	if _, err := client.Write([]byte{socksVersion, 0x00}); err != nil {
		log.Printf("Failed to write auth response: %v", err)
		return
	}

	// Read connection request header
	buf = make([]byte, 4)
	if _, err := io.ReadFull(client, buf); err != nil {
		s.reply(client, failure, mErrorsRequestHeader, "Failed to read request header", err)
		return
	}

	// Verify command (only CONNECT is supported)
	if buf[0] != socksVersion || buf[1] != 0x01 {
		s.reply(client, commandNotSupported, mErrorsCommand, "Unsupported request version or command", nil)
		return
	}

	// Parse target address and establish connection
	var destIP net.IP

	switch buf[3] {
	case 0x01: // IPv4 address
		buf = make([]byte, 4)
		if _, err := io.ReadFull(client, buf); err != nil {
			s.reply(client, failure, mErrorsAddressType, "Failed to read IPv4 address", err)
			return
		}

		destIP = buf

	case 0x03: // Domain name
		buf = make([]byte, 1)
		if _, err := io.ReadFull(client, buf); err != nil {
			s.reply(client, failure, mErrorsAddressType, "Failed to read domain length", err)
			return
		}

		domain := make([]byte, buf[0])
		if _, err := io.ReadFull(client, domain); err != nil {
			s.reply(client, failure, mErrorsAddressType, "Failed to read domain", err)
			return
		}

		ip, err := s.dnsResolver.Resolve(string(domain))
		if err != nil {
			s.reply(client, connectionRefused, mErrorsDNSResolve, "DNS resolve failed", err)
			return
		}

		destIP = ip

	case 0x04: // IPv6 address
		buf = make([]byte, 16)
		if _, err := io.ReadFull(client, buf); err != nil {
			s.reply(client, failure, mErrorsAddressType, "Failed to read IPv6 address", err)
			return
		}

		destIP = buf

	default:
		s.reply(client, addressTypeNotSupported, mErrorsAddressType, "Unsupported address type", nil)
		return
	}

	// Read target port
	buf = make([]byte, 2)
	if _, err := io.ReadFull(client, buf); err != nil {
		s.reply(client, failure, mErrorsPort, "Failed to read port", err)
		return
	}

	dest := destIP.String() + ":" + strconv.Itoa(int(binary.BigEndian.Uint16(buf)))

	t := s.metrics.Timer(mTargetDialDuration)

	// Connect to the target address
	target, err := fasthttp.DialTimeout(dest, connectionTimeout)
	if err != nil {
		s.reply(client, connectionRefused, mErrorsConnectionRefused, "Failed to connect to target", err)
		return
	}
	defer target.Close()

	t()

	// Send success response to client
	if _, err = client.Write(success); err != nil {
		log.Printf("Failed to write success response: %v", err)
		return
	}

	client.SetDeadline(time.Now().Add(idleTimeout))
	target.SetDeadline(time.Now().Add(idleTimeout))

	// Bidirectional data transfer
	var wg sync.WaitGroup

	wg.Add(2)

	go func() {
		defer wg.Done()

		bytesCopied, err := io.Copy(target, client)
		if err != nil && isCriticalNetworkError(err) {
			s.reply(client, nil, mErrorsCopyToTarget, "Error copying from client to target", err)
		}

		s.metrics.Count(mBytesSent, bytesCopied)

		if conn, ok := target.(closeWriter); ok {
			conn.CloseWrite()
		}
	}()

	go func() {
		defer wg.Done()

		bytesCopied, err := io.Copy(client, target)
		if err != nil && isCriticalNetworkError(err) {
			s.reply(client, nil, mErrorsCopyToClient, "Error copying from target to client", err)
		}

		s.metrics.Count(mBytesReceived, bytesCopied)

		if conn, ok := client.(closeWriter); ok {
			conn.CloseWrite()
		}
	}()

	wg.Wait()
}

// reply sends a response to the client, increments the error
// metric if present, and logs the message and error if applicable.
func (s *Server) reply(client net.Conn, response []byte, metric string, message string, err error) {
	if errors.Is(err, io.EOF) {
		return
	}

	if response != nil {
		client.Write(response)
	}

	s.metrics.Increment(metric)

	if err != nil {
		log.Printf("%s: %v", message, err)
	} else {
		log.Println(message)
	}
}

// startMonitoring monitors active connections.
func (s *Server) startMonitoring() {
	for {
		s.metrics.Gauge(mActiveConnections, float64(s.activeConnections.Load()))

		time.Sleep(1 * time.Second)
	}
}

// isCriticalNetworkError checks if an error is critical.
func isCriticalNetworkError(err error) bool {
	if err == nil {
		return false
	}

	var ne net.Error
	if errors.As(err, &ne) && ne.Timeout() {
		return false
	}

	if strings.Contains(err.Error(), "connection reset by peer") || strings.Contains(err.Error(), "broken pipe") {
		return false
	}

	return true
}
