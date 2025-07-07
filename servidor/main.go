package main

import (
	"context"
	"fmt"
	"log"
	"net"

	pb "ForcaGame/proto"

	"google.golang.org/grpc"
)

type gameServer struct {
	pb.UnimplementedGameServiceServer
	// aqui você pode adicionar mapas para armazenar jogos, jogadores, etc.
}

func (s *gameServer) CriarJogo(ctx context.Context, req *pb.CriarJogoRequest) (*pb.CriarJogoResponse, error) {
	// TODO: lógica para criar jogo
	return &pb.CriarJogoResponse{
		CodigoJogo: "abc123",
		Mensagem:   "Jogo criado com sucesso",
	}, nil
}

func (s *gameServer) EntrarJogo(ctx context.Context, req *pb.EntrarJogoRequest) (*pb.EntrarJogoResponse, error) {
	// TODO: lógica para entrar em jogo
	return &pb.EntrarJogoResponse{
		Mensagem: "Entrou no jogo com sucesso",
		Sucesso:  true,
	}, nil
}

func (s *gameServer) PalpitarLetra(ctx context.Context, req *pb.PalpitarLetraRequest) (*pb.AtualizacaoResponse, error) {
	// TODO: lógica para palpite de letra
	return &pb.AtualizacaoResponse{
		Mensagem: "Letra processada",
	}, nil
}

func (s *gameServer) PalpitarPalavra(ctx context.Context, req *pb.PalpitarPalavraRequest) (*pb.AtualizacaoResponse, error) {
	// TODO: lógica para palpite da palavra
	return &pb.AtualizacaoResponse{
		Mensagem: "Palavra processada",
	}, nil
}

func (s *gameServer) PedirDica(ctx context.Context, req *pb.DicaRequest) (*pb.AtualizacaoResponse, error) {
	// TODO: lógica para dica
	return &pb.AtualizacaoResponse{
		Mensagem: "Dica enviada",
	}, nil
}

func (s *gameServer) ObterEstado(ctx context.Context, req *pb.EstadoRequest) (*pb.AtualizacaoResponse, error) {
	// TODO: lógica para retornar estado atual
	return &pb.AtualizacaoResponse{
		Mensagem: "Estado atual do jogo",
	}, nil
}

func main() {
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Erro ao escutar: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterGameServiceServer(grpcServer, &gameServer{})
	fmt.Println("Servidor gRPC rodando na porta 50051...")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Erro ao rodar servidor: %v", err)
	}
}
