package utils

import (
	"encoding/json"
	"log"
	"os"
)

type Config struct {
	LiveUrl string `json:"live_url"`
	AuthKey string `json:"auth_key"`
}

var CONFIG Config

func init() {
	// 打开文件
	file, err := os.Open("config.json")
	if err != nil {
		log.Fatalln("配置文件读取失败")
	}
	// 关闭文件
	defer file.Close()
	//NewDecoder创建一个从file读取并解码json对象的*Decoder，解码器有自己的缓冲，并可能超前读取部分json数据。
	decoder := json.NewDecoder(file)
	//Decode从输入流读取下一个json编码值并保存在v指向的值里
	err = decoder.Decode(&CONFIG)
	if err != nil {
		log.Fatalln("配置文件转换失败")
	}
}
