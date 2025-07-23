// totalCliente.go
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Jogo struct {
	Codigo         string
	PalavraVisivel []rune
	Erros          int
	DicaUsada      bool
	LetrasErradas  []string
	JogadorDaVez   string
	VencedorID     string
	Status         int
}

const widthGame = 44
const (
	MENU               = 1
	AGUARDANDO_JOGADORES = 2
	INGRESSO           = 3
	EM_ANDAMENTO       = 4
	FIM_DE_JOGO        = 5
)

const (
	NAO_INICIADO       = 0
	PENDENTE_JOGADORES = 1
	EM_CURSO           = 2
	FINALIZADO         = 3
)

func desenharForca(parte int) string {
	forcas := []string{
		`
  _______
 |/      |
 |
 |
 |
 |
_|___
`,
		`
  _______
 |/      |
 |      (_)
 |
 |
 |
_|___
`,
		`
  _______
 |/      |
 |      (_)
 |       |
 |
 |
_|___
`,
		`
  _______
 |/      |
 |      (_)
 |      \|
 |
 |
_|___
`,
		`
  _______
 |/      |
 |      (_)
 |      \|/
 |
 |
_|___
`,
		`
  _______
 |/      |
 |      (_)
 |      \|/
 |       |
 |
_|___
`,
		`
  _______
 |/      |
 |      (_)
 |      \|/
 |       |
 |      / \
_|___
`,
	}
	if parte < 0 || parte >= len(forcas) {
		parte = len(forcas) - 1
	}
	return forcas[parte]
}

func printLinhaGame(msg string, placeholder rune) {
	size := widthGame - 2
	freeSpace := size - len(msg)
	placeholderString := strings.Repeat(string(placeholder), freeSpace/2)
	if freeSpace%2 == 0 {
		msg = placeholderString + msg + placeholderString
	} else {
		msg = placeholderString + msg + placeholderString + string(placeholder)
	}
	fmt.Println(msg)
}

func printGame(jogo *Jogo, jogadorId string) {
	fmt.Println(strings.Repeat("-", widthGame))
	fmt.Printf("Palavra: %s\n", string(jogo.PalavraVisivel))
	fmt.Printf("Letras erradas: %s\n", strings.Join(jogo.LetrasErradas, ", "))
	fmt.Println(desenharForca(jogo.Erros))
	if jogo.Status == FINALIZADO {
		if jogo.VencedorID == jogadorId {
			fmt.Println("Parabéns! Você venceu!")
		} else {
			fmt.Println("Você perdeu.")
		}
	}
	fmt.Println(strings.Repeat("-", widthGame))
}

type (
	CriarJogoRequest struct {
		JogadorId string `json:"jogador_id"`
		Solo      bool   `json:"solo"`
		ComAmigos bool   `json:"com_amigos"`
	}

	CriarJogoResponse struct {
		CodigoJogo string `json:"codigo_jogo"`
		Mensagem   string `json:"mensagem"`
	}

	EntrarJogoRequest struct {
		JogadorId  string `json:"jogador_id"`
		CodigoJogo string `json:"codigo_jogo"`
	}

	EntrarJogoResponse struct {
		Mensagem string `json:"mensagem"`
		Sucesso  bool   `json:"sucesso"`
	}

	PalpitarLetraRequest struct {
		JogadorId  string `json:"jogador_id"`
		CodigoJogo string `json:"codigo_jogo"`
		Letra      string `json:"letra"`
	}

	PalpitarPalavraRequest struct {
		JogadorId  string `json:"jogador_id"`
		CodigoJogo string `json:"codigo_jogo"`
		Palavra    string `json:"palavra"`
	}

	DicaRequest struct {
		JogadorId  string `json:"jogador_id"`
		CodigoJogo string `json:"codigo_jogo"`
	}

	EstadoRequest struct {
		JogadorId  string `json:"jogador_id"`
		CodigoJogo string `json:"codigo_jogo"`
	}

	AtualizacaoResponse struct {
		Mensagem       string   `json:"mensagem"`
		PalavraVisivel string   `json:"palavra_visivel"`
		ErrosJogador   int      `json:"erros_jogador"`
		LetrasErradas  []string `json:"letras_erradas"`
		JogadorDaVez   string   `json:"jogador_da_vez"`
		JogoStatus     int      `json:"jogo_status"`
		VencedorId     string   `json:"vencedor_id"`
	}
)

func mqttRequest[T any](client mqtt.Client, topicReq, topicResp string, payload []byte) (*T, error) {
	respChan := make(chan *T)
	client.Subscribe(topicResp, 0, func(_ mqtt.Client, msg mqtt.Message) {
		var resp T
		if err := json.Unmarshal(msg.Payload(), &resp); err == nil {
			respChan <- &resp
		}
	})
	client.Publish(topicReq, 0, false, payload)
	select {
	case resp := <-respChan:
		client.Unsubscribe(topicResp)
		return resp, nil
	case <-time.After(5 * time.Second):
		client.Unsubscribe(topicResp)
		return nil, fmt.Errorf("timeout")
	}
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Digite seu ID de jogador: ")
	jogadorId, _ := reader.ReadString('\n')
	jogadorId = strings.TrimSpace(jogadorId)

	clientID := fmt.Sprintf("forca-client-%d", time.Now().UnixNano())
	opts := mqtt.NewClientOptions().AddBroker("tcp://localhost:1883").SetClientID(clientID)
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("Erro ao conectar ao broker: %v", token.Error())
	}
	defer client.Disconnect(250)

	var codigoJogo string

	for {
		fmt.Println("\n--- MENU ---")
		fmt.Println("1 - Criar jogo solo")
		fmt.Println("2 - Criar jogo com amigos")
		fmt.Println("3 - Entrar em um jogo existente")
		fmt.Println("4 - Palpitar letra")
		fmt.Println("5 - Palpitar palavra")
		fmt.Println("6 - Pedir dica")
		fmt.Println("7 - Ver estado do jogo")
		fmt.Println("0 - Sair")
		fmt.Print("Escolha uma opção: ")

		opcao, _ := reader.ReadString('\n')
		opcao = strings.TrimSpace(opcao)

		switch opcao {
		case "1", "2":
			solo := opcao == "1"
			req := CriarJogoRequest{JogadorId: jogadorId, Solo: solo, ComAmigos: !solo}
			payload, _ := json.Marshal(req)
			topicResp := fmt.Sprintf("forca/resp/criar_jogo/%s", jogadorId)
			resp, err := mqttRequest[CriarJogoResponse](client, "forca/criar_jogo", topicResp, payload)
			if err != nil {
				fmt.Println("Erro ao criar jogo:", err)
				continue
			}
			fmt.Println(resp.Mensagem)
			codigoJogo = resp.CodigoJogo

		case "3":
			fmt.Print("Digite o código do jogo: ")
			codigoJogo, _ = reader.ReadString('\n')
			codigoJogo = strings.TrimSpace(codigoJogo)
			req := EntrarJogoRequest{JogadorId: jogadorId, CodigoJogo: codigoJogo}
			payload, _ := json.Marshal(req)
			topicResp := fmt.Sprintf("forca/resp/entrar_jogo/%s", jogadorId)
			resp, err := mqttRequest[EntrarJogoResponse](client, "forca/entrar_jogo", topicResp, payload)
			if err != nil || !resp.Sucesso {
				fmt.Println("Erro ao entrar:", err)
				continue
			}
			fmt.Println(resp.Mensagem)

		case "4":
			fmt.Print("Digite a letra: ")
			letra, _ := reader.ReadString('\n')
			req := PalpitarLetraRequest{JogadorId: jogadorId, CodigoJogo: codigoJogo, Letra: strings.TrimSpace(letra)}
			payload, _ := json.Marshal(req)
			topicResp := fmt.Sprintf("forca/resp/palpitar_letra/%s", jogadorId)
			resp, err := mqttRequest[AtualizacaoResponse](client, "forca/palpitar_letra", topicResp, payload)
			if err != nil {
				fmt.Println("Erro:", err)
			} else {
				jogo := Jogo{
					PalavraVisivel: []rune(resp.PalavraVisivel),
					Erros:          resp.ErrosJogador,
					LetrasErradas:  resp.LetrasErradas,
					VencedorID:     resp.VencedorId,
					Status:         resp.JogoStatus,
				}
				printGame(&jogo, jogadorId)
			}

		case "5":
			fmt.Print("Digite a palavra: ")
			palavra, _ := reader.ReadString('\n')
			req := PalpitarPalavraRequest{JogadorId: jogadorId, CodigoJogo: codigoJogo, Palavra: strings.TrimSpace(palavra)}
			payload, _ := json.Marshal(req)
			topicResp := fmt.Sprintf("forca/resp/palpitar_palavra/%s", jogadorId)
			resp, err := mqttRequest[AtualizacaoResponse](client, "forca/palpitar_palavra", topicResp, payload)
			if err != nil {
				fmt.Println("Erro:", err)
			} else {
				jogo := Jogo{
					PalavraVisivel: []rune(resp.PalavraVisivel),
					Erros:          resp.ErrosJogador,
					LetrasErradas:  resp.LetrasErradas,
					VencedorID:     resp.VencedorId,
					Status:         resp.JogoStatus,
				}
				printGame(&jogo, jogadorId)
			}

		case "6":
			req := DicaRequest{JogadorId: jogadorId, CodigoJogo: codigoJogo}
			payload, _ := json.Marshal(req)
			topicResp := fmt.Sprintf("forca/resp/pedir_dica/%s", jogadorId)
			resp, err := mqttRequest[AtualizacaoResponse](client, "forca/pedir_dica", topicResp, payload)
			if err != nil {
				fmt.Println("Erro ao pedir dica:", err)
			} else {
				fmt.Println(resp.Mensagem)
			}

		case "7":
			req := EstadoRequest{JogadorId: jogadorId, CodigoJogo: codigoJogo}
			payload, _ := json.Marshal(req)
			topicResp := fmt.Sprintf("forca/resp/obter_estado/%s", jogadorId)
			resp, err := mqttRequest[AtualizacaoResponse](client, "forca/obter_estado", topicResp, payload)
			if err != nil {
				fmt.Println("Erro ao obter estado:", err)
			} else {
				jogo := Jogo{
					PalavraVisivel: []rune(resp.PalavraVisivel),
					Erros:          resp.ErrosJogador,
					LetrasErradas:  resp.LetrasErradas,
					VencedorID:     resp.VencedorId,
					Status:         resp.JogoStatus,
				}
				printGame(&jogo, jogadorId)
			}

		case "0":
			fmt.Println("Saindo...")
			return

		default:
			fmt.Println("Opção inválida.")
		}
	}
}
