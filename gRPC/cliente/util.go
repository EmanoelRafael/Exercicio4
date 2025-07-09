package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

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

func desenharBoneco(erros int) {
	bonecos := []string{
		`
  +---+
  |   |
      |
      |
      |
      |
=========`,
		`
  +---+
  |   |
  O   |
      |
      |
      |
=========`,
		`
  +---+
  |   |
  O   |
  |   |
      |
      |
=========`,
		`
  +---+
  |   |
  O   |
 /|   |
      |
      |
=========`,
		`
  +---+
  |   |
  O   |
 /|\  |
      |
      |
=========`,
		`
  +---+
  |   |
  O   |
 /|\  |
 /    |
      |
=========`,
		`
  +---+
  |   |
  O   |
 /|\  |
 / \  |
      |
=========`,
	}

	if erros < 0 {
		erros = 0
	} else if erros > 6 {
		erros = 6
	}

	fmt.Println(bonecos[erros])
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
		desenharBoneco(jogo.Erros)
		printLinhaGame(" ", ' ')

		if jogadorId == jogo.JogadorDaVez {
			printLinhaGame("1 - CHUTAR LETRA", ' ')
			printLinhaGame("2 - CHUTAR PALAVRA", ' ')
			printLinhaGame("3 - PEDIR DICA", ' ')
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
		if jogadorId == jogo.VencedorID {
			printLinhaGame("", ' ')
			printLinhaGame("PARABENS, VOCE VENCEU!", ' ')
		} else {
			printLinhaGame("", ' ')
			printLinhaGame("VOCE PERDEU!", ' ')
			desenharBoneco(jogo.Erros)
		}
		printLinhaGame("", ' ')
		printLinhaGame("", '-')
	}

	return ret
}
