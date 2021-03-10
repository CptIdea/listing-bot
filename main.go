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
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func init() {
	var exist bool
	var err error

	if err := godotenv.Load(); err != nil {
		log.Fatal("No .env file found")
	}

	dsn, exist = os.LookupEnv("DSN")
	if !exist {
		log.Fatal(fmt.Errorf(".env DSN not exist"))
	}

	rawUseSQLite, exist := os.LookupEnv("USE_SQLITE")
	if !exist {
		log.Fatal(fmt.Errorf(".env USE_SQLITE not exist"))
	}
	if strings.ToLower(rawUseSQLite) == "true" {
		useSQLite = true
	}

	sqLite, exist = os.LookupEnv("SQLLITE")
	if !exist {
		log.Fatal(fmt.Errorf(".env SQLLITE not exist"))
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

	rawAdmins, exist := os.LookupEnv("ADMINS")
	if !exist {
		log.Fatal(fmt.Errorf(".env ADMINS not exist"))
	}

	for _, v := range strings.Split(rawAdmins, ",") {
		i, _ := strconv.Atoi(v)
		if i > 0 {
			admins = append(admins, i)
		}
	}

	logFile, err := os.OpenFile("list.log", os.O_CREATE|os.O_RDWR, 0777)
	if err != nil {
		log.Fatal(err)
	}

	logFile.Truncate(0)

	log.SetOutput(logFile)

}

var admins []int
var (
	useSQLite = false
	dsn       = ""
	sqLite    = ""
	token     = ""
	groupID   = 0
	version   = ""
)

func main() {
	var db *gorm.DB
	var err error

	if useSQLite {
		fmt.Println(1)
		db, err = gorm.Open(sqlite.Open(sqLite), &gorm.Config{})
		if err != nil {
			log.Fatalf("Error %s", err.Error())
		}

	} else {
		fmt.Println(2)
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err != nil {
			log.Fatalf("Error %s", err.Error())
		}

	}

	s := vk.NewSession(token, version)
	for {
		nUpd, err := s.UpdateCheck(groupID)
		handle(err)

		for _, upd := range nUpd.Updates {
			if strings.HasPrefix(strings.ToLower(upd.Object.MessageNew.Text), "список") && canCreate(upd.Object.MessageNew.FromId) {

				if len(strings.Split(upd.Object.MessageNew.Text, " ")) < 2 {
					s.SendMessage(upd.Object.MessageNew.PeerId, "Используйте: \"список <название>\"")
					continue
				}
				newList := list{Name: strings.Replace(upd.Object.MessageNew.Text, strings.Split(upd.Object.MessageNew.Text, " ")[0]+" ", "", 1), PeerID: upd.Object.MessageNew.PeerId}
				db.Create(&newList)

				kb := vk.GenerateKeyBoard(fmt.Sprintf("Запись %d", newList.ID), false, false)
				kb.Buttons[0][0].Action.Payload = strconv.Itoa(newList.ID)
				_, err := s.SendKeyboard(upd.Object.MessageNew.PeerId, kb, fmt.Sprintf("Создан список %q\n\nЗаписаться:\nЗапись %d", newList.Name, newList.ID))
				handle(err)

			}
			if strings.Contains(strings.ToLower(upd.Object.MessageNew.Text), "запись") {
				if len(strings.Split(upd.Object.MessageNew.Text, " ")) < 2 {

					if canCreate(upd.Object.MessageNew.FromId) {
						s.SendMessage(upd.Object.MessageNew.PeerId, "Используйте: \"запись <ID>\"")
					}

					continue
				}
				var n int
				l := list{}
				if upd.Object.MessageNew.Payload != "" {
					n, err = strconv.Atoi(upd.Object.MessageNew.Payload)
					if handle(err) {
						continue
					}
				} else {
					n, err = strconv.Atoi(strings.Split(upd.Object.MessageNew.Text, " ")[1])
					if handle(err) {
						continue
					}
				}

				db.First(&l, n)
				if l.Name == "" {
					continue
				}
				if strings.Contains(l.Users, strconv.Itoa(upd.Object.MessageNew.FromId)) {
					continue
				}

				l.Users += fmt.Sprintf("%d;", upd.Object.MessageNew.FromId)

				ids := []int{}
				for _, v := range strings.Split(l.Users, ";") {
					n, err := strconv.Atoi(v)
					if err != nil {
						log.Println(err)
						continue
					}
					ids = append(ids, n)
				}

				usrs, err := s.GetUsersInfo(ids)
				handle(err)

				textToSend := fmt.Sprintf("Ты успешно записался в %q!\n\nВот весь список:\n", l.Name)

				for i, v := range usrs {
					textToSend += fmt.Sprintf("%d. %s %s\n", i+1, v.FirstName, v.LastName)
				}

				textToSend += fmt.Sprintf("\nЗаписаться: запись %d", l.ID)

				_, err = s.SendMessage(upd.Object.MessageNew.PeerId, textToSend)
				handle(err)

				db.Save(&l)
			}
			if strings.Contains(strings.ToLower(upd.Object.MessageNew.Text), "выход") {
				if len(strings.Split(upd.Object.MessageNew.Text, " ")) < 2 {

					if canCreate(upd.Object.MessageNew.FromId) {
						s.SendMessage(upd.Object.MessageNew.PeerId, "Используйте: \"выход <ID>\"")
					}

					continue
				}

				l := list{}

				n, err := strconv.Atoi(strings.Split(upd.Object.MessageNew.Text, " ")[1])
				if handle(err) {
					continue
				}

				db.First(&l, n)

				if !strings.Contains(l.Users, strconv.Itoa(upd.Object.MessageNew.FromId)) {
					continue
				}

				l.Users = strings.Replace(l.Users, strconv.Itoa(upd.Object.MessageNew.FromId)+";", "", 1)

				ids := []int{}
				for _, v := range strings.Split(l.Users, ";") {
					n, err := strconv.Atoi(v)
					if err != nil {
						log.Println(err)
						continue
					}
					ids = append(ids, n)
				}

				usrs, err := s.GetUsersInfo(ids)
				handle(err)
				textToSend := fmt.Sprintf("Ты успешно выписался из %q!\n\nВот весь список:\n", l.Name)

				for i, v := range usrs {
					textToSend += fmt.Sprintf("%d. %s %s\n", i+1, v.FirstName, v.LastName)
				}

				_, err = s.SendMessage(upd.Object.MessageNew.PeerId, textToSend)
				handle(err)

				db.Save(&l)
			}
			if upd.Object.MessageNew.Text == "/отмена" {
				_, err = s.SendKeyboard(upd.Object.MessageNew.PeerId, vk.GenerateEmptyKeyBoard(""), "Удаляю")
				if err != nil {
					log.Println(err)
				}
			}
			if strings.HasPrefix(strings.ToLower(upd.Object.MessageNew.Text), "удалить") && canCreate(upd.Object.MessageNew.FromId) {

				if len(strings.Split(upd.Object.MessageNew.Text, " ")) < 2 {
					s.SendMessage(upd.Object.MessageNew.PeerId, "Используйте: \"удалить <номер>\"")
					continue
				}
				id, err := strconv.Atoi(strings.Replace(upd.Object.MessageNew.Text, strings.Split(upd.Object.MessageNew.Text, " ")[0]+" ", "", 1))
				if handle(err) {
					s.SendMessage(upd.Object.MessageNew.PeerId, fmt.Sprintf("%q - не номер списка", strings.Replace(upd.Object.MessageNew.Text, strings.Split(upd.Object.MessageNew.Text, " ")[0]+" ", "", 1)))
					continue
				}
				newList := list{ID: id, PeerID: upd.Object.MessageNew.PeerId}
				tx := db.First(&newList)
				if tx.Error != nil {
					s.SendMessage(upd.Object.MessageNew.PeerId, fmt.Sprintf("Невозможно найти список №%d", id))
					continue
				}
				db.Delete(&newList)

				textToSend := fmt.Sprintf("Удалён список %q!\n\nВот кто успел записаться:\n", newList.Name)

				ids := []int{}
				for _, v := range strings.Split(newList.Users, ";") {
					n, err := strconv.Atoi(v)
					if err != nil {
						log.Println(err)
						continue
					}
					ids = append(ids, n)
				}

				usrs, err := s.GetUsersInfo(ids)

				for i, v := range usrs {
					textToSend += fmt.Sprintf("%d. %s %s\n", i+1, v.FirstName, v.LastName)
				}

				_, err = s.SendKeyboard(upd.Object.MessageNew.PeerId, vk.GenerateEmptyKeyBoard(""), textToSend)
				handle(err)

			}
			if strings.Contains(strings.ToLower(upd.Object.MessageNew.Text), "выписать") && canCreate(upd.Object.MessageNew.FromId) {
				if len(strings.Split(upd.Object.MessageNew.Text, " ")) < 3 {

					if canCreate(upd.Object.MessageNew.FromId) {
						s.SendMessage(upd.Object.MessageNew.PeerId, "Используйте: \"выписать <ID списка> <номер в списке>\"")
					}

					continue
				}

				l := list{}

				n, err := strconv.Atoi(strings.Split(upd.Object.MessageNew.Text, " ")[1])
				if handle(err) {
					continue
				}

				db.First(&l, n)

				nom, err := strconv.Atoi(strings.Split(upd.Object.MessageNew.Text, " ")[2])
				if handle(err) {
					continue
				}

				if len(strings.Split(l.Users, ";")) <= nom {
					s.SendMessage(upd.Object.MessageNew.PeerId, "Такого номера нет :0")
					continue
				}

				ids := []int{}
				for i, v := range strings.Split(l.Users, ";") {
					n, err := strconv.Atoi(v)
					if err != nil {
						log.Println(err)
						continue
					}
					if i == nom-1 {
						l.Users = strings.Replace(l.Users, strconv.Itoa(upd.Object.MessageNew.FromId)+";", "", 1)
						continue
					}
					ids = append(ids, n)
				}

				usrs, err := s.GetUsersInfo(ids)
				handle(err)
				textToSend := fmt.Sprintf("Кого-то успешно выписали из %q!\n\nВот весь список:\n", l.Name)

				for i, v := range usrs {
					textToSend += fmt.Sprintf("%d. %s %s\n", i+1, v.FirstName, v.LastName)
				}

				_, err = s.SendMessage(upd.Object.MessageNew.PeerId, textToSend)
				handle(err)

				db.Save(&l)
			}
		}
	}
}

type list struct {
	ID     int
	Users  string
	PeerID int
	Name   string
}

func handle(err error) bool {
	if err != nil {
		log.Print(err)
		return true
	}
	return false
}

func canCreate(id int) bool {
	for _, v := range admins {
		if v == id {
			return true
		}
	}
	return false
}
