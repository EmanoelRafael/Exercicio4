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

- [Go](https://go.dev/dl/) instalado (v1.16 ou superior)
- [Docker](https://docs.docker.com/get-docker/) instalado (para rodar o RabbitMQ)

---

## ⚙️ Como rodar o projeto

### 1. Clone este repositório

```bash
git clone https://github.com/seu-usuario/jogo-forca.git
cd jogo-forca
```

### 2. Instale as dependências do Go

```bash
go mod tidy
```

### 3. Inicie o RabbitMQ com Docker

```bash
sudo docker run -d --hostname rabbit --name rabbitmq \
  -p 5672:5672 -p 15672:15672 rabbitmq:3-management
```

> 💡 A interface web do RabbitMQ estará disponível em [http://localhost:15672](http://localhost:15672)  
> Login: `guest` | Senha: `guest`

---

## 🚀 Como jogar

### 🧠 Etapas:

1. Abra **três terminais**: um para o servidor e dois para os clientes.

---

### 📡 Terminal 1 — Servidor

```bash
go run servidor.go
```

---

### 🧑 Terminal 2 — Jogador 1

```bash
go run cliente.go
# Quando solicitado, digite: cliente1
```

---

### 🧑 Terminal 3 — Jogador 2

```bash
go run cliente.go
# Quando solicitado, digite: cliente2
```

---

### 🎮 Jogando

- O servidor define a palavra secreta.
- Os jogadores se alternam enviando letras.
- O jogo termina quando:
  - A palavra for adivinhada corretamente
  - O número máximo de tentativas (6) for atingido

> O servidor **garante a ordem dos turnos**. Palpites fora de hora são ignorados.

---

## 📝 Estrutura do Projeto

```
jogo-forca/
├── cliente.go         # Código do cliente
├── servidor.go        # Código do servidor
├── go.mod             # Dependências do projeto
├── go.sum             # Sums de verificação
└── README.md          # Este arquivo
```

---

## 🧪 Exemplo de Execução

```text
=== FORCA ===
Palavra: __a___
Tentativas: 0/6
Letras usadas: [a]
Sua vez! Digite uma letra:
```

---

## ✨ Possíveis melhorias

- Gerar palavras aleatórias a cada rodada
- Permitir reinício automático do jogo
- Interface gráfica (opcional)
- Suporte a mais jogadores/salas

---

## 👩‍💻 Autoria

Projeto desenvolvido por **Rebecca Lima Sousa** como atividade acadêmica.  
Implementado em Golang com uso de RabbitMQ via Docker.

---

## 📄 Licença

Uso livre para fins educacionais.
