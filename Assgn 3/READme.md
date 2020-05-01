Koe Jia Yee 1003017

# PSET3
Total number of clients in network (excl.CM) = 10\
Node 999 = CM\
Node 99x = Replicas\
This value can be changed in the `main` function, `var numNodes`

File with timings can be found at `DS PSET3 Timings`

## Basic Ivy
1. `go run Ivy.go [numRequests]` where `numRequests` = number of clients simultaneously requesting

## Fault Tolerant Ivy 
### Without Faults
1. `cd Ivy1`
2. `go run CM1.go Ivy1.go [numRequests]` where `numRequests` = number of clients simultaneously requesting 

### With Faults
1. `cd Ivy2`
2. `go run CM1.go Ivy1.go 20` as at least 20 nodes are needed to help demonstrate the full simulation
