package main

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
