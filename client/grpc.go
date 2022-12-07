package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	pb "github.com/tonx22/lancktask/pb"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/oauth"
	"io"
	"io/ioutil"
	"os"
)

type searchService struct {
	GRPCClient pb.SearchSvcClient
}

func NewGRPCClient(token string) (*searchService, error) {
	defaultHost, ok := os.LookupEnv("GRPC_HOST")
	if !ok {
		defaultHost = "localhost"
	}
	defaultPort, ok := os.LookupEnv("GRPC_PORT")
	if !ok {
		defaultPort = "50051"
	}

	// Set up the credentials for the connection.
	perRPC := oauth.NewOauthAccess(&oauth2.Token{AccessToken: token})

	cert, err := tls.LoadX509KeyPair("./data/x509/client-cert.pem", "./data/x509/client-key.pem")
	if err != nil {
		return nil, fmt.Errorf("failed to load key pair: %v", err)
	}

	ca := x509.NewCertPool()
	caBytes, err := ioutil.ReadFile("./data/x509/ca-cert.pem")
	if err != nil {
		return nil, fmt.Errorf("failed to read ca cert: %v", err)
	}
	if ok := ca.AppendCertsFromPEM(caBytes); !ok {
		return nil, fmt.Errorf("failed to parse ca cert")
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      ca,
	}

	opts := []grpc.DialOption{
		grpc.WithPerRPCCredentials(perRPC),
		grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)),
	}

	serverAddr := fmt.Sprintf("%s:%s", defaultHost, defaultPort)
	conn, err := grpc.Dial(serverAddr, opts...)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("fail to dial: %s", err))
	}
	svc := searchService{GRPCClient: pb.NewSearchSvcClient(conn)}
	return &svc, nil
}

func (svc searchService) GetCodeByNumber(ctx context.Context, number string) (*string, error) {
	resp, err := svc.GRPCClient.GetCodeByNumber(ctx, &pb.PhoneNumberRequest{PhoneNumber: number})
	if err != nil {
		return nil, err
	}
	return &resp.MccmncCode, nil
}

func (svc searchService) StreamingGetCodeByNumber(ctx context.Context, numbers []string) (*[]string, error) {
	stream, err := svc.GRPCClient.StreamingGetCodeByNumber(ctx)
	if err != nil {
		return nil, fmt.Errorf("Stream err: %v", err)
	}

	for _, number := range numbers {
		err = stream.Send(&pb.PhoneNumberRequest{PhoneNumber: number})
		if err != nil {
			return nil, fmt.Errorf("Send error: %v", err)
		}
	}
	stream.CloseSend()

	var res []string
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to receive response due to error: %v", err)
		}
		res = append(res, resp.MccmncCode)
	}
	return &res, nil
}
