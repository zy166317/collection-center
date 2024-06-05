package signClient

import (
	"collection-center/internal/logger"
	pb "collection-center/internal/signClient/pb/offlineSign"
	"context"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"log"
)

type RemoteSigner struct {
	Host       string
	Port       string
	TlsPemPath string
	User       string
	Pass       string
}

var SignerConfig = &RemoteSigner{}

// NewClient 返回 client, conn, error
// 在使用完毕后需要调用 conn.Close() 关闭链接
func NewClient() (pb.OfflineSignClient, *grpc.ClientConn, error) {
	if SignerConfig.Host == "" || SignerConfig.Port == "" {
		return nil, nil, errors.New("no signerConfig found")
	}
	creds, err := credentials.NewClientTLSFromFile(SignerConfig.TlsPemPath, "orca.org")
	if err != nil {
		log.Fatalf("failed to load credentials: %v", err)
	}
	conn, err := grpc.Dial(
		SignerConfig.Host+":"+SignerConfig.Port,
		grpc.WithTransportCredentials(creds),
		grpc.WithPerRPCCredentials(&Authentication{
			User:     SignerConfig.User,
			Password: SignerConfig.Pass,
		}),
	)
	if err != nil {
		logger.Errorf("failed to connect: %v", err)
		return nil, nil, err
	}

	client := pb.NewOfflineSignClient(conn)
	return client, conn, err
}

type Authentication struct {
	User     string
	Password string
}

func (a *Authentication) GetRequestMetadata(context.Context, ...string) (map[string]string, error) {
	return map[string]string{"user": a.User, "password": a.Password}, nil
}

func (a *Authentication) RequireTransportSecurity() bool {
	return false
}
