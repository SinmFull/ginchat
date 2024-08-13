package models

import (
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"sync"

	"net/http"

	"github.com/gorilla/websocket"
	"gopkg.in/fatih/set.v0"
	"gorm.io/gorm"
)

type Message struct {
	gorm.Model
	FromId   int64  //发送着
	TargetId int64  //接受者
	Type     int    //消息类型 群聊 私聊 广播
	Media    int    //消息类型 文字 图片 音频
	Content  string //消息内容
	Pic      string
	Url      string
	Desc     string
	Amount   int //其他数字统计

}

func (table *Message) TableName() string {
	return "message"
}

type Node struct {
	Conn      *websocket.Conn
	DataQueue chan []byte
	GroupSets set.Interface
}

// 映射关系
var clientMap map[int64]*Node = make(map[int64]*Node, 0)

// 读写锁
var rwLocker sync.RWMutex

func Chat(writer http.ResponseWriter, request *http.Request) {
	// 1.获取参数并校验token 等合法性
	query := request.URL.Query()
	Id := query.Get("userId")
	userId, _ := strconv.ParseInt(Id, 10, 64)
	//token := query.Get("token")
	//msgType := query.Get("type")
	//targetId := query.Get("targetId")
	//context := query.Get("context")

	isValid := true //checkToken()
	conn, err := (&websocket.Upgrader{
		//	token 校验
		CheckOrigin: func(r *http.Request) bool {
			return isValid
		},
	}).Upgrade(writer, request, nil)

	if err != nil {
		fmt.Println(err)
	}

	// 2. 获取连接
	node := &Node{
		Conn:      conn,
		DataQueue: make(chan []byte, 50),
		GroupSets: set.New(set.ThreadSafe),
	}

	// 3.用户关系

	//4.userId跟node绑定 并加锁
	rwLocker.Lock()
	clientMap[userId] = node
	rwLocker.Unlock()

	//5.完成发送的逻辑
	go sendProc(node)
	//6.完成接受逻辑
	go recvProc(node)
	fmt.Println("sendMsg >>>> userId: ", userId)
	sendMsg(userId, []byte("欢迎进入聊天室"))
}

func sendProc(node *Node) {
	for {
		select {
		case data := <-node.DataQueue:
			fmt.Println("[ws] sendProc >>>> msg: ", string(data))
			err := node.Conn.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	}
}

func recvProc(node *Node) {
	for {
		_, data, err := node.Conn.ReadMessage()
		if err != nil {
			fmt.Println(err)
			return
		}
		broadMsg(data)
		fmt.Println("[ws] recvProc <<<<< ", string(data))
	}
}

var udpsenChan chan []byte = make(chan []byte, 1024)

func broadMsg(data []byte) {
	udpsenChan <- data
}

func init() {
	go udpSendProc()
	go udpRecvProc()
	fmt.Println("init goroutine: ")
}

// 完成udp数据发送协程
func udpSendProc() {
	con, err := net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   net.IPv4(10, 249, 155, 193),
		Port: 3000,
	})
	defer con.Close()
	if err != nil {
		fmt.Println(err)
	}

	for {
		select {
		case data := <-udpsenChan:
			fmt.Println("udpSendProc data : ", string(data))
			_, err := con.Write(data)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	}
}

// udp数据接受协程
func udpRecvProc() {
	con, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.IPv4zero,
		Port: 3000,
	})
	if err != nil {
		fmt.Println(err)
	}
	defer con.Close()
	for {
		var buf [512]byte
		n, err := con.Read(buf[0:])
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("udpRecvProc data : ", string(buf[0:n]))
		dispatch(buf[0:n])
	}
}

// 后端调度逻辑
func dispatch(data []byte) {
	msg := Message{}
	err := json.Unmarshal(data, &msg)
	if err != nil {
		fmt.Println(err)
		return
	}
	switch msg.Type {
	case 1: //私信
		fmt.Println("dispatch data : ", string(data))
		sendMsg(int64(msg.TargetId), data)
		// case 2: //群发
		// 	sendGroupMsg()
		// case 3: //广播
		// 	sendAllMsg()
		// case 4:
	}
}

func sendMsg(userId int64, msg []byte) {
	fmt.Println("sendMsg >> userId:", userId, "msg:", string(msg))
	rwLocker.RLock()
	node, ok := clientMap[userId]
	rwLocker.RUnlock()
	if ok {
		node.DataQueue <- msg
	}
}
