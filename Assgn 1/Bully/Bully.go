// 1003017 PSET 1
package main

import (
	"fmt"

	// "math/rand"

	"time"
)

//Timeout ...
const Timeout time.Duration = 2 * time.Second

// Node ...
type Node struct {
	ID         int
	Channel    chan Message
	Pals       []*Node
	Coord      int
	Election   bool            // true if it has started an election
	electReply map[*Node]*Node // nodes that ACK election
}

// Message ...
type Message struct {
	Type   string
	Sender *Node
}

func createNodes(id int) *Node {
	n := &Node{}
	n.Channel = make(chan Message)
	n.Election = false
	n.electReply = make(map[*Node]*Node)
	return n
}

// Run ...
func (n *Node) Run(fc chan bool, allNodes []*Node) {

	// list other nodes
	for _, node := range allNodes {
		if node.ID != n.ID {
			n.Pals = append(n.Pals, node)
		}
	}
	fmt.Println("Node", n.ID, "has started")

	// set coordinator as the highestID node
	n.Coord = allNodes[len(allNodes)-1].ID
	fmt.Println("Coordinator set to node", n.Coord, "for Node", n.ID)

	// set up channels
	for {
		select {
		case msg := <-n.Channel:
			n.receive(msg)
		case <-fc:
			return
		}
	}
}

// starts an election by setting election state to True and sending election requests to higher IDs
func (n *Node) startElection() {
	fmt.Println("Node", n.ID, "started an election")
	n.Election = true
	msg := Message{"Elect", n}
	highest := true

	for _, node := range n.Pals {
		if node.ID > n.ID {
			n.electReply[node] = node
			highest = false
			go n.send(msg, node)
		}
	}
	// if n is the highest ID, wins election
	if highest {
		n.broadcastWinner()
		return
	}
}

// check if all requested nodes have replied to election request
func (n *Node) checkElectResult(node *Node) {

	// if someone replied me, i lose election
	if _, found := n.electReply[node]; found {
		delete(n.electReply, node)
	}

	// if all election requests timeout, wins election
	if len(n.electReply) == 0 {
		n.broadcastWinner()
	}
}

// announce the new coordinator and reset election state
func (n *Node) broadcastWinner() {
	msg := Message{"Coordinator", n}
	n.Election = false
	n.Coord = n.ID
	fmt.Println("Node", n.ID, "sets itself as Coordinator")

	for _, node := range n.Pals {
		go n.send(msg, node)
	}
}

// handles messages received
func (n *Node) receive(msg Message) {
	switch msg.Type {
	case "Elect":
		// since only higherIDs are sent election messages, reply with rejection
		reply := Message{"Reject", n}
		n.send(reply, msg.Sender)
		if !n.Election {
			n.startElection()
		}
	case "Coordinator":
		// end election and update new coordinator
		fmt.Println("Node", n.ID, "receives coordinator broadcast msg from Node", msg.Sender.ID)
		if n.Election {
			fmt.Println("Node", n.ID, "stops its election")
			n.Election = false
			n.electReply = make(map[*Node]*Node)
		}
		n.Coord = msg.Sender.ID
		fmt.Println("Node", n.ID, "sets", n.Coord, "as new Coordinator")
	case "Timeout":
		// check whether an election is going on
		switch n.Election {
		case true:
			n.checkElectResult(msg.Sender)
		case false:
			// is the faulty node a coordinator
			if n.Coord == msg.Sender.ID {
				fmt.Println("Node", n.ID, "detects that the coordinator Node", msg.Sender.ID, "is faulty")
				n.startElection()
			} else {
				fmt.Println("Node", msg.Sender.ID, "timed out")
			}
		}
	case "Reject":
		// stop election, reset non-election state
		if n.Election {
			fmt.Println("Node", n.ID, "rejected by", msg.Sender.ID)
			fmt.Println("Node", n.ID, "stops election")
			n.Election = false
			n.electReply = make(map[*Node]*Node)
		}
	case "Hello":
	}
}

// sends message to other nodes and take note of timeout
func (n *Node) send(msg Message, dest *Node) {
	fmt.Println("Node", n.ID, "sends", msg.Type, "message to Node", dest.ID)
	go func() {
		timeout := make(chan bool, 1)
		go func() {
			time.Sleep(Timeout)
			timeout <- true
		}()

		select {
		case dest.Channel <- msg:
			return
		case <-timeout:
			n.Channel <- Message{"Timeout", dest}
		}
		// if <-timeout {
		// 	fmt.Println("Node", n.ID, "receive timeout message from node", dest.ID)
		// 	n.Channel <- Message{"Timeout", dest}
		// }
	}()
}

// Reset ...
func Reset(nodes []*Node) {
	// reset coordinate for everyone
	for _, n := range nodes {
		n.Coord = len(nodes) - 1
	}
	fmt.Println("Nodes have been reset. Simulating faulty coordinator..")
}

func main() {
	var s string
	var numNodes = 6
	var faultChannels []chan bool
	var nodes []*Node

	// create Nodes
	for i := 0; i < numNodes; i++ {
		n := createNodes(i)
		n.ID = i
		nodes = append(nodes, n)
	}
	fmt.Println(numNodes, "have been created")

	// run Nodes
	for i := 0; i < numNodes; i++ {
		fc := make(chan bool)
		faultChannels = append(faultChannels, fc)
		go nodes[i].Run(fc, nodes)
	}
	time.Sleep(time.Duration(2 * time.Second))

	// kill the coordinator and simulate detection
	faultChannels[len(faultChannels)-1] <- true
	fmt.Println("!!KILLING THE COORDINATOR!! which is Node", nodes[numNodes-1].ID)

	fmt.Println("==========")
	fmt.Println("BEST CASE: 2nd highest ID detects faulty Coordinator.", nodes[numNodes-2].ID, "pings faulty Coordinator")
	fmt.Println("==========")
	nodes[numNodes-2].send(Message{"Hello", nodes[numNodes-2]}, nodes[numNodes-1])

	time.Sleep(time.Duration(10 * time.Second))
	fmt.Println("Press 'Enter' to start next scenario")
	fmt.Scanln(&s)
	Reset(nodes)

	fmt.Println("==========")
	fmt.Println("WORST CASE: Lowest ID detects faulty Coordinator.", nodes[0].ID, "pings faulty Coordinator")
	fmt.Println("==========")
	nodes[0].send(Message{"Hello", nodes[0]}, nodes[numNodes-1])

	time.Sleep(time.Duration(12 * time.Second))
	fmt.Println("Press 'Enter' to start next scenario")
	fmt.Scanln(&s)
	Reset(nodes)

	fmt.Println("==========")
	fmt.Println("2nd highest ID node fails after election has started by Lowest ID node.")
	fmt.Println("==========")
	nodes[0].startElection()
	time.Sleep(time.Duration(1 * time.Second))
	faultChannels[numNodes-2] <- true
	fmt.Println("Killed 2nd highest ID node", nodes[numNodes-2].ID)

	time.Sleep(time.Duration(8 * time.Second))
	fmt.Println("Press 'Enter' to start next scenario")
	fmt.Scanln(&s)
	Reset(nodes)

	fmt.Println("==========")
	fmt.Println("Lowest 3 nodes simultaneously start election process")
	fmt.Println("==========")
	nodes[0].startElection()
	nodes[1].startElection()
	nodes[2].startElection()

	time.Sleep(time.Duration(9 * time.Second))
	fmt.Println("Finish")
	fmt.Scanln(&s)
}
