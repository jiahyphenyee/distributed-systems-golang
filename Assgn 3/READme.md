Koe Jia Yee 1003017

# PSET3
Total number of clients in network (excl.CM) = 10 to 20\
This value can be changed in the `main` function, `var numNodes`

Node 999 = CM\
Node 99x = Replicas\
Currently there are 3 Replicas. This value can be changed in the `main` function, `var numNodes`

File with timings can be found at `DS PSET3 Timings`

## Basic Ivy
1. `go run Ivy.go [numRequests]` where `numRequests` = number of clients simultaneously requesting

## Fault Tolerant Ivy 
### Without Faults
1. `cd Ivy1`
2. `go run CM1.go Ivy1.go [numRequests]` where `numRequests` = number of clients simultaneously requesting 

### With Faults - Pri CM dies
1. `cd Ivy2`
2. `go run CM1.go Ivy1.go 20` as at least 20 nodes are needed to help demonstrate the full simulation

### With Faults - Pri CM dies and revives
1. Run steps 1 & 2 as above
2. When the print statement

`==================
99X WON ELECTION. START COMPLETING REQUESTS
==================`

appears, press `Enter` to simulate nonde revival.
