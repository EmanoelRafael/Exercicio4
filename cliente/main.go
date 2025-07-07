package main

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "ForcaGame/proto"

	"google.golang.org/grpc"
)

func main() {
	// Conectar ao servidor
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Erro ao conectar: %v", err)
	}
	defer conn.Close()

	client := pb.NewGameServiceClient(conn)

	// Exemplo 1: Criar jogo
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	fmt.Println("-> Criando jogo...")
	res, err := client.CriarJogo(ctx, &pb.CriarJogoRequest{
		JogadorId: "jogador123",
		ComAmigos: false,
		Solo:      true,
	})
	if err != nil {
		log.Fatalf("Erro ao criar jogo: %v", err)
	}
	fmt.Printf("Jogo criado: código = %s | mensagem = %s\n", res.CodigoJogo, res.Mensagem)

	// Exemplo 2: Palpitar uma letra
	fmt.Println("-> Enviando palpite de letra...")
	respLetra, err := client.PalpitarLetra(ctx, &pb.PalpitarLetraRequest{
		JogadorId:  "jogador123",
		CodigoJogo: res.CodigoJogo,
		Letra:      "a",
	})
	if err != nil {
		log.Fatalf("Erro ao palpitar letra: %v", err)
	}
	fmt.Printf("Resposta: %s | Palavra visível: %s\n", respLetra.Mensagem, respLetra.PalavraVisivel)

	// Você pode testar outros métodos aqui:
	// - client.EntrarJogo(...)
	// - client.PalpitarPalavra(...)
	// - client.PedirDica(...)
	// - client.ObterEstado(...)
}
