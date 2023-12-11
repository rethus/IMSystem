package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	conn       net.Conn
	flag       int
}

// 创建一个客户端
func NewClient(serverIp string, serverPort int) *Client {
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		flag:       999,
	}

	// 连接服务器
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if err != nil {
		fmt.Println("net.Dial error:", err)
		return nil
	}

	client.conn = conn

	return client
}

// 将服务端的消息打印到客户端
func (client *Client) DealResponse() {
	// 一旦 client.conn 有数据,copy 到 stdout 上, 永久阻塞监听
	io.Copy(os.Stdout, client.conn)
}

// 展示功能目录
func (client *Client) menu() bool {
	var flag int
	fmt.Println("1. public chat")
	fmt.Println("2. private chat")
	fmt.Println("3. update username")
	fmt.Println("0. quit")

	fmt.Scanln(&flag)

	if 0 <= flag && flag <= 3 {
		client.flag = flag
		return true
	} else {
		fmt.Println(">>>>>>>>> please input a legal number ...")
		return false
	}
}

// 公聊模式
func (client *Client) PublicChat() {
	var chatMsg string

	fmt.Println(">>>>>>>>>> please input chat content, \"exit\" for quit")
	fmt.Scanln(&chatMsg)

	for chatMsg != "exit" {
		// 发送给服务器
		if len(chatMsg) != 0 {
			sendMsg := chatMsg + "\n"
			_, err := client.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("conn Write err:", err)
				break
			}
		}
		chatMsg = ""
		fmt.Println(">>>>>>>>>> please input chat content, \"exit\" for quit")
		fmt.Scanln(&chatMsg)
	}
}

// 查询在线用户
func (client *Client) SelectUsers() {
	sendMsg := "who\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn Write err:", err)
		return
	}
}

// 私聊模式
func (client *Client) PrivateChat() {
	var remoteName string
	var chatMsg string

	client.SelectUsers()
	fmt.Println(">>>>>>>>>> please input chat userName, \"exit\" for quit")
	fmt.Scanln(&remoteName)

	for remoteName != "exit" {
		fmt.Println(">>>>>>>>>> please input chat content, \"exit\" for quit")
		fmt.Scanln(&chatMsg)

		for chatMsg != "exit" {
			if len(chatMsg) > 0 {
				sendMsg := "to|" + remoteName + "|" + chatMsg + "\n"
				_, err := client.conn.Write([]byte(sendMsg))
				if err != nil {
					fmt.Println("conn Write err:", err)
					break
				}
			}
			chatMsg = ""
			fmt.Println(">>>>>>>>>> please input chat content, \"exit\" for quit")
			fmt.Scanln(&chatMsg)
		}

		client.SelectUsers()
		fmt.Println(">>>>>>>>>> please input chat userName, \"exit\" for quit")
		fmt.Scanln(&remoteName)
	}
}

// 修改当前客户端用户名
func (client *Client) UpdateName() bool {
	fmt.Println(">>>>>>>>>> please input username:")
	fmt.Scanln(&client.Name)

	sendMsg := "rename|" + client.Name + "\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write err:", err)
		return false
	}
	return true
}

// 运行
func (client *Client) Run() {
	for client.flag != 0 {
		for client.menu() != true {

		}
		// 根据不同模式处理不同业务
		switch client.flag {
		case 1:
			// 公聊模式
			client.PublicChat()
			break
		case 2:
			// 私聊模式
			client.PrivateChat()
			break
		case 3:
			// 更新用户名
			client.UpdateName()
			break
		}
	}
}

var serverIp string
var serverPort int

// ./client -ip 127.0.0.1
func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "设置服务器ip地址（默认 127.0.0.1）")
	flag.IntVar(&serverPort, "port", 8888, "设置服务器端口（默认 8888）")
}

func main() {
	// client.exe -ip 127.0.0.1 -port 8888
	// 命令行解析
	flag.Parse()

	client := NewClient(serverIp, serverPort)
	if client == nil {
		fmt.Println(">>>>>>>>>> link to server failed ...")
		return
	}

	// 单独开启一个 goroutine 处理 server 的回执消息

	go client.DealResponse()

	fmt.Println(">>>>>>>>>> link to server succeed ...")

	// 启动客户端业务
	client.Run()
}
