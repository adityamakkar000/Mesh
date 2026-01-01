package main

import (
	cli "github.com/adityamakkar000/Mesh/internal/parse"
)

func main() {
	// cli.Execute()

	_, err := cli.Clusters()
	if err != nil {
		panic(err)
	}

}

// import (
// 	cli "github.com/adityamakkar000/Mesh/internal/cli"
// )

// func main() {
// 	cli.Execute()
// }