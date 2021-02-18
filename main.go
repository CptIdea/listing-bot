package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/CptIdea/go-vk-api-2"
	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func init() {
	logFile, err := os.OpenFile("list.log", os.O_CREATE|os.O_RDWR, 0777)
	if err != nil {
		log.Fatal(err)
	}

	logFile.Truncate(0)

	log.SetOutput(logFile)

	if err := godotenv.Load(); err != nil {
		log.Fatal("No .env file found")
	}

	var exist bool

	dsn, exist = os.LookupEnv("DSN")
	if !exist {
		log.Fatal(fmt.Errorf(".env DSN not exist"))
	}

	token, exist = os.LookupEnv("VK_TOKEN")
	if !exist {
		log.Fatal(fmt.Errorf(".env VK_TOKEN not exist"))
	}

	version, exist = os.LookupEnv("VK_VERSION")
	if !exist {
		log.Fatal(fmt.Errorf(".env VK_VERSION not exist"))
	}

	rawGID, exist := os.LookupEnv("VK_GROUP")
	if !exist {
		log.Fatal(fmt.Errorf(".env VK_GROUP not exist"))
	}

	groupID, err = strconv.Atoi(rawGID)
	if err != nil {
		log.Fatal(fmt.Errorf("failed convert VK_GROUP"))
	}
}

var (
	dsn     = ""
	token   = ""
	groupID = 0
	version = ""
)

func main() {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Error %s", err.Error())
	}

	s := vk.NewSession(token, version)
	for {
		nUpd, err := s.UpdateCheck(groupID)
		if err != nil {
			log.Println(err)
		}
		for _, upd := range nUpd.Updates {
			if upd.Object.MessageNew.Text == "/список" {
				newList := List{}
				db.Create(&newList)

				kb := vk.GenerateKeyBoard("!pЗаписаться в список", false, false)
				kb.Buttons[0][0].Action.Payload = strconv.Itoa(newList.Id)
				s.SendKeyboard(upd.Object.MessageNew.PeerId, kb, "Создан новый список!")
			}
			if strings.Contains(upd.Object.MessageNew.Text, "Записаться в список") && upd.Object.MessageNew.Payload != "" {
				l := List{}
				n, err := strconv.Atoi(upd.Object.MessageNew.Payload)
				if err != nil {
					log.Println(err)
					continue
				}

				db.First(&l, n)

				if l.Users == "" {
					l.Users = strconv.Itoa(upd.Object.MessageNew.FromId)
				} else {
					l.Users += "," + strconv.Itoa(upd.Object.MessageNew.FromId)
				}

				ids := []int{}
				for _, v := range strings.Split(l.Users, ",") {
					n, err := strconv.Atoi(v)
					if err != nil {
						log.Println(err)
						continue
					}
					ids = append(ids, n)
				}

				usrs, err := s.GetUsersInfo(ids)
				if err != nil {
					log.Println(err)
				}

				textToSend := "Ты успешно записался!\n\nВот список:\n"

				for i, v := range usrs {
					textToSend += fmt.Sprintf("%d. %s %s\n", i+1, v.FirstName, v.LastName)
				}

				_, err = s.SendMessage(upd.Object.MessageNew.PeerId, textToSend)
				if err != nil {
					log.Println(err)
				}

				db.Save(&l)
			}
			if upd.Object.MessageNew.Text == "/отмена" {
				_, err = s.SendKeyboard(upd.Object.MessageNew.PeerId, vk.GenerateEmptyKeyBoard(""), "Удаляю")
				if err != nil {
					log.Println(err)
				}
			}
		}
	}
}

type List struct {
	Id    int
	Users string
}
