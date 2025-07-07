package main

import (
	"fmt"
	"log"
	"net"
	"time"

	pb "ForcaGame/proto"

	"google.golang.org/grpc"
)

func main() {
	// Inicializa aleatoriedade para geração de código
	// (Importante se usar rand!)
	time.Sleep(1 * time.Millisecond)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Erro ao escutar: %v", err)
	}

	grpcServer := grpc.NewServer()
	gameSrv := NewGameServer()
	pb.RegisterGameServiceServer(grpcServer, gameSrv)

	fmt.Println("Servidor gRPC rodando na porta 50051...")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Erro ao rodar servidor: %v", err)
	}
}
