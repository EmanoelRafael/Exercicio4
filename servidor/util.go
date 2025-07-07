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
