package main

import (
	"fmt"
	"time"
)

//TimeoutDur ...
const TimeoutDur time.Duration = 2 * time.Second

// CM ...
type CM struct {
	ID         int
	ReqList    []*Message
	AllPages   map[int]*PageInfo // pageID:PageInfo
	Channel    chan Message
	Pals       map[int]*CM
	Pri        int // ID of pri CM
	Election   bool
	electReply map[int]int // replicas that ACK election
}

func initReplica(id int, CMPri int, page PageInfo) *CM {
	cm := &CM{}
	cm.ID = id
	cm.AllPages = make(map[int]*PageInfo)
	cm.AllPages[page.ID] = &page
	cm.Channel = make(chan Message)
	cm.Pri = CMPri
	cm.Election = false
	cm.Pals = make(map[int]*CM)
	return cm
}

// Run ...
func (cm *CM) Run(cms []*CM) {

	// fill up Pals
	for j := 0; j < len(cms); j++ {
		cm.Pals[cms[j].ID] = cms[j]
	}

	// periodic ping
	// ticker := time.NewTicker(500 * time.Millisecond)
	// go func() {
	// 	for range ticker.C {
	// 		fmt.Printf("CM %d ping Pri", cm.ID)

	// 		timeout := make(chan bool, 1)
	// 		go func() {
	// 			time.Sleep(TimeoutDur)
	// 			timeout <- true
	// 		}()

	// 		select {
	// 		case pri.Channel <- pingMsg:
	// 			return
	// 		case <-timeout:
	// 			// start election
	// 			timeoutMsg := Message{
	// 				Sender: cm.Pri,
	// 				Type:   Timeout,
	// 			}
	// 			cm.send(timeoutMsg, cm)
	// 		}
	// 	}
	// }()
}

// handle messages received
func (cm *CM) listen(msg Message) {
	fmt.Println(">>>>>CM", cm.ID, "received", msg.Type, "from CM", msg.Sender)

	switch msg.Type {
	case Ping:
		// send request list
		consistencyMsg := Message{
			Req:     msg.Req,
			Sender:  cm.ID,
			Type:    ReplyPing,
			Content: cm.ReqList,
		}
		cm.send(consistencyMsg, cm.Pals[msg.Sender])

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
			if cm.Pri == msg.Sender {
				fmt.Println("Replica", cm.ID, "detects Pri CM", msg.Sender, "Timed out!")
				cm.startElection()
			}
		}

	case Reject:
		// stop election
		cm.Election = false
		cm.electReply = make(map[int]int)

	case PriCM:
		// announcement of new Pri CM
		cm.Pri = msg.Sender

		if cm.Election {
			fmt.Println("Replica", cm.ID, "stops its election")
			cm.Election = false
			cm.electReply = make(map[int]int)
		}

	case IWon:
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
	if cm.Pri == cm.ID {
		return
	}

	fmt.Println("CM", cm.ID, "started an election")
	cm.electReply = make(map[int]int)
	cm.Election = true
	highest := true

	electMsg := Message{
		Sender: cm.ID,
		Type:   Elect,
	}

	for _, pal := range cm.Pals {
		if pal.ID > cm.ID {
			cm.electReply[pal.ID] = pal.ID
			highest = false
			go cm.send(electMsg, pal)
		}
	}

	// if n is the highest ID, wins election
	if highest {
		cm.broadcastWinner()
		return
	}
}

// announce the new Pri CM and reset election state
func (cm *CM) broadcastWinner() {
	broadcastMsg := Message{
		Sender: cm.ID,
		Type:   PriCM,
	}

	iWonMsg := Message{
		Sender: cm.ID,
		Type:   IWon,
	}

	go cm.send(iWonMsg, cm)

	cm.Election = false
	cm.Pri = cm.ID
	fmt.Println("Replica", cm.ID, "is now the new Pri CM")

	for _, pal := range cm.Pals {
		go cm.send(broadcastMsg, pal)
	}
}

// check if all requested nodes have replied to election request
func (cm *CM) checkElectResult(palID int) {

	// delete people who timed out my election msg
	if _, found := cm.electReply[palID]; found {
		delete(cm.electReply, palID)
	}

	// if all election requests timeout, wins election
	if len(cm.electReply) == 0 {
		cm.broadcastWinner()
	}
}
