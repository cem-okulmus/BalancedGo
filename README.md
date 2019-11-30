# BalancedGo
Compute Generalized Hypertree Decompositions via balanced separators, in Go with a focus on parallelism. 

Takes as input a hypergraph in [HyperBench format](http://hyperbench.dbai.tuwien.ac.at/downloads/manual.pdf) or [PACE Challenge 2019 format](https://pacechallenge.org/2019/htd/htd_format/), and a width parameter (positive non-zero integer). 
[HyperBench](http://hyperbench.dbai.tuwien.ac.at/) is a benchmark library, containing over 3000 hypergraphs from CQ and CSP intances, from industry and research. 

## Installation
Needs Go >= 1.12, look [here](https://golang.org/dl/) for Linux, MacOS or Windows versions.   
Simply run `make`, alternatitvely on platforms without the make tool, run `go build`

## Usage 
**Usage of BalancedGo:**

-exact      
> Compute exact width (width flag not needed)  

-graph      	<string>  
> the file path to a hypergraph   
>	(see http://hyperbench.dbai.tuwien.ac.at/downloads/manual.pdf, 1.3 for correct format)  

-width      	<int>
>	a positive, non-zero integer indicating the width of the GHD to search for  

**Optional Arguments:**

-akatov     
>	compute balanced decomposition   

-balDet     	<int>  
>	Test hybrid balSep and DetK algorithm  
  
-balfactor  	<int>  
>	Determines the factor that balanced separator check uses  
  
-bench  
>	Benchmark mode, reduces unneeded output (incompatible with -log flag)  

-cpu        	<int>  
>	Set number of CPUs to use  
  
-cpuprofile 	<string>  
>	write cpu profile to file  
  
-det         
>	Test out DetKDecomp  

-divide      
>	Test for divideKDecomp  

-g          
>	perform a GYÖ reduct and show the resulting graph  

-global     
>	Test out global BalSep  

-gml        	<string>
>	Output the produced decomposition into the specified gml file  
  
-heuristic  	<int>
>	turn on to activate edge ordering  
> 1 ... Degree Ordering  
>	2 ... Max. Separator Ordering  
>	3 ... MCSO    
  
-local      
>	Test out local BalSep  

-localbip   
>	To be used in combination with "det": turns on subedge handling  

-log        
>	turn on extensive logs  

-pace       
>	Use PACE 2019 format for graphs  
>	(see https://pacechallenge.org/2019/htd/htd_format/ for correct format)  

-sub        
>	Compute the subedges of the graph and print it out 

-t          
>	perform a Type Collapse and show the resulting graph  



