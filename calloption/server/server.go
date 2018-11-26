package main

import (
	"crypto/tls"
	"crypto/x509"
	"io"
	"io/ioutil"
	"log"
	"net"

	"flag"

	pb "github.com/smallnest/grpc-examples/calloption/pb"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/reflection"
)

var (
	port = flag.String("p", ":8972", "port")

	rootFile = "./config/tls-config/ca/myssl/myssl_root.cer"
	certFile = "./config/tls-config/localhost/server/cert.pem"
	keyFile  = "./config/tls-config/localhost/server/key.pem"
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

type server struct{}

func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	if p, ok := peer.FromContext(ctx); ok {
		log.Printf("unary receive Peer: %+v", p)
	}

	if md, ok := metadata.FromIncomingContext(ctx); ok {
		log.Printf("unary receive MD: %+v", md)
	}
	return &pb.HelloReply{Message: "Hello " + in.Name}, nil
}

func (s *server) SayHello1(gs pb.Greeter_SayHello1Server) error {
	if md, ok := metadata.FromIncomingContext(gs.Context()); ok {
		log.Printf("streaming receive MD: %+v", md)
	}

	for {
		in, err := gs.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		gs.Send(&pb.HelloReply{Message: "Hello " + in.Name})
	}

	return nil
}

func main() {
	flag.Parse()
	log.SetFlags(log.Llongfile)

	// 将 根证书 放进证书池
	rootBuf, err := ioutil.ReadFile(rootFile)
	if err != nil {
		log.Fatalln(err.Error())
	}

	certPool := x509.NewCertPool()
	if ok := certPool.AppendCertsFromPEM(rootBuf); !ok {
		log.Fatalln("failed to append test CA")
	}

	lis, err := net.Listen("tcp", *port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	tlsConfig := getTLSConfig(certFile, keyFile)
	tlsConfig.ClientCAs = certPool
	s := grpc.NewServer(grpc.Creds(credentials.NewTLS(tlsConfig)))
	pb.RegisterGreeterServer(s, &server{})

	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
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
