package main


func main() {
	serve := NewServer("127.0.0.1", 8888)
	serve.Start()
}
