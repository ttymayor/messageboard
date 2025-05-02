package main

import (
	"messageboard/routers"
)

func main() {
	// Initialize the router
	r := routers.InitRouter()

	// Start the server on port 8080
	if err := r.Run(":8080"); err != nil {
		panic(err)
	}
}
