package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Digite seu apelido: ")
	nickInput, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf("Erro ao ler apelido: %v", err)
	}

	nick := strings.TrimSpace(nickInput)
	if nick == "" {
		nick = "Anonimo"
	}

	conn, err := net.Dial("tcp", "localhost:4444")
	if err != nil {
		log.Fatalf("Não foi possível conectar ao servidor em localhost:4444: %v", err)
	}
	defer conn.Close()

	// Envia handshake inicial de registro do usuário
	fmt.Fprintf(conn, "JOIN_USER %s\n", nick)

	fmt.Println("\n--- Conectado ao Chat ---")
	fmt.Println("Comandos disponíveis:")
	fmt.Println("  \\msg <mensagem>             - Mensagem pública")
	fmt.Println("  \\msg @<apelido> <mensagem>  - Mensagem privada")
	fmt.Println("  \\changenick <novo_apelido>  - Trocar de apelido")
	fmt.Println("  \\exit                       - Sair do chat\n")

	done := make(chan struct{})

	// Gorrotina para escutar e exibir mensagens vindas do servidor
	go func() {
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
		log.Println("Conexão com o servidor encerrada.")
		close(done)
	}()

	// Loop principal para captura de entrada no terminal
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if line == "\\exit" {
			fmt.Fprintln(conn, "EXIT")
			break
		} else if strings.HasPrefix(line, "\\changenick ") {
			newNick := strings.TrimSpace(strings.TrimPrefix(line, "\\changenick "))
			if newNick != "" {
				fmt.Fprintf(conn, "NICK %s\n", newNick)
			}
		} else if strings.HasPrefix(line, "\\msg ") {
			rest := strings.TrimSpace(strings.TrimPrefix(line, "\\msg "))
			if strings.HasPrefix(rest, "@") {
				parts := strings.SplitN(rest, " ", 2)
				target := parts[0]
				msgText := ""
				if len(parts) == 2 {
					msgText = parts[1]
				}
				fmt.Fprintf(conn, "PRIVMSG %s %s\n", target, msgText)
			} else {
				fmt.Fprintf(conn, "PUBMSG %s\n", rest)
			}
		} else {
			// Envia como mensagem pública caso o usuário digite direto
			fmt.Fprintf(conn, "PUBMSG %s\n", line)
		}
	}

	<-done
}
