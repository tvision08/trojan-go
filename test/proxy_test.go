package test

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"sync"
	"testing"
	"time"

	"net/http"
	_ "net/http/pprof"

	"github.com/p4gefau1t/trojan-go/api"
	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
	_ "github.com/p4gefau1t/trojan-go/log/golog"
	tp "github.com/p4gefau1t/trojan-go/proxy"
	"github.com/p4gefau1t/trojan-go/proxy/client"
	"github.com/p4gefau1t/trojan-go/proxy/server"
	_ "github.com/p4gefau1t/trojan-go/router/mixed"
	_ "github.com/p4gefau1t/trojan-go/stat/memory"
	_ "github.com/p4gefau1t/trojan-go/stat/mysql"
	"golang.org/x/net/proxy"
	"golang.org/x/net/websocket"
	"google.golang.org/grpc"
)

var cert string = `
-----BEGIN CERTIFICATE-----
MIIDZTCCAk0CFFphZh018B5iAD9F5fV4y0AlD0LxMA0GCSqGSIb3DQEBCwUAMG8x
CzAJBgNVBAYTAlVTMQ0wCwYDVQQIDARNYXJzMRMwEQYDVQQHDAppVHJhbnN3YXJw
MRMwEQYDVQQKDAppVHJhbnN3YXJwMRMwEQYDVQQLDAppVHJhbnN3YXJwMRIwEAYD
VQQDDAlsb2NhbGhvc3QwHhcNMjAwMzMxMTAwMDUxWhcNMzAwMzI5MTAwMDUxWjBv
MQswCQYDVQQGEwJVUzENMAsGA1UECAwETWFyczETMBEGA1UEBwwKaVRyYW5zd2Fy
cDETMBEGA1UECgwKaVRyYW5zd2FycDETMBEGA1UECwwKaVRyYW5zd2FycDESMBAG
A1UEAwwJbG9jYWxob3N0MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA
ml44fThYMkCcT627o7ibEs7mq2WOhImjDwYijYJ1684BatrCsHJNcw8PJGTuP+tg
GdngmALjA3l+RipjaE/UK4FJrAjruphA/hOCjZfWqk8KBR4qk0OltxCMWJlp/XCM
9ny1ogFdWUlBbqThs4NWSOUESgxf/Be2njeiOrngGR31qxSiLCLBvafIhKqq/4av
Rlx0Ht770uvF97MlAj1ASAvzTZICHAfUZxEdWl0J4MBbG7SNcnMBbyAF+s60eFTa
4RGMfRGnUa2Fzz/gfjhvfSIGeLQ3JRG6sl6jkc5xe0PZzhq3UNpK0gtQ48yy9CSP
neZnrynoKks7XC2bizsr3QIDAQABMA0GCSqGSIb3DQEBCwUAA4IBAQAHS/xuG5+F
yGU3N6V4kv+HbKqHaXNOq4zKVsCc1k7vg4MFFpKUJKxtJYooCI8n2ypp5XRUTIGQ
bmEbVcIPqm9Rf/4vHtF0falNCwieAbXDkiEHoykRmmU1UE/ccPA7X8NO9aVLJAJO
N2Li8MH0Ixgs02pQH56eyGKoRBWPR5C3ETQ9Leqvazg6Dn1iJWvmfF0mOte5228s
mZJOntF9t8MZOJdIWGdrUHn6euRfhd0btkmL/NUDzeCTwJcuPORLxkBbCP5mTC6G
GnLS5Z4oRYgCgvT2pLtcM0r48hYjwgjXFQ4zalkW6YI9LPpqwwMhhOzINlXjBaDi
Haz8uKI4EciU
-----END CERTIFICATE-----
`

var key string = `
-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAml44fThYMkCcT627o7ibEs7mq2WOhImjDwYijYJ1684BatrC
sHJNcw8PJGTuP+tgGdngmALjA3l+RipjaE/UK4FJrAjruphA/hOCjZfWqk8KBR4q
k0OltxCMWJlp/XCM9ny1ogFdWUlBbqThs4NWSOUESgxf/Be2njeiOrngGR31qxSi
LCLBvafIhKqq/4avRlx0Ht770uvF97MlAj1ASAvzTZICHAfUZxEdWl0J4MBbG7SN
cnMBbyAF+s60eFTa4RGMfRGnUa2Fzz/gfjhvfSIGeLQ3JRG6sl6jkc5xe0PZzhq3
UNpK0gtQ48yy9CSPneZnrynoKks7XC2bizsr3QIDAQABAoIBAFpYUo9W7qdakSFA
+NS1Mm0rkm01nteLBlfAq3BOrl030DSNm+xQuWthoOcX+yiFxVTb40qURfC+plzC
ajOepPphTJDXF7+5ZDBPktTzzLsYTzD3mstdiBtAICOqhhHCUX3hNxx91/htm1H6
Re4eK921y3DbFUIhTswCm3vrVXDc4yTXtURGllVzo40K/1Of39CpufKFdpJ81HV+
h/VW++h3o+sFV4KqcqIjClxBfDxoJpBaRlOCunTiHqZNvqO+EPqPR5zdn34werjU
xQEvPzmz+ClwnaEXQxYWgIcYQii9VNsHogDxEw4R31S7lVrUt0f0atDmGJip1lPb
E7IomAECgYEAzKQ3PzBV46nUNfVO9SODpf14Z+xYfLKouPC+Qnepwp0V0JS6zY1+
Wzskyb80drjnoQraWSEvGsX+tEWeLcnjN7JuMu/U8DPKRcQ+Q2dsVo/q4sfBOgvl
VhPNMZLfa7NIkRUx2KXku++Ep0Xtak0dskrfQrZnvhymRPyWuIMM6IECgYEAwRwL
Gt/ZZdUueE/hwT3c1hNn6igeDLOwK2t6frib+Ofw5oCAQxtTROvP1ljlnWUPkeIS
uzTusmqucalcK3lCHIsyHLwApOI/B31M971pxMVBRZ0wIbBaoarCGND7gi8JUPFR
VErGcAB5YnpRlmfLPEgw2o7DpjsDc2KmdE9oNV0CgYEAmfNEWLYtNztxGTK1treD
96ELLutf2lexlIgQKgLJ5E22tpbdPXwfvdRtpZTBjDsojj+S6hCL1lFzfv0MtZe2
5xTF0G4avKXJmti6moy4tRpJ81ehZuDCJBJ7gLrkd6qFghf2yuxqenQDUK/Lnvfq
ylGHSjHdM+lrsGRxotd8I4ECgYBoo4GA9nseqv2bQ+3YgGUBu1I7l7FwwI1decfO
ksoxfb0Tqd3WfyAH4J+mTlVdjD17lzz/JBeTpisQe+ztwa8JOIPW/ih7L/1nWYYz
V/fQH/LWfe5u0tjJcXXrbJJcYJBzw8+GFV6hoiAkNJOxJF0ENToDtAhgMuoTxAje
TYjyIQKBgQCmHkLLq0Bj3FpIOVrwo2gNvQteNPa7jkkGp4lljO8JQUHhCHDGWKEH
MUJ0EFsxS/EaQa+rW6jHhs3GyBA2TxmC783stAOOEX+hO/zpcbzdCWgp6eZ0aGMW
WS94/5WE/lwHJi8ZPSjH1AURCzXhUi4fGvBrNBtry95e+jcEvP5c0g==
-----END RSA PRIVATE KEY-----
`

func getKeyPair() []tls.Certificate {
	cert, err := tls.X509KeyPair([]byte(cert), []byte(key))
	common.Must(err)
	return []tls.Certificate{cert}
}

func getTLSConfig() conf.TLSConfig {
	KeyPair := getKeyPair()
	pool := x509.NewCertPool()
	if ok := pool.AppendCertsFromPEM([]byte(cert)); !ok {
		panic("invalid cert")
	}
	c := conf.TLSConfig{
		SNI:             "localhost",
		CertPool:        pool,
		KeyPair:         KeyPair,
		Verify:          true,
		ReuseSession:    true,
		SessionTicket:   true,
		FallbackAddress: common.NewAddress("127.0.0.1", 10080, "tcp"),
		ALPN: []string{
			"http/1.1",
			"h2",
		},
		Fingerprint: "firefox",
	}
	return c
}

func getHash(password string) map[string]string {
	hash := common.SHA224String(password)
	m := make(map[string]string)
	m[hash] = password
	return m
}

func getPasswords(password string) []string {
	return []string{password}
}

func getBasicServerConfig() *conf.GlobalConfig {
	config := &conf.GlobalConfig{
		LocalAddress:  common.NewAddress("127.0.0.1", 4445, "tcp"),
		RemoteAddress: common.NewAddress("127.0.0.1", 10080, "tcp"),
		TLS:           getTLSConfig(),
		Hash:          getHash("trojanpassword"),
		Passwords:     getPasswords("trojanpassword"),
		BufferSize:    512 * 1024,
	}
	return config
}

func getBasicClientConfig() *conf.GlobalConfig {
	config := &conf.GlobalConfig{
		LocalAddress:  common.NewAddress("127.0.0.1", 4444, "tcp"),
		RemoteAddress: common.NewAddress("127.0.0.1", 4445, "tcp"),
		TLS:           getTLSConfig(),
		Hash:          getHash("trojanpassword"),
		Passwords:     getPasswords("trojanpassword"),
		BufferSize:    512 * 1024,
	}
	return config
}

func addWsConfig(config *conf.GlobalConfig) *conf.GlobalConfig {
	config.Websocket = conf.WebsocketConfig{
		Enabled:             true,
		HostName:            "127.0.0.1",
		Path:                "/websocket",
		ObfuscationPassword: "123456789",
		DoubleTLS:           true,
		TLS:                 getTLSConfig(),
	}
	hash := md5.New()
	hash.Write([]byte(config.Websocket.ObfuscationPassword))
	config.Websocket.ObfuscationKey = hash.Sum(nil)
	return config
}

func addMuxConfig(config *conf.GlobalConfig) *conf.GlobalConfig {
	config.Mux = conf.MuxConfig{
		Enabled:     true,
		Concurrency: 8,
		IdleTimeout: 30,
	}
	return config
}

func addRouterConfig(config *conf.GlobalConfig) *conf.GlobalConfig {
	config.Router = conf.RouterConfig{
		Enabled:       true,
		BypassList:    []byte("127.0.0.1\nlocalhost"),
		DefaultPolicy: "proxy",
	}
	return config
}

func addTCPOption(config *conf.GlobalConfig) *conf.GlobalConfig {
	config.TCP = conf.TCPConfig{
		KeepAlive:    true,
		FastOpen:     true,
		NoDelay:      true,
		FastOpenQLen: 5,
	}
	return config
}

func addMySQLConfig(t *testing.T, config *conf.GlobalConfig) *conf.GlobalConfig {
	database := os.Getenv("mysql_database")
	username := os.Getenv("mysql_username")
	password := os.Getenv("mysql_password")
	if database == "" || username == "" || password == "" {
		t.Skip("skipping mysql test")
		database = "trojan"
		username = "root"
		password = "password"
	}
	config.MySQL = conf.MySQLConfig{
		Enabled:    true,
		ServerHost: "127.0.0.1",
		ServerPort: 3306,
		Database:   database,
		Username:   username,
		Password:   password,
		CheckRate:  1,
	}
	return config
}

func addAPIConfig(config *conf.GlobalConfig) *conf.GlobalConfig {
	config.API = conf.APIConfig{
		Enabled:    true,
		APIAddress: common.NewAddress("127.0.0.1", 10000, "tcp"),
	}
	return config
}

func addDNSConfig(config *conf.GlobalConfig) *conf.GlobalConfig {
	config.DNS = []string{
		"dot://223.5.5.5",
		"8.8.8.8",
	}
	return config
}

func RunClient(ctx context.Context, config *conf.GlobalConfig) {
	c := client.Client{}
	r, err := c.Build(config)
	common.Must(err)
	go r.Run()
	<-ctx.Done()
	r.Close()
}

func RunForward(ctx context.Context, config *conf.GlobalConfig) {
	c := client.Forward{}
	r, err := c.Build(config)
	common.Must(err)
	go r.Run()
	<-ctx.Done()
	r.Close()
}

func RunServer(ctx context.Context, config *conf.GlobalConfig) {
	s := server.Server{}
	r, err := s.Build(config)
	common.Must(err)
	go r.Run()
	<-ctx.Done()
	r.Close()
}

func CheckClientServer(t *testing.T, clientConfig *conf.GlobalConfig, serverConfig *conf.GlobalConfig) {
	ctx, cancel := context.WithCancel(context.Background())
	go RunEchoTCPServer(ctx)
	go RunServer(ctx, serverConfig)
	go RunClient(ctx, clientConfig)

	time.Sleep(time.Second)

	payloadSize := 1024
	sendBuf := GeneratePayload(payloadSize)
	recvBuf := make([]byte, payloadSize)

	dialer, err := proxy.SOCKS5("tcp", "127.0.0.1:4444", nil, nil)
	common.Must(err)
	conn, err := dialer.Dial("tcp", "127.0.0.1:5000")
	common.Must(err)
	common.Must2(conn.Write(sendBuf))
	common.Must2(conn.Read(recvBuf))
	if !bytes.Equal(sendBuf, recvBuf) {
		t.Fatal("not equal")
	}
	conn.Close()
	cancel()
	time.Sleep(time.Second)
}

func CheckForwardServer(t *testing.T, clientConfig *conf.GlobalConfig, serverConfig *conf.GlobalConfig) {
	time.Sleep(time.Second)
	ctx, cancel := context.WithCancel(context.Background())
	clientConfig.TargetAddress = common.NewAddress("127.0.0.1", 5000, "tcp")
	go RunEchoTCPServer(ctx)
	go RunEchoUDPServer(ctx)
	go RunServer(ctx, serverConfig)
	go RunForward(ctx, clientConfig)

	time.Sleep(time.Second)

	payloadSize := 1024
	sendBuf := GeneratePayload(payloadSize)
	recvBuf := make([]byte, payloadSize)

	conn, err := net.Dial("tcp", "127.0.0.1:4444")
	common.Must(err)
	common.Must2(conn.Write(sendBuf))
	common.Must2(conn.Read(recvBuf))
	if !bytes.Equal(sendBuf, recvBuf) {
		t.Fatal("not equal")
	}
	conn.Close()

	conn, err = net.Dial("udp", "127.0.0.1:4444")
	common.Must(err)
	common.Must2(conn.Write(sendBuf))
	common.Must2(conn.Read(recvBuf))
	if !bytes.Equal(sendBuf, recvBuf) {
		t.Fatal("not equal")
	}
	conn.Close()
	cancel()
}

func SingleThreadSpeedTestClientServer(b *testing.B, clientConfig *conf.GlobalConfig, serverConfig *conf.GlobalConfig) {
	time.Sleep(time.Second)
	ctx, cancel := context.WithCancel(context.Background())
	go RunBlackHoleTCPServer(ctx)
	go RunServer(ctx, serverConfig)
	go RunClient(ctx, clientConfig)

	time.Sleep(time.Second)
	dialer, err := proxy.SOCKS5("tcp", "127.0.0.1:4444", nil, nil)
	common.Must(err)
	conn, err := dialer.Dial("tcp", "127.0.0.1:5000")
	common.Must(err)
	mbytes := 2048
	payload := GeneratePayload(1024 * 1024 * mbytes)
	t1 := time.Now()
	common.Must2(conn.Write(payload))
	t2 := time.Now()
	speed := float64(mbytes) / t2.Sub(t1).Seconds()
	b.Log("single-thread link speed:", speed, "MiB/s")
	conn.Close()
	cancel()
}

func MultiThreadSpeedTestClientServer(b *testing.B, clientConfig *conf.GlobalConfig, serverConfig *conf.GlobalConfig) {
	time.Sleep(time.Second)
	ctx, cancel := context.WithCancel(context.Background())
	go RunBlackHoleTCPServer(ctx)
	go RunServer(ctx, serverConfig)
	go RunClient(ctx, clientConfig)

	time.Sleep(time.Second)
	dialer, err := proxy.SOCKS5("tcp", "127.0.0.1:4444", nil, nil)
	common.Must(err)
	mbytes := 2048
	threads := 16
	payload := GeneratePayload(1024 * 1024 * mbytes / threads)

	wg := sync.WaitGroup{}
	wg.Add(threads)
	t1 := time.Now()
	for i := 0; i < threads; i++ {
		go func() {
			conn, err := dialer.Dial("tcp", "127.0.0.1:5000")
			common.Must(err)
			common.Must2(conn.Write(payload))
			wg.Done()
			conn.Close()
		}()
	}
	wg.Wait()
	t2 := time.Now()
	speed := float64(mbytes) / t2.Sub(t1).Seconds()

	b.Log("multi-thread link speed:", speed, "MiB/s")
	cancel()
}

func TestRealProxy(t *testing.T) {
	if os.Getenv("real_test") == "" {
		t.Skip("skipping real proxy test")
	}
	clientConfig := getBasicClientConfig()
	serverConfig := getBasicServerConfig()
	go RunClient(context.Background(), clientConfig)
	go RunHelloHTTPServer(context.Background())
	RunServer(context.Background(), serverConfig)
}

func TestRealClient(t *testing.T) {
	if os.Getenv("real_test") == "" {
		t.Skip("skipping real proxy test")
	}
	b, err := ioutil.ReadFile("/etc/trojan-go/config.json")
	common.Must(err)
	config, err := conf.ParseJSON(b)
	common.Must(err)
	RunClient(context.Background(), config)
}

func TestNormal(t *testing.T) {
	CheckClientServer(t, getBasicClientConfig(), getBasicServerConfig())
	CheckForwardServer(t, getBasicClientConfig(), getBasicServerConfig())
}

func TestMux(t *testing.T) {
	clientConfig := addMuxConfig(getBasicClientConfig())
	serverConfig := getBasicServerConfig()
	CheckClientServer(t, clientConfig, serverConfig)
	CheckForwardServer(t, clientConfig, serverConfig)
}

func TestWebsocket(t *testing.T) {
	clientConfig := addWsConfig(getBasicClientConfig())
	serverConfig := addWsConfig(getBasicServerConfig())
	CheckClientServer(t, clientConfig, serverConfig)
	CheckForwardServer(t, clientConfig, serverConfig)
}

func TestWebsocketMux(t *testing.T) {
	clientConfig := addMuxConfig(addWsConfig(getBasicClientConfig()))
	serverConfig := addWsConfig(getBasicServerConfig())
	CheckClientServer(t, clientConfig, serverConfig)
	CheckForwardServer(t, clientConfig, serverConfig)
}

func BenchmarkNormal(b *testing.B) {
	clientConfig := getBasicClientConfig()
	serverConfig := getBasicServerConfig()
	SingleThreadSpeedTestClientServer(b, clientConfig, serverConfig)
	MultiThreadSpeedTestClientServer(b, clientConfig, serverConfig)
}

func BenchmarkMux(b *testing.B) {
	clientConfig := addMuxConfig(getBasicClientConfig())
	serverConfig := getBasicServerConfig()
	SingleThreadSpeedTestClientServer(b, clientConfig, serverConfig)
	MultiThreadSpeedTestClientServer(b, clientConfig, serverConfig)
}

func BenchmarkWebsocket(b *testing.B) {
	clientConfig := addWsConfig(getBasicClientConfig())
	serverConfig := addWsConfig(getBasicServerConfig())
	SingleThreadSpeedTestClientServer(b, clientConfig, serverConfig)
	MultiThreadSpeedTestClientServer(b, clientConfig, serverConfig)
}

func BenchmarkMuxWebsocket(b *testing.B) {
	clientConfig := addMuxConfig(addWsConfig(getBasicClientConfig()))
	serverConfig := addWsConfig(getBasicServerConfig())
	SingleThreadSpeedTestClientServer(b, clientConfig, serverConfig)
	MultiThreadSpeedTestClientServer(b, clientConfig, serverConfig)
}

func TestWebsocketShadow(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	go RunHelloHTTPServer(ctx)
	serverConfig := addWsConfig(getBasicServerConfig())
	go RunServer(ctx, serverConfig)
	time.Sleep(time.Second)

	//test http
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	resp, err := httpClient.Get("https://127.0.0.1:4445")
	common.Must(err)
	body, err := ioutil.ReadAll(resp.Body)
	common.Must(err)
	if string(body) != "HelloWorld" {
		t.Fatal("http shadow")
	}

	//test websocket
	conn, err := tls.Dial("tcp", "127.0.0.1:4445", &tls.Config{InsecureSkipVerify: true})
	common.Must(err)
	wsConfig, err := websocket.NewConfig("wss://127.0.0.1:65535/websocket", "https://127.0.0.1:65535")
	common.Must(err)
	wsClient, err := websocket.NewClient(wsConfig, conn)
	common.Must(err)
	buf := [100]byte{}
	common.Must2(wsClient.Write([]byte("I'm GFW1231231231231212391273871283719823791237912398721933123")))
	n, err := wsClient.Read(buf[:])
	common.Must(err)
	if string(buf[:n]) != "HelloWorld" {
		t.Fatal("ws shadow")
	}
	conn.Close()
	cancel()
}

func TestShadow(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	go RunHelloHTTPServer(ctx)
	serverConfig := getBasicServerConfig()
	go RunServer(ctx, serverConfig)
	time.Sleep(time.Second)

	//test http
	httpClient := &http.Client{
		//some config
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	resp, err := httpClient.Get("https://127.0.0.1:4445")
	common.Must(err)
	body, err := ioutil.ReadAll(resp.Body)
	common.Must(err)
	if string(body) != "HelloWorld" {
		t.Fatal("http shadow")
	}

	//fallback
	resp, err = http.Get("http://127.0.0.1:4445")
	common.Must(err)
	body, err = ioutil.ReadAll(resp.Body)
	common.Must(err)
	if string(body) != "HelloWorld" {
		t.Fatal("http shadow")
	}
	cancel()
}

func TestAutoClientID(t *testing.T) {
	serverConfig := getBasicServerConfig()
	clientConfig := getBasicClientConfig()
	clientConfig.TLS.Fingerprint = "auto"
	CheckClientServer(t, clientConfig, serverConfig)
}

func TestTCPOptions(t *testing.T) {
	serverConfig := addTCPOption(getBasicServerConfig())
	clientConfig := addTCPOption(getBasicClientConfig())
	CheckClientServer(t, clientConfig, serverConfig)
}

func TestMySQL(t *testing.T) {
	serverConfig := addMySQLConfig(t, getBasicServerConfig())
	clientConfig := getBasicClientConfig()
	clientConfig.Passwords = getPasswords("mysqlpassword")
	clientConfig.Hash = getHash("mysqlpassword")
	CheckClientServer(t, clientConfig, serverConfig)
}

func TestServerAPI(t *testing.T) {
	serverConfig := addAPIConfig(getBasicServerConfig())
	clientConfig := getBasicClientConfig()
	clientConfig.Hash = getHash("apitest")
	clientConfig.Passwords = getPasswords("apitest")

	ctx, cancel := context.WithCancel(context.Background())

	go RunBlackHoleTCPServer(ctx)
	go RunServer(ctx, serverConfig)
	go RunClient(ctx, clientConfig)

	time.Sleep(time.Second * 2)
	grpcConn, err := grpc.Dial("127.0.0.1:10000", grpc.WithInsecure())
	common.Must(err)
	server := api.NewTrojanServerServiceClient(grpcConn)

	listUserStream, err := server.ListUsers(ctx, &api.ListUserRequest{})
	common.Must(err)
	defer listUserStream.CloseSend()
	for {
		resp, err := listUserStream.Recv()
		if err != nil {
			break
		}
		fmt.Println(resp.User.Hash)
		fmt.Println(*resp.SpeedCurrent)
		fmt.Println(*resp.SpeedLimit)
	}
	listUserStream.CloseSend()
	setUserStream, err := server.SetUsers(ctx)
	setUserStream.Send(&api.SetUserRequest{
		User: &api.User{
			Hash: common.SHA224String("apitest"),
		},
		SpeedLimit: &api.Speed{
			UploadSpeed: 1024 * 1024 * 2,
		},
		Operation: api.SetUserRequest_Add,
	})
	resp3, err := setUserStream.Recv()
	if err != nil || !resp3.Success {
		t.Fail()
	}
	setUserStream.CloseSend()

	go func() {
		dialer, err := proxy.SOCKS5("tcp", "127.0.0.1:4444", nil, nil)
		common.Must(err)
		conn, err := dialer.Dial("tcp", "127.0.0.1:5000")
		common.Must(err)
		mbytes := 16
		payload := GeneratePayload(1024 * 1024 * mbytes)
		t1 := time.Now()
		conn.Write(payload)
		t2 := time.Now()
		speed := float64(mbytes) / t2.Sub(t1).Seconds()
		t.Log("single-thread link speed:", speed, "MiB/s")
		conn.Close()
	}()

	time.Sleep(time.Second * 5)
	listUserStream, err = server.ListUsers(ctx, &api.ListUserRequest{})
	common.Must(err)
	defer listUserStream.CloseSend()
	for {
		resp, err := listUserStream.Recv()
		if err != nil {
			break
		}
		fmt.Println(resp.User.Hash)
		fmt.Println(resp.SpeedCurrent.UploadSpeed)
		fmt.Println(resp.SpeedLimit.UploadSpeed)
	}
	listUserStream.CloseSend()
	cancel()
}

func TestDNS(t *testing.T) {
	serverConfig := addDNSConfig(getBasicServerConfig())
	clientConfig := getBasicClientConfig()
	ctx, cancel := context.WithCancel(context.Background())
	go RunServer(ctx, serverConfig)
	go RunClient(ctx, clientConfig)
	time.Sleep(time.Second)
	dialer, err := proxy.SOCKS5("tcp", "127.0.0.1:4444", nil, nil)
	common.Must(err)
	conn, err := dialer.Dial("tcp", "www.baidu.com:80")
	common.Must(err)
	httpReq, err := http.NewRequest("GET", "http://www.baidu.com", nil)
	common.Must(err)
	httpReq.Write(conn)
	buf := [1024]byte{}
	common.Must2(conn.Read(buf[:]))
	fmt.Println(string(buf[:]))
	conn.Close()
	cancel()
}

func TestJSON(t *testing.T) {
	if os.Getenv("CONFIG") == "" {
		t.Skip("skip json test")
	}
	configFile1 := "/etc/trojan-go/config.json"
	configBytes1, err := ioutil.ReadFile(configFile1)
	common.Must(err)
	config1, err := conf.ParseJSON(configBytes1)
	common.Must(err)
	r, err := tp.NewProxy(config1)
	common.Must(err)
	r.Run()
}
