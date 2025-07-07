# Enunciado

Implemente um jogo multiplayer baseado em texto, como Forca, utilizando uma arquitetura cliente-servidor, 
e implementado com o gRPC e com o RabbitMQ, em que o servidor é responsável por coordenar o jogo, gerenciar 
o estado compartilhado e garantir a sincronização entre os jogadores. Os clientes se conectam ao servidor 
para realizar ações no jogo, como enviar palpites, posicionar peças ou receber atualizações de estado, sempre 
respeitando a ordem dos turnos. O servidor deve ser capaz de controlar partidas entre dois, validar as jogadas 
recebidas e enviar respostas apropriadas de acordo com as regras do jogo. Este exercício não precisa realizar 
avaliação de desempenho [Equipe 10]

## Entregáveis

Gerar um único arquivo (e.g., zip) com o (1) Código Go de todas as versões, e (2) slides ( pdf ) com a 
apresentação de detalhes da aplicaçãão cliente-servidor implementada ( Não pode ser link ).