package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
)

// Client representa a conexão de um usuário ou bot no servidor
type Client struct {
	conn  net.Conn
	ch    chan string
	nick  string
	isBot bool
}

type PubMessage struct {
	sender *Client
	text   string
}

type PrivMessage struct {
	sender *Client
	target string
	text   string
}

type NickChange struct {
	client  *Client
	newNick string
}

var (
	entering     = make(chan *Client)
	leaving      = make(chan *Client)
	pubMessages  = make(chan PubMessage)
	privMessages = make(chan PrivMessage)
	nickChanges  = make(chan NickChange)
)

func broadcaster() {
	clients := make(map[string]*Client)

	for {
		select {
		case cli := <-entering:
			// Garante nickname único caso haja duplicata no registro
			baseNick := cli.nick
			counter := 1
			for {
				if _, exists := clients[cli.nick]; !exists {
					break
				}
				cli.nick = fmt.Sprintf("%s_%d", baseNick, counter)
				counter++
			}

			clients[cli.nick] = cli

			var announce string
			if cli.isBot {
				announce = fmt.Sprintf("Bot @%s acabou de entrar", cli.nick)
			} else {
				announce = fmt.Sprintf("Usuário @%s acabou de entrar", cli.nick)
			}

			// Exibe no console do servidor
			fmt.Println(announce)

			// Broadcast apenas para usuários humanos (bots não recebem msgs públicas/sistema)
			for _, c := range clients {
				if !c.isBot {
					c.ch <- announce
				}
			}

		case cli := <-leaving:
			if existing, ok := clients[cli.nick]; ok && existing == cli {
				delete(clients, cli.nick)
				close(cli.ch)

				var announce string
				if cli.isBot {
					announce = fmt.Sprintf("Bot @%s saiu", cli.nick)
				} else {
					announce = fmt.Sprintf("Usuário @%s saiu", cli.nick)
				}

				// Exibe no console do servidor
				fmt.Println(announce)

				// Broadcast para todos os usuários humanos restantes
				for _, c := range clients {
					if !c.isBot {
						c.ch <- announce
					}
				}
			}

		case pm := <-pubMessages:
			formattedMsg := fmt.Sprintf("@%s disse: %s", pm.sender.nick, pm.text)
			// Mensagens públicas são enviadas apenas para usuários humanos (não bots)
			for _, c := range clients {
				if !c.isBot {
					c.ch <- formattedMsg
				}
			}

		case priv := <-privMessages:
			targetNick := strings.TrimPrefix(priv.target, "@")
			formattedMsg := fmt.Sprintf("@%s disse em privado: %s", priv.sender.nick, priv.text)

			// Exibe no console do servidor
			fmt.Println(formattedMsg)

			// Envia exclusivamente para o destinatário
			if recipient, ok := clients[targetNick]; ok {
				recipient.ch <- formattedMsg
			} else {
				if !priv.sender.isBot {
					priv.sender.ch <- fmt.Sprintf("Sistema: Usuário ou bot @%s não foi encontrado.", targetNick)
				}
			}

		case nc := <-nickChanges:
			oldNick := nc.client.nick
			newNick := strings.TrimPrefix(strings.TrimSpace(nc.newNick), "@")

			if newNick == "" {
				nc.client.ch <- "Sistema: Apelido inválido."
				continue
			}

			if _, exists := clients[newNick]; exists {
				nc.client.ch <- fmt.Sprintf("Sistema: Apelido @%s já está em uso.", newNick)
				continue
			}

			delete(clients, oldNick)
			nc.client.nick = newNick
			clients[newNick] = nc.client

			announce := fmt.Sprintf("Usuário @%s agora é @%s", oldNick, newNick)

			// Exibe no console do servidor
			fmt.Println(announce)

			// Broadcast da troca de apelido para os usuários humanos
			for _, c := range clients {
				if !c.isBot {
					c.ch <- announce
				}
			}
		}
	}
}

func clientWriter(conn net.Conn, ch <-chan string) {
	for msg := range ch {
		fmt.Fprintln(conn, msg)
	}
}

func handleConn(conn net.Conn) {
	scanner := bufio.NewScanner(conn)
	if !scanner.Scan() {
		conn.Close()
		return
	}

	firstLine := scanner.Text()
	parts := strings.SplitN(firstLine, " ", 2)
	if len(parts) < 2 {
		conn.Close()
		return
	}

	cmd := parts[0]
	nick := strings.TrimPrefix(strings.TrimSpace(parts[1]), "@")

	isBot := false
	if cmd == "JOIN_BOT" {
		isBot = true
	} else if cmd != "JOIN_USER" {
		conn.Close()
		return
	}

	cli := &Client{
		conn:  conn,
		ch:    make(chan string, 10),
		nick:  nick,
		isBot: isBot,
	}

	go clientWriter(conn, cli.ch)

	entering <- cli

	defer func() {
		leaving <- cli
		conn.Close()
	}()

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		lineParts := strings.SplitN(line, " ", 3)
		action := lineParts[0]

		switch action {
		case "PUBMSG":
			if len(lineParts) >= 2 {
				text := strings.TrimPrefix(line, "PUBMSG ")
				pubMessages <- PubMessage{sender: cli, text: text}
			}
		case "PRIVMSG":
			if len(lineParts) >= 3 {
				target := lineParts[1]
				text := strings.TrimPrefix(line, "PRIVMSG "+target+" ")
				privMessages <- PrivMessage{sender: cli, target: target, text: text}
			}
		case "NICK":
			if len(lineParts) >= 2 {
				newNick := lineParts[1]
				nickChanges <- NickChange{client: cli, newNick: newNick}
			}
		case "EXIT":
			return
		}
	}
}

func main() {
	fmt.Println("Iniciando servidor de chat...")
	listener, err := net.Listen("tcp", "localhost:4444")
	if err != nil {
		log.Fatalf("Erro ao iniciar servidor na porta 4444: %v", err)
	}
	defer listener.Close()

	fmt.Println("Servidor escutando em localhost:4444")

	go broadcaster()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Erro ao aceitar conexão: %v", err)
			continue
		}
		go handleConn(conn)
	}
}
