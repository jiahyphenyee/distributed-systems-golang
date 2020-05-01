// 1003017 PSET 1
package main

import (
	"fmt"
	"math/rand"
	"time"
)

var cv bool
var cvLog []map[string][]uint

// History ...
type History struct {
	sender   string
	receiver string
	lamport  []uint
	pid      uint
}

// func (h *History) lower(other History) bool {
// 	// if logical clock is lower
// 	// OR if same clock value, look at pid
// 	result := h.lamport < other.lamport || (h.lamport == other.lamport && h.pid < other.pid)
// 	return result
// }

// Msg ...
type Msg struct {
	senderID uint
	sender   string
	lamport  []uint
}

// Server ...
type Server struct {
	Clients  []*Client
	Clock    []uint
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

			s.Clock = syncClock(msg.lamport, s.Clock, msg.sender, "server")
			fmt.Println("Server's clock updated to", s.Clock)
			s.histChan <- History{msg.sender, "server", s.Clock, 0}
			s.Clock[0]++

			// broadcast message to each registered client
			send := Msg{0, "Server", s.Clock}

			for _, c := range s.Clients {
				if c.Name != msg.sender {
					fmt.Println("Broadcasted to ", c.Name)
					c.Channel <- send
					s.Clock[0]++
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
	Clock    []uint
	Channel  chan Msg
	histChan chan History
}

// New ... new client and registration
func New(s Server, id uint, histChan chan History, numClient int) *Client {
	c := &Client{"", id, make([]uint, numClient+1), make(chan Msg), histChan}
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

	send := Msg{c.ID, c.Name, c.Clock}
	ticker := time.NewTicker(5 * time.Second)

	go func() {
		for range ticker.C {
			fmt.Printf("Client %d sending message to Server", c.ID)
			serverChan <- send
			c.Clock[c.ID]++
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
		c.Clock = syncClock(msg.lamport, c.Clock, msg.sender, c.Name)
		c.histChan <- History{msg.sender, c.Name, c.Clock, c.ID}
		c.Clock[c.ID]++
		fmt.Println("Client", c.ID, "clock updated to", c.Clock)
	}
}

func syncClock(clkSender []uint, clkReceiver []uint, sender string, receiver string) []uint {
	before := false
	after := false

	for i := 0; i < len(clkSender); i++ {
		if clkSender[i] > clkReceiver[i] {
			before = true
			clkReceiver[i] = clkSender[i]
		} else if clkSender[i] < clkReceiver[i] {
			after = true
		}
	}

	if !before && after {
		info := map[string][]uint{
			"sender":   clkSender,
			"receiver": clkReceiver,
		}
		cv = true
		cvLog = append(cvLog, info)
		fmt.Println("Potential Causality Violation where receiver", receiver, "clock is", clkReceiver, "and sender", sender, "clock is", clkSender)
	}

	return clkReceiver
}

func main() {

	fmt.Println("\n\n===========")
	fmt.Println("After nodes start running, press 'Enter' to stop processes and check for potential causality violation")
	fmt.Println("===========")
	time.Sleep(time.Duration(1 * time.Second))

	cv = false
	stopChan := make(chan bool)
	var numClient = 3
	var clients []*Client

	var hist []History
	histChan := make(chan History, numClient+3)
	go func() {
		for {
			item := <-histChan
			hist = append(hist, item)
		}
	}()

	server := Server{[]*Client{}, make([]uint, numClient+1), make(chan Msg), histChan, 0}

	for i := 1; i < numClient+1; i++ {
		fmt.Println("Creating client", i)
		c := New(server, uint(i), histChan, numClient)
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

	if cv == false {
		fmt.Println("There are no potential Causality Violations")
	} else {
		fmt.Println("Potential Causality Violation occurred:")
		for _, i := range cvLog {
			fmt.Println("Potential Causality Violation where receiver clock is", i["receiver"], "and sender clock is", i["sender"])
		}
	}

}
