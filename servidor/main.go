package main

import (
	"context"
	"fmt"
	"log"
	"net"

	pb "meu-grpc-projeto/jogodaforca"

	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedGreeterServer
}

func (s *server) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloReply, error) {
	return &pb.HelloReply{Message: "Olá, " + req.GetName()}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Erro ao escutar: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterGreeterServer(s, &server{})
	fmt.Println("Servidor gRPC rodando na porta 50051...")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Erro ao rodar servidor: %v", err)
	}
}
