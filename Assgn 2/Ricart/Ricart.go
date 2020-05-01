package main

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Request ...
type Request struct {
	// for tracking own requests only
	Timestamp uint
	NodeID    int
	reqChan   chan Message
	replies   []int //nodeID
}

// Message ...
type Message struct {
	ID        int64
	NodeID    int
	Timestamp uint
	Type      string
	Req       *Request
}

// Node ...
type Node struct {
	ID         int
	Channel    chan Message
	Pals       []*Node
	Queue      map[int]Message
	Requests   []*Request
	hasReq     bool
	inCS       bool
	clock      uint
	reportChan chan int
}

// Run ...
func (n *Node) Run(allNodes []*Node) {
	// list other nodes
	for _, node := range allNodes {
		if node.ID != n.ID {
			n.Pals = append(n.Pals, node)
		}
	}
	fmt.Println("Node", n.ID, "has started")

	// set up channels
	for {
		select {
		case msg := <-n.Channel:
			n.onReceive(msg)
		default:
			continue
		}
	}
}

// send msgs
func (n *Node) send(msg Message, dest *Node) {

	go func() {
		select {
		case dest.Channel <- msg:
			return
		}
	}()
}

// create and send request
func (n *Node) request() {
	n.clock++

	req := Request{
		NodeID:    n.ID,
		Timestamp: n.clock,
		reqChan:   make(chan Message),
	}

	reqMsg := Message{
		ID:        int64(req.NodeID),
		NodeID:    n.ID,
		Timestamp: n.clock,
		Type:      "Request",
		Req:       &req,
	}

	fmt.Println("\nTimestamp of Req from Node", n.ID, ":", n.clock)

	n.hasReq = true
	n.Requests = append(n.Requests, &req)

	for _, pal := range n.Pals {
		fmt.Printf("Node %d -------REQ-------> node %d\n", n.ID, pal.ID)
		n.send(reqMsg, pal)
	}

	//wait for replies
	for len(req.replies) < len(n.Pals) {
		select {
		case okMsg := <-req.reqChan:
			req.replies = append(req.replies, okMsg.NodeID)
			n.send(okMsg, n)
		}
		fmt.Println("Replies to node", n.ID, "received: ", req.replies)
	}

	fmt.Printf("Node %d received all %d replies.\n\n========= Node %d ENTERING CS!!=========\n", n.ID, len(req.replies), n.ID)
	time.Sleep(time.Duration(1 * time.Second))
	fmt.Printf("\n========= Node %d EXITING CS!!=========\n", n.ID)
	n.reportChan <- 1
	n.inCS = false

	if len(n.Queue) > 0 {
		n.onRelease(reqMsg)
	} else {
		fmt.Printf("Node %d has no one to send REPLY messages to \n", n.ID)
		fmt.Println("All requests completed.")
	}
	n.hasReq = false

}

// exit CS and send REPLY/RELEASE msgs
func (n *Node) onRelease(msg Message) {
	for id, req := range n.Queue {
		go n.ok(req)
		delete(n.Queue, id)
	}

	if len(n.Queue) > 0 {
		fmt.Println("ERROR: Node", n.ID, "didn't finish sending out RELEASE")
		for k := range n.Queue {
			fmt.Println("Still in queue: Node", k)
		}
	}
}

// send OK reply
func (n *Node) ok(msg Message) {
	n.clock++
	okMsg := Message{
		ID:        msg.ID,
		NodeID:    n.ID,
		Timestamp: n.clock,
		Type:      "OK",
		Req:       msg.Req,
	}

	fmt.Println("Node", n.ID, "sent OK msg to node", msg.NodeID)
	msg.Req.reqChan <- okMsg
}

// listen and handles messages received
func (n *Node) onReceive(msg Message) {
	n.clock = syncClock(msg.Timestamp, n.clock)
	switch msg.Type {
	case "Request":
		if n.hasReq {
			n.checkReq(msg)
		} else {
			go n.ok(msg)
		}

	case "OK":

	default:
		fmt.Println("ERROR: Unknown msg received: ", msg.Type)
	}
}

// process request received / next request after releasing another
func (n *Node) checkReq(msg Message) {

	// if received request is later than my own request. Add to queue
	if n.Requests[0].lower(msg) {
		if _, ok := n.Queue[msg.NodeID]; !ok {
			n.Queue[msg.NodeID] = msg
			fmt.Println("Node", n.ID, "added REQ from Node", msg.NodeID, "to queue")
		}

	} else {
		// make sure own request is not in CS
		go func() {
			for n.inCS {
			}
			go n.ok(msg)
		}()
	}
}

func (m *Request) lower(other Message) bool {
	// if logical clock is lower
	// OR if same clock value, look at pid
	result := m.Timestamp < other.Timestamp || (m.Timestamp == other.Timestamp && m.NodeID < other.NodeID)
	return result
}

func syncClock(clkSender uint, clkReceiver uint) uint {
	if clkReceiver <= clkSender {
		return clkSender + uint(1)
	}
	return clkReceiver + uint(1)
}

func createNodes(id int) *Node {
	n := &Node{}
	n.ID = id
	n.Queue = make(map[int]Message)
	n.Channel = make(chan Message)
	n.hasReq = false
	n.inCS = false
	return n
}

func main() {
	numRequests, _ := strconv.Atoi(os.Args[1])
	var s string
	var numNodes = 10
	var nodes []*Node
	checkChan := make(chan int)

	// create Nodes
	for i := 0; i < numNodes; i++ {
		n := createNodes(i)
		n.reportChan = checkChan
		nodes = append(nodes, n)
	}
	fmt.Println(numNodes, "nodes have been created")

	// run Nodes
	for i := 0; i < numNodes; i++ {
		go nodes[i].Run(nodes)
	}
	time.Sleep(time.Duration(1 * time.Second))

	fmt.Println("========================================")
	fmt.Println("Starting Ricart & Agrawala Simulation. Number of requests:", numRequests)
	fmt.Println("========================================")
	time.Sleep(time.Duration(2 * time.Second))

	start := time.Now()
	for i := 0; i < numRequests; i++ {
		go nodes[i].request()
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

	fmt.Scanln(&s)
}
