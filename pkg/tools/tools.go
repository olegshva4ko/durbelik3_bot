package tools

import (
	"Durbelik3/pkg/models"
	"fmt"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

//CheckAdminAndPatrulChats check if chat is not an admin or patrul
func CheckAdminAndPatrulChats(message *tgbotapi.Message) bool {
	var (
		isChat     bool
		countChats int
	)
l:
	for _, v := range models.BuffChat.Chat {
		//ONE LOOP FOR PATRUL AND ADMINS
		if len(v.PatrulChats) <= len(v.AdminChats) {
			countChats = len(v.PatrulChats)

			for j, k := range v.AdminChats { //avoid firstMessageBan in adminchat
				if message.Chat.ID == k {
					isChat = true
					break l
				}
				if j < countChats {
					if v.PatrulChats[j] == message.Chat.ID {
						isChat = true
						break l
					}
				}

			}
		} else {
			countChats = len(v.AdminChats)

			for j, k := range v.PatrulChats {
				if message.Chat.ID == k {
					isChat = true
					break l
				}
				if j < countChats {
					if v.AdminChats[j] == message.Chat.ID {
						isChat = true
						break l
					}
				}

			}
		}

	}
	return isChat
}

var firstnameCrash = [...]byte{
	byte('['),
	byte(']'),
}

//AvoidMarkdownCrashFirstname removes '[' ']' symbols from firstname
func AvoidMarkdownCrashFirstname(firstname []byte) string {
	var ret []byte
l:
	for _, b := range firstname {
		for _, symbol := range firstnameCrash {
			if b == symbol {
				ret = append(ret, '.')
				continue l
			}
		}
		ret = append(ret, b)
	}
	return string(ret)
}

var pattern = [...]byte{
	byte('_'),
	byte('*'),
	byte('~'),
	byte('`'),
	byte('['),
}

//AvoidMarkdownCrashReason is a function to replace regexp ReplaceAll
func AvoidMarkdownCrashReason(str []byte) string {
	var ret []byte
	for _, b := range str {
		for _, symbol := range pattern {
			if b == symbol {
				ret = append(ret, byte('\\'))
				break
			}
		}
		ret = append(ret, b)
	}
	return string(ret)
}

//TimeInGroup counts time beetwen 2 dates not in UNIX
func TimeInGroup(date int64, language string) (string, error) {
	var (
		dayss    string
		hourss   string
		minutess string
	)

	before := time.Unix(date, 0)
	dif := time.Now().Sub(before)

	days := int(dif.Hours() / 24)
	hours := int(dif.Hours()) - (24 * days)
	minutes := int(dif.Minutes()) - (60 * int(dif.Hours()))

	lastNum := strings.Split(strconv.Itoa(days), "")
	switch lastNum[len(lastNum)-1] {
	case "1":
		if language == "ua" {
			dayss = "день"
		} else {
			dayss = "day"
		}
	case "2":
		if language == "ua" {
			dayss = "дні"
		} else {
			dayss = "days"
		}
	case "3":
		if language == "ua" {
			dayss = "дні"
		} else {
			dayss = "days"
		}
	case "4":
		if language == "ua" {
			dayss = "дні"
		} else {
			dayss = "days"
		}
	default:
		if language == "ua" {
			dayss = "днів"
		} else {
			dayss = "days"
		}
	}

	lastNum = strings.Split(strconv.Itoa(hours), "")
	switch lastNum[len(lastNum)-1] {
	case "1":
		if language == "ua" {
			hourss = "годину"
		} else {
			hourss = "hour"
		}
	case "2":
		if language == "ua" {
			hourss = "години"
		} else {
			hourss = "hours"
		}
	case "3":
		if language == "ua" {
			hourss = "години"
		} else {
			hourss = "hours"
		}
	case "4":
		if language == "ua" {
			hourss = "години"
		} else {
			hourss = "hours"
		}
	default:
		if language == "ua" {
			hourss = "годин"
		} else {
			hourss = "hours"
		}
	}

	lastNum = strings.Split(strconv.Itoa(minutes), "")
	switch lastNum[len(lastNum)-1] {
	case "1":
		if language == "ua" {
			minutess = "хвилину"
		} else {
			minutess = "minute"
		}
	case "2":
		if language == "ua" {
			minutess = "хвилини"
		} else {
			minutess = "minutes"
		}
	case "3":
		if language == "ua" {
			minutess = "хвилини"
		} else {
			minutess = "minutes"
		}
	case "4":
		if language == "ua" {
			minutess = "хвилини"
		} else {
			minutess = "minutes"
		}
	default:
		if language == "ua" {
			minutess = "хвилин"
		} else {
			minutess = "minutes"
		}
	}

	difference := fmt.Sprintf("%d %s, %d %s, %d %s", days, dayss, hours, hourss, minutes, minutess)
	return difference, nil
}
