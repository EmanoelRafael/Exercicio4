package main

import "math/rand"

const letras = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const NAO_INICIADO = 0
const PENDENTE_JOGADORES = 1
const EM_CURSO = 2
const FINALIZADO = 3

var IDX_PALAVRA = 0
var palavras = []string{"Desasnado", "Filantropo", "Idiossincrasia", "Juvenilizante", "Odiento", "Quimera", "Verossimilhança", "Xaropear", "Yanomami", "Vicissitude"}

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

func letrasErradasSlice(m map[string]bool) string {
	letras := ""
	for l := range m {
		letras += l
	}
	return letras
}
