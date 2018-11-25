package main

import (
	"context"
	"crypto/tls"
	"flag"
	"google.golang.org/grpc/credentials"
	"io/ioutil"
	"log"

	pb "github.com/smallnest/grpc-examples/calloption/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

var (
	address  = flag.String("addr", "localhost:8972", "address")
	name     = flag.String("n", "world", "name")
	certFile = "./config/tls-config/localhost/cert.pem"
	keyFile  = "./config/tls-config/localhost/key.pem"
)

type perrpccred struct{}

func (p *perrpccred) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	m := make(map[string]string)
	m["k1"] = "val1"
	return m, nil
}

func (p *perrpccred) RequireTransportSecurity() bool {
	return false
}

type authInfo struct{}

func (ai *authInfo) AuthType() string {
	return "test"
}

func main() {
	flag.Parse()

	// 连接服务器
	conn, err := grpc.Dial(*address, grpc.WithTransportCredentials(credentials.NewTLS(getTLSConfig(certFile, keyFile))))
	if err != nil {
		log.Fatalf("faild to connect: %v", err)
	}
	defer conn.Close()

	c := pb.NewGreeterClient(conn)

	//unary
	ctx := context.Background()

	p := &peer.Peer{
		AuthInfo: &authInfo{},
	}
	callOption := grpc.Peer(p)

	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: *name}, callOption)

	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Greeting: %s", r.Message)
	log.Printf("peer: %+v", p)

}

func getTLSConfig(certFile, keyFile string) *tls.Config {
	cert, _ := ioutil.ReadFile(certFile)
	key, _ := ioutil.ReadFile(keyFile)
	var demoKeyPair *tls.Certificate
	pair, err := tls.X509KeyPair(cert, key)
	if err != nil {
		log.Fatalf("TLS KeyPair err: %v\n", err)
	}
	demoKeyPair = &pair
	return &tls.Config{
		Certificates: []tls.Certificate{*demoKeyPair},
		// NextProtos:   []string{http2.NextProtoTLS}, // HTTP2 TLS支持
	}
}
