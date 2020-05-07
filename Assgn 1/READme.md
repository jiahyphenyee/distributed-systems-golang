## Lamport Logical Vlock
1. Open terminal and navigate to directory containing Lamport.go
2. In main(), edit `var numClient` to change number of clients in network (minimum 3)
3. In terminal, enter `go run Lamport.go`
4. Press the 'return' key to stop processes and output Total Order

## Vector Clock
1. Open terminal and navigate to directory containing Lamport.go
2. In main(), edit `var numClient` to change number of clients in network (minimum 3)
3. In terminal, enter `go run VectorClock.go`
4. Press the 'return' key to stop processes and output potential causality violations

## Bully Election
1. Open terminal and navigate to directory containing Bully.go
2. In main(), edit `var numNodes` to change number of nodes in network (minimum 5)
3. In terminal, enter `go run Bully.go`
4. When prompted, press the 'return' key for the next scenario

Included:
- Worst Case: lowest ID node detects coordinator fault
- Best Case: second highest ID node detects coordinator fault
- Second highest ID node fails after election has started by Lowest ID node
- Lowest 3 nodes start election process simultaneously
