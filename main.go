package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"
)

// Сетевой адрес.
//
// Служба будет слушать запросы на всех IP-адресах
// компьютера на порту 12345.
// Например, 127.0.0.1:12345
const addr = "0.0.0.0:12345"

// Протокол сетевой службы.
const proto = "tcp4"

func main() {

	proverbs, err := getProverbs()
	if len(proverbs) == 0 || err != nil {
		log.Fatal("Не удалось получить список поговорок\n", err)
	}

	// Запуск сетевой службы по протоколу TCP
	// на порту 12345.
	listener, err := net.Listen(proto, addr)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	// Подключения обрабатываются в бесконечном цикле.
	// Иначе после обслуживания первого подключения сервер
	//завершит работу.
	for {
		// Принимаем подключение.
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}
		// Вызов обработчика подключения.
		go handleConn(conn, proverbs)
	}
}

func getProverbs() ([]string, error) {
	var clearProverb []string
	resp, err := http.Get("https://go-proverbs.github.io/")
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	//fmt.Println(string(body))

	re := regexp.MustCompile(`<h3><a[^>]*>([^<]+)</a></h3>`)
	matches := re.FindAllStringSubmatch(string(body), -1)
	for _, match := range matches {
		clearProverb = append(clearProverb, match[1])
	}

	return clearProverb, nil
}

// Обработчик. Вызывается для каждого соединения.
func handleConn(conn net.Conn, proverbs []string) {
	// Закрытие соединения.
	stopChan := make(chan bool)
	defer func() {
		conn.Close()
		stopChan <- true
	}()

	go sendProverbs(stopChan, proverbs, conn)

	// Чтение сообщения от клиента.
	for {

		reader := bufio.NewReader(conn)
		b, err := reader.ReadBytes('\n')
		if err != nil {
			log.Println(err)
			return
		}

		// Удаление символов конца строки.
		msg := strings.TrimSuffix(string(b), "\n")
		msg = strings.TrimSuffix(msg, "\r")
		// Если получили "time" - пишем время в соединение.

		if msg == "time" {
			conn.Write([]byte(time.Now().String() + "\n"))
			msg = ""
			return
		}

		if msg == "close" {
			stopChan <- true
			conn.Write([]byte("Connection closed\n"))
			conn.Close()
			return
		}
	}
}

func sendProverbs(stopChan chan bool, proverbs []string, conn net.Conn) {
	var mutex sync.Mutex
	for {

		select {
		case <-stopChan:
			// Сигнал для остановки выполнения функции
			return
		default:
			// Выполняем действия по отправке сообщения
			mutex.Lock()
			fmt.Println("Функция выполняется каждые 3 секунды.")
			conn.Write([]byte(proverbs[rand.Intn(len(proverbs)-1)] + "\n \r"))
			mutex.Unlock()
			time.Sleep(3 * time.Second)
		}
	}
}
