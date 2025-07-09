package main

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	pb "ForcaGame/proto"
)

type GameServer struct {
	pb.UnimplementedGameServiceServer
	jogos map[string]*Jogo
	mu    sync.Mutex
}

func NewGameServer() *GameServer {
	return &GameServer{
		jogos: make(map[string]*Jogo),
	}
}

func (s *GameServer) monitorarTimeout(codigo string) {
	for {
		time.Sleep(5 * time.Second)

		s.mu.Lock()
		jogo, ok := s.jogos[codigo]
		if !ok || jogo.Status != EM_CURSO {
			s.mu.Unlock()
			return
		}

		if time.Since(jogo.UltimaJogada) > 60*time.Second {
			jogador := jogo.JogadorDaVez
			fmt.Printf("Jogador %s foi eliminado por inatividade\n", jogador)
			jogo.Eliminados[jogador] = true
			jogo.Erros[jogador] = 6

			restantes := jogadoresRestantes(jogo)
			if len(restantes) == 1 {
				jogo.Status = FINALIZADO
				jogo.VencedorID = restantes[0]
				s.mu.Unlock()
				return
			} else if len(restantes) == 0 {
				jogo.Status = FINALIZADO
				jogo.VencedorID = "nil"
				s.mu.Unlock()
				return
			}

			trocarTurno(jogo)
			jogo.UltimaJogada = time.Now()
		}
		s.mu.Unlock()
	}
}

func (s *GameServer) CriarJogo(ctx context.Context, req *pb.CriarJogoRequest) (*pb.CriarJogoResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Palavra padrão por enquanto
	palavra := obterPalavra()
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

		Status:     EM_CURSO,
		VencedorID: "",

		UltimaJogada: time.Now(),
	}

	msg := "Jogo criado com sucesso"
	if req.Solo {
		msg = "Jogo solo criado com sucesso"
	} else if req.ComAmigos {
		jogo.Status = PENDENTE_JOGADORES
		msg = "Jogo com amigos criado. Compartilhe o código."
	}

	s.jogos[codigo] = jogo
	go s.monitorarTimeout(codigo)

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

	if jogo.Status == FINALIZADO {
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
		jogo.Status = EM_CURSO
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
	if !existe || jogo.Status == FINALIZADO {
		jogo.UltimaJogada = time.Now()
		return &pb.AtualizacaoResponse{Mensagem: "Jogo não encontrado ou já finalizado"}, nil
	}

	if jogo.JogadorDaVez != req.JogadorId {
		jogo.UltimaJogada = time.Now()
		return &pb.AtualizacaoResponse{Mensagem: "Não é sua vez"}, nil
	}

	letra := []rune(req.Letra)
	if len(letra) != 1 {
		jogo.UltimaJogada = time.Now()
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
				jogo.Status = FINALIZADO
				jogo.VencedorID = restantes[0]
				jogo.UltimaJogada = time.Now()
				return &pb.AtualizacaoResponse{
					Mensagem:       fmt.Sprintf("Jogador %s perdeu. Último jogador restante venceu!", req.JogadorId),
					PalavraVisivel: string(jogo.PalavraVisivel),
					JogoStatus:     FINALIZADO,
					VencedorId:     restantes[0],
				}, nil
			} else if len(restantes) == 0 {
				jogo.Status = FINALIZADO
				jogo.VencedorID = "nil"
				fmt.Println("Palpitar letra - Jogador perdeu")
				jogo.UltimaJogada = time.Now()
				return &pb.AtualizacaoResponse{
					Mensagem:       fmt.Sprintf("Jogador %s perdeu. Jogo encerrado!", req.JogadorId),
					PalavraVisivel: string(jogo.PalavraVisivel),
					JogoStatus:     FINALIZADO,
					VencedorId:     "nil",
				}, nil
			}
		}
	} else if palavraCompleta(jogo.PalavraVisivel) {
		jogo.Status = FINALIZADO
		jogo.VencedorID = req.JogadorId
		jogo.UltimaJogada = time.Now()
		return &pb.AtualizacaoResponse{
			Mensagem:       "Parabéns! Você completou a palavra e venceu!",
			PalavraVisivel: string(jogo.PalavraVisivel),
			JogoStatus:     FINALIZADO,
			VencedorId:     req.JogadorId,
		}, nil
	}

	fmt.Println("O jogador da vez antes era: ", jogo.JogadorDaVez)
	trocarTurno(jogo)
	fmt.Println("O jogador da vez agora eh: ", jogo.JogadorDaVez)
	jogo.UltimaJogada = time.Now()
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
	if !existe || jogo.Status == FINALIZADO {
		jogo.UltimaJogada = time.Now()
		return &pb.AtualizacaoResponse{Mensagem: "Jogo não encontrado ou finalizado"}, nil
	}

	if jogo.JogadorDaVez != req.JogadorId {
		jogo.UltimaJogada = time.Now()
		return &pb.AtualizacaoResponse{Mensagem: "Não é sua vez"}, nil
	}

	if strings.EqualFold(req.Palavra, jogo.Palavra) {
		jogo.PalavraVisivel = []rune(jogo.Palavra)
		jogo.Status = FINALIZADO
		jogo.VencedorID = req.JogadorId
		jogo.UltimaJogada = time.Now()
		return &pb.AtualizacaoResponse{
			Mensagem:       "Você acertou a palavra! Vitória!",
			PalavraVisivel: string(jogo.PalavraVisivel),
			JogoStatus:     FINALIZADO,
			VencedorId:     req.JogadorId,
		}, nil
	}

	jogo.Erros[req.JogadorId]++
	if jogo.Erros[req.JogadorId] >= 6 {
		jogo.Eliminados[req.JogadorId] = true
		restantes := jogadoresRestantes(jogo)
		if len(restantes) == 1 {
			jogo.Status = FINALIZADO
			jogo.VencedorID = restantes[0]
			jogo.UltimaJogada = time.Now()
			return &pb.AtualizacaoResponse{
				Mensagem:       fmt.Sprintf("Jogador %s perdeu. Último jogador restante venceu!", req.JogadorId),
				PalavraVisivel: string(jogo.PalavraVisivel),
				JogoStatus:     FINALIZADO,
				VencedorId:     restantes[0],
			}, nil
		}
	}

	trocarTurno(jogo)
	jogo.UltimaJogada = time.Now()
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
		jogo.UltimaJogada = time.Now()
		return &pb.AtualizacaoResponse{Mensagem: "Jogo inválido ou finalizado"}, nil
	}

	if jogo.DicasUsadas[req.JogadorId] {
		jogo.UltimaJogada = time.Now()
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

	jogo.UltimaJogada = time.Now()
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
