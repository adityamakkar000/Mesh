package main

import (
	"github.com/adityamakkar000/Mesh/internal/parse"
)

func main() {
	clusters, err := parse.ParseClusters("./cluster.yaml") 

	if err != nil {
		panic(err)
	}

	_ = clusters 
}
