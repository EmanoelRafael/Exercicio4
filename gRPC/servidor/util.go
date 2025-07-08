package main

import "math/rand"

const letras = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func GerarCodigoJogo() string {
	b := make([]byte, 6)
	for i := range b {
		b[i] = letras[rand.Intn(len(letras))]
	}
	return string(b)
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

	// Gira até encontrar o próximo não eliminado
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

func letrasErradasSlice(m map[string]bool) []string {
	letras := []string{}
	for l := range m {
		letras = append(letras, l)
	}
	return letras
}
