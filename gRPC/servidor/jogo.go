package main

type Jogo struct {
	Codigo         string
	Palavra        string
	PalavraVisivel []rune
	Jogadores      []string
	Erros          map[string]int
	DicasUsadas    map[string]bool
	LetrasErradas  map[string]bool
	JogadorDaVez   string
	VencedorID     string
	Status         int
	Eliminados     map[string]bool
}
