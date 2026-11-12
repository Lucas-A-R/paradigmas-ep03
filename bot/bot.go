package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

// Inverte a string preservando caracteres acentuados/UTF-8
func reverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func main() {
	botNick := "Inversor"
	if len(os.Args) > 1 {
		botNick = os.Args[1]
	}

	conn, err := net.Dial("tcp", "localhost:4444")
	if err != nil {
		log.Fatalf("Bot não conseguiu conectar ao servidor: %v", err)
	}
	defer conn.Close()

	// Envia handshake identificando-se como Bot
	fmt.Fprintf(conn, "JOIN_BOT %s\n", botNick)

	log.Printf("Bot @%s conectado e aguardando mensagens...\n\n", botNick)

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()

		// Trata exclusivamente mensagens privadas recebidas
		if strings.Contains(line, " disse em privado: ") {
			parts := strings.SplitN(line, " disse em privado: ", 2)
			sender := parts[0] // Exemplo: "@Zezinho" ou "@Hack3r"
			receivedMsg := parts[1]

			responseMsg := reverseString(receivedMsg)

			// Impressão obrigatória no console do Bot
			fmt.Printf("Mensagem de: %s\n", sender)
			fmt.Printf("Recebi: %s\n", receivedMsg)
			fmt.Printf("Resposta: %s\n\n", responseMsg)

			// Devolve a resposta em privado para quem enviou
			fmt.Fprintf(conn, "PRIVMSG %s %s\n", sender, responseMsg)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Erro na leitura da conexão do bot: %v", err)
	}
}
