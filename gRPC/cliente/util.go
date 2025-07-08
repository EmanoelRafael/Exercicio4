package main

import (
	"fmt"
	"strings"
)

const widthGame = 44

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

func printGame(msg string) {
	if tipoMenu == 1 {
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
	} else if tipoMenu == 2 {
		printLinhaGame("FORCA GAME", '-')
		printLinhaGame("", ' ')
		printLinhaGame("AGUARDANDO A ENTRADA DOS JOGADORES", ' ')
		printLinhaGame(" ", ' ')
		printLinhaGame(" ", ' ')
		printLinhaGame(" ", ' ')
		printLinhaGame("", ' ')
		printLinhaGame("", '-')
	} else if tipoMenu == 3 {
		printLinhaGame("FORCA GAME", '-')
		printLinhaGame("", ' ')
		printLinhaGame("INSIRA O CODIGO DO JOGO", ' ')
		printLinhaGame(" ", ' ')
		printLinhaGame(" ", ' ')
		printLinhaGame(" ", ' ')
		printLinhaGame("", ' ')
		printLinhaGame("", '-')
		fmt.Print("=> ")
	} else if tipoMenu == 4 {
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
		}
	} else if tipoMenu == 5 {
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
			printLinhaGame("BOM JOGO!", ' ')
		}
		printLinhaGame("", ' ')
		printLinhaGame("", '-')
	}

}
