package main

import (
	"context"
	"log"
	"time"

	pb "ForcaGame/proto"

	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Não foi possível conectar: %v", err)
	}
	defer conn.Close()

	client := pb.NewGreeterClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	resp, err := client.SayHello(ctx, &pb.HelloRequest{Name: "Fulano"})
	if err != nil {
		log.Fatalf("Erro ao chamar SayHello: %v", err)
	}
	log.Printf("Resposta do servidor: %s", resp.GetMessage())
}
