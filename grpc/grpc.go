package grpc

import (
	"time"

	"github.com/Alberto-Izquierdo/RPIHomeServer-go/configuration_loader"
	"google.golang.org/grpc"
)

var connection *grpc.ClientConn

func ConnectToGrpcServer(config configuration_loader.InitialConfiguration) (err error) {
	connection, err = grpc.Dial(config.GRPCServerIp, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(time.Second))
	return err
}

func CloseConnection() {
	if connection != nil {
		connection.Close()
	}
}
