package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	pb "ForcaGame/proto"

	"google.golang.org/grpc"
)

var jogadorId string
var codigoJogo string
var gameStage = MENU
var jogo *Jogo
var flag = false
var jogoStatusAntigo = NAO_INICIADO

func main() {

	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Erro ao conectar: %v", err)
	}
	defer conn.Close()
	client := pb.NewGameServiceClient(conn)

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

	fmt.Print("Digite seu ID de jogador: ")
	jogadorId, _ = reader.ReadString('\n')
	jogadorId = strings.TrimSpace(jogadorId)

	for {

		opcao := ""

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel()

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

		if gameStage == MENU {
			switch opcao {
			case "1":
				resp, err := client.CriarJogo(ctx, &pb.CriarJogoRequest{
					JogadorId: jogadorId,
					Solo:      true,
				})
				if err != nil {
					log.Println("Erro:", err)
					continue
				}
				codigoJogo = resp.CodigoJogo
				jogo.Codigo = resp.CodigoJogo
				gameStage = EM_ANDAMENTO
			case "2":
				resp, err := client.CriarJogo(ctx, &pb.CriarJogoRequest{
					JogadorId: jogadorId,
					ComAmigos: true,
				})
				if err != nil {
					log.Println("Erro:", err)
					continue
				}
				codigoJogo = resp.CodigoJogo
				jogo.Codigo = resp.CodigoJogo
				gameStage = AGUARDANDO_JOGADORES
			case "3":
				fmt.Print("Digite o codigo do jogo=> ")
				codigoJogo, _ = reader.ReadString('\n')
				codigoJogo = strings.TrimSpace(codigoJogo)
				resp, err := client.EntrarJogo(ctx, &pb.EntrarJogoRequest{
					JogadorId:  jogadorId,
					CodigoJogo: codigoJogo,
				})
				if err != nil {
					log.Println("Erro:", err)
					continue
				}
				if resp.Sucesso {
					gameStage = AGUARDANDO_JOGADORES
					jogo.Codigo = codigoJogo
				}
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

				_, err := client.PalpitarLetra(ctx, &pb.PalpitarLetraRequest{
					JogadorId:  jogadorId,
					CodigoJogo: codigoJogo,
					Letra:      letra,
				})
				if err != nil {
					log.Println("Erro:", err)
					continue
				}
			case "2":
				fmt.Print("Digite a palavra: ")
				palpite, _ := reader.ReadString('\n')
				palpite = strings.TrimSpace(palpite)

				_, err := client.PalpitarPalavra(ctx, &pb.PalpitarPalavraRequest{
					JogadorId:  jogadorId,
					CodigoJogo: codigoJogo,
					Palavra:    palpite,
				})
				if err != nil {
					log.Println("Erro:", err)
					continue
				}
			case "3":
				_, err := client.PedirDica(ctx, &pb.DicaRequest{
					JogadorId:  jogadorId,
					CodigoJogo: codigoJogo,
				})
				if err != nil {
					log.Println("Erro:", err)
					continue
				}
			case "0":
				fmt.Println("Saindo...")
				return

			default:
				fmt.Println("Opção inválida.")
			}
		}

		if gameStage == AGUARDANDO_JOGADORES || gameStage == EM_ANDAMENTO {
			resp, err := client.ObterEstado(ctx, &pb.EstadoRequest{
				CodigoJogo: codigoJogo,
				JogadorId:  jogadorId,
			})
			if err != nil {
				log.Println("Erro:", err)
				continue
			}

			jogo.PalavraVisivel = []rune(resp.PalavraVisivel)
			jogo.Erros = int(resp.ErrosJogador)
			jogo.LetrasErradas = resp.LetrasErradas
			jogo.JogadorDaVez = resp.JogadorDaVez
			jogo.Status = int(resp.JogoStatus)
			jogo.VencedorID = resp.VencedorId
			if jogo.Status == FINALIZADO {
				gameStage = FIM_DE_JOGO
			}
		}
	}
}
