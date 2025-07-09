# 🕹️ Jogo da Forca Multiplayer (Go + RabbitMQ)

Este é um projeto de **jogo da forca baseado em texto**, com **modo multiplayer para 2 jogadores**, implementado em **Go (Golang)** com arquitetura **cliente-servidor** usando **RabbitMQ** para comunicação.

---

## 🎯 Funcionalidades

- Jogo textual baseado em turnos
- Comunicação assíncrona entre jogadores via RabbitMQ
- Sincronização do estado do jogo para ambos os jogadores
- Controle de turno garantido pelo servidor
- Encerramento automático por vitória ou derrota

---

## 📦 Requisitos

- Go instalado (v1.16+ recomendado)
- Docker instalado (para rodar o RabbitMQ)

---

## ⚙️ Como rodar o projeto

### 1. Clone este repositório

```bash
git clone https://github.com/seu-usuario/jogo-forca.git
cd jogo-forca
