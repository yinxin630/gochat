package main

import (
	"fmt"
	"os"
	"strconv"
	"net"
	"bufio"
	"math/rand"
	"time"
)

//数据包头，标识数据内容
var reflectString = map[string]string {
	"连接": 		"connect  :",
	"在线": 		"online   :",
	"聊天": 		"chat     :",
	"在线用户": 	"get      :",
}

//服务器端口
const CLIENTPORT = 1616
//数据缓冲区
const BUFFSIZE = 1024
var buff = make([]byte, BUFFSIZE)

//错误消息处理
func HandleError(err error) {
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(2)
	}
}
//发送消息
func SendMessage(udpConn *net.UDPConn) {
	scaner := bufio.NewScanner(os.Stdin)

	for scaner.Scan() {
		if scaner.Text() == "exit" {
			return
		}
		udpConn.Write([]byte(scaner.Text()))
	}
}
//接收消息
func HandleMessage(udpListener *net.UDPConn) {
	for {
		n, _, err := udpListener.ReadFromUDP(buff)
		HandleError(err)

		if n > 0 {
			fmt.Println(string(buff[:n]))
		}
	}
}
/*
func AnalyzeMessage(buff []byte, len int) ([]string) {
    analMsg := make([]string, 0)
    strNow := ""
    for i := 0; i < len; i++ {
        if string(buff[i:i + 1]) == ":" {
            analMsg = append(analMsg, strNow)
            strNow = ""
        } else {
            strNow += string(buff[i:i + 1])
        }
    }
    analMsg = append(analMsg, strNow)
    return analMsg
}*/
//发送心跳包
func SendOnlineMessage(udpConn *net.UDPConn) {
	for {
		//每间隔1s向服务器发送一次在线信息
		udpConn.Write([]byte(reflectString["在线"]))
		sleepTimer := time.NewTimer(time.Second)
		<- sleepTimer.C
	}
}

func main() {
	//判断命令行参数，参数应该为服务器ip
	if len(os.Args) != 2 {
		fmt.Println("程序命令行参数错误！")
		os.Exit(2)
	}
	//获取ip
	host := os.Args[1]

	//udp地址
	udpAddr, err := net.ResolveUDPAddr("udp4", host + ":" + strconv.Itoa(CLIENTPORT))
	HandleError(err)

	//udp连接
	udpConn, err := net.DialUDP("udp4", nil, udpAddr)
	HandleError(err)

	//本地监听端口
	newSeed := rand.NewSource(int64(time.Now().Second()))
	newRand := rand.New(newSeed)
	randPort := newRand.Intn(30000) + 10000

	//本地监听udp地址
	udpLocalAddr, err := net.ResolveUDPAddr("udp4", "127.0.0.1:" + strconv.Itoa(randPort))
	HandleError(err)

	//本地监听udp连接
	udpListener, err := net.ListenUDP("udp4", udpLocalAddr)
	HandleError(err)

	//fmt.Println("监听", randPort, "端口")

	//用户昵称
	userName := ""
	fmt.Printf("请输入昵称：")
	fmt.Scanf("%s", &userName)
	
	//向服务器发送连接信息（昵称+本地监听端口）
	udpConn.Write([]byte(reflectString["连接"] + userName + ":" + strconv.Itoa(randPort)))

	//关闭端口
	defer udpConn.Close()
	defer udpListener.Close()

	//发送心跳包
	go SendOnlineMessage(udpConn)
	//接收消息
	go HandleMessage(udpListener)

	command := ""
	
	for {
		//获取命令
		fmt.Scanf("%s", &command)
		switch command {
			case "chat" :
				people := ""
				//fmt.Printf("请输入对方昵称：")
				fmt.Scanf("%s", &people)
				//向服务器发送聊天对象信息
				udpConn.Write([]byte(reflectString["聊天"] + people))
				//进入会话
				SendMessage(udpConn)
				//退出会话
				fmt.Println("退出与" + people + "的会话")
			case "get" :
				//请求在线情况信息
				udpConn.Write([]byte(reflectString["在线用户"]))
		}
	}
}
