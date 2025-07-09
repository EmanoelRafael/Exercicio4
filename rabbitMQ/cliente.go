package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/streadway/amqp"
)

type GameState struct {
	Palavra      string
	Progresso    string
	Tentativas   int
	MaxErros     int
	LetrasUsadas []string
	Turno        int
	Encerrado    bool
	Vencedor     string
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Digite seu ID (cliente1 ou cliente2): ")
	clienteID, _ := reader.ReadString('\n')
	clienteID = strings.TrimSpace(clienteID)
	fmt.Printf("Cliente iniciado com ID: %s\n", clienteID)

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Conexão RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Canal")
	defer ch.Close()

	ch.QueueDeclare(clienteID, false, false, false, false, nil)
	msgs, err := ch.Consume(clienteID, "", true, false, false, false, nil)
	failOnError(err, "Consumo")
	fmt.Println("Aguardando mensagens do servidor...")

	for msg := range msgs {
		var state GameState
		json.Unmarshal(msg.Body, &state)

		fmt.Printf("\n=== FORCA ===\nPalavra: %s\nTentativas: %d/%d\nLetras usadas: %v\n",
			state.Progresso, state.Tentativas, state.MaxErros, state.LetrasUsadas)

		if state.Encerrado {
			if state.Vencedor == clienteID {
				fmt.Println("Parabéns, você venceu!")
			} else if state.Vencedor == "forca" {
				fmt.Println("Vocês perderam! A palavra era:", state.Palavra)
			} else {
				fmt.Println("Você perdeu!")
			}
			break
		}

		if clienteID == fmt.Sprintf("cliente%d", state.Turno+1) {
			fmt.Print("Sua vez! Digite uma letra: ")
			letra, _ := reader.ReadString('\n')
			letra = strings.TrimSpace(letra)

			ch.Publish("", "palpites", false, false, amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(letra),
				Headers: amqp.Table{
					"player": clienteID,
				},
			})
		}
	}
}
