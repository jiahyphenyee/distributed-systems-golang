Koe Jia Yee 1003017

# PSET2
Total number of nodes in network (excl.server) = 10
This value can be changed in the `main` function, `var numNodes`

## Lamport Priority Queue without Ricart
1. `go run PriorityQueue.go [numRequests]` where `numRequest` = number of nodes simultaneously requesting

## Lamport Priority Queue with Ricart & Agrawala
1. `cd Ricart`
2. `go run Ricart.go [numRequests]` where `numRequest` = number of nodes simultaneously requesting 

## Centralised Server
1. `cd Central`
2. `go run Central.go [numRequests]` where `numRequest` = number of nodes simultaneously requesting 