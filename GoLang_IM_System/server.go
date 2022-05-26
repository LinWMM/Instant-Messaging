package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	Ip   string
	Port int

	OnlineMap map[string]*User
	mapLock   sync.RWMutex

	Message chan string
}

//创建server的接口
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server
}

//监听Message广播消息，发消息给全部的user
func (this *Server) ListenMessage() {
	for {
		msg := <-this.Message

		this.mapLock.Lock() //加锁防止有新用户加入
		for _, cli := range this.OnlineMap {
			cli.C <- msg
		}
		this.mapLock.Unlock()
	}
}

//将用户上线的消息发到Server的channel
func (this *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	this.Message <- sendMsg
}

//处理业务
func (this *Server) Handler(conn net.Conn) {
	//fmt.Println("连接建立成功.")

	//将用户加到OnlineMap中
	user := NewUser(conn, this)
	user.Online()

	//监听用户是否活跃的channel
	isLive := make(chan bool)

	//接受客户端传递的消息
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				user.Offline()
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("Server read user err:", err)
				return
			}
			//读取用户消息，去处'\n'
			msg := string(buf[:n-1])

			//将得到的消息广播
			user.DoMessage(msg)

			//活跃
			isLive <- true
		}
	}()

	//定时器，踢不活跃用户
	for {
		select {
		case <-isLive:
			//当前用户时活跃的，重置定时器
			//不做任何事情，更新定时器
			//会判断下一个case，达到更新的目的
		case <-time.After(60 * time.Second): //go的定时器实际上是一个channel，执行到后自动重置
			//进来代表已经超时，将用户提出
			user.SendMsg("你被踢了。\n")

			this.mapLock.Lock()
			delete(this.OnlineMap, user.Name)
			this.mapLock.Unlock()
			//销毁gorouting
			close(user.C)

			//关闭连接
			conn.Close()

			//关闭isLive管道
			close(isLive)

			//销毁当前handler
			return
		}
	}
}

//启动服务
func (this *Server) Start() {
	//socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port)) //"127.0.0.1:888"
	if err != nil {
		fmt.Println("net.listen err!")
		return
	}
	//记得关闭listener
	defer listener.Close()

	//启动监听Message的go routing
	go this.ListenMessage()

	for {
		//accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("accept error!")
			continue
		}

		//do handler
		go this.Handler(conn)
	}

	//close listen socket
}
