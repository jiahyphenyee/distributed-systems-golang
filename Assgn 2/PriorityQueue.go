package main

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"time"
)

// Request ...
type Request struct {
	// for tracking own requests only
	ID      int64
	reqChan chan Message
	replies []int //nodeID
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
	Queue      []Message
	ToReply    map[*Request]bool // ppl I have yet to reply
	Requests   []*Request
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
		ID:      int64(n.ID) + int64(n.clock),
		reqChan: make(chan Message),
	}

	reqMsg := Message{
		ID:        req.ID,
		NodeID:    n.ID,
		Timestamp: n.clock,
		Type:      "Request",
		Req:       &req,
	}

	fmt.Println("\nTimestamp of Req from Node", n.ID, ":", n.clock)

	n.Requests = append(n.Requests, &req)

	// TODO add to queue properly
	n.Queue = append(n.Queue, reqMsg)

	for _, pal := range n.Pals {
		fmt.Printf("Node %d -------REQ-------> node %d\n", n.ID, pal.ID)
		n.send(reqMsg, pal)
	}

	//wait for replies
	for len(req.replies) < len(n.Pals) || n.Queue[0].NodeID != n.ID {
		select {
		case okMsg := <-req.reqChan:
			req.replies = append(req.replies, okMsg.NodeID)
			n.send(okMsg, n)
		}
		fmt.Println("Replies to Node", n.ID, "received: ", req.replies)
	}

	n.inCS = true
	fmt.Printf("Node %d received all %d replies.\n\n========= Node %d ENTERING CS!!=========\n", n.ID, len(req.replies), n.ID)
	time.Sleep(time.Duration(1 * time.Second))
	fmt.Printf("\n========= Node %d EXITING CS!!=========\n", n.ID)
	n.reportChan <- 1
	n.inCS = false
	n.release(reqMsg)

}

// exit CS and send release msg. Release from own queue
func (n *Node) release(msg Message) {
	n.clock++
	var x Message

	relMsg := Message{
		ID:        msg.ID,
		NodeID:    n.ID,
		Timestamp: n.clock,
		Type:      "Release",
	}

	for _, pal := range n.Pals {
		fmt.Printf("Node %d --------RELEASE------->node %d \n", n.ID, pal.ID)
		go n.send(relMsg, pal)
	}

	// release from own queue
	if msg.ID != n.Queue[0].ID {
		panic(fmt.Sprintln("ERROR: Wrong request released from queue!"))
	} else {
		x, n.Queue = n.Queue[0], n.Queue[1:]
		fmt.Print(x)
	}

	if len(n.Queue) > 0 {
		n.checkToReply()
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

	fmt.Println("Node", n.ID, "-------OK-------> Node", msg.NodeID)
	msg.Req.reqChan <- okMsg

	// remove from toReply queue
	delete(n.ToReply, msg.Req)
}

// listen and handles messages received
func (n *Node) onReceive(msg Message) {
	n.clock = syncClock(msg.Timestamp, n.clock)
	// fmt.Println("Node", n.ID, "Clock updated to", n.clock)
	var x Message

	switch msg.Type {
	case "Request":
		n.addToQueue(msg)

	case "Release":
		// release from queue and reset head
		x, n.Queue = n.Queue[0], n.Queue[1:]
		fmt.Printf("Releasing node %d from Queue %d\n", msg.NodeID, n.ID)

		if msg.NodeID != x.NodeID {
			fmt.Println("ERROROTHERS: Wrong request released from queue!")
		}
		n.checkToReply()
	}
}

// process request received / next request after releasing another
func (n *Node) addToQueue(msg Message) {
	n.Queue = append(n.Queue, msg)
	currentHead := n.Queue[0]

	if len(n.Queue) > 1 {
		sort.SliceStable(n.Queue,
			func(i, j int) bool {
				return n.Queue[i].lower(n.Queue[j])
			})

		fmt.Println("Node", n.ID, "added Node", msg.NodeID, "to queue")
		n.ToReply[msg.Req] = false

	}

	// if req received is head of queue
	if len(n.Queue) == 1 || n.Queue[0] == msg {
		fmt.Println("Node", n.ID, "Queue Head = Node", msg.NodeID)

		go func() {
			for n.inCS {
			}
			go n.ok(msg)
		}()

	} else {
		// check that head stays the same
		if currentHead != n.Queue[0] {
			panic(fmt.Sprintln("head of queue changed randomly to", n.Queue[0].NodeID))
		}
	}
}

func (m *Message) lower(other Message) bool {
	// if logical clock is lower
	// OR if same clock value, look at pid
	result := m.Timestamp < other.Timestamp || (m.Timestamp == other.Timestamp && m.NodeID < other.NodeID)
	return result
}

// on Release, checks who to reply next
func (n *Node) checkToReply() {
	if len(n.Queue) > 0 {
		if _, ok := n.ToReply[n.Queue[0].Req]; ok {
			go n.ok(n.Queue[0])
		}
	} else {
		fmt.Println("Node", n.ID, "completed")
	}
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
	n.Channel = make(chan Message)
	n.ToReply = make(map[*Request]bool)
	n.inCS = false
	return n
}

func main() {
	numRequests, _ := strconv.Atoi(os.Args[1])
	var s string
	var numNodes = 10
	var nodes []*Node
	clock := uint(0)
	checkChan := make(chan int)

	// create Nodes
	for i := 0; i < numNodes; i++ {
		n := createNodes(i)
		n.clock = clock
		n.reportChan = checkChan
		nodes = append(nodes, n)
	}
	fmt.Println(numNodes, "nodes have been created")

	// run Nodes
	for i := 0; i < numNodes; i++ {
		go nodes[i].Run(nodes)
	}
	time.Sleep(time.Duration(2 * time.Second))

	fmt.Println("==========")
	fmt.Println("Starting Lamport Shared Priority Queue Simulation. Number of requests:", numRequests)
	fmt.Println("==========")
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
