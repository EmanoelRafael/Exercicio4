package main

import (
	"context"
	"fmt"
	"strings"
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
		Codigo:         codigo,                  // string gerada por GerarCodigoJogo()
		Palavra:        palavra,                 // string da palavra escolhida
		PalavraVisivel: visivel,                 // slice de '_' com o mesmo tamanho da palavra
		Jogadores:      []string{req.JogadorId}, // jogador que criou o jogo
		JogadorDaVez:   req.JogadorId,           // começa com quem criou

		Erros:         map[string]int{req.JogadorId: 0},
		DicasUsadas:   map[string]bool{req.JogadorId: false},
		LetrasErradas: make(map[string]bool), // compartilhado entre todos os jogadores
		Eliminados:    make(map[string]bool), // ninguém eliminado no início

		Status:     2,
		VencedorID: "",
	}

	msg := "Jogo criado com sucesso"
	if req.Solo {
		msg = "Jogo solo criado com sucesso"
	} else if req.ComAmigos {
		jogo.Status = 1
		msg = "Jogo com amigos criado. Compartilhe o código."
	}

	s.jogos[codigo] = jogo

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

	if jogo.Status == 3 {
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
	jogo.Erros[req.JogadorId] = 0
	jogo.DicasUsadas[req.JogadorId] = false

	if len(jogo.Jogadores) == 4 {
		jogo.Status = 2
	}

	return &pb.EntrarJogoResponse{
		Mensagem: "Jogador adicionado ao jogo com sucesso",
		Sucesso:  true,
	}, nil
}

func (s *GameServer) PalpitarLetra(ctx context.Context, req *pb.PalpitarLetraRequest) (*pb.AtualizacaoResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	jogo, existe := s.jogos[req.CodigoJogo]
	if !existe || jogo.Status == 3 {
		return &pb.AtualizacaoResponse{Mensagem: "Jogo não encontrado ou já finalizado"}, nil
	}

	if jogo.JogadorDaVez != req.JogadorId {
		return &pb.AtualizacaoResponse{Mensagem: "Não é sua vez"}, nil
	}

	letra := []rune(req.Letra)
	if len(letra) != 1 {
		return &pb.AtualizacaoResponse{Mensagem: "Letra inválida"}, nil
	}

	acertou := false
	for i, l := range jogo.Palavra {
		if l == letra[0] {
			jogo.PalavraVisivel[i] = l
			acertou = true
		}
	}

	if !acertou {
		jogo.Erros[req.JogadorId]++
		jogo.LetrasErradas[strings.ToUpper(req.Letra)] = true

		if jogo.Erros[req.JogadorId] >= 6 {
			jogo.Eliminados[req.JogadorId] = true

			// Verifica se restou só um jogador
			restantes := jogadoresRestantes(jogo)
			if len(restantes) == 1 {
				jogo.Status = 3
				jogo.VencedorID = restantes[0]
				return &pb.AtualizacaoResponse{
					Mensagem:       fmt.Sprintf("Jogador %s perdeu. Último jogador restante venceu!", req.JogadorId),
					PalavraVisivel: string(jogo.PalavraVisivel),
					JogoStatus:     3,
					VencedorId:     restantes[0],
				}, nil
			} else if len(restantes) == 0 {
				jogo.Status = 3
				jogo.VencedorID = "nil"
				fmt.Println("Palpitar letra - Jogador perdeu")
				return &pb.AtualizacaoResponse{
					Mensagem:       fmt.Sprintf("Jogador %s perdeu. Jogo encerrado!", req.JogadorId),
					PalavraVisivel: string(jogo.PalavraVisivel),
					JogoStatus:     3,
					VencedorId:     "nil",
				}, nil
			}
		}
	} else if palavraCompleta(jogo.PalavraVisivel) {
		jogo.Status = 3
		jogo.VencedorID = req.JogadorId
		return &pb.AtualizacaoResponse{
			Mensagem:       "Parabéns! Você completou a palavra e venceu!",
			PalavraVisivel: string(jogo.PalavraVisivel),
			JogoStatus:     3,
			VencedorId:     req.JogadorId,
		}, nil
	}

	fmt.Println("O jogador da vez antes era: ", jogo.JogadorDaVez)
	trocarTurno(jogo)
	fmt.Println("O jogador da vez agora eh: ", jogo.JogadorDaVez)
	return &pb.AtualizacaoResponse{
		Mensagem:       "Letra processada",
		PalavraVisivel: string(jogo.PalavraVisivel),
		ErrosJogador:   int32(jogo.Erros[req.JogadorId]),
		LetrasErradas:  letrasErradasSlice(jogo.LetrasErradas),
		JogadorDaVez:   jogo.JogadorDaVez,
	}, nil
}

func (s *GameServer) PalpitarPalavra(ctx context.Context, req *pb.PalpitarPalavraRequest) (*pb.AtualizacaoResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	jogo, existe := s.jogos[req.CodigoJogo]
	if !existe || jogo.Status == 3 {
		return &pb.AtualizacaoResponse{Mensagem: "Jogo não encontrado ou finalizado"}, nil
	}

	if jogo.JogadorDaVez != req.JogadorId {
		return &pb.AtualizacaoResponse{Mensagem: "Não é sua vez"}, nil
	}

	if strings.EqualFold(req.Palavra, jogo.Palavra) {
		jogo.PalavraVisivel = []rune(jogo.Palavra)
		jogo.Status = 3
		jogo.VencedorID = req.JogadorId
		return &pb.AtualizacaoResponse{
			Mensagem:       "Você acertou a palavra! Vitória!",
			PalavraVisivel: string(jogo.PalavraVisivel),
			JogoStatus:     3,
			VencedorId:     req.JogadorId,
		}, nil
	}

	// Errou a palavra
	jogo.Erros[req.JogadorId]++
	if jogo.Erros[req.JogadorId] >= 6 {
		jogo.Eliminados[req.JogadorId] = true
		restantes := jogadoresRestantes(jogo)
		if len(restantes) == 1 {
			jogo.Status = 3
			jogo.VencedorID = restantes[0]
			return &pb.AtualizacaoResponse{
				Mensagem:       fmt.Sprintf("Jogador %s perdeu. Último jogador restante venceu!", req.JogadorId),
				PalavraVisivel: string(jogo.PalavraVisivel),
				JogoStatus:     3,
				VencedorId:     restantes[0],
			}, nil
		}
	}

	trocarTurno(jogo)
	return &pb.AtualizacaoResponse{
		Mensagem:       "Palpite errado. Próximo jogador.",
		PalavraVisivel: string(jogo.PalavraVisivel),
		ErrosJogador:   int32(jogo.Erros[req.JogadorId]),
		LetrasErradas:  letrasErradasSlice(jogo.LetrasErradas),
		JogadorDaVez:   jogo.JogadorDaVez,
	}, nil
}

func (s *GameServer) PedirDica(ctx context.Context, req *pb.DicaRequest) (*pb.AtualizacaoResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	jogo, existe := s.jogos[req.CodigoJogo]
	if !existe || jogo.Status == 3 {
		return &pb.AtualizacaoResponse{Mensagem: "Jogo inválido ou finalizado"}, nil
	}

	if jogo.DicasUsadas[req.JogadorId] {
		return &pb.AtualizacaoResponse{Mensagem: "Você já usou sua dica"}, nil
	}

	// Revela uma letra oculta
	for i, r := range jogo.Palavra {
		if jogo.PalavraVisivel[i] == '_' {
			jogo.PalavraVisivel[i] = r
			jogo.DicasUsadas[req.JogadorId] = true
			break
		}
	}

	return &pb.AtualizacaoResponse{
		Mensagem:       "Dica revelada. Você joga novamente.",
		PalavraVisivel: string(jogo.PalavraVisivel),
		ErrosJogador:   int32(jogo.Erros[req.JogadorId]),
		LetrasErradas:  letrasErradasSlice(jogo.LetrasErradas),
		JogadorDaVez:   req.JogadorId,
	}, nil
}

func (s *GameServer) ObterEstado(ctx context.Context, req *pb.EstadoRequest) (*pb.AtualizacaoResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	jogo, ok := s.jogos[req.CodigoJogo]
	if !ok {
		return &pb.AtualizacaoResponse{
			Mensagem: "Jogo não encontrado",
		}, nil
	}

	// Converte o slice de runes para string
	palavraVisivel := string(jogo.PalavraVisivel)

	resp := &pb.AtualizacaoResponse{
		PalavraVisivel: palavraVisivel,
		Mensagem:       "Estado atual do jogo",
		JogadorDaVez:   jogo.JogadorDaVez,
		JogoStatus:     int32(jogo.Status),
		VencedorId:     jogo.VencedorID,
		LetrasErradas:  letrasErradasSlice(jogo.LetrasErradas),
		ErrosJogador:   int32(jogo.Erros[req.JogadorId]),
	}

	return resp, nil
}
