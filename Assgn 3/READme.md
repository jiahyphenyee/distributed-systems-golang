Koe Jia Yee 1003017

# PSET3
Total number of clients in network (excl.CM) = 10\
Node 999 = CM\
Node 99x = Replicas\
This value can be changed in the `main` function, `var numNodes`\

File with timings can be found at `DS PSET3 Timings`\

## Basic Ivy
1. `go run Ivy.go [numRequests]` where `numRequests` = number of clients simultaneously requesting

## Fault Tolerant Ivy 
### Without Faults
1. `cd IvyFT`
2. `go run CMElection.go IvyFT.go [numRequests]` where `numRequests` = number of clients simultaneously requesting 

### With Faults


