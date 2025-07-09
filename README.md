# ForcaGame
Repositório para armazenamento do código desenvolvido durante atividade avaliativa da cadeira de Programação Concorrente e Distribuída

# Para a versão gRPC

## Ferramentas Necessarias

- protobuf
Pode ser obtido em https://github.com/protocolbuffers/protobuf/releases/tag/v31.1 
- go 
Pode ser obtido em https://go.dev/
- gRPC
Pode ser instalado a partir do tutorial https://grpc.io/docs/languages/go/quickstart/


## Para rodar o programa
- Na pasta gRPC execute
    - go run servidor/jogo.go servidor/server.go servidor/util.go servidor/main.go para rodar o servidor
    - go run cliente/jogo.go cliente/util.go cliente/main.go para rodar o cliente
