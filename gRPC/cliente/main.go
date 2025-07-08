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

func main() {
	// Conexão com o servidor
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Erro ao conectar: %v", err)
	}
	defer conn.Close()
	client := pb.NewGameServiceClient(conn)

	reader := bufio.NewReader(os.Stdin)

	var jogadorId string
	var codigoJogo string

	fmt.Print("Digite seu ID de jogador: ")
	jogadorId, _ = reader.ReadString('\n')
	jogadorId = strings.TrimSpace(jogadorId)

	for {
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

		opcao, _ := reader.ReadString('\n')
		opcao = strings.TrimSpace(opcao)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

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
			fmt.Println(resp.Mensagem)

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
			fmt.Println(resp.Mensagem)

		case "3":
			fmt.Print("Digite o código do jogo: ")
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
			fmt.Println(resp.Mensagem)

		case "4":
			fmt.Print("Digite uma letra: ")
			letra, _ := reader.ReadString('\n')
			letra = strings.TrimSpace(letra)

			resp, err := client.PalpitarLetra(ctx, &pb.PalpitarLetraRequest{
				JogadorId:  jogadorId,
				CodigoJogo: codigoJogo,
				Letra:      letra,
			})
			if err != nil {
				log.Println("Erro:", err)
				continue
			}
			fmt.Println("Resposta:", resp.Mensagem)
			fmt.Println("Palavra:", resp.PalavraVisivel)

		case "5":
			fmt.Print("Digite a palavra: ")
			palpite, _ := reader.ReadString('\n')
			palpite = strings.TrimSpace(palpite)

			resp, err := client.PalpitarPalavra(ctx, &pb.PalpitarPalavraRequest{
				JogadorId:  jogadorId,
				CodigoJogo: codigoJogo,
				Palavra:    palpite,
			})
			if err != nil {
				log.Println("Erro:", err)
				continue
			}
			fmt.Println("Resposta:", resp.Mensagem)

		case "6":
			resp, err := client.PedirDica(ctx, &pb.DicaRequest{
				JogadorId:  jogadorId,
				CodigoJogo: codigoJogo,
			})
			if err != nil {
				log.Println("Erro:", err)
				continue
			}
			fmt.Println("Dica:", resp.Mensagem)
			fmt.Println("Palavra:", resp.PalavraVisivel)

		case "7":
			resp, err := client.ObterEstado(ctx, &pb.EstadoRequest{
				CodigoJogo: codigoJogo,
				JogadorId:  jogadorId,
			})
			if err != nil {
				log.Println("Erro:", err)
				continue
			}
			fmt.Println("Palavra:", resp.PalavraVisivel)
			fmt.Println("Erros:", resp.ErrosJogador)
			fmt.Println("Letras erradas:", resp.LetrasErradas)
			fmt.Println("Jogador da vez:", resp.JogadorDaVez)
			fmt.Println("Mensagem:", resp.Mensagem)

		case "0":
			fmt.Println("Saindo...")
			return

		default:
			fmt.Println("Opção inválida.")
		}
	}
}
