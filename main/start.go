package main

import (
	"chatroom/common"
	"chatroom/utils"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net/http"
)

// Configure the upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

//存放所有房间
var rooms = make(map[string]*common.RoomInfo)

func main() {
	var port string
	flag.StringVar(&port, "port", "8000", "端口号，默认值 8000")
	flag.Parse()
	http.Handle("/", http.FileServer(http.Dir("static")))

	http.HandleFunc("/key", handleKey)
	http.HandleFunc("/auth", handleAuth)
	http.HandleFunc("/members", handleMembers)
	http.HandleFunc("/ws", handleConnections)

	log.Println("http server started on :", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

/**
获取房间成员列表
 */
func handleMembers(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	room := r.Form.Get("room")
	ret := new(common.Ret)
	if room == "" {
		ret.Code = -1
		ret.Message = "错误的参数"
	} else {
		roomInfo, ok := rooms[room]
		if (ok) {
			clients := roomInfo.Clients
			members := make(map[string]string)
			for _, v := range clients {
				members[v.Id] = v.Username
			}
			ret.Data = members
		}
	}
	ret_json, _ := json.Marshal(ret)
	io.WriteString(w, string(ret_json))
}

/**
获取key
 */
func handleKey(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	room := r.Form.Get("room")
	username := r.Form.Get("username")
	//liveUrl := r.Form.Get("liveUrl")
	ret := new(common.Ret)
	if room == "" || username == "" {
		ret.Code = -1
		ret.Message = "错误的key"
	} else {
		p := common.Content{
			Room:     room,
			Username: username,
			LiveUrl:  fmt.Sprintf("http://10.1.125.49:8930/live/%s.flv", room),
		}
		b, _ := json.Marshal(p)
		result, err := utils.AesEncrypt(b, []byte(utils.CONFIG.AuthKey))
		if err != nil {
			ret.Code = -1
			ret.Message = "鉴权失败"
		}
		ret.Data = base64.StdEncoding.EncodeToString(result)
	}
	ret_json, _ := json.Marshal(ret)
	io.WriteString(w, string(ret_json))
}

/**
鉴权
 */
func handleAuth(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	key := r.Form.Get("key")
	p, err := parseKey(key)
	ret := new(common.Ret)
	if err != nil || p.Room == "" || p.Username == "" || p.LiveUrl == "" {
		ret.Code = -1
		ret.Message = "错误的key"
	} else {
		if p.Id == "" {
			p.Id = common.CreateRandomString(15)
		}
		b, _ := json.Marshal(p)
		result, err := utils.AesEncrypt(b, []byte(utils.CONFIG.AuthKey))
		if err != nil {
			ret.Code = -1
			ret.Message = "鉴权失败"
		}
		p.WsKey = base64.StdEncoding.EncodeToString(result)
		data_json, _ := json.Marshal(p)
		ret.Data = string(data_json)
	}
	ret_json, _ := json.Marshal(ret)
	io.WriteString(w, string(ret_json))
}

/**
解析key
 */
func parseKey(key string) (content common.Content, err error) {
	var p common.Content
	if key == "" {
		return p, errors.New("参数为空")
	}
	ak, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		log.Println(err.Error())
		return p, errors.New("参数错误")
	}
	parseString, err := utils.AesDecrypt(ak, []byte(utils.CONFIG.AuthKey))
	if err != nil {
		log.Println(err.Error())
		return p, errors.New("参数错误")
	}
	err = json.Unmarshal(parseString, &p)
	if err != nil {
		log.Println(err.Error())
		return p, errors.New("参数错误")
	}
	p.LiveUrl = fmt.Sprintf(utils.CONFIG.LiveUrl, p.Room)
	return p, nil
}

/**
websocket连接
 */
func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	r.ParseForm()
	key := r.Form.Get("auth")
	p, err := parseKey(key)
	if err != nil || p.Id == "" || p.Room == "" || p.Username == "" || p.LiveUrl == "" {
		ws.Close()
		return
	}
	roomInfo, ok := rooms[p.Room]
	if (!ok) {
		roomInfo = common.NewRoom()
		roomInfo.Roomname = p.Room
		rooms[p.Room] = roomInfo
		go handleMessages(roomInfo)
	}
	roomInfo.Clients[ws] = p

	onlineMessage(roomInfo, p.Username, p.Id)
	go listenMessage(p.Id, p.Username, ws, roomInfo)
}

/**
上线消息
 */
func onlineMessage(roomInfo *common.RoomInfo, username string, id string) {
	msg := common.SystemMessage(username + "  进入了房间")
	msg.Id = id
	msg.Username = username
	msg.Action = common.JOIN_ACTION
	roomInfo.Broadcast <- msg
}

/**
离线消息
 */
func offlineMessage(roomInfo *common.RoomInfo, username string, id string) {
	msg := common.SystemMessage(username + "  离开了房间")
	msg.Id = id
	msg.Username = username
	msg.Action = common.LEAVE_ACTION
	roomInfo.Broadcast <- msg
}

/**
监听客户端
 */
func listenMessage(id string, username string, ws *websocket.Conn, roomInfo *common.RoomInfo) {
	for {
		_, message, err := ws.ReadMessage()
		if err != nil {
			log.Printf("error: %v", err)
			delete(roomInfo.Clients, ws)
			offlineMessage(roomInfo, username, id)
			break
		}
		msg := common.UserMessage(username, string(message))
		msg.Id = id
		roomInfo.Broadcast <- msg
	}
}

/**
监听房间消息
 */
func handleMessages(roomInfo *common.RoomInfo) {
	for {
		msg := <-roomInfo.Broadcast
		for client := range roomInfo.Clients {
			err := client.WriteJSON(msg)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				p := roomInfo.Clients[client]
				delete(roomInfo.Clients, client)
				offlineMessage(roomInfo, p.Username, p.Id)
			}
		}
		if len(roomInfo.Clients) == 0 {
			//房间里面没人的时候删除房间信息
			delete(rooms, roomInfo.Roomname)
		}
	}
}
