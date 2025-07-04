# Jogo da Forca

### Regras do jogo:
A ordem dos jogadores é definida aleatóriamente
Na sua vez o jogador pode:
1. Dar um palpite sobre uma letra
    - Caso a palavra contenha a letra a posição da letra na palavra é revelada.
    - Caso a palavra não contenha a letra uma parte do corpo do boneco do jogador é desenhada.
    - Caso o boneco do jogador for completado o jogador perde.
2. Dar um chute da palavra
    - Caso o chute esteja correto o jogador vence
    - Caso o chute esteja incorreto uma parte do corpo do boneco do jogador é desenhada.
    - Caso o boneco do jogador for completado o jogador perde
3. Pedir dica
    - Uma letra da palavra é revelada
    - Após a dica o jogador deve jogar novamente

Obs: O boneco de cada jogador tem 6 partes (ou seja, se um jogador errar 5 vezes perde)
Obs: Cada jogador tem direito a uma dica

## Sobre o Sistema

### Funcionalidades do Cliente:
- Criar um novo jogo
    - Criar um jogo com amigos gerando código para compartilhamento.
    - Criar um jogo solo
- Ingressar em um jogo
    - Por meio de código de compartilhamento
    - Ingressar em um jogo aleatório
- Enviar palpite sobre uma letra
- Enviar palpite sobre a palavra
- Pedir dica

### Funcionalidades do Servidor:
- Iniciar um novo jogo 
    - Escolher a palavra que será utilizada e
        - Iniciar um novo jogo e gerar um código para compartilhamento ou
        - Adicionar o jogador a um jogo aleatório ou
        - Iniciar um jogo solo
- Ingressar o jogador em um jogo por meio de código de compartilhamento
- Administrar um jogo
    - Definir e administrar a ordem de jogada
    - Obter palpite de uma letra ou da palavra
    - Obter a dica para um dado jogo
- Administrar a fila de jogos aleatórios
    - Criar um jogo aleatório sempre que tiver jogadores suficientes na fila

### Estruturas importantes do sistema

#### Jogo
Atributos:
- Palavra
- Jogadores (cada jogador com um boneco)

#### Jogador
Atributos:
- Identificador
- Jogo

#### Fila de jogos aleatórios
Quando um jogador solicita a entrada em um jogo aleatório ele entra na fila de jogos aleatórios e sai quando um novo jogo aleatório é criado (os jogos aleatórios são de 2 participantes)

### Stack Utilizada:
- gRPC
- RabbitMQ
- Go
- Protobuf