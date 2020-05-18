package main

import (
	"chatroom/common"
	"chatroom/utils"
	"encoding/base64"
	"encoding/json"
	"errors"
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

var rooms = make(map[string]*common.RoomInfo)

func main() {
	http.Handle("/", http.FileServer(http.Dir("static")))

	http.HandleFunc("/key", handleKey)
	http.HandleFunc("/auth", handleAuth)
	http.HandleFunc("/ws", handleConnections)

	log.Println("http server started on :8000")
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
func handleKey(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	room := r.Form.Get("room")
	username := r.Form.Get("username")
	//liveUrl := r.Form.Get("liveUrl")
	ret := new(common.Ret)
	if  room == "" || username == "" {
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

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	r.ParseForm()
	key := r.Form.Get("auth")
	p, err := parseKey(key)
	if err != nil || p.Id == "" ||p.Room == "" || p.Username == "" || p.LiveUrl == "" {
		ws.Close()
		return
	}
	roomInfo, ok := rooms[p.Room]
	if (!ok) {
		roomInfo = common.NewRoom()
		rooms[p.Room] = roomInfo
		go handleMessages(roomInfo)
	}
	roomInfo.Clients[ws] = p.Username
	/*if p.Id == "" {
		p.Id = common.CreateRandomString(15)
	}

	initMsg := common.ConnectMessage(p)
	err = ws.WriteJSON(initMsg)
	if err != nil {
		log.Printf("error: %v", err)
	}*/

	onlineMessage(roomInfo, p.Username)
	go listenMessage(p.Id, p.Username, ws, roomInfo)
}

/**
上线消息
 */
func onlineMessage(roomInfo *common.RoomInfo, username string) {
	msg := common.SystemMessage(username + "  进入了房间")
	roomInfo.Broadcast <- msg
}

/**
离线消息
 */
func offlineMessage(roomInfo *common.RoomInfo, username string) {
	msg := common.SystemMessage(username + "  离开了房间")
	roomInfo.Broadcast <- msg
}

func listenMessage(id string, username string, ws *websocket.Conn, roomInfo *common.RoomInfo) {
	for {
		_, message, err := ws.ReadMessage()
		if err != nil {
			log.Printf("error: %v", err)
			delete(roomInfo.Clients, ws)
			offlineMessage(roomInfo, username)
			break
		}
		msg := common.UserMessage(username, string(message))
		msg.Id = id
		roomInfo.Broadcast <- msg
	}
}

func handleMessages(roomInfo *common.RoomInfo) {
	for {
		msg := <-roomInfo.Broadcast
		for client := range roomInfo.Clients {
			err := client.WriteJSON(msg)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				username := roomInfo.Clients[client]
				delete(roomInfo.Clients, client)
				offlineMessage(roomInfo, username)
			}
		}
	}
}