module github.com/cem-okulmus/BalancedGo

go 1.12

require (
	github.com/alecthomas/participle v0.3.0
	github.com/cem-okulmus/BalancedGo/tools/HyperParse v0.0.0
	github.com/google/go-cmp v0.3.1
	github.com/spakin/disjoint v0.0.0-20170506060253-925e67a26b59
)

replace github.com/cem-okulmus/BalancedGo/tools/HyperParse v0.0.0 => ./tools/HyperParse
