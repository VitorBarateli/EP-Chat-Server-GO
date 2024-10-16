package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
)

func main() {
	fmt.Println("Tentando conectar ao servidor como bot...")

	conn, err := net.Dial("tcp", "localhost:3000")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	fmt.Println("Bot conectado ao servidor")

	// Receber a mensagem de "Escolha um apelido" e ignorar
	scanner := bufio.NewScanner(conn)
	if scanner.Scan() {
		fmt.Println("Bot recebeu:", scanner.Text())
	}

	// Enviar o apelido do bot para o servidor
	botName := "BotInversor"
	fmt.Fprintln(conn, botName)
	fmt.Println("Bot nome enviado:", botName)

	// Continuar recebendo mensagens e respondendo
	go func() {
		for scanner.Scan() {
			msg := scanner.Text()
			fmt.Println("Bot recebeu mensagem:", msg)

			// Verificar se a mensagem é direcionada ao bot (privada)
			if strings.HasPrefix(msg, "Mensagem privada de @") {
				fmt.Println("Mensagem privada detectada, processando...")

				// Extrair o texto da mensagem privada e inverter
				parts := strings.SplitN(msg, ": ", 2)
				if len(parts) > 1 {
					originalMsg := parts[1]
					invertedMsg := reverseString(originalMsg)
					response := fmt.Sprintf("Resposta do bot: %s", invertedMsg)
					fmt.Fprintln(conn, response)
					fmt.Println("Bot enviou resposta:", response)
				}
			}
		}
	}()

	// Manter o bot ativo
	select {}
}

// Função para inverter uma string
func reverseString(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}
