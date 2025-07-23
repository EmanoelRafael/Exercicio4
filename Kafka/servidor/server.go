package main

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/IBM/sarama"
)

type GameServer struct {
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

func (s *GameServer) CriarJogo(JogadorId string, Solo bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	palavra := obterPalavra()
	visivel := make([]rune, len(palavra))
	for i := range visivel {
		visivel[i] = '_'
	}

	codigo := GerarCodigoJogo()

	jogo := &Jogo{
		Codigo:         codigo,
		Palavra:        palavra,
		PalavraVisivel: visivel,
		Jogadores:      []string{JogadorId},
		JogadorDaVez:   JogadorId,

		Erros:         map[string]int{JogadorId: 0},
		DicasUsadas:   map[string]bool{JogadorId: false},
		LetrasErradas: make(map[string]bool),
		Eliminados:    make(map[string]bool),

		Status:     EM_CURSO,
		VencedorID: "",

		UltimaJogada: time.Now(),
	}

	msg := "Jogo criado com sucesso"
	if Solo {
		msg = "Jogo solo criado com sucesso"
	} else {
		jogo.Status = PENDENTE_JOGADORES
		msg = "Jogo com amigos criado. Compartilhe o código."
	}

	topicDetail := &sarama.TopicDetail{
		NumPartitions:     1,
		ReplicationFactor: 1,
	}

	topicGame := "jogada-" + codigo

	err = Admin.CreateTopic(topicGame, topicDetail, false)
	if err != nil {
		if err, ok := err.(*sarama.TopicError); ok && err.Err == sarama.ErrTopicAlreadyExists {
			fmt.Printf("Tópico '%s' já existe.\n", topicGame)
		} else {
			log.Fatalf("Erro ao criar o tópico '%s': %v", topicGame, err)
		}
	} else {
		fmt.Printf("Tópico '%s' criado com sucesso.\n", topicGame)
	}

	topicGameStatus := "status-jogo-" + codigo

	err = Admin.CreateTopic(topicGameStatus, topicDetail, false)
	if err != nil {
		if err, ok := err.(*sarama.TopicError); ok && err.Err == sarama.ErrTopicAlreadyExists {
			fmt.Printf("Tópico '%s' já existe.\n", topicGameStatus)
		} else {
			log.Fatalf("Erro ao criar o tópico '%s': %v", topicGameStatus, err)
		}
	} else {
		fmt.Printf("Tópico '%s' criado com sucesso.\n", topicGameStatus)
	}

	topicPlayer := "resposta-" + JogadorId

	err = Admin.CreateTopic(topicPlayer, topicDetail, false)
	if err != nil {
		if err, ok := err.(*sarama.TopicError); ok && err.Err == sarama.ErrTopicAlreadyExists {
			fmt.Printf("Tópico '%s' já existe.\n", topicPlayer)
		} else {
			log.Fatalf("Erro ao criar o tópico '%s': %v", topicPlayer, err)
		}
	} else {
		fmt.Printf("Tópico '%s' criado com sucesso.\n", topicPlayer)
	}

	s.jogos[codigo] = jogo
	go s.monitorarTimeout(codigo)

	msgRet := ResponseKafka{
		TipoMensagem:   1,
		PalavraVisivel: "",
		LetrasErradas:  "",
		ErrosJogador:   int(jogo.Erros[JogadorId]),
		JogoStatus:     0,
		Mensagem:       msg,
		JogadorId:      JogadorId,
		JogoId:         codigo,
	}
	postarMsg(msgRet)
}

func (s *GameServer) EntrarJogo(CodigoJogo string, JogadorId string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	jogo := s.jogos[CodigoJogo]
	msgRet := ResponseKafka{
		TipoMensagem:   1,
		PalavraVisivel: "",
		LetrasErradas:  "",
		ErrosJogador:   int(jogo.Erros[JogadorId]),
		JogoStatus:     0,
		Mensagem:       "",
		JogadorId:      JogadorId,
		JogoId:         CodigoJogo,
	}

	jogo, existe := s.jogos[CodigoJogo]
	if !existe {
		msgRet.Mensagem = "Código de jogo inválido"
		postarMsg(msgRet)
	}

	if jogo.Status == FINALIZADO {
		msgRet.Mensagem = "O jogo já foi finalizado"
		postarMsg(msgRet)
	}

	for _, j := range jogo.Jogadores {
		if j == JogadorId {
			msgRet.Mensagem = "Você já está participando deste jogo"
			postarMsg(msgRet)
		}
	}

	if len(jogo.Jogadores) >= 4 {
		msgRet.Mensagem = "O jogo já possui quatro jogadores"
		postarMsg(msgRet)
	}

	jogo.Jogadores = append(jogo.Jogadores, JogadorId)
	jogo.Erros[JogadorId] = 0
	jogo.DicasUsadas[JogadorId] = false

	if len(jogo.Jogadores) == 4 {
		jogo.Status = EM_CURSO
	}

	topicPlayer := "resposta-" + JogadorId

	topicDetail := &sarama.TopicDetail{
		NumPartitions:     1,
		ReplicationFactor: 1,
	}

	err = Admin.CreateTopic(topicPlayer, topicDetail, false)
	if err != nil {
		if err, ok := err.(*sarama.TopicError); ok && err.Err == sarama.ErrTopicAlreadyExists {
			fmt.Printf("Tópico '%s' já existe.\n", topicPlayer)
		} else {
			log.Fatalf("Erro ao criar o tópico '%s': %v", topicPlayer, err)
		}
	} else {
		fmt.Printf("Tópico '%s' criado com sucesso.\n", topicPlayer)
	}

	msgRet.Mensagem = "Jogador adicionado com sucesso"
	msgRet.JogoStatus = 1
	postarMsg(msgRet)
}

func (s *GameServer) PalpitarLetra(CodigoJogo string, JogadorId string, Letra string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	jogo := s.jogos[CodigoJogo]

	msgRet := ResponseKafka{
		TipoMensagem:   1,
		PalavraVisivel: "",
		LetrasErradas:  "",
		ErrosJogador:   int(jogo.Erros[JogadorId]),
		JogoStatus:     0,
		Mensagem:       "",
		JogadorId:      JogadorId,
		JogoId:         CodigoJogo,
	}

	jogo, existe := s.jogos[CodigoJogo]
	if !existe || jogo.Status == FINALIZADO {
		jogo.UltimaJogada = time.Now()
		msgRet.Mensagem = "Jogo não encontrado ou já finalizado"
		postarMsg(msgRet)
	}

	if jogo.JogadorDaVez != JogadorId {
		jogo.UltimaJogada = time.Now()
		msgRet.Mensagem = "Não é sua vez"
		postarMsg(msgRet)
	}

	letra := []rune(Letra)
	if len(letra) != 1 {
		jogo.UltimaJogada = time.Now()
		msgRet.Mensagem = "Letra inválida"
		postarMsg(msgRet)
	}

	acertou := false
	for i, l := range jogo.Palavra {
		if l == letra[0] {
			jogo.PalavraVisivel[i] = l
			acertou = true
		}
	}

	if !acertou {
		jogo.Erros[JogadorId]++
		jogo.LetrasErradas[strings.ToUpper(Letra)] = true

		if jogo.Erros[JogadorId] >= 6 {
			jogo.Eliminados[JogadorId] = true

			restantes := jogadoresRestantes(jogo)
			if len(restantes) == 1 {
				jogo.Status = FINALIZADO
				jogo.VencedorID = restantes[0]
				jogo.UltimaJogada = time.Now()
				msgRet.Mensagem = "Jogador " + JogadorId + " perdeu. Último jogador restante venceu!"
				postarMsg(msgRet)
			} else if len(restantes) == 0 {
				jogo.Status = FINALIZADO
				jogo.VencedorID = "nil"
				jogo.UltimaJogada = time.Now()
				msgRet.Mensagem = "Jogador " + JogadorId + " perdeu. Jogo encerrado!"
				postarMsg(msgRet)
			}
		}
	} else if palavraCompleta(jogo.PalavraVisivel) {
		jogo.Status = FINALIZADO
		jogo.VencedorID = JogadorId
		jogo.UltimaJogada = time.Now()
		msgRet.Mensagem = "Parabéns! Você completou a palavra e venceu!"
		postarMsg(msgRet)
	}

	trocarTurno(jogo)
	jogo.UltimaJogada = time.Now()
	msgRet.Mensagem = "Letra processada"
	postarMsg(msgRet)
}

func (s *GameServer) PalpitarPalavra(CodigoJogo string, JogadorId string, Palavra string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	jogo := s.jogos[CodigoJogo]

	msgRet := ResponseKafka{
		TipoMensagem:   1,
		PalavraVisivel: "",
		LetrasErradas:  "",
		ErrosJogador:   int(jogo.Erros[JogadorId]),
		JogoStatus:     0,
		Mensagem:       "",
		JogadorId:      JogadorId,
		JogoId:         CodigoJogo,
	}

	jogo, existe := s.jogos[CodigoJogo]
	if !existe || jogo.Status == FINALIZADO {
		jogo.UltimaJogada = time.Now()
		msgRet.Mensagem = "Jogo não encontrado ou finalizado"
		postarMsg(msgRet)
	}

	if jogo.JogadorDaVez != JogadorId {
		jogo.UltimaJogada = time.Now()
		msgRet.Mensagem = "Não é sua vez"
		postarMsg(msgRet)
	}

	if strings.EqualFold(Palavra, jogo.Palavra) {
		jogo.PalavraVisivel = []rune(jogo.Palavra)
		jogo.Status = FINALIZADO
		jogo.VencedorID = JogadorId
		jogo.UltimaJogada = time.Now()
		msgRet.Mensagem = "Você acertou a palavra! Vitória!"
		postarMsg(msgRet)
	}

	jogo.Erros[JogadorId]++
	if jogo.Erros[JogadorId] >= 6 {
		jogo.Eliminados[JogadorId] = true
		restantes := jogadoresRestantes(jogo)
		if len(restantes) == 1 {
			jogo.Status = FINALIZADO
			jogo.VencedorID = restantes[0]
			jogo.UltimaJogada = time.Now()
			msgRet.Mensagem = "Jogador " + JogadorId + " perdeu. Último jogador restante venceu!"
			postarMsg(msgRet)
		}
	}

	trocarTurno(jogo)
	jogo.UltimaJogada = time.Now()
	msgRet.Mensagem = "Palpite errado. Próximo jogador."
	postarMsg(msgRet)
}

func (s *GameServer) PedirDica(CodigoJogo string, JogadorId string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	jogo := s.jogos[CodigoJogo]

	msgRet := ResponseKafka{
		TipoMensagem:   1,
		PalavraVisivel: "",
		LetrasErradas:  "",
		ErrosJogador:   int(jogo.Erros[JogadorId]),
		JogoStatus:     0,
		Mensagem:       "",
		JogadorId:      JogadorId,
		JogoId:         CodigoJogo,
	}

	jogo, existe := s.jogos[CodigoJogo]
	if !existe || jogo.Status == 3 {
		jogo.UltimaJogada = time.Now()
		msgRet.Mensagem = "Jogo inválido ou finalizado"
		postarMsg(msgRet)
	}

	if jogo.DicasUsadas[JogadorId] {
		jogo.UltimaJogada = time.Now()
		msgRet.Mensagem = "Você já usou sua dica"
		postarMsg(msgRet)
	}

	for i, r := range jogo.Palavra {
		if jogo.PalavraVisivel[i] == '_' {
			jogo.PalavraVisivel[i] = r
			jogo.DicasUsadas[JogadorId] = true
			break
		}
	}

	jogo.UltimaJogada = time.Now()
	msgRet.Mensagem = "Dica revelada. Você joga novamente."
	postarMsg(msgRet)
}

func (s *GameServer) ObterEstado(CodigoJogo string, JogadorId string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	jogo := s.jogos[CodigoJogo]

	msgRet := ResponseKafka{
		TipoMensagem:   2,
		PalavraVisivel: "",
		LetrasErradas:  "",
		ErrosJogador:   int(jogo.Erros[JogadorId]),
		JogoStatus:     0,
		Mensagem:       "",
		JogadorId:      JogadorId,
		JogoId:         CodigoJogo,
	}

	jogo, ok := s.jogos[CodigoJogo]
	if !ok {
		msgRet.Mensagem = "Jogo não encontrado"
		postarMsg(msgRet)
	}

	palavraVisivel := string(jogo.PalavraVisivel)

	msgRet.PalavraVisivel = palavraVisivel
	msgRet.Mensagem = "Estado Atual do Jogo"
	msgRet.JogadorId = jogo.JogadorDaVez
	msgRet.JogoStatus = int(jogo.Status)
	msgRet.LetrasErradas = letrasErradasSlice(jogo.LetrasErradas)
}
