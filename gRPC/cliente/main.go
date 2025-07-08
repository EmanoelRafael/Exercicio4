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

/*
	menu principal (tela 1)
		- se iniciou um jogo solo
			- iniciar jogo solo
		- se criou um jogo com amigos
			- tela de espera para inicio do jogo (tela 2)
				- confirma inicio do jogo
					- inicia jogo
		- se entrar em um jogo
			- tela para inserir o codigo do jogo (tela 3)
				- entra no jogo
	tela do jogo (tela 4)
		- palavra
		- boneco do jogador
		- msg de status (esperando o jogador x jogar)
		- opcoes (se for a vez do jogador - chutar letra, chutar palavra, pedir dica)
	tela de fim de jogo
		- tem mensagem de parabens ou derrota (tela 5)


		fmt.Println("\n------ MENU ------")

fmt.Println("1. Criar jogo solo")
fmt.Println("2. Criar jogo com amigos")
fmt.Println("3. Entrar em jogo")
fmt.Println("4. Palpitar letra")
fmt.Println("5. Palpitar palavra")
fmt.Println("6. Pedir dica")
fmt.Println("7. Obter estado do jogo") //Quando o rabbitMQ for implementado isso vai ser desnecessario?
fmt.Println("0. Sair")
fmt.Print("Escolha uma opção: ")
*/
var jogadorId string
var codigoJogo string
var tipoMenu = 1
var jogo *Jogo

func main() {
	// Conexão com o servidor
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
		Status:         0,
	}

	fmt.Print("Digite seu ID de jogador: ")
	jogadorId, _ = reader.ReadString('\n')
	jogadorId = strings.TrimSpace(jogadorId)

	for {

		printGame("")

		opcao := ""

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		if tipoMenu == 1 {
			opcao, _ = reader.ReadString('\n')
			opcao = strings.TrimSpace(opcao)
			if opcao != "1" && opcao != "2" && opcao != "3" && opcao != "0" {
				fmt.Println("Opção inválida.")
				continue
			}
		} else if tipoMenu != 5 {
			resp, err := client.ObterEstado(ctx, &pb.EstadoRequest{
				CodigoJogo: codigoJogo,
				JogadorId:  jogadorId,
			})
			cancel()
			if err != nil {
				log.Println("Erro:", err)
				continue
			}

			jogo.PalavraVisivel = []rune(resp.PalavraVisivel)
			jogo.Erros = int(resp.ErrosJogador)
			jogo.LetrasErradas = resp.LetrasErradas
			jogo.JogadorDaVez = resp.JogadorDaVez
			jogo.Status = int(resp.JogoStatus)
			fmt.Println("Mensagem:", resp.Mensagem)

			// se o estado do jogo for em espera pela entrada de novos integrantes (dar a opcao de continuar se tiver mais de 2 ou mais integrantes e
			// espera 10s para atualizar o estado do jogo, se tiver 4 integrantes o jogo inicia automaticamente)
			// se o estado do jogo for ativo e nao estiver na vez do jogador atual espera 5s para atualizar o estado do jogo
			// se o estado do jogo for ativo e estiver na vez do jogador atual apresenta as opcoes pro jogador (pedir dica, chutar letra e chutar palavra)

			if jogo.Status == 1 {
				tipoMenu = 2
			} else if jogo.Status == 2 {
				tipoMenu = 4
				if jogo.JogadorDaVez == jogadorId {
					opcao, _ = reader.ReadString('\n')
					opcao = strings.TrimSpace(opcao)
					if opcao != "1" && opcao != "2" && opcao != "3" {
						fmt.Println("Opção inválida.")
						continue
					}
				}
			} else if jogo.Status == 3 {
				tipoMenu = 5
			}

		} else {
			tipoMenu = 1
		}

		if tipoMenu == 1 {
			switch opcao {
			case "1":
				resp, err := client.CriarJogo(ctx, &pb.CriarJogoRequest{
					JogadorId: jogadorId,
					Solo:      true,
				})
				cancel()
				if err != nil {
					log.Println("Erro:", err)
					continue
				}
				codigoJogo = resp.CodigoJogo
				jogo.Codigo = resp.CodigoJogo
				tipoMenu = 4
				fmt.Println(resp.Mensagem)

			case "2":
				resp, err := client.CriarJogo(ctx, &pb.CriarJogoRequest{
					JogadorId: jogadorId,
					ComAmigos: true,
				})
				cancel()
				if err != nil {
					log.Println("Erro:", err)
					continue
				}
				codigoJogo = resp.CodigoJogo
				jogo.Codigo = resp.CodigoJogo
				tipoMenu = 2
				fmt.Println(resp.Mensagem)

			case "3":
				fmt.Print("Digite o código do jogo: ")
				codigoJogo, _ = reader.ReadString('\n')
				codigoJogo = strings.TrimSpace(codigoJogo)

				resp, err := client.EntrarJogo(ctx, &pb.EntrarJogoRequest{
					JogadorId:  jogadorId,
					CodigoJogo: codigoJogo,
				})
				cancel()
				if err != nil {
					log.Println("Erro:", err)
					continue
				}
				tipoMenu = 4
				jogo.Codigo = codigoJogo
				fmt.Println(resp.Mensagem)
			case "0":
				fmt.Println("Saindo...")
				return

			default:
				fmt.Println("Opção inválida.")
			}
		} else if tipoMenu == 4 && jogo.JogadorDaVez == jogadorId {
			switch opcao {
			case "1":
				fmt.Print("Digite uma letra: ")
				letra, _ := reader.ReadString('\n')
				letra = strings.TrimSpace(letra)

				resp, err := client.PalpitarLetra(ctx, &pb.PalpitarLetraRequest{
					JogadorId:  jogadorId,
					CodigoJogo: codigoJogo,
					Letra:      letra,
				})
				cancel()
				if err != nil {
					log.Println("Erro:", err)
					continue
				}
				fmt.Println("Resposta:", resp.Mensagem)
				fmt.Println("Palavra:", resp.PalavraVisivel)

			case "2":
				fmt.Print("Digite a palavra: ")
				palpite, _ := reader.ReadString('\n')
				palpite = strings.TrimSpace(palpite)

				resp, err := client.PalpitarPalavra(ctx, &pb.PalpitarPalavraRequest{
					JogadorId:  jogadorId,
					CodigoJogo: codigoJogo,
					Palavra:    palpite,
				})
				cancel()
				if err != nil {
					log.Println("Erro:", err)
					continue
				}
				fmt.Println("Resposta:", resp.Mensagem)

			case "3":
				resp, err := client.PedirDica(ctx, &pb.DicaRequest{
					JogadorId:  jogadorId,
					CodigoJogo: codigoJogo,
				})
				cancel()
				if err != nil {
					log.Println("Erro:", err)
					continue
				}
				fmt.Println("Dica:", resp.Mensagem)
				fmt.Println("Palavra:", resp.PalavraVisivel)

			case "0":
				fmt.Println("Saindo...")
				return

			default:
				fmt.Println("Opção inválida.")
			}
		}

	}
}
