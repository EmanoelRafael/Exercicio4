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
	LetrasErradas  []string
	JogadorDaVez   string
	VencedorId     string // padronizado para I minúsculo
	Status         int
}

const widthGame = 44
const MENU = 1
const AGUARDANDO_JOGADORES = 2
const INGRESSO = 3
const EM_ANDAMENTO = 4
const FIM_DE_JOGO = 5

const NAO_INICIADO = 0
const PENDENTE_JOGADORES = 1
const EM_CURSO = 2
const FINALIZADO = 3

// Structs para request/response via MQTT (JSON)
type CriarJogoRequest struct {
	JogadorId string `json:"jogador_id"`
	Solo      bool   `json:"solo"`
	ComAmigos bool   `json:"com_amigos"`
}

type CriarJogoResponse struct {
	CodigoJogo string `json:"codigo_jogo"`
	Mensagem   string `json:"mensagem"`
}

type EntrarJogoRequest struct {
	JogadorId  string `json:"jogador_id"`
	CodigoJogo string `json:"codigo_jogo"`
}

type EntrarJogoResponse struct {
	Mensagem string `json:"mensagem"`
	Sucesso  bool   `json:"sucesso"`
}

type PalpitarLetraRequest struct {
	JogadorId  string `json:"jogador_id"`
	CodigoJogo string `json:"codigo_jogo"`
	Letra      string `json:"letra"`
}

type PalpitarPalavraRequest struct {
	JogadorId  string `json:"jogador_id"`
	CodigoJogo string `json:"codigo_jogo"`
	Palavra    string `json:"palavra"`
}

type EstadoRequest struct {
	JogadorId  string `json:"jogador_id"`
	CodigoJogo string `json:"codigo_jogo"`
}

type AtualizacaoResponse struct {
	Mensagem       string   `json:"mensagem"`
	PalavraVisivel string   `json:"palavra_visivel"`
	ErrosJogador   int      `json:"erros_jogador"`
	LetrasErradas  []string `json:"letras_erradas"`
	JogadorDaVez   string   `json:"jogador_da_vez"`
	JogoStatus     int      `json:"jogo_status"`
	VencedorId     string   `json:"vencedor_id"`
	DesenhoForca   string   `json:"desenho_forca"`
}

func desenharForca(parte int) string {
	if parte < 0 || parte > 6 {
		return "Número inválido. Use um valor entre 0 e 6."
	}

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

	return forcas[parte]
}

func printLinhaGame(msg string, placeholder rune) {
	size := widthGame - 2
	freeSpace := size - len(msg)
	msgFinal := msg
	placeholderString := strings.Repeat(string(placeholder), freeSpace/2)

	if freeSpace%2 == 0 {
		msgFinal = placeholderString + msgFinal + placeholderString
	} else {
		msgFinal = placeholderString + msgFinal + placeholderString + string(placeholder)
	}

	fmt.Println(msgFinal)
}

func printGame() string {

	ret := ""

	reader := bufio.NewReader(os.Stdin)

	if gameStage == 1 {
		printLinhaGame("FORCA GAME", '-')
		printLinhaGame("", ' ')
		printLinhaGame("MENU", ' ')
		printLinhaGame("1 - JOGO SOLO", ' ')
		printLinhaGame("2 - JOGO COM AMIGOS", ' ')
		printLinhaGame("3 - ENTRAR EM UM JOGO", ' ')
		printLinhaGame("0 - SAIR", ' ')
		printLinhaGame("", ' ')
		printLinhaGame("", '-')
		fmt.Print("=> ")
		for ret != "1" && ret != "2" && ret != "3" && ret != "0" {
			ret, _ = reader.ReadString('\n')
			ret = strings.TrimSpace(ret)
			if ret != "1" && ret != "2" && ret != "3" && ret != "0" {
				fmt.Println("Opção inválida. Digite novamente. \n=> ")
			}
		}
	} else if gameStage == 2 {
		printLinhaGame("FORCA GAME", '-')
		printLinhaGame("", ' ')
		printLinhaGame("AGUARDANDO A ENTRADA DOS JOGADORES", ' ')
		printLinhaGame(" ", ' ')
		printLinhaGame("CODIGO DO JOGO: "+jogo.Codigo, ' ')
		printLinhaGame(" ", ' ')
		printLinhaGame("", ' ')
		printLinhaGame("", '-')
	} else if gameStage == 3 {
		printLinhaGame("FORCA GAME", '-')
		printLinhaGame("", ' ')
		printLinhaGame("INSIRA O CODIGO DO JOGO", ' ')
		printLinhaGame(" ", ' ')
		printLinhaGame(" ", ' ')
		printLinhaGame(" ", ' ')
		printLinhaGame("", ' ')
		printLinhaGame("", '-')
		fmt.Print("=> ")
		ret, _ = reader.ReadString('\n')
		ret = strings.TrimSpace(ret)
	} else if gameStage == 4 {
		printLinhaGame("FORCA GAME", '-')
		printLinhaGame("", ' ')
		printLinhaGame(string(jogo.PalavraVisivel), ' ')
		printLinhaGame(" ", ' ')
		if len(jogo.LetrasErradas) > 0 {
			printLinhaGame("LETRAS ERRADAS:", ' ')
			printLinhaGame(strings.Join(jogo.LetrasErradas, ", "), ' ')
		}
		printLinhaGame(" ", ' ')
		desenharForca(jogo.Erros)
		printLinhaGame(" ", ' ')

		if jogadorId == jogo.JogadorDaVez {
			printLinhaGame("1 - CHUTAR LETRA", ' ')
			printLinhaGame("2 - CHUTAR PALAVRA", ' ')
		} else {
			printLinhaGame(" ", ' ')
			printLinhaGame(" ", ' ')
			printLinhaGame(" ", ' ')
		}
		printLinhaGame("", ' ')
		printLinhaGame("", '-')
		if jogadorId == jogo.JogadorDaVez {
			fmt.Print("=> ")
			for ret != "1" && ret != "2" && ret != "3" && ret != "0" {
				ret, _ = reader.ReadString('\n')
				ret = strings.TrimSpace(ret)
				if ret != "1" && ret != "2" && ret != "3" && ret != "0" {
					fmt.Println("Opção inválida. Digite novamente. \n=> ")
				}
			}
		}
	} else if gameStage == 5 {
		printLinhaGame("FORCA GAME", '-')
		printLinhaGame("", ' ')
		printLinhaGame(string(jogo.PalavraVisivel), ' ')
		printLinhaGame(" ", ' ')
		printLinhaGame("FIM DE JOGO", ' ')
		if strings.ToLower(jogadorId) == strings.ToLower(jogo.VencedorId) {
			printLinhaGame("", ' ')
			printLinhaGame("PARABENS, VOCE VENCEU!", ' ')
		} else {
			printLinhaGame("", ' ')
			printLinhaGame("VOCE PERDEU!", ' ')
			desenharForca(jogo.Erros)
		}
		printLinhaGame("", ' ')
		printLinhaGame("", '-')
	}

	return ret
}

func criarJogoMQTT(client mqtt.Client, jogadorId string, solo bool, comAmigos bool) (*CriarJogoResponse, error) {
	// Monta o request
	req := CriarJogoRequest{
		JogadorId: jogadorId,
		Solo:      solo,
		ComAmigos: comAmigos,
	}
	payload, _ := json.Marshal(req)

	// Canal para receber a resposta
	respChan := make(chan *CriarJogoResponse)

	// Handler para resposta
	topicResp := fmt.Sprintf("forca/resp/criar_jogo/%s", jogadorId)
	token := client.Subscribe(topicResp, 0, func(_ mqtt.Client, msg mqtt.Message) {
		var resp CriarJogoResponse
		err := json.Unmarshal(msg.Payload(), &resp)
		if err == nil {
			respChan <- &resp
		}
	})
	token.Wait()

	// Publica o pedido
	client.Publish("forca/criar_jogo", 0, false, payload)

	// Aguarda resposta (timeout de 5s)
	select {
	case resp := <-respChan:
		client.Unsubscribe(topicResp)
		return resp, nil
	case <-time.After(5 * time.Second):
		client.Unsubscribe(topicResp)
		return nil, fmt.Errorf("timeout ao aguardar resposta do servidor")
	}
}

func entrarJogoMQTT(client mqtt.Client, jogadorId string, codigoJogo string) (*EntrarJogoResponse, error) {
	req := EntrarJogoRequest{
		JogadorId:  jogadorId,
		CodigoJogo: codigoJogo,
	}
	payload, _ := json.Marshal(req)

	respChan := make(chan *EntrarJogoResponse)
	topicResp := fmt.Sprintf("forca/resp/entrar_jogo/%s", jogadorId)
	token := client.Subscribe(topicResp, 0, func(_ mqtt.Client, msg mqtt.Message) {
		var resp EntrarJogoResponse
		err := json.Unmarshal(msg.Payload(), &resp)
		if err == nil {
			respChan <- &resp
		}
	})
	token.Wait()

	client.Publish("forca/entrar_jogo", 0, false, payload)

	select {
	case resp := <-respChan:
		client.Unsubscribe(topicResp)
		return resp, nil
	case <-time.After(5 * time.Second):
		client.Unsubscribe(topicResp)
		return nil, fmt.Errorf("timeout ao aguardar resposta do servidor")
	}
}

func palpitarLetraMQTT(client mqtt.Client, jogadorId, codigoJogo, letra string) (*AtualizacaoResponse, error) {
	req := PalpitarLetraRequest{
		JogadorId:  jogadorId,
		CodigoJogo: codigoJogo,
		Letra:      letra,
	}
	payload, _ := json.Marshal(req)
	respChan := make(chan *AtualizacaoResponse)
	topicResp := fmt.Sprintf("forca/resp/palpitar_letra/%s", jogadorId)
	client.Subscribe(topicResp, 0, func(_ mqtt.Client, msg mqtt.Message) {
		var resp AtualizacaoResponse
		if err := json.Unmarshal(msg.Payload(), &resp); err == nil {
			respChan <- &resp
		}
	})
	client.Publish("forca/palpitar_letra", 0, false, payload)
	select {
	case resp := <-respChan:
		client.Unsubscribe(topicResp)
		return resp, nil
	case <-time.After(5 * time.Second):
		client.Unsubscribe(topicResp)
		return nil, fmt.Errorf("timeout ao aguardar resposta do servidor")
	}
}

func palpitarPalavraMQTT(client mqtt.Client, jogadorId, codigoJogo, palavra string) (*AtualizacaoResponse, error) {
	req := PalpitarPalavraRequest{
		JogadorId:  jogadorId,
		CodigoJogo: codigoJogo,
		Palavra:    palavra,
	}
	payload, _ := json.Marshal(req)
	respChan := make(chan *AtualizacaoResponse)
	topicResp := fmt.Sprintf("forca/resp/palpitar_palavra/%s", jogadorId)
	client.Subscribe(topicResp, 0, func(_ mqtt.Client, msg mqtt.Message) {
		var resp AtualizacaoResponse
		if err := json.Unmarshal(msg.Payload(), &resp); err == nil {
			respChan <- &resp
		}
	})
	client.Publish("forca/palpitar_palavra", 0, false, payload)
	select {
	case resp := <-respChan:
		client.Unsubscribe(topicResp)
		return resp, nil
	case <-time.After(5 * time.Second):
		client.Unsubscribe(topicResp)
		return nil, fmt.Errorf("timeout ao aguardar resposta do servidor")
	}
}

func obterEstadoMQTT(client mqtt.Client, jogadorId, codigoJogo string) (*AtualizacaoResponse, error) {
	req := EstadoRequest{
		JogadorId:  jogadorId,
		CodigoJogo: codigoJogo,
	}
	payload, _ := json.Marshal(req)
	respChan := make(chan *AtualizacaoResponse)
	topicResp := fmt.Sprintf("forca/resp/obter_estado/%s", jogadorId)
	client.Subscribe(topicResp, 0, func(_ mqtt.Client, msg mqtt.Message) {
		var resp AtualizacaoResponse
		if err := json.Unmarshal(msg.Payload(), &resp); err == nil {
			respChan <- &resp
		}
	})
	client.Publish("forca/obter_estado", 0, false, payload)
	select {
	case resp := <-respChan:
		client.Unsubscribe(topicResp)
		return resp, nil
	case <-time.After(5 * time.Second):
		client.Unsubscribe(topicResp)
		return nil, fmt.Errorf("timeout ao aguardar resposta do servidor")
	}
}

func aguardarJogadores(client mqtt.Client, jogadorId, codigoJogo string) {
	for {
		resp, err := obterEstadoMQTT(client, jogadorId, codigoJogo)
		if err != nil {
			fmt.Println("Erro ao obter estado do jogo:", err)
			return
		}
		if resp.JogoStatus == 2 { // EM_CURSO
			// Aviso para o segundo jogador
			if resp.JogadorDaVez != jogadorId {
				fmt.Println("Aguarde, é a vez do outro jogador. Você só poderá jogar depois do jogador " + resp.JogadorDaVez)
			}
			return
		}
		time.Sleep(1 * time.Second)
	}
}

var jogadorId string
var codigoJogo string
var gameStage = MENU
var jogo *Jogo
var flag = false
var jogoStatusAntigo = NAO_INICIADO

func main() {
	broker := "tcp://localhost:1883"
	clientID := fmt.Sprintf("forca-client-%d", time.Now().UnixNano())
	opts := mqtt.NewClientOptions().AddBroker(broker).SetClientID(clientID)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("Erro ao conectar ao broker MQTT: %v", token.Error())
	}

	fmt.Println("Cliente MQTT conectado ao broker em", broker)

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Digite seu ID de jogador: ")
	jogadorId, _ = reader.ReadString('\n')
	jogadorId = strings.TrimSpace(jogadorId)
	jogadorId = strings.ToLower(jogadorId)

	var codigoJogo string
	var emJogo bool

	for {
		var opcao string
		if !emJogo {
			gameStage = MENU
			opcao = printGame()
		} else {
			// Buscar o estado atual do jogo
			estado, err := obterEstadoMQTT(client, jogadorId, codigoJogo)
			if err == nil {
				if estado.JogoStatus == FINALIZADO {
					gameStage = FIM_DE_JOGO
					jogo = &Jogo{
						PalavraVisivel: []rune(estado.PalavraVisivel),
						LetrasErradas:  estado.LetrasErradas,
						Erros:          estado.ErrosJogador,
						VencedorId:     estado.VencedorId,
					}
					printGame()
					emJogo = false
					continue
				} else if estado.JogadorDaVez == jogadorId {
					gameStage = EM_ANDAMENTO
					jogo = &Jogo{
						PalavraVisivel: []rune(estado.PalavraVisivel),
						LetrasErradas:  estado.LetrasErradas,
						Erros:          estado.ErrosJogador,
						JogadorDaVez:   estado.JogadorDaVez,
						VencedorId:     estado.VencedorId,
					}
					opcao = printGame()
				} else {
					// fmt.Println("Aguarde, é a vez do outro jogador. Você só poderá jogar depois do jogador " + estado.JogadorDaVez)
					// // Espera até ser a vez do jogador
					// for {
					// 	time.Sleep(2 * time.Second)
					// 	estado, err = obterEstadoMQTT(client, jogadorId, codigoJogo)
					// 	if err == nil && estado.JogoStatus != FINALIZADO && estado.JogadorDaVez == jogadorId {
					// 		break
					// 	}
					// 	if err == nil && estado.JogoStatus == FINALIZADO {
					// 		break
					// 	}
					// }
					// Após sair do loop, volta ao início do for para exibir printGame ou fim de jogo
					continue
				}
			}
		}

		switch opcao {
		case "1":
			if emJogo {
				estado, err := obterEstadoMQTT(client, jogadorId, codigoJogo)
				if err != nil {
					fmt.Println("Erro ao obter estado do jogo:", err)
					continue
				}
				if estado.JogoStatus == FINALIZADO {
					// fmt.Println("\n--- FIM DE JOGO ---")
					// fmt.Println(estado.Mensagem)
					// fmt.Println("Palavra correta:", estado.PalavraVisivel)
					// fmt.Println("Vencedor:", estado.VencedorId)
					gameStage = FIM_DE_JOGO
					jogo = &Jogo{
						PalavraVisivel: []rune(estado.PalavraVisivel),
						LetrasErradas:  estado.LetrasErradas,
						Erros:          estado.ErrosJogador,
						VencedorId:     estado.VencedorId,
					}
					printGame()
					emJogo = false
					continue
				}
				if estado.JogadorDaVez != jogadorId {
					fmt.Printf("Aguarde sua vez! Agora é a vez de: %s\n", estado.JogadorDaVez)
					continue
				}
				fmt.Print("Digite a letra: ")
				letra, _ := reader.ReadString('\n')
				letra = strings.TrimSpace(letra)
				resp, err := palpitarLetraMQTT(client, jogadorId, codigoJogo, letra)
				if err != nil {
					fmt.Println("Erro ao palpitar letra:", err)
				} else {
					fmt.Println(resp.Mensagem)
					if resp.DesenhoForca != "" {
						fmt.Println(resp.DesenhoForca)
						fmt.Printf("Erros: %d/7\n", resp.ErrosJogador)
					}
					if resp.JogoStatus == FINALIZADO {
						gameStage = FIM_DE_JOGO
						jogo = &Jogo{
							PalavraVisivel: []rune(resp.PalavraVisivel),
							LetrasErradas:  resp.LetrasErradas,
							Erros:          resp.ErrosJogador,
							VencedorId:     resp.VencedorId,
						}
						printGame()
						emJogo = false
						continue
					}
				}
				continue
			}
			resp, err := criarJogoMQTT(client, jogadorId, true, false)
			if err != nil {
				fmt.Println("Erro ao criar jogo solo:", err)
				continue
			}
			fmt.Println(resp.Mensagem)
			codigoJogo = resp.CodigoJogo
			emJogo = true
		case "2":
			if emJogo {
				estado, err := obterEstadoMQTT(client, jogadorId, codigoJogo)
				if err != nil {
					fmt.Println("Erro ao obter estado do jogo:", err)
					continue
				}
				if estado.JogoStatus == FINALIZADO {
					// fmt.Println("\n--- FIM DE JOGO ---")
					// fmt.Println(estado.Mensagem)
					// fmt.Println("Palavra correta:", estado.PalavraVisivel)
					// fmt.Println("Vencedor:", estado.VencedorId)
					gameStage = FIM_DE_JOGO
					jogo = &Jogo{
						PalavraVisivel: []rune(estado.PalavraVisivel),
						LetrasErradas:  estado.LetrasErradas,
						Erros:          estado.ErrosJogador,
						VencedorId:     estado.VencedorId,
					}
					printGame()
					emJogo = false
					continue
				}
				if estado.JogadorDaVez != jogadorId {
					fmt.Printf("Aguarde sua vez! Agora é a vez de: %s\n", estado.JogadorDaVez)
					continue
				}
				fmt.Print("Digite a palavra: ")
				palavra, _ := reader.ReadString('\n')
				palavra = strings.TrimSpace(palavra)
				resp, err := palpitarPalavraMQTT(client, jogadorId, codigoJogo, palavra)
				if err != nil {
					fmt.Println("Erro ao palpitar palavra:", err)
				} else {
					fmt.Println(resp.Mensagem)
					if resp.JogoStatus == FINALIZADO {
						gameStage = FIM_DE_JOGO
						jogo = &Jogo{
							PalavraVisivel: []rune(resp.PalavraVisivel),
							LetrasErradas:  resp.LetrasErradas,
							Erros:          resp.ErrosJogador,
							VencedorId:     resp.VencedorId,
						}
						printGame()
						emJogo = false
						continue
					}
				}
				continue
			}
			resp, err := criarJogoMQTT(client, jogadorId, false, true)
			if err != nil {
				fmt.Println("Erro ao criar jogo com amigos:", err)
				continue
			}
			fmt.Println(resp.Mensagem)
			codigoJogo = resp.CodigoJogo
			emJogo = true
			aguardarJogadores(client, jogadorId, codigoJogo)
		case "3":
			if emJogo {
				fmt.Println("Opção inválida.")
				continue
			}
			fmt.Print("Digite o código do jogo: ")
			codigoJogo, _ = reader.ReadString('\n')
			codigoJogo = strings.TrimSpace(codigoJogo)
			resp, err := entrarJogoMQTT(client, jogadorId, codigoJogo)
			if err != nil {
				fmt.Println("Erro ao entrar no jogo:", err)
				continue
			}
			fmt.Println(resp.Mensagem)
			if resp.Sucesso {
				emJogo = true
				aguardarJogadores(client, jogadorId, codigoJogo)
			}
		case "0":
			fmt.Println("Saindo...")
			return
		default:
			fmt.Println("Opção inválida.")
		}
	}
}
