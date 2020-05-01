package main

import (
	"fmt"
	"math/rand"
	"sort"
	"time"
)

// History ...
type History struct {
	sender   string
	receiver string
	lamport  uint
	pid      uint
}

func (h *History) lower(other History) bool {
	// if logical clock is lower
	// OR if same clock value, look at pid
	result := h.lamport < other.lamport || (h.lamport == other.lamport && h.pid < other.pid)
	return result
}

// Msg ...
type Msg struct {
	sender  string
	lamport uint
}

// Server ...
type Server struct {
	Clients  []*Client
	Clock    uint
	Channel  chan Msg
	histChan chan History
	pID      uint
}

func (s *Server) start(stopChan chan bool) {
	for {
		select {
		case <-stopChan:
			return

		default:
			// on receive by Server
			msg := <-s.Channel
			fmt.Println("Server received message from", msg.sender)

			s.Clock = syncClock(msg.lamport, s.Clock)
			fmt.Println("Server's clock updated to", s.Clock)
			s.histChan <- History{msg.sender, "server", s.Clock, 0}

			// broadcast message to each registered client
			send := Msg{"Server", s.Clock}

			for _, c := range s.Clients {
				if c.Name != msg.sender {
					fmt.Println("Broadcasted to ", c.Name)
					c.Channel <- send
					s.Clock++
					//sleep sporadically
					amt := time.Duration(rand.Intn(20))
					time.Sleep(time.Millisecond * amt)
				}
			}
		}
	}
}

// Client ...
type Client struct {
	Name     string
	ID       uint
	Clock    uint
	Channel  chan Msg
	histChan chan History
}

// New ... new client and registration
func New(s Server, id uint, histChan chan History) *Client {
	c := &Client{"", id, 0, make(chan Msg), histChan}
	c.Name = "Client" + fmt.Sprint(id)
	c.register(s)

	return c
}

func (c *Client) register(s Server) {
	s.Clients = append(s.Clients, c)
	fmt.Println("Server registered client ", c.ID)
}

// client starts - periodically sends messages and handles msg on receipt
func (c *Client) start(serverChan chan Msg) {

	send := Msg{c.Name, c.Clock}
	ticker := time.NewTicker(5 * time.Second)

	go func() {
		for range ticker.C {
			fmt.Printf("Client %d sending message to Server", c.ID)
			serverChan <- send
			c.Clock++
		}
	}()

	go c.receive()
}

// handle incoming messages
func (c *Client) receive() {
	for {
		msg := <-c.Channel
		fmt.Println("Client", c.ID, "received message from", msg.sender)
		// a read from ch has occurred
		c.Clock = syncClock(msg.lamport, c.Clock)
		c.histChan <- History{msg.sender, c.Name, c.Clock, c.ID}
		fmt.Println("Client", c.ID, "clock updated to", c.Clock)
	}
}

func syncClock(clkSender uint, clkReceiver uint) uint {
	if clkReceiver <= clkSender {
		return clkSender + uint(1)
	}
	return clkReceiver + uint(1)
}

func main() {

	clock := uint(0)
	checkChan := make(chan int)
	var numClient = 3
	var clients []*Client

	fmt.Println("\n\n===========")
	fmt.Println("Start Central Server")
	fmt.Println("===========")
	time.Sleep(time.Duration(1 * time.Second))

	var hist []History
	histChan := make(chan History, numClient+3)
	go func() {
		for {
			item := <-histChan
			hist = append(hist, item)
		}
	}()

	server := Server{[]*Client{}, 0, make(chan Msg), histChan, 0}

	for i := 1; i < numClient+1; i++ {
		fmt.Println("Creating client", i)
		c := New(server, uint(i), histChan)
		clients = append(clients, c)
	}

	server.Clients = clients

	for _, c := range clients {
		go c.start(server.Channel)
	}

	go server.start(stopChan)

	var s string
	fmt.Scanln(&s)

	// stop running program
	stopChan <- true
	// get total order
	sort.SliceStable(hist,
		func(i, j int) bool {
			return hist[i].lower(hist[j])
		})

	fmt.Println("\n\n===========")
	fmt.Println("Total Order")
	fmt.Println("===========")
	for _, h := range hist {
		fmt.Println("Clock", h.lamport, ":", h.sender, "sent message to", h.receiver)
	}

}
