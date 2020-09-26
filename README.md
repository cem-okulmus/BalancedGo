# BalancedGo
Compute Generalized Hypertree Decompositions via the use of balanced separators, in Go with a focus on parallelism. 

Takes as input a hypergraph in [HyperBench format](http://hyperbench.dbai.tuwien.ac.at/downloads/manual.pdf) or [PACE Challenge 2019 format](https://pacechallenge.org/2019/htd/htd_format/), and a width parameter (positive non-zero integer). 
[HyperBench](http://hyperbench.dbai.tuwien.ac.at/) is a benchmark library, containing over 3000 hypergraphs from CQ and CSP instances, from industry and research. 

## Installation
Needs Go >= 1.12, look [here](https://golang.org/dl/) for Linux, MacOS or Windows versions.   
Simply run `make`, alternatively on platforms without the make tool, run `go build`
  



