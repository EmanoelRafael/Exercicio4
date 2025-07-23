package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	//pb "ForcaGame/proto"

	"github.com/IBM/sarama"
)

var jogadorId string
var codigoJogo string
var gameStage = MENU
var jogo *Jogo
var flag = false
var jogoStatusAntigo = NAO_INICIADO

const (
	brokerAddress = "localhost:29092"
	topicEntrada  = "jogo-entrada"
)

var MsgChan chan string
var Producer sarama.SyncProducer

type RequestKafka struct {
	TipoMensagem    int
	IdJogador       string
	TipoSolicitacao int
	ConteudoReq     string
	jogoCod         string
}

type ResponseKafka struct {
	TipoMensagem   int
	PalavraVisivel string
	LetrasErradas  string
	ErrosJogador   int
	JogoStatus     int
	Mensagem       string
	JogadorId      string
	JogoId         string
}

func escutarTopico(consumer sarama.Consumer, topico string, msgChan chan<- string) {
	pc, err := consumer.ConsumePartition(topico, 0, sarama.OffsetNewest)
	if err != nil {
		log.Printf("Erro ao consumir tópico %s: %v", topico, err)
		return
	}
	defer pc.Close()

	for msg := range pc.Messages() {
		msgChan <- string(msg.Value)
		fmt.Printf("[JOGO %s] Mensagem: %s\n", topico, string(msg.Value))
	}

}

func postarMsg(msg RequestKafka) {
	jsonBytes, _ := json.Marshal(msg)
	var mensagem *sarama.ProducerMessage
	switch msg.TipoMensagem {
	case 0:
		mensagem = &sarama.ProducerMessage{
			Topic: topicEntrada,
			Value: sarama.ByteEncoder(jsonBytes),
		}
	case 3:
		mensagem = &sarama.ProducerMessage{
			Topic: "jogada-" + jogo.Codigo,
			Value: sarama.ByteEncoder(jsonBytes),
		}
	}
	partition, offset, err := Producer.SendMessage(mensagem)
	if err != nil {
		log.Fatalf("Erro ao enviar mensagem: %v", err)
	}

	fmt.Printf("Mensagem enviada para a partição %d com offset %d\n", partition, offset)

}

func stringParaArray(s string) []string {
	var resultado []string
	for _, letra := range s {
		resultado = append(resultado, string(letra))
	}
	return resultado
}

func processarMsg(msg string) string {

	var msgKafka ResponseKafka
	err := json.Unmarshal([]byte(msg), &msgKafka)
	if err != nil {
		fmt.Println("Erro ao converter JSON:", err)
		return ""
	}

	switch msgKafka.TipoMensagem {
	case 1:
		fmt.Println(msgKafka.Mensagem)
		jogo.Erros = msgKafka.ErrosJogador
	case 2:
		jogo.PalavraVisivel = []rune(msgKafka.PalavraVisivel)
		jogo.LetrasErradas = stringParaArray(msgKafka.LetrasErradas)
		jogo.JogadorDaVez = msgKafka.JogadorId
		jogo.Status = msgKafka.JogoStatus
		jogo.VencedorID = msgKafka.JogadorId
		if jogo.Status == FINALIZADO {
			gameStage = FIM_DE_JOGO
		}
	}

	fmt.Printf("Processando mensagem: %s\n", msg)
	return ""
}

func main() {

	reader := bufio.NewReader(os.Stdin)

	jogo = &Jogo{
		Codigo:         "",
		PalavraVisivel: nil,
		Erros:          0,
		DicaUsada:      false,
		LetrasErradas:  []string{},
		JogadorDaVez:   "",
		VencedorID:     "",
		Status:         NAO_INICIADO,
	}

	MsgChan = make(chan string)

	fmt.Print("Digite seu ID de jogador: ")
	jogadorId, _ = reader.ReadString('\n')
	jogadorId = strings.TrimSpace(jogadorId)

	config := sarama.NewConfig()
	config.Version = sarama.V2_1_0_0

	config.Consumer.Return.Errors = true
	consumer, err := sarama.NewConsumer([]string{brokerAddress}, config)
	if err != nil {
		log.Fatalf("Erro ao criar consumidor: %v", err)
	}
	defer consumer.Close()

	config.Producer.Return.Successes = true

	Producer, err = sarama.NewSyncProducer([]string{brokerAddress}, config)
	if err != nil {
		log.Fatalf("Erro ao criar producer: %v", err)
	}
	defer Producer.Close()

	flagL := false
	for {
		select {
		case msg := <-MsgChan:
			processarMsg(msg)

		case <-time.After(1 * time.Second):
			fmt.Println("Aguardando novas mensagens...")
		}
		opcao := ""

		if !flag {
			flag = false
			opcao = printGame()
		}

		if gameStage != FIM_DE_JOGO && gameStage != INGRESSO && gameStage != MENU {
			if jogo.Status == PENDENTE_JOGADORES {
				if jogoStatusAntigo == NAO_INICIADO {
					gameStage = AGUARDANDO_JOGADORES
					flag = true
					jogoStatusAntigo = PENDENTE_JOGADORES
				}
				time.Sleep(3 * time.Second)
			} else if jogo.Status == EM_CURSO {
				if jogoStatusAntigo == PENDENTE_JOGADORES {
					gameStage = EM_ANDAMENTO
					jogoStatusAntigo = EM_CURSO
				} else if jogoStatusAntigo == NAO_INICIADO {
					gameStage = EM_ANDAMENTO
				}
				flag = false
				if jogo.JogadorDaVez != jogadorId {
					time.Sleep(3 * time.Second)
				}
			} else if jogo.Status == FINALIZADO {
				gameStage = FIM_DE_JOGO
			}

		} else if gameStage == FIM_DE_JOGO {
			gameStage = MENU
			jogoStatusAntigo = NAO_INICIADO
			continue
		}

		msgRet := RequestKafka{
			TipoMensagem:    0,
			IdJogador:       jogadorId,
			TipoSolicitacao: 0,
			ConteudoReq:     "",
			jogoCod:         jogo.Codigo,
		}

		if gameStage == MENU {
			switch opcao {
			case "1":
				msgRet.TipoSolicitacao = 1
				postarMsg(msgRet)
				time.Sleep(2 * time.Second)
			case "2":
				msgRet.TipoSolicitacao = 2
				postarMsg(msgRet)
			case "3":
				fmt.Print("Digite o codigo do jogo=> ")
				codigoJogo, _ = reader.ReadString('\n')
				codigoJogo = strings.TrimSpace(codigoJogo)
				msgRet.TipoSolicitacao = 3
				msgRet.ConteudoReq = codigoJogo
				postarMsg(msgRet)
			case "0":
				fmt.Println("Saindo...")
				return

			default:
				fmt.Println("Opção inválida.")
			}
		} else if gameStage == EM_ANDAMENTO && jogo.JogadorDaVez == jogadorId {
			switch opcao {
			case "1":
				fmt.Print("Digite uma letra: ")
				letra, _ := reader.ReadString('\n')
				letra = strings.TrimSpace(letra)

				msgRet.TipoMensagem = 3
				msgRet.TipoSolicitacao = 1
				msgRet.ConteudoReq = letra
				postarMsg(msgRet)
			case "2":
				fmt.Print("Digite a palavra: ")
				palpite, _ := reader.ReadString('\n')
				palpite = strings.TrimSpace(palpite)

				msgRet.TipoMensagem = 3
				msgRet.TipoSolicitacao = 2
				msgRet.ConteudoReq = palpite
				postarMsg(msgRet)
			case "3":
				msgRet.TipoMensagem = 3
				msgRet.TipoSolicitacao = 3
				postarMsg(msgRet)
			case "0":
				fmt.Println("Saindo...")
				return

			default:
				fmt.Println("Opção inválida.")
			}
		}

		if !flagL && jogo.Codigo != "" {
			fmt.Println("Iniciando a escuta dos topicos")
			go escutarTopico(consumer, "status-jogo-"+jogo.Codigo, MsgChan)
			go escutarTopico(consumer, "resposta-"+jogadorId, MsgChan)
			flagL = true
		}
	}
}
