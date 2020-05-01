package main

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"time"
)

func (m *Msg) lower(other *Msg) bool {
	// if logical clock is lower
	// OR if same clock value, look at pid
	result := m.lamport < other.lamport || (m.lamport == other.lamport && m.senderID < other.senderID)
	return result
}

// Msg ...
type Msg struct {
	senderID uint
	lamport  uint
	Type     string
}

// Server ...
type Server struct {
	Nodes   map[uint]*Client
	Clock   uint
	Channel chan Msg
	Queue   []*Msg
	inCS    bool
}

func (s *Server) start() {
	for {
		select {
		case msg := <-s.Channel:
			switch msg.Type {
			case "Request":
				s.serverReceive(msg)
				fmt.Println("Server received message from Node", msg.senderID)

			case "Release":
				s.Clock = syncClock(msg.lamport, s.Clock)
				// if s.Queue[0].senderID != msg.senderID {
				// 	fmt.Println("Head is Node", s.Queue[0].senderID)
				// 	fmt.Println("But releasing Node", msg.senderID)
				// 	panic(fmt.Sprintln("Removing wrong request from queue"))
				// }

				for ind, req := range s.Queue {
					if req.senderID == msg.senderID {
						copy(s.Queue[ind:], s.Queue[ind+1:])
						s.Queue = s.Queue[:len(s.Queue)-1]
					}
				}

				if len(s.Queue) > 0 {
					go s.send("OK", s.Nodes[s.Queue[0].senderID].Channel)
				} else {
					fmt.Println("Completed all CS entry requests")
				}
			}

		}
	}
}

func (s *Server) serverReceive(msg Msg) {
	s.Clock = syncClock(msg.lamport, s.Clock)
	fmt.Println("Server's clock updated to", s.Clock)

	switch msg.Type {
	case "Request":
		s.Queue = append(s.Queue, &msg)
		fmt.Println("Server added node", msg.senderID)

		sort.SliceStable(s.Queue,
			func(i, j int) bool {
				return s.Queue[i].lower(s.Queue[j])
			})

		// grant access if it is head
		if s.Queue[0].senderID == msg.senderID {
			go func() {
				for s.inCS {
				}
				go s.send("OK", s.Nodes[msg.senderID].Channel)
			}()
		}
	}
}

func (s *Server) send(msgType string, receiver chan Msg) {
	s.Clock++
	msg := Msg{999, s.Clock, msgType}

	switch msgType {
	case "OK":
		s.inCS = true
	case "Release":
		s.inCS = false
	}

	receiver <- msg
}

// Client ...
type Client struct {
	ID         uint
	Clock      uint
	Channel    chan Msg
	reportChan chan int
	Server     Server
}

// handle incoming messages
func (c *Client) listen() {
	for {
		select {
		case msg := <-c.Channel:
			fmt.Println("Client", c.ID, "received OK from Server")
			c.Clock = syncClock(msg.lamport, c.Clock)

			fmt.Println("\n========== Node", c.ID, "ENTERING CS ========")
			time.Sleep(time.Duration(2 * time.Second))
			fmt.Println("\n========== Node", c.ID, "EXITING CS ========")
			c.reportChan <- 1

			c.send("Release")
		}

	}
}

// request to enter CS
func (c *Client) send(msgType string) {
	c.Clock++
	msg := Msg{c.ID, c.Clock, msgType}

	switch msgType {
	case "Request":
		fmt.Println("Node", c.ID, "--------REQ-------> Server")
		c.Server.Channel <- msg

	case "Release":
		fmt.Println("Node", c.ID, "--------RELEASE-------> Server")
		c.Server.Channel <- msg
	}
}

func syncClock(clkSender uint, clkReceiver uint) uint {
	if clkReceiver <= clkSender {
		return clkSender + uint(1)
	}
	return clkReceiver + uint(1)
}

func main() {
	numRequests, _ := strconv.Atoi(os.Args[1])
	checkChan := make(chan int)
	var numNodes = 10
	var nodes []*Client

	server := Server{make(map[uint]*Client), 0, make(chan Msg), []*Msg{}, false}

	// create nodes
	for i := 0; i < numNodes; i++ {
		fmt.Println("Creating client", i)
		c := &Client{uint(i), 0, make(chan Msg), checkChan, server}
		server.Nodes[uint(i)] = c
		nodes = append(nodes, c)
	}

	go server.start()

	// start clients
	for _, c := range nodes {
		go c.listen()
	}

	fmt.Println("==========")
	fmt.Println("Starting Central Server Simulation. Number of requests:", numRequests)
	fmt.Println("==========")
	time.Sleep(time.Duration(2 * time.Second))

	start := time.Now()
	for i := 0; i < numRequests; i++ {
		go nodes[i].send("Request")
	}

	checkDone := 0
	go func() {
		for checkDone < numRequests {
			select {
			case <-checkChan:
				checkDone++
			}
		}

		end := time.Now()
		elapsed := end.Sub(start)
		fmt.Println("\n\n==========")
		fmt.Println("Time elapsed:", elapsed)
		fmt.Println("==========")

	}()

	var s string
	fmt.Scanln(&s)

}
