package main

type Jogo struct {
	Codigo         string
	Palavra        string
	PalavraVisivel []rune
	Jogadores      []string
	Finalizado     bool
}
