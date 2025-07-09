package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/streadway/amqp"
)

type GameState struct {
	Palavra      string   `json:"palavra"`
	Progresso    string   `json:"progresso"`
	Tentativas   int      `json:"tentativas"`
	MaxErros     int      `json:"max_erros"`
	LetrasUsadas []string `json:"letras_usadas"`
	Turno        int      `json:"turno"`
	Encerrado    bool     `json:"encerrado"`
	Vencedor     string   `json:"vencedor"`
}

var jogadores = []string{"cliente1", "cliente2"}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Conexão RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Canal")
	defer ch.Close()

	_, err = ch.QueueDeclare("palpites", false, false, false, false, nil)
	failOnError(err, "Fila de palpites")

	msgs, err := ch.Consume("palpites", "", true, false, false, false, nil)
	failOnError(err, "Consumo de palpites")

	// Estado inicial do jogo
	palavra := "golang"
	progresso := strings.Repeat("_", len(palavra))
	tentativas := 0
	letras := []string{}
	turno := 0
	maxErros := 6
	encerrado := false
	vencedor := ""

	fmt.Println("Servidor iniciado. Enviando estado inicial...")

	// Enviar estado inicial
	estadoInicial := GameState{
		Palavra:      palavra,
		Progresso:    progresso,
		Tentativas:   tentativas,
		MaxErros:     maxErros,
		LetrasUsadas: letras,
		Turno:        turno,
		Encerrado:    encerrado,
		Vencedor:     vencedor,
	}
	estadoBytes, _ := json.Marshal(estadoInicial)

	for _, cliente := range jogadores {
		ch.QueueDeclare(cliente, false, false, false, false, nil)
		ch.Publish("", cliente, false, false, amqp.Publishing{
			ContentType: "application/json",
			Body:        estadoBytes,
		})
	}

	fmt.Println("Estado inicial enviado. Aguardando palpites...")

	for msg := range msgs {
		fmt.Println("Mensagem recebida no servidor!")
		fmt.Printf("Body: %s\n", msg.Body)
		fmt.Printf("Headers: %v\n", msg.Headers)

		if encerrado {
			continue
		}

		player := msg.Headers["player"].(string)

		if player != jogadores[turno] {
			fmt.Printf("Palpite fora de turno: %s (turno atual: %s)\n", player, jogadores[turno])
			continue
		}

		letra := strings.ToLower(string(msg.Body))

		if strings.Contains(strings.Join(letras, ""), letra) {
			fmt.Println("Letra repetida. Ignorando.")
			continue
		}

		letras = append(letras, letra)

		acertou := false
		runProgresso := []rune(progresso)
		for i, c := range palavra {
			if string(c) == letra {
				runProgresso[i] = c
				acertou = true
			}
		}
		progresso = string(runProgresso)

		if !acertou {
			tentativas++
		}

		if progresso == palavra {
			encerrado = true
			vencedor = jogadores[turno]
		} else if tentativas >= maxErros {
			encerrado = true
			vencedor = "forca"
		} else {
			turno = (turno + 1) % 2
		}

		state := GameState{
			Palavra:      palavra,
			Progresso:    progresso,
			Tentativas:   tentativas,
			MaxErros:     maxErros,
			LetrasUsadas: letras,
			Turno:        turno,
			Encerrado:    encerrado,
			Vencedor:     vencedor,
		}

		stateBytes, _ := json.Marshal(state)

		for _, cliente := range jogadores {
			ch.QueueDeclare(cliente, false, false, false, false, nil)
			ch.Publish("", cliente, false, false, amqp.Publishing{
				ContentType: "application/json",
				Body:        stateBytes,
			})
		}

		if encerrado {
			fmt.Println("Jogo encerrado. Vencedor:", vencedor)
			time.Sleep(3 * time.Second)
			break
		}
	}
}
