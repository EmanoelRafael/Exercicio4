package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Jogo struct {
	Codigo         string
	Palavra        string
	PalavraVisivel []rune
	Jogadores      []string
	Erros          map[string]int
	LetrasErradas  map[string]bool
	JogadorDaVez   string
	VencedorID     string
	Status         int
	Eliminados     map[string]bool
	UltimaJogada   time.Time
}

const letras = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const NAO_INICIADO = 0
const PENDENTE_JOGADORES = 1
const EM_CURSO = 2
const FINALIZADO = 3

// Adicionar constante para máximo de erros
const MAX_ERROS = 7

var IDX_PALAVRA = 0
var palavras = []string{"Desasnado", "Filantropo", "Idiossincrasia", "Juvenilizante", "Odiento", "Quimera", "Verossimilhança", "Xaropear", "Yanomami", "Vicissitude"}

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

// Mapa global de jogos e mutex
type JogosManager struct {
	jogos map[string]*Jogo
	mu    sync.Mutex
}

var jogosManager = &JogosManager{
	jogos: make(map[string]*Jogo),
}

func GerarCodigoJogo() string {
	b := make([]byte, 6)
	for i := range b {
		b[i] = letras[rand.Intn(len(letras))]
	}
	return string(b)
}

func obterPalavra() string {
	palavra := palavras[IDX_PALAVRA]
	IDX_PALAVRA += 1
	if IDX_PALAVRA >= len(palavras) {
		IDX_PALAVRA = 0
	}
	return palavra
}

func trocarTurno(j *Jogo) {
	jogadores := j.Jogadores
	atual := j.JogadorDaVez
	var idx int
	for i, id := range jogadores {
		if id == atual {
			idx = i
			break
		}
	}

	for {
		idx = (idx + 1) % len(jogadores)
		if !j.Eliminados[jogadores[idx]] {
			j.JogadorDaVez = jogadores[idx]
			break
		}
	}
}

func jogadoresRestantes(j *Jogo) []string {
	var restantes []string
	for _, id := range j.Jogadores {
		if !j.Eliminados[id] {
			restantes = append(restantes, id)
		}
	}
	return restantes
}

func palavraCompleta(palavraVisivel []rune) bool {
	for _, c := range palavraVisivel {
		if c == '_' {
			return false
		}
	}
	return true
}

func letrasErradasSlice(m map[string]bool) []string {
	letras := []string{}
	for l := range m {
		letras = append(letras, l)
	}
	return letras
}

func notificarFimDeJogo(client mqtt.Client, jogo *Jogo, mensagem string) {
	for _, jogador := range jogo.Jogadores {
		resp := AtualizacaoResponse{
			Mensagem:       mensagem,
			PalavraVisivel: string(jogo.PalavraVisivel),
			ErrosJogador:   jogo.Erros[jogador],
			LetrasErradas:  letrasErradasSlice(jogo.LetrasErradas),
			JogadorDaVez:   jogo.JogadorDaVez,
			JogoStatus:     jogo.Status,
			VencedorId:     jogo.VencedorID,
			DesenhoForca:   desenharForca(jogo.Erros[jogador]),
		}
		topicRespLetra := fmt.Sprintf("forca/resp/palpitar_letra/%s", jogador)
		client.Publish(topicRespLetra, 0, false, mustJson(resp))
		topicRespPalavra := fmt.Sprintf("forca/resp/palpitar_palavra/%s", jogador)
		client.Publish(topicRespPalavra, 0, false, mustJson(resp))
	}
}

func mustJson(v interface{}) []byte {
	b, _ := json.Marshal(v)
	return b
}

// Handler MQTT para criar jogo
func criarJogoHandler(client mqtt.Client, msg mqtt.Message) {
	var req CriarJogoRequest
	err := json.Unmarshal(msg.Payload(), &req)
	if err != nil {
		log.Println("Erro ao decodificar CriarJogoRequest:", err)
		return
	}

	jogosManager.mu.Lock()
	defer jogosManager.mu.Unlock()

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
		Jogadores:      []string{strings.ToLower(strings.TrimSpace(req.JogadorId))},
		JogadorDaVez:   strings.ToLower(strings.TrimSpace(req.JogadorId)),
		Erros:          map[string]int{strings.ToLower(strings.TrimSpace(req.JogadorId)): 0},
		LetrasErradas:  make(map[string]bool),
		Eliminados:     make(map[string]bool),
		Status:         EM_CURSO,
		VencedorID:     "",
		UltimaJogada:   time.Now(),
	}

	msgResp := "Jogo criado com sucesso"
	fmt.Println("Jogo criado com sucesso. A palavra escolhida foi " + palavra)
	if req.Solo {
		msgResp = "Jogo solo criado com sucesso"
	} else if req.ComAmigos {
		jogo.Status = PENDENTE_JOGADORES
		msgResp = "Jogo com amigos criado. Compartilhe o código: " + codigo
	}

	jogosManager.jogos[codigo] = jogo
	// (Opcional) go monitorarTimeout(codigo)

	resp := CriarJogoResponse{
		CodigoJogo: codigo,
		Mensagem:   msgResp,
	}
	respBytes, _ := json.Marshal(resp)

	topicResp := fmt.Sprintf("forca/resp/criar_jogo/%s", req.JogadorId)
	client.Publish(topicResp, 0, false, respBytes)
}

func entrarJogoHandler(client mqtt.Client, msg mqtt.Message) {
	var req EntrarJogoRequest
	err := json.Unmarshal(msg.Payload(), &req)
	if err != nil {
		log.Println("Erro ao decodificar EntrarJogoRequest:", err)
		return
	}

	jogosManager.mu.Lock()
	defer jogosManager.mu.Unlock()

	jogo, existe := jogosManager.jogos[req.CodigoJogo]
	if !existe {
		resp := EntrarJogoResponse{
			Mensagem: "Código de jogo inválido",
			Sucesso:  false,
		}
		respBytes, _ := json.Marshal(resp)
		topicResp := fmt.Sprintf("forca/resp/entrar_jogo/%s", req.JogadorId)
		client.Publish(topicResp, 0, false, respBytes)
		return
	}

	if jogo.Status == FINALIZADO {
		resp := EntrarJogoResponse{
			Mensagem: "O jogo já foi finalizado",
			Sucesso:  false,
		}
		respBytes, _ := json.Marshal(resp)
		topicResp := fmt.Sprintf("forca/resp/entrar_jogo/%s", req.JogadorId)
		client.Publish(topicResp, 0, false, respBytes)
		return
	}

	for _, j := range jogo.Jogadores {
		if j == strings.ToLower(strings.TrimSpace(req.JogadorId)) {
			resp := EntrarJogoResponse{
				Mensagem: "Você já está participando deste jogo",
				Sucesso:  true,
			}
			respBytes, _ := json.Marshal(resp)
			topicResp := fmt.Sprintf("forca/resp/entrar_jogo/%s", req.JogadorId)
			client.Publish(topicResp, 0, false, respBytes)
			return
		}
	}

	if len(jogo.Jogadores) >= 2 {
		resp := EntrarJogoResponse{
			Mensagem: "O jogo já possui 2 jogadores",
			Sucesso:  false,
		}
		respBytes, _ := json.Marshal(resp)
		topicResp := fmt.Sprintf("forca/resp/entrar_jogo/%s", req.JogadorId)
		client.Publish(topicResp, 0, false, respBytes)
		return
	}

	jogo.Jogadores = append(jogo.Jogadores, strings.ToLower(strings.TrimSpace(req.JogadorId)))
	jogo.Erros[strings.ToLower(strings.TrimSpace(req.JogadorId))] = 0

	// Alterar status para EM_CURSO quando houver pelo menos 2 jogadores
	if len(jogo.Jogadores) >= 2 && jogo.Status == PENDENTE_JOGADORES {
		jogo.Status = EM_CURSO
	}

	resp := EntrarJogoResponse{
		Mensagem: "Jogador adicionado ao jogo com sucesso",
		Sucesso:  true,
	}
	respBytes, _ := json.Marshal(resp)
	topicResp := fmt.Sprintf("forca/resp/entrar_jogo/%s", req.JogadorId)
	client.Publish(topicResp, 0, false, respBytes)
}

func obterEstadoHandler(client mqtt.Client, msg mqtt.Message) {
	var req EstadoRequest
	err := json.Unmarshal(msg.Payload(), &req)
	if err != nil {
		log.Println("Erro ao decodificar EstadoRequest:", err)
		return
	}

	jogosManager.mu.Lock()
	defer jogosManager.mu.Unlock()

	jogo, existe := jogosManager.jogos[req.CodigoJogo]
	if !existe {
		return
	}

	resp := AtualizacaoResponse{
		PalavraVisivel: string(jogo.PalavraVisivel),
		Mensagem:       "Estado atual do jogo",
		JogadorDaVez:   jogo.JogadorDaVez,
		JogoStatus:     jogo.Status,
		VencedorId:     jogo.VencedorID,
		LetrasErradas:  letrasErradasSlice(jogo.LetrasErradas),
		ErrosJogador:   jogo.Erros[strings.ToLower(strings.TrimSpace(req.JogadorId))],
		DesenhoForca:   desenharForca(jogo.Erros[strings.ToLower(strings.TrimSpace(req.JogadorId))]),
	}
	respBytes, _ := json.Marshal(resp)
	topicResp := fmt.Sprintf("forca/resp/obter_estado/%s", req.JogadorId)
	client.Publish(topicResp, 0, false, respBytes)
}

func palpitarLetraHandler(client mqtt.Client, msg mqtt.Message) {
	var req PalpitarLetraRequest
	err := json.Unmarshal(msg.Payload(), &req)
	if err != nil {
		log.Println("Erro ao decodificar PalpitarLetraRequest:", err)
		return
	}

	jogosManager.mu.Lock()
	defer jogosManager.mu.Unlock()

	jogo, existe := jogosManager.jogos[req.CodigoJogo]
	if !existe || jogo.Status == FINALIZADO {
		return
	}

	if jogo.JogadorDaVez != strings.ToLower(strings.TrimSpace(req.JogadorId)) {
		resp := AtualizacaoResponse{
			Mensagem:       "Não é sua vez",
			PalavraVisivel: string(jogo.PalavraVisivel),
			JogoStatus:     jogo.Status,
			JogadorDaVez:   jogo.JogadorDaVez,
			DesenhoForca:   desenharForca(jogo.Erros[strings.ToLower(strings.TrimSpace(req.JogadorId))]),
		}
		respBytes, _ := json.Marshal(resp)
		topicResp := fmt.Sprintf("forca/resp/palpitar_letra/%s", req.JogadorId)
		client.Publish(topicResp, 0, false, respBytes)
		return
	}

	letra := []rune(req.Letra)
	if len(letra) != 1 {
		resp := AtualizacaoResponse{
			Mensagem:       "Letra inválida",
			PalavraVisivel: string(jogo.PalavraVisivel),
			JogoStatus:     jogo.Status,
			JogadorDaVez:   jogo.JogadorDaVez,
			DesenhoForca:   desenharForca(jogo.Erros[strings.ToLower(strings.TrimSpace(req.JogadorId))]),
		}
		respBytes, _ := json.Marshal(resp)
		topicResp := fmt.Sprintf("forca/resp/palpitar_letra/%s", req.JogadorId)
		client.Publish(topicResp, 0, false, respBytes)
		return
	}

	palpite := strings.ToUpper(req.Letra)
	acertou := false
	resultado := "Letra correta"
	for i, l := range jogo.Palavra {
		if strings.ToUpper(string(l)) == palpite {
			jogo.PalavraVisivel[i] = l
			acertou = true
		}
	}

	if !acertou {
		jogo.Erros[strings.ToLower(strings.TrimSpace(req.JogadorId))]++
		jogo.LetrasErradas[palpite] = true
		if jogo.Erros[strings.ToLower(strings.TrimSpace(req.JogadorId))] >= MAX_ERROS {
			jogo.Eliminados[strings.ToLower(strings.TrimSpace(req.JogadorId))] = true
			// Verifica se restou apenas um jogador
			restantes := jogadoresRestantes(jogo)
			if len(restantes) == 1 {
				jogo.VencedorID = restantes[0]
				jogo.Status = FINALIZADO
				notificarFimDeJogo(client, jogo, "FIM DE JOGO! Vencedor: "+restantes[0])
				return
			} else {
				resultado = "Você errou 5 vezes e foi eliminado!"
			}
		} else {
			resultado = "Letra incorreta. Restam " + fmt.Sprint(MAX_ERROS-jogo.Erros[strings.ToLower(strings.TrimSpace(req.JogadorId))]) + " tentativas"
		}
	}

	if palavraCompleta(jogo.PalavraVisivel) {
		jogo.VencedorID = strings.ToLower(strings.TrimSpace(req.JogadorId))
		jogo.Status = FINALIZADO
		notificarFimDeJogo(client, jogo, "FIM DE JOGO! Vencedor: "+strings.ToLower(strings.TrimSpace(req.JogadorId)))
		return
	} else if !acertou && jogo.Erros[strings.ToLower(strings.TrimSpace(req.JogadorId))] >= MAX_ERROS {
		return // já notificou acima
	} else {
		trocarTurno(jogo)
	}

	resp := AtualizacaoResponse{
		Mensagem:       resultado,
		PalavraVisivel: string(jogo.PalavraVisivel),
		ErrosJogador:   jogo.Erros[strings.ToLower(strings.TrimSpace(req.JogadorId))],
		LetrasErradas:  letrasErradasSlice(jogo.LetrasErradas),
		JogadorDaVez:   jogo.JogadorDaVez,
		JogoStatus:     jogo.Status,
		VencedorId:     jogo.VencedorID,
		DesenhoForca:   desenharForca(jogo.Erros[strings.ToLower(strings.TrimSpace(req.JogadorId))]),
	}
	respBytes, _ := json.Marshal(resp)
	topicResp := fmt.Sprintf("forca/resp/palpitar_letra/%s", req.JogadorId)
	client.Publish(topicResp, 0, false, respBytes)
}

func palpitarPalavraHandler(client mqtt.Client, msg mqtt.Message) {
	var req PalpitarPalavraRequest
	err := json.Unmarshal(msg.Payload(), &req)
	if err != nil {
		log.Println("Erro ao decodificar PalpitarPalavraRequest:", err)
		return
	}

	jogosManager.mu.Lock()
	defer jogosManager.mu.Unlock()

	jogo, existe := jogosManager.jogos[req.CodigoJogo]
	if !existe || jogo.Status == FINALIZADO {
		return
	}

	if jogo.JogadorDaVez != strings.ToLower(strings.TrimSpace(req.JogadorId)) {
		resp := AtualizacaoResponse{
			Mensagem:       "Não é sua vez",
			PalavraVisivel: string(jogo.PalavraVisivel),
			JogoStatus:     jogo.Status,
			JogadorDaVez:   jogo.JogadorDaVez,
		}
		respBytes, _ := json.Marshal(resp)
		topicResp := fmt.Sprintf("forca/resp/palpitar_palavra/%s", req.JogadorId)
		client.Publish(topicResp, 0, false, respBytes)
		return
	}

	palpite := strings.ToUpper(strings.TrimSpace(req.Palavra))
	palavraCerta := strings.ToUpper(jogo.Palavra)
	acertou := palpite == palavraCerta
	resultado := "Palpite incorreto. Você foi eliminado."

	if acertou {
		for i, l := range jogo.Palavra {
			jogo.PalavraVisivel[i] = l
		}
		jogo.VencedorID = strings.ToLower(strings.TrimSpace(req.JogadorId))
		jogo.Status = FINALIZADO
		notificarFimDeJogo(client, jogo, "FIM DE JOGO! Vencedor: "+strings.ToLower(strings.TrimSpace(req.JogadorId)))
		return
	} else {
		jogo.Eliminados[strings.ToLower(strings.TrimSpace(req.JogadorId))] = true
		// Verifica se restou apenas um jogador
		restantes := jogadoresRestantes(jogo)
		if len(restantes) == 1 {
			jogo.VencedorID = restantes[0]
			jogo.Status = FINALIZADO
			notificarFimDeJogo(client, jogo, "FIM DE JOGO! Vencedor: "+restantes[0])
			return
		} else {
			trocarTurno(jogo)
		}
	}

	resp := AtualizacaoResponse{
		Mensagem:       resultado,
		PalavraVisivel: string(jogo.PalavraVisivel),
		ErrosJogador:   jogo.Erros[strings.ToLower(strings.TrimSpace(req.JogadorId))],
		LetrasErradas:  letrasErradasSlice(jogo.LetrasErradas),
		JogadorDaVez:   jogo.JogadorDaVez,
		JogoStatus:     jogo.Status,
		VencedorId:     jogo.VencedorID,
	}
	respBytes, _ := json.Marshal(resp)
	topicResp := fmt.Sprintf("forca/resp/palpitar_palavra/%s", req.JogadorId)
	client.Publish(topicResp, 0, false, respBytes)
}

// Função para desenhar a forca
func desenharForca(parte int) string {
	if parte < 0 || parte > 5 {
		parte = 5
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

func main() {
	// Inicializa aleatoriedade para geração de código
	time.Sleep(1 * time.Millisecond)

	// Configuração MQTT
	broker := "tcp://localhost:1883"
	clientID := fmt.Sprintf("forca-server-%d", time.Now().UnixNano())
	opts := mqtt.NewClientOptions().AddBroker(broker).SetClientID(clientID)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("Erro ao conectar ao broker MQTT: %v", token.Error())
	}

	fmt.Println("Servidor MQTT conectado ao broker em", broker)

	// Subscreve no tópico de criar jogo
	client.Subscribe("forca/criar_jogo", 0, criarJogoHandler)
	client.Subscribe("forca/entrar_jogo", 0, entrarJogoHandler)
	client.Subscribe("forca/obter_estado", 0, obterEstadoHandler)
	client.Subscribe("forca/palpitar_letra", 0, palpitarLetraHandler)
	client.Subscribe("forca/palpitar_palavra", 0, palpitarPalavraHandler)

	// Mantém o servidor rodando
	select {}
}
