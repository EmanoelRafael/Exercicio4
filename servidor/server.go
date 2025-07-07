package main

import (
	"context"
	"sync"

	pb "ForcaGame/proto"
)

type GameServer struct {
	pb.UnimplementedGameServiceServer
	jogos map[string]*Jogo
	mu    sync.Mutex
	// aqui você pode adicionar mapas para armazenar jogos, jogadores, etc.
}

func NewGameServer() *GameServer {
	return &GameServer{
		jogos: make(map[string]*Jogo),
	}
}

func (s *GameServer) CriarJogo(ctx context.Context, req *pb.CriarJogoRequest) (*pb.CriarJogoResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Palavra padrão por enquanto
	palavra := "golang"
	visivel := make([]rune, len(palavra))
	for i := range visivel {
		visivel[i] = '_'
	}

	codigo := GerarCodigoJogo()

	jogo := &Jogo{
		Codigo:         codigo,
		Palavra:        palavra,
		PalavraVisivel: visivel,
		Jogadores:      []string{req.JogadorId},
		Finalizado:     false,
	}

	s.jogos[codigo] = jogo

	msg := "Jogo criado com sucesso"
	if req.Solo {
		msg = "Jogo solo criado com sucesso"
	} else if req.ComAmigos {
		msg = "Jogo com amigos criado. Compartilhe o código."
	}

	return &pb.CriarJogoResponse{
		CodigoJogo: codigo,
		Mensagem:   msg,
	}, nil
}

func (s *GameServer) EntrarJogo(ctx context.Context, req *pb.EntrarJogoRequest) (*pb.EntrarJogoResponse, error) {
	// TODO: lógica para entrar em jogo
	return &pb.EntrarJogoResponse{
		Mensagem: "Entrou no jogo com sucesso",
		Sucesso:  true,
	}, nil
}

func (s *GameServer) PalpitarLetra(ctx context.Context, req *pb.PalpitarLetraRequest) (*pb.AtualizacaoResponse, error) {
	// TODO: lógica para palpite de letra
	return &pb.AtualizacaoResponse{
		Mensagem: "Letra processada",
	}, nil
}

func (s *GameServer) PalpitarPalavra(ctx context.Context, req *pb.PalpitarPalavraRequest) (*pb.AtualizacaoResponse, error) {
	// TODO: lógica para palpite da palavra
	return &pb.AtualizacaoResponse{
		Mensagem: "Palavra processada",
	}, nil
}

func (s *GameServer) PedirDica(ctx context.Context, req *pb.DicaRequest) (*pb.AtualizacaoResponse, error) {
	// TODO: lógica para dica
	return &pb.AtualizacaoResponse{
		Mensagem: "Dica enviada",
	}, nil
}

func (s *GameServer) ObterEstado(ctx context.Context, req *pb.EstadoRequest) (*pb.AtualizacaoResponse, error) {
	// TODO: lógica para retornar estado atual
	return &pb.AtualizacaoResponse{
		Mensagem: "Estado atual do jogo",
	}, nil
}
