Week 8 assignment for MSDS 431

performs bootstrap sampling across a variety of data shapes in both R and Go

run-bootstrap-median.R is the example R code given by the assignment, modified to create different data shapes.

bootstrap_sim contains the go project.

Outputs of R and Go code is in R_output.txt and Go_output.txt respectively

run as: 
  cd bootstrap_sim/cmd/bootstrap/
  go run main.go

build as:
  cd bootstrap_sim/cmd/bootstrap/
  go build main.go

test as:
  cd bootstrap_sim/internal/bootstrap
  go test -v
  cd ../distrib/
  go test -v
  cd ../permtest/
  go test -v
  cd ../stats/
  go test -v
  
The results from both the R and Go code are similar, however, the Go code ran in 20% of the time it took the R code

Go has a particular advantage over R in this use case because of the overhead associated with the large amount of loops that are run in the bootstrapping proces
Go's compiled, statically typed execution is much quicker than R's interpreted loop execution.

Go has an additional benefit over R in that it is designed to be run as a service or API

For the cost comparison, I am using AWS t3.medium, which has an on-demand hourly rate of $0.0416, 2 vCPUs, 4 GiB memory, EBS storage, and up to 5 Gigabit network performance

                    |     R     |    Go     |  
Job Wall Time       |  76.680 s | 13.3127 s |  
Wall Time (hrs)     | 0.0213 hr | 0.00337 hr|  
Compute Cost/hr     | $0.00089  | $0.00015  |  

The cost to perform the same bootstrapping calculation on the same AWS service is ~83% cheaper using Go instead of R

For this assignment, I used claude to add additional data shapes to the provided R code, and then to refactor that R code into go code. I then added benchmarking to be able to compare the two.
Claude logs can be found in claude_log.txt
