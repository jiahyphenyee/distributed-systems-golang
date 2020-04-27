package main

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"time"
)

// MessageType ...
type MessageType string

// AcessType ...
type AcessType string

const (
	// ReadQuery reader to CM
	ReadQuery MessageType = "ReadQuery"
	// ReadForward CM to Owner
	ReadForward MessageType = "ReadForward"
	// ReadData Owner to Reader
	ReadData MessageType = "ReadData"
	// ReadConfirm Reader to Owner
	ReadConfirm MessageType = "ReadConfirm"
	// WriteQuery Reader to CM
	WriteQuery MessageType = "WriteQuery"
	// WriteForward CM to Owner
	WriteForward MessageType = "WriteForward"
	// WriteData Owner to Writer
	WriteData MessageType = "WriteData"
	// WriteConfirm Writer to CM
	WriteConfirm MessageType = "WriteConfirm"
	// Invalidate CM to CopySet
	Invalidate MessageType = "Invalidate"
	// InvalidConfirm CopySet to CM
	InvalidConfirm MessageType = "InvalidConfirm"

	// Ping primary CM
	Ping MessageType = "Ping"
	// ReplyPing ...
	ReplyPing MessageType = "ReplyPing"
	// Elect when starting election
	Elect MessageType = "Elect"
	// Reject election message from lower ID
	Reject MessageType = "Reject"
	// Timeout to detect failed CM
	Timeout MessageType = "Timeout"
	// PriCM newly elected PriCM
	PriCM MessageType = "PriCM"

	// ReadCopy represents Read access to copy of Page
	ReadCopy AcessType = "ReadCopy"
	// WriteOwner represents Write access by owner
	WriteOwner AcessType = "WriteOwner"
	// ReadOwner represents Read access by owner
	ReadOwner AcessType = "ReadOwner"
)

// Request ...
type Request struct {
	ID        int
	Requester int // nodeID
	ReqType   MessageType
	Channel   chan bool //channel to notify req complete
	Page      int
}

// Message ...
type Message struct {
	Req       Request
	Sender    int
	Timestamp uint
	Type      MessageType
	Content   []*Message
}

// Node ...
type Node struct {
	ID         int
	Channel    chan Message
	Pals       map[int]*Node
	Lock       bool
	clock      uint
	PageAccess map[int]AcessType // pageID:accesstype
	CMID       int
	CMStore    CM
	reportChan chan int
}

// PageInfo ...
type PageInfo struct {
	ID      int
	Lock    bool
	OwnerID int
	CopySet []int // list of nodeID owning copies
}

// Run ...
func (n *Node) Run(allNodes []*Node) {
	// fill up Pals
	for j := 0; j < len(allNodes); j++ {
		n.Pals[allNodes[j].ID] = allNodes[j]
	}
	fmt.Println("Node", n.ID, "has started")

	n.Lock = false
	n.PageAccess = make(map[int]AcessType)

	// set up channels
	for {
		select {
		case msg := <-n.Channel:
			n.listen(msg)
		default:
			continue
		}
	}
}

// send msgs
func (n *Node) send(msg Message, dest *Node) {
	n.clock++
	msg.Timestamp = n.clock

	go func() {
		for {
			select {
			case dest.Channel <- msg:
				return
			default:
				continue
			}
		}
	}()
}

// create and send request
func (n *Node) request(msgType MessageType) {

	r := Request{
		ID:        n.ID + int(n.clock),
		Requester: n.ID,
		ReqType:   msgType,
		Channel:   make(chan (bool)),
		Page:      0,
	}

	reqMsg := Message{
		Req:    r,
		Sender: n.ID,
		Type:   msgType,
	}

	// Owner is reading
	if j, ok := n.PageAccess[r.Page]; ok {
		if r.ReqType == ReadQuery && (j == ReadCopy || j == ReadOwner) {
			n.reportChan <- 1
			return
		}
	}

	var confirmType MessageType
	if msgType == ReadQuery {
		confirmType = ReadConfirm
	} else {
		confirmType = WriteConfirm
	}

	fmt.Println("\n------ New", msgType, "from Node", n.ID, "at clock:", n.clock)
	n.Lock = true
	n.send(reqMsg, n.Pals[n.CMID])

	//wait for replies
	for {
		select {
		case <-r.Channel:
			// R/W request done
			confirmMsg := Message{
				Req:    reqMsg.Req,
				Sender: n.ID,
				Type:   confirmType,
			}
			n.send(confirmMsg, n.Pals[n.CMID])
			n.reportChan <- 1
			return

		default:
			continue
		}
	}
}

// listen and handles messages received
func (n *Node) listen(msg Message) {
	n.clock = syncClock(msg.Timestamp, n.clock)
	// fmt.Println("Node", n.ID, "Clock updated to", n.clock)
	fmt.Println("Node", n.ID, "received", msg.Type, "from Node", msg.Sender)
	switch msg.Type {
	case ReadQuery:
		n.addToReqList(msg)

	case WriteQuery:
		n.addToReqList(msg)

	case ReadForward:
		n.Lock = true

		dataMsg := Message{
			Req:    msg.Req,
			Sender: n.ID,
			Type:   ReadData,
		}
		n.send(dataMsg, n.Pals[msg.Req.Requester])
		n.Lock = false

	case WriteForward:
		n.Lock = true
		var empty AcessType
		n.PageAccess[msg.Req.Page] = empty

		dataMsg := Message{
			Req:    msg.Req,
			Sender: n.ID,
			Type:   WriteData,
		}
		n.send(dataMsg, n.Pals[msg.Req.Requester])
		n.Lock = false

	case ReadData:
		n.PageAccess[msg.Req.Page] = ReadCopy
		n.Lock = false
		fmt.Println("\n ------ Node", n.ID, "request complete")
		msg.Req.Channel <- true
		time.Sleep(time.Duration(1 * time.Second))

	case WriteData:
		n.PageAccess[msg.Req.Page] = WriteOwner
		fmt.Println("Node", n.ID, "is now the new Owner of Page", msg.Req.Page)
		fmt.Println("\n ------ Node", n.ID, "request complete")
		n.Lock = false
		msg.Req.Channel <- true
		time.Sleep(time.Duration(1 * time.Second))

	case ReadConfirm:
		// remove from CM ReqList
		for ind, reqMsg := range n.CMStore.ReqList {
			if reqMsg.Sender == msg.Sender {
				if len(n.CMStore.ReqList) == 1 {
					n.CMStore.ReqList = []*Message{}
				} else {
					copy(n.CMStore.ReqList[ind:], n.CMStore.ReqList[ind+1:])
					n.CMStore.ReqList = n.CMStore.ReqList[:len(n.CMStore.ReqList)-1]
				}
			}
		}
		n.CMStore.AllPages[msg.Req.Page].Lock = false
		n.startNextRequest()

	case WriteConfirm:
		n.CMStore.AllPages[msg.Req.Page].OwnerID = msg.Req.Requester

		// remove from CM ReqList
		for ind, reqMsg := range n.CMStore.ReqList {
			if reqMsg.Sender == msg.Sender {
				if len(n.CMStore.ReqList) == 1 {
					n.CMStore.ReqList = []*Message{}
				} else {
					copy(n.CMStore.ReqList[ind:], n.CMStore.ReqList[ind+1:])
					n.CMStore.ReqList = n.CMStore.ReqList[:len(n.CMStore.ReqList)-1]
				}
			}
		}
		n.CMStore.AllPages[msg.Req.Page].Lock = false
		n.startNextRequest()

	case Invalidate:
		n.Lock = true
		var empty AcessType
		n.PageAccess[msg.Req.Page] = empty

		invCMsg := Message{
			Req:    msg.Req,
			Sender: n.ID,
			Type:   InvalidConfirm,
		}
		n.send(invCMsg, n.Pals[msg.Sender]) // aka CM
		n.Lock = false

	case InvalidConfirm:
	}
}

// process request received / next request after releasing another
func (n *Node) addToReqList(msg Message) {
	n.CMStore.ReqList = append(n.CMStore.ReqList, &msg)
	currentHead := n.CMStore.ReqList[0]

	if len(n.CMStore.ReqList) > 1 {
		sort.SliceStable(n.CMStore.ReqList,
			func(i, j int) bool {
				return n.CMStore.ReqList[i].lower(*n.CMStore.ReqList[j])
			})

		fmt.Println("Added Node", msg.Sender, "to queue. Length: ", len(n.CMStore.ReqList))
	}

	// if req received is head of queue
	if len(n.CMStore.ReqList) == 1 || n.CMStore.ReqList[0] == &msg {
		fmt.Println("Queue Head = Node", msg.Sender)

		n.startNextRequest()

	} else {
		// check that head stays the same
		if currentHead != n.CMStore.ReqList[0] {
			panic(fmt.Sprintln("Head of queue changed randomly to", n.CMStore.ReqList[0].Sender))
		}
	}
}

// check which Request next
func (n *Node) startNextRequest() {

	if len(n.CMStore.ReqList) > 0 {
		nextReq := n.CMStore.ReqList[0]
		n.Pals[nextReq.Sender].Lock = true
		ownerID := n.CMStore.AllPages[nextReq.Req.Page].OwnerID

		if nextReq.Type == ReadQuery {
			// if ReadQuery request
			fmt.Println("\nStarting ReadQuery by Node", nextReq.Sender)
			n.CMStore.AllPages[nextReq.Req.Page].CopySet = append(n.CMStore.AllPages[nextReq.Req.Page].CopySet, nextReq.Sender)
			n.CMStore.AllPages[nextReq.Req.Page].Lock = true
			forwardMsg := Message{
				Req:    nextReq.Req,
				Sender: n.ID,
				Type:   ReadForward,
			}
			n.send(forwardMsg, n.Pals[ownerID])

		} else if nextReq.Type == WriteQuery {
			// if WriteQuery request
			fmt.Println("\nStarting WriteQuery by Node", nextReq.Sender)
			n.CMStore.AllPages[nextReq.Req.Page].Lock = true
			invMsg := Message{
				Req:    nextReq.Req,
				Sender: n.ID,
				Type:   Invalidate,
			}
			for _, copyID := range n.CMStore.AllPages[nextReq.Req.Page].CopySet {
				n.send(invMsg, n.Pals[copyID])
			}

			// wait for invalidation confirmation
			if len(n.CMStore.AllPages[nextReq.Req.Page].CopySet) != 0 {
				var replies int
				n.CMStore.AllPages[nextReq.Req.Page].Lock = false
				for replies < len(n.CMStore.AllPages[nextReq.Req.Page].CopySet) {
					select {
					case invConfirm := <-n.Channel:
						if invConfirm.Type == InvalidConfirm {
							fmt.Println("Node", n.ID, "received", invConfirm.Type, "from Node", invConfirm.Sender)
							replies++
						}
					}
				}
			} else {
				fmt.Println("No one in copyset just WRITE")
			}

			// clear current CopySet
			n.CMStore.AllPages[nextReq.Req.Page].Lock = true
			n.CMStore.AllPages[nextReq.Req.Page].CopySet = []int{}

			forwardMsg := Message{
				Req:    nextReq.Req,
				Sender: n.ID,
				Type:   WriteForward,
			}
			n.send(forwardMsg, n.Pals[ownerID])
		}

	} else {
		fmt.Println("All requests completed.")
	}
}

func createNode(id int) *Node {
	n := &Node{}
	n.ID = id
	n.Channel = make(chan Message, 10)
	n.CMID = 999 // default primary CM
	n.PageAccess = make(map[int]AcessType)
	n.Pals = make(map[int]*Node)

	return n
}

// ---- Total Program Order ---- //

func (m *Message) lower(other Message) bool {
	// if logical clock is lower
	// OR if same clock value, look at pid
	result := m.Timestamp < other.Timestamp || (m.Timestamp == other.Timestamp && m.Sender < other.Sender)
	return result
}

func syncClock(clkSender uint, clkReceiver uint) uint {
	if clkReceiver <= clkSender {
		return clkSender + uint(1)
	}
	return clkReceiver + uint(1)
}

// ---- MAIN ---- //
func main() {
	numRequests, _ := strconv.Atoi(os.Args[1])
	var s string
	var numNodes = 10
	var numReplicas = 3
	var nodes []*Node
	var cms []*CM
	clock := uint(0)
	checkChan := make(chan int)

	// create Nodes
	for i := 0; i < numNodes; i++ {
		n := createNode(i)
		n.clock = clock
		n.reportChan = checkChan
		nodes = append(nodes, n)
	}
	fmt.Println(numNodes-1, "Clients have been created")

	// create Page
	page := PageInfo{0, false, numNodes - 1, []int{}}
	nodes[numNodes-1].PageAccess[0] = ReadOwner // store page with Owner

	// create CM nodes
	for i := 0; i < numReplicas+1; i++ {
		n := createNode(999 - i)
		n.clock = clock
		n.reportChan = checkChan
		nodes = append(nodes, n)
		n.CMStore = *initReplica(n.ID, n.CMID, page)
		cms = append(cms, &n.CMStore)
	}

	// run Nodes
	for i := 0; i < len(nodes); i++ {
		go nodes[i].Run(nodes)
	}

	// run CM
	for j := 0; j < len(cms); j++ {
		go cms[j].Run(cms)
	}

	time.Sleep(time.Duration(1 * time.Second))

	fmt.Println("==========")
	fmt.Println("Starting IVY with fault tolerance.\nNumber of requests:", numRequests)
	fmt.Println("CM: Node 999 \nPage owner: Node", numNodes-1)
	fmt.Println("==========")
	time.Sleep(time.Duration(2 * time.Second))

	start := time.Now()
	for i := 0; i < numRequests; i++ {
		r := ReadQuery
		if i == 2 || i == 5 || i == 8 {
			r = WriteQuery
		}
		go nodes[i].request(r)
	}

	checkDone := 0
	go func() {
		for checkDone < numRequests {
			select {
			case <-checkChan:
				checkDone++
				fmt.Println(checkDone, "requests done")
			}
		}

		end := time.Now()
		elapsed := end.Sub(start)
		time.Sleep(time.Duration(1 * time.Second))
		fmt.Println("\n\n==========")
		fmt.Println("Time elapsed:", elapsed)
		fmt.Println("==========")

	}()

	fmt.Scanln(&s)
}
