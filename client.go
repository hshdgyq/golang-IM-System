package main

import (
	"fmt"
	"net"
	"flag"
	"io"
	"os"
)

type Client struct {
	ServerIp string
	ServerPort int
	Name string
	conn net.Conn
	flag int //当前客户端选择的模式
}

func NewClient(serverIp string, serverPort int) *Client {
	//创建客户端对象
	client := &Client{
		ServerIp: serverIp,
		ServerPort: serverPort,
		flag: 999,
	}

	//链接server
	conn, err :=net.Dial("tcp", fmt.Sprintf("%s:%d",serverIp,serverPort))
	if err != nil {
		fmt.Println("net.Dial error\n")
		return nil
	}

	client.conn = conn

	//返回对象
	return client
}

//处理server回应的消息,直接显示
func (client *Client) DealResponse() {
	//一旦client.conn有数据，就copy到stdout标准输出上，永久阻塞监听
	io.Copy(os.Stdout, client.conn)
}

//显示所有在线用户的方法
func (client *Client) ShowOnlineUser() {
	sendMsg := "who\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn Write err:", err)
		return
	}
}

//公聊模式方法
func (client *Client) PublicChat() {
	var chatMsg string
	//提示用户输入消息
	fmt.Println(">>>>>>公共聊天室，请注意言行哦(exit表示退出)<<<<<<<<<")
	fmt.Scanln(&chatMsg)
	for chatMsg != "exit" {
		//发给服务器

		if len(chatMsg) != 0 {
			sendMsg := chatMsg + "\n"
			_,err := client.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("conn Write err:", err)
				break
			}
		}

		chatMsg = ""
		fmt.Scanln(&chatMsg)
	}
	
}

//私聊模式方法
func (client *Client) PrivateChat() {
	var remoteName string
	var chatMsg string
	fmt.Println("请输入聊天对象[用户名]，exit退出")
	fmt.Scanln(&remoteName)
	for remoteName != "exit" {
		fmt.Println("这是您与",remoteName,"的私聊哦")
		fmt.Scanln(&chatMsg)

		for chatMsg != "exit" {
			if len(chatMsg) != 0 {
				sendMsg := "to|" + remoteName + "|" + chatMsg + "\n"
				_,err := client.conn.Write([]byte(sendMsg))
				if err != nil {
					fmt.Println("conn Write err:", err)
					break
				}
			}
	
			chatMsg = ""
			fmt.Scanln(&chatMsg)
		}

		fmt.Println("请输入聊天对象[用户名]，exit退出\n")
		fmt.Scanln(&remoteName)
	}

}

//更新用户名的方法
func (client *Client) UpdateName() bool {
	fmt.Println(">>>>>>请输入用户名<<<<<<<<")
	fmt.Scanln(&client.Name)

	sendMsg := "rename|" + client.Name +"\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write err:", err)
		return false
	}

	return true
}

//菜单
func (client *Client) menu() bool {
	var flag int
	fmt.Println("1.公聊模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.更新用户名")
	fmt.Println("4.查看当前在线用户")
	fmt.Println("0.退出")

	fmt.Scanln(&flag)

	if flag >= 0 && flag <=4 {
		client.flag = flag
		return true
	}else{
		fmt.Println(">>>>>>>>>>>请输入合法数字<<<<<<<<<<")
		return false
	}
}

//业务执行
func (client *Client) Run() {
	for client.flag !=0 {
		for client.menu() != true {

		}
		//根据不同模式处理不同业务
		switch client.flag {
		case 1:
			//公聊模式
			fmt.Println("公聊模式选择...")
			client.PublicChat()
			break
		case 2:
			//私聊模式
			fmt.Println("私聊模式选择...")
			client.PrivateChat()
			break
		case 3:
			//更新用户名
			fmt.Println("更新用户名选择...")
			client.UpdateName()
			break
		case 4:
			//查看当前在线用户
			client.ShowOnlineUser()
			break
		}
	}
}


var serverIp string
var serverPort int

func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "设置服务器地址(默认是127.0.0.1)")
	flag.IntVar(&serverPort, "port", 8888, "设置服务器端口(默认是8888)")

}


func main() {
	//命令行解析
	flag.Parse()

	client :=NewClient(serverIp,serverPort)
	if client == nil {
		fmt.Println(">>>>>>> 连接服务器失败......")
		return
	}

	//单独开启goroutine处理server的回执
	go client.DealResponse()

	fmt.Println(">>>>>>>> 连接服务器成功......")

	//启动客户端业务
	client.Run()
}