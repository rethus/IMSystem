package main

func main() {
	// http://127.0.0.1:8888/
	server := NewServer("127.0.0.1", 8888)
	server.Start()
}
