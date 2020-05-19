package common

import (
	"bytes"
	"crypto/rand"
	"github.com/gorilla/websocket"
	"math/big"
	"time"
)

const (
	USER_MSG_TYPE   = "ut"    //用户消息类型
	SYSTEM_MSG_TYPE = "st"    //系统消息类型
	SAY_ACTION      = "say"   //发言行为
	LEAVE_ACTION    = "leave" //离开行为
	JOIN_ACTION     = "join"  //加入行为
)

/**
消息
 */
type Message struct {
	Id       string `json:"id"`
	Action   string `json:"action"`
	Username string `json:"username"`
	Message  string `json:"message"`
	Time     string `json:"time"`
	Mtype    string `json:"type"`
}

/**
用户消息
 */
func UserMessage(username string, message string) *Message {
	return &Message{
		Username: username,
		Message:  message,
		Time:     time.Now().Format("2006-01-02 15:04:05"),
		Mtype:    USER_MSG_TYPE,
		Action:   SAY_ACTION,
	}
}

/**
系统消息
 */
func SystemMessage(message string) *Message {
	return &Message{
		Message: message,
		Time:    time.Now().Format("2006-01-02 15:04:05"),
		Mtype:   SYSTEM_MSG_TYPE,
	}
}

/**
首次连接响应消息
 */
/*func ConnectMessage(content Content) *Message {
	pj, err := json.Marshal(content)
	if err != nil {
		return &Message{
			Mtype: initMsgType,
		}
	}
	return &Message{
		Message: string(pj),
		Time:    time.Now().Format("2006-01-02 15:04:05"),
		Mtype:   initMsgType,
	}
}*/

type RoomInfo struct {
	Clients   map[*websocket.Conn]Content
	Broadcast chan *Message
	Roomname  string
}

func NewRoom() *RoomInfo {
	return &RoomInfo{
		Clients:   make(map[*websocket.Conn]Content),
		Broadcast: make(chan *Message),
	}
}

/**
鉴权信息
 */
type Content struct {
	Id       string `json:"id"`
	Room     string `json:"room"`
	Username string `json:"username"`
	LiveUrl  string `json:"live_url"`
	WsKey    string `json:"ws_key"`
}

type Ret struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func CreateRandomString(len int) string {
	var container string
	var str = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
	b := bytes.NewBufferString(str)
	length := b.Len()
	bigInt := big.NewInt(int64(length))
	for i := 0; i < len; i++ {
		randomInt, _ := rand.Int(rand.Reader, bigInt)
		container += string(str[randomInt.Int64()])
	}
	return container
}
