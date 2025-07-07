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
	s.mu.Lock()
	defer s.mu.Unlock()

	jogo, existe := s.jogos[req.CodigoJogo]
	if !existe {
		return &pb.EntrarJogoResponse{
			Mensagem: "Código de jogo inválido",
			Sucesso:  false,
		}, nil
	}

	if jogo.Finalizado {
		return &pb.EntrarJogoResponse{
			Mensagem: "O jogo já foi finalizado",
			Sucesso:  false,
		}, nil
	}

	// Verifica se jogador já está no jogo
	for _, j := range jogo.Jogadores {
		if j == req.JogadorId {
			return &pb.EntrarJogoResponse{
				Mensagem: "Você já está participando deste jogo",
				Sucesso:  true,
			}, nil
		}
	}

	// Limite de jogadores: 4
	if len(jogo.Jogadores) >= 4 {
		return &pb.EntrarJogoResponse{
			Mensagem: "O jogo já possui quatro jogadores",
			Sucesso:  false,
		}, nil
	}

	// Adiciona jogador ao jogo
	jogo.Jogadores = append(jogo.Jogadores, req.JogadorId)

	return &pb.EntrarJogoResponse{
		Mensagem: "Jogador adicionado ao jogo com sucesso",
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
