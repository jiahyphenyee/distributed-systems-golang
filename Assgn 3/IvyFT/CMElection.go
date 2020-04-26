package main

import (
	"fmt"
	"time"
)

//Timeout ...
const TimeoutDur time.Duration = 2 * time.Second

// CM ...
type CM struct {
	ID       int
	ReqList  []*Message
	AllPages map[int]*PageInfo // pageID:PageInfo
	Channel  chan Message
	Pals     []*CM
	Pri      int // ID of pri CM
	Election bool
}

func initReplica(id int, CMPri int, page PageInfo) *CM {
	cm := &CM{}
	cm.ID = id
	cm.AllPages[page.ID] = &page
	cm.Channel = make(chan Message)
	cm.Pri = CMPri
	cm.Election = false
	return cm
}

// Run ...
func (cm *CM) Run(cms []*CM) {
	cm.Pals = append(cm.Pals, cms...)

	// get primary CM
	var pri *CM
	for j := 0; j < len(cm.Pals); j++ {
		if cm.Pals[j].ID == cm.Pri {
			pri = cm.Pals[j]
		}
	}

	r := Request{
		Requester: cm.ID,
		ReqType:   PingCM,
		Channel:   make(chan (bool)),
	}

	pingMsg := Message{
		Req:    r,
		Sender: cm.ID,
		Type:   PingCM,
	}

	// periodic ping
	ticker := time.NewTicker(5 * time.Second)
	go func() {
		for range ticker.C {
			fmt.Printf("CM %d ping Pri", cm.ID)

			timeout := make(chan bool, 1)
			go func() {
				time.Sleep(TimeoutDur)
				timeout <- true
			}()

			select {
			case pri.Channel <- pingMsg:
				return
			case <-timeout:
				// start election
				timeoutMsg := Message{
					Sender: cm.Pri,
					Type:   Timeout,
				}

				cm.send(timeoutMsg, cm)

			}
		}
	}()

	// set up channels
	for {
		select {
		case msg := <-cm.Channel:
			cm.listen(msg)
		default:
			continue
		}
	}

}

func (cm *CM) listen(msg Message) {
	fmt.Println("CM", cm.ID, "received", msg.Type, "from CM", msg.Sender)

	switch msg.Type {
	case PingCM:
		// send request list
		broadcastMsg := Message{
			Req:     msg.Req,
			Sender:  cm.ID,
			Type:    ReplyPing,
			Content: cm.ReqList,
		}
		cm.send(broadcastMsg, cm.Pals[msg.Sender])

	case ReplyPing:
		cm.ReqList = msg.Content

	case Elect:
		// since only higherIDs are sent election messages, reply with rejection
		rejectMsg := Message{
			Sender: cm.ID,
			Type:   Reject,
		}
		cm.send(rejectMsg, cm.Pals[msg.Sender])

		if !cm.Election {
			cm.startElection()
		}

	case Timeout:
		// can either be no reply during election
		// or timeout from ping to Pri CM
		switch cm.Election {
		case true:
			cm.checkElectResult(msg.Sender)
		case false:
			// is the faulty node a coordinator
			if cm.Coord == msg.Sender {
				fmt.Println("Replica", cm.ID, "detects Pri CM", msg.Sender, "Timed out!"))
				n.startElection()
			} else {
				panic(fmt.Sprintln("who the heck timeout me"))
			}
		}

	case Reject:
		// stop election
		cm.Election = false

	}
}

// send msgs
func (cm *CM) send(msg Message, dest *CM) {
	go func() {
		dest.Channel <- msg
		return
	}()
}

// start election
func (cm *CM) startElection() {
	fmt.Println("CM", cm.ID, "started an election")
	cm.Election = true
	highest := true

	electMsg := Message {
		Sender: cm.ID,
		Type: Elect,
	}

	for _, pal := range cm.Pals {
		if pal.ID > cm.ID {
			n.electReply[node] = node
			highest = false
			go n.send(msg, node)
		}
	}

}

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