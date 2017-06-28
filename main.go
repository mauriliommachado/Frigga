package main

import (
	"./server"
)

func main() {
	server.Start(server.ServerProperties{Address: "/goroom", Port: "8081"})
}
