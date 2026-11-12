# 🐹 Servidor de Chat Concorrente em Go (TCP)

Este repositório contém a implementação do **EP03 da disciplina de Paradigmas de Programação**. Trata-se de um sistema de chat completo baseado em arquitetura Cliente-Servidor via sockets TCP, desenvolvido puramente em Go (Golang) explorando seus primitivos nativos de concorrência.

## 📌 Funcionalidades e Arquitetura

O sistema é composto por três programas autônomos que se comunicam através da porta `4444`:

*   **Servidor (`server.go`):** Atua como o nó central de roteamento. Ele aceita conexões TCP, diferencia humanos de robôs (bots) logo no aperto de mão (*handshake*) e gerencia o envio de mensagens em broadcast ou unicast (privadas).
*   **Cliente (`client.go`):** Interface interativa via terminal (CLI) que traduz os comandos de texto do usuário para o protocolo de rede.
*   **Bot Reativo (`bot.go`):** Um agente autônomo que se conecta ao chat. Para evitar *loops* e poluição visual, o servidor possui uma barreira lógica (`isBot`) que garante que os bots recebam **apenas** mensagens privadas. O bot implementado (`@Inversor`) lê a mensagem, inverte os caracteres (respeitando a codificação UTF-8 via `[]rune`) e devolve ao remetente.

## 🚀 Modelo de Concorrência (Sem Mutexes)

O grande diferencial deste projeto é o tratamento de concorrência. Seguindo o provérbio do Go: *"Não se comunique compartilhando memória; compartilhe memória se comunicando"*:
*   **Goroutines:** Cada cliente roda em goroutines dedicadas de I/O assíncrono.
*   **Channels:** A lista global de clientes não utiliza travas de memória (`sync.Mutex`). Todo o controle de estado é orquestrado por uma única goroutine central (`broadcaster`) que consome canais (`channels`) de eventos de entrada, saída, mensagens públicas e privadas, eliminando *Race Conditions*.

## ⚙️ Como Executar

Certifique-se de ter o [Go instalado](https://go.dev/dl/). 
Abra **4 abas de terminal** e execute na seguinte ordem:

**1. Suba o Servidor (Terminal 1):**
```bash
go run ./server/server.go
# Ou se os arquivos estiverem na raiz: go run server.go
```
**2. Conecte o Bot Inversor (Terminal 2):Bashgo run ./bot/bot.go**

```bash
go run ./bot/bot.go
```


# Ou se estiver na raiz: go run bot.go
**3. Conecte os Clientes (Terminais 3 e 4):Bashgo run ./client/client.go**

```bash
go run ./client/client.go
```

# Ou se estiver na raiz: go run client.go

*(O terminal pedirá para você digitar um apelido ao conectar).*

## 📖 Protocolo e Comandos do Cliente

Uma vez conectado, o cliente pode usar os seguintes comandos no terminal:

| Comando no CLI | Protocolo de Rede (TCP) | Descrição |
| :--- | :--- | :--- |
| `\msg <texto>` | `PUBMSG <texto>` | Envia uma mensagem pública para todos. |
| `\msg @<apelido> <texto>` | `PRIVMSG <apelido> <texto>` | Envia mensagem privada (ex: para o bot). |
| `\changenick <novo>` | `NICK <novo>` | Troca o seu apelido em tempo real. |
| `\exit` | `EXIT` | Desconecta do servidor de forma segura. |


### Próximos passos para subir para o GitHub

Se você já criou o repositório em branco lá no site do GitHub, abra o terminal na pasta do projeto no VS Code e rode estes comandos para enviar seu código e esse README:

```bash
git init
git add .
git commit -m "Entrega EP03: Servidor de Chat Concorrente em Go"
git branch -M main
git remote add origin https://github.com/SEU_USUARIO/NOME_DO_REPOSITORIO.git
git push -u origin main
