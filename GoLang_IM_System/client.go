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
	flag       int //客户端模式
}

func NewClient(serverIp string, serverPort int) *Client {
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		flag:       999,
	}

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if err != nil {
		fmt.Println("net.Dail err :", err)
		return nil
	}

	client.conn = conn
	return client
}

//接受服务器返回的消息
func (client *Client) DealResponse() {
	io.Copy(os.Stdout, client.conn) //类似重定向，和下面三行等价
	/*for{
		buf := make(chan string)
		client.conn.Read([]byte(<-buf))
		fmt.Println(buf)
	}*/
}

func (client *Client) menu() bool {
	var flag int
	fmt.Println("1.公聊模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.更新用户名")
	fmt.Println("0.退出")
	fmt.Scanln(&flag)

	if flag >= 0 && flag <= 3 {
		client.flag = flag
		return true
	} else {
		fmt.Println("<<<<请输入合法数字<<<<")
		return false
	}
}

func (client *Client) SelectUsers() {
	sendMsg := "who\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn Write err:", err)
		return
	}
}

func (client *Client) PrivateChat() {
	client.SelectUsers()
	var remoteName string
	fmt.Println(">>>>请输入[用户名],exit退出:")
	fmt.Scanln(&remoteName)
	for remoteName != "exit" {

		var chatMsg string
		fmt.Println("请输入消息面内容,exit退出:")
		fmt.Scanln(&chatMsg)
		for chatMsg != "exit" {

			if len(chatMsg) != 0 {
				sendMsg := "to|" + remoteName + "|" + chatMsg + "\n"
				_, err := client.conn.Write([]byte(sendMsg))
				if err != nil {
					fmt.Println("conn.Write err:", err)
					break
				}
			}
			chatMsg = ""
			fmt.Println("请输入消息面内容,exit退出:")
			fmt.Scanln(&chatMsg)
		}

		remoteName = ""
		client.SelectUsers()
		fmt.Println(">>>>请输入[用户名],exit退出:")
		fmt.Scanln(&remoteName)
	}
}

func (client *Client) PublicChat() {
	//提示用户输入消息
	var chatMsg string
	fmt.Println(">>>>请输入聊天内容,exit退出.")
	fmt.Scanln(&chatMsg)

	for chatMsg != "exit" {
		if len(chatMsg) != 0 {
			sendMsg := chatMsg + "\n"
			_, err := client.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("conn.Write err:", err)
				break
			}
		}

		chatMsg = ""
		fmt.Println(">>>>请输入聊天内容，exit退出.")
		fmt.Scanln(&chatMsg)
	}
}

func (client *Client) UpdateName() bool {
	fmt.Println(">>>请输入用户名:")
	fmt.Scanln(&client.Name)

	sendMsg := "rename|" + client.Name + "\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write err:", err)
		return false
	}
	return true
}

func (client *Client) Run() {
	for client.flag != 0 {
		for client.menu() != true {
		}

		switch client.flag {
		case 1:
			//fmt.Println("公聊模式选择...")
			client.PublicChat()
			break
		case 2:
			//fmt.Println("私聊模式选择...")
			client.PrivateChat()
			break
		case 3:
			//fmt.Println("更新用户名选择...")
			client.UpdateName()
			break
		}
	}
}

var serverIp string
var serverPort int

// ./client -ip 127.0.0.1
func init() { //把形参绑定到flag包中
	//./client -h会提示说明
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "设置服务器ip地址,默认是(127.0.0.1)")
	flag.IntVar(&serverPort, "port", 8888, "服务器端口号，默认是(8888)")
}

func main() {
	//命令行解析
	flag.Parse()

	client := NewClient(serverIp, serverPort)
	if client == nil {
		fmt.Println(">>>>>> 连接服务器失败...")
		return
	}

	//读消息
	go client.DealResponse()

	fmt.Println(">>>>>> 连接服务器成功.")

	//发消息，启动服务端业务
	client.Run()
}
