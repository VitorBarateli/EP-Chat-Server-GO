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
	fmt.Println("Tentando conectar ao servidor...")
	conn, err := net.Dial("tcp", "localhost:3000")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	fmt.Println("Conectado ao servidor")

	// Receber mensagens do servidor
	go func() {
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}()

	// Enviar mensagens para o servidor
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	nickname := scanner.Text()
	nickname = strings.TrimSpace(nickname)
	// Enviar apelido para o servidor
	fmt.Fprintf(conn, nickname+"\n")

	for scanner.Scan() {
		fmt.Print("> ")
		text := scanner.Text()

		if text == "\\exit" {
			fmt.Println("Desconectando do servidor...")
			conn.Close()
			break
		}

		if strings.HasPrefix(text, "\\msg") || strings.HasPrefix(text, "\\changenick") {
			fmt.Fprintln(conn, text) // Enviar a mensagem para o servidor
		} else {
			// Informar ao usuário que deve usar os comandos válidos
			fmt.Println("Erro: Você deve começar a mensagem com '\\msg ' ou '\\changenick ' para enviar.")
		}
	}
}
