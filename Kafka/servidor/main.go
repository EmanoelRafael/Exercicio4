package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"slices"

	"github.com/IBM/sarama"
)

const (
	brokerAddress = "localhost:29092"
	topicEntrada  = "jogo-entrada"
)

var MsgChan chan string
var Producer sarama.SyncProducer
var Admin sarama.ClusterAdmin
var err error
var serverGame *GameServer

type RequestKafka struct {
	TipoMensagem    int `json:"tipoMensagem"`
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

var jogosAux []string

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

func escutarTopicosJogos(consumer sarama.Consumer) {
	for {
		for codigo, _ := range serverGame.jogos {
			if !slices.Contains(jogosAux, codigo) {
				go escutarTopico(consumer, "jogada-"+codigo, MsgChan)
			}
		}
		time.Sleep(2 * time.Second)
	}
}

func postarMsg(msg ResponseKafka) {
	jsonBytes, _ := json.Marshal(msg)
	var mensagem *sarama.ProducerMessage
	switch msg.TipoMensagem {
	case 1:
		mensagem = &sarama.ProducerMessage{
			Topic: "resposta-" + msg.JogadorId,
			Value: sarama.ByteEncoder(jsonBytes),
		}
	case 2:
		mensagem = &sarama.ProducerMessage{
			Topic: "status-jogo-" + msg.JogoId,
			Value: sarama.ByteEncoder(jsonBytes),
		}
	}
	partition, offset, err := Producer.SendMessage(mensagem)
	if err != nil {
		log.Fatalf("Erro ao enviar mensagem: %v", err)
	}

	fmt.Printf("Mensagem enviada para a partição %d com offset %d\n", partition, offset)

}

func processarMsg(msg string) string {

	var msgKafka RequestKafka
	err := json.Unmarshal([]byte(msg), &msgKafka)
	if err != nil {
		fmt.Println("Erro ao converter JSON:", err)
		return ""
	}

	switch msgKafka.TipoMensagem {
	case 0:
		switch msgKafka.TipoSolicitacao {
		case 1:
			serverGame.CriarJogo(msgKafka.IdJogador, true)
		case 2:
			serverGame.CriarJogo(msgKafka.IdJogador, false)
		case 3:
			serverGame.EntrarJogo(msgKafka.ConteudoReq, msgKafka.IdJogador)
		}
	case 3:
		switch msgKafka.TipoSolicitacao {
		case 1:
			serverGame.PalpitarLetra(msgKafka.jogoCod, msgKafka.IdJogador, msgKafka.ConteudoReq)
		case 2:
			serverGame.PalpitarPalavra(msgKafka.jogoCod, msgKafka.IdJogador, msgKafka.ConteudoReq)
		case 3:
			serverGame.PedirDica(msgKafka.jogoCod, msgKafka.IdJogador)
		}
	}

	fmt.Printf("Processando mensagem: %s\n", msg)
	return ""
}

func main() {
	time.Sleep(1 * time.Millisecond)
	serverGame = NewGameServer()
	MsgChan = make(chan string)

	config := sarama.NewConfig()
	config.Version = sarama.V2_1_0_0

	Admin, err = sarama.NewClusterAdmin([]string{brokerAddress}, config)
	if err != nil {
		log.Fatalf("Erro ao criar ClusterAdmin do Kafka: %v", err)
	}
	defer Admin.Close()

	topicDetail := &sarama.TopicDetail{
		NumPartitions:     1,
		ReplicationFactor: 1,
	}

	err = Admin.CreateTopic(topicEntrada, topicDetail, false)
	if err != nil {
		if err, ok := err.(*sarama.TopicError); ok && err.Err == sarama.ErrTopicAlreadyExists {
			fmt.Printf("Tópico '%s' já existe.\n", topicEntrada)
		} else {
			log.Fatalf("Erro ao criar o tópico '%s': %v", topicEntrada, err)
		}
	} else {
		fmt.Printf("Tópico '%s' criado com sucesso.\n", topicEntrada)
	}

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

	go escutarTopico(consumer, topicEntrada, MsgChan)
	go escutarTopicosJogos(consumer)

	for {
		select {
		case msg := <-MsgChan:
			processarMsg(msg)

		case <-time.After(1 * time.Second):
			fmt.Println("Aguardando novas mensagens...")
		}
	}

}
