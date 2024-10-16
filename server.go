package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
)

type client struct {
	name string
	ch   chan<- string
}

var (
	messages = make(chan string)     // canal para mensagens públicas
	entering = make(chan client)     // canal para novos clientes
	leaving  = make(chan client)     // canal para clientes que saem
	clients  = make(map[client]bool) // todos os clientes conectados
	mu       sync.Mutex              // mutex para proteger o mapa de clientes
)

func broadcaster() {
	for {
		select {
		case msg := <-messages:
			fmt.Println("Broadcasting mensagem:", msg)
			for cli := range clients {
				cli.ch <- msg // broadcast para todos os clientes
			}
		case cli := <-entering:
			mu.Lock()
			clients[cli] = true
			fmt.Println("Cliente entrou: ", cli.name) // Adicionado para verificar nomes
			mu.Unlock()
		case cli := <-leaving:
			mu.Lock()
			delete(clients, cli)
			close(cli.ch)
			fmt.Println("Cliente saiu: ", cli.name) // Adicionado para verificar saídas
			mu.Unlock()
		}
	}
}

func handleConn(conn net.Conn) {
	ch := make(chan string)
	go clientWriter(conn, ch)

	// Pedir apelido ao cliente
	fmt.Fprintln(conn, "Escolha um apelido: ") // Enviando a mensagem ao cliente
	fmt.Println("Mensagem de apelido enviada ao cliente")

	input := bufio.NewScanner(conn)
	if input.Scan() {
		nickname := input.Text()
		fmt.Println("Apelido Escolhido:", nickname)
		cli := client{name: nickname, ch: ch}

		// Anunciar entrada do usuário
		messages <- fmt.Sprintf("Usuário @%s acabou de entrar", cli.name)
		entering <- cli

		// Continuar processando mensagens do cliente
		for input.Scan() {
			msg := input.Text()
			fmt.Println("Recebido de", cli.name+":", msg)

			if strings.HasPrefix(msg, "\\changenick") {
				parts := strings.SplitN(msg, " ", 2)
				if len(parts) > 1 {
					newNick := parts[1]
					messages <- fmt.Sprintf("Usuário @%s agora é @%s", cli.name, newNick)
					cli.name = newNick
				}
			} else if strings.HasPrefix(msg, "\\msg @") {
				parts := strings.SplitN(msg, " ", 3)
				if len(parts) > 2 {
					// Remover o caractere "@" do nome do destinatário para comparação
					targetNick := strings.TrimPrefix(parts[1], "@")
					privateMsg := parts[2]
					sendPrivateMessage(cli, targetNick, privateMsg)
				}
			} else if strings.HasPrefix(msg, "\\msg ") {
				// Mensagem pública
				publicMsg := msg[5:] // Remover o prefixo "\msg "
				messages <- fmt.Sprintf("@%s disse: %s", cli.name, publicMsg)
			} else {
				messages <- fmt.Sprintf("@%s disse: %s", cli.name, msg)
			}
		}

		leaving <- cli
		messages <- fmt.Sprintf("Usuário @%s saiu", cli.name)
		conn.Close()
	}
}

func sendPrivateMessage(sender client, targetNick, privateMsg string) {
	mu.Lock()
	defer mu.Unlock()

	// Tornar o apelido do destinatário minúsculo para comparação
	targetNick = strings.ToLower(targetNick)
	fmt.Println("Procurando destinatário:", targetNick) // Debug para verificar o destinatário

	for cli := range clients {
		fmt.Println("Verificando cliente:", cli.name)
		if strings.ToLower(cli.name) == targetNick { // Comparar apelidos sem distinção de maiúsculas/minúsculas
			cli.ch <- fmt.Sprintf("Mensagem privada de @%s: %s", sender.name, privateMsg)
			fmt.Println("Mensagem privada enviada para", cli.name)
			return
		}
	}
	// Enviar mensagem de erro de volta ao remetente se o destinatário não for encontrado
	sender.ch <- fmt.Sprintf("Erro: Destinatário @%s não encontrado.", targetNick)
	fmt.Println("Destinatário não encontrado:", targetNick)
}

func clientWriter(conn net.Conn, ch <-chan string) {
	for msg := range ch {
		fmt.Fprintln(conn, msg)
	}
}

func main() {
	fmt.Println("Servidor Iniciado!")
	listener, err := net.Listen("tcp", "localhost:3000")
	if err != nil {
		log.Fatal(err)
	}
	go broadcaster()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err)
			continue
		}
		fmt.Println("Nova conexão recebida")
		go handleConn(conn)
	}
}
