# BalancedGo
[![Actions Status](https://github.com/cem-okulmus/BalancedGo/workflows/Go/badge.svg)](https://github.com/cem-okulmus/BalancedGo/actions)
[![](https://img.shields.io/github/v/release/cem-okulmus/BalancedGo)](https://github.com/cem-okulmus/BalancedGo/releases/latest)
[![Go Reference](https://pkg.go.dev/badge/github.com/cem-okulmus/BalancedGo.svg)](https://pkg.go.dev/github.com/cem-okulmus/BalancedGo)
[![Go Report Card](https://goreportcard.com/badge/github.com/cem-okulmus/BalancedGo)](https://goreportcard.com/report/github.com/cem-okulmus/BalancedGo)

Compute Generalized Hypertree Decompositions via the use of balanced separators, in Go with a focus on parallelism. 

Takes as input a hypergraph in [HyperBench format](http://hyperbench.dbai.tuwien.ac.at/downloads/manual.pdf) or [PACE Challenge 2019 format](https://pacechallenge.org/2019/htd/htd_format/), and a width parameter (positive non-zero integer). 
[HyperBench](http://hyperbench.dbai.tuwien.ac.at/) is a benchmark library, containing over 3000 hypergraphs from CQ and CSP instances, from industry and research. 

## Installation
Needs Go >= 1.12, look [here](https://golang.org/dl/) for Linux, MacOS or Windows versions.   
Simply run `make`, alternatively on platforms without the make tool, run `go build`
  


## Usage 
No fixed command-line interface. Use "BalancedGo -h" to see the currently supported commands. 
Generally, any run will require 1) a valid hypergraph, according to the formats specified above, 2) a specified width (unless the "exact" or "approx" flags are used) and 3) an algorithm to actually compute an HD or GHD (depending on the type of algorithm). 
