package transport

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	pb "github.com/tonx22/lancktask/pb"
	"github.com/tonx22/lancktask/pkg/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"io"
	"io/ioutil"
	"log"
	"net"
	"strings"
	"time"
)

var (
	errMissingMetadata = status.Errorf(codes.InvalidArgument, "missing metadata")
	errInvalidToken    = status.Errorf(codes.Unauthenticated, "invalid token")
)

type server struct {
	pb.UnimplementedSearchSvcServer
	service service.SearchService
}

func (s *server) GetCodeByNumber(ctx context.Context, in *pb.PhoneNumberRequest) (*pb.MCCMNCCodeReply, error) {
	resp, err := s.service.GetCodeByNumber(ctx, in.PhoneNumber)
	if err != nil {
		return nil, err
	}
	return &pb.MCCMNCCodeReply{MccmncCode: resp}, err
}

func (s *server) StreamingGetCodeByNumber(stream pb.SearchSvc_StreamingGetCodeByNumberServer) error {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		resp, err := s.service.GetCodeByNumber(context.Background(), in.PhoneNumber)
		if err != nil {
			return err
		}
		err = stream.Send(&pb.MCCMNCCodeReply{MccmncCode: resp})
		if err != nil {
			return err
		}
	}
	return nil
}

func StartNewGRPCServer(s interface{}, grpcPort int, token string) error {
	svc := s.(service.SearchService)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	cert, err := tls.LoadX509KeyPair("./data/x509/server-cert.pem", "./data/x509/server-key.pem")
	if err != nil {
		return fmt.Errorf("failed to load key pair: %v", err)
	}

	ca := x509.NewCertPool()
	caBytes, err := ioutil.ReadFile("./data/x509/ca-cert.pem")
	if err != nil {
		return fmt.Errorf("failed to read ca cert: %v", err)
	}
	if ok := ca.AppendCertsFromPEM(caBytes); !ok {
		return fmt.Errorf("failed to parse ca cert")
	}

	tlsConfig := &tls.Config{
		ClientAuth:   tls.RequireAndVerifyClientCert,
		Certificates: []tls.Certificate{cert},
		ClientCAs:    ca,
	}

	interceptor := NewAuthInterceptor(token)
	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(interceptor.Unary()),
		grpc.StreamInterceptor(interceptor.Stream()),
		grpc.Creds(credentials.NewTLS(tlsConfig)),
	}

	grpcServer := grpc.NewServer(opts...)
	pb.RegisterSearchSvcServer(grpcServer, &server{service: svc})

	ch := make(chan error)
	go func() {
		ch <- grpcServer.Serve(lis)
	}()

	var e error
	select {
	case e = <-ch:
		return e
	case <-time.After(time.Second * 1):
	}
	log.Printf("GRPC server listening at %v", lis.Addr())
	return nil
}

type authInterceptor struct {
	validToken string
}

func NewAuthInterceptor(token string) *authInterceptor {
	return &authInterceptor{validToken: token}
}

func (interceptor *authInterceptor) Valid(authorization []string) bool {
	if len(authorization) < 1 {
		return false
	}
	token := strings.TrimPrefix(authorization[0], "Bearer ")
	return token == interceptor.validToken
}

func (interceptor *authInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// authentication (token verification)
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, errMissingMetadata
		}
		if !interceptor.Valid(md["authorization"]) {
			return nil, errInvalidToken
		}
		return handler(ctx, req)
	}
}

func (interceptor *authInterceptor) Stream() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// authentication (token verification)
		md, ok := metadata.FromIncomingContext(ss.Context())
		if !ok {
			return errMissingMetadata
		}
		if !interceptor.Valid(md["authorization"]) {
			return errInvalidToken
		}
		return handler(srv, ss)
	}
}
