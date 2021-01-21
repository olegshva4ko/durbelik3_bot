package models

import (
	"fmt"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

//ChatConfig keeps information about chat
type ChatConfig struct {
	ID             int64     `toml:"ID"` //simple chat id
	Title          string    `toml:"title"`
	UserName       string    `toml:"username"`
	Type           string    `toml:"type"`        //Type of chat, can be either “private”, “group”, “supergroup” or “channel”
	AdminChats     []int64   `toml:"adminchats"`  //list of chats in which durbelik is going to send messages
	PatrulChats    []int64   `toml:"patrulchats"` //chat where durbelik will send some of messages
	Language       string    `toml:"language"`
	BannedStickers []string  `toml:"bannedstickers"`
	BroadCastKey   string    `toml:"broadcast_key"`
	IsMain         bool      `toml:"is_main"`     //show if it is main chat
	FatherChat     int64     `toml:"father_chat"` //if chat is not main(admin or patrul) it keeps link on main
	Admins         []*User   `toml:"admins"`      //list of admins in every chat
	Patruls        []*User   `toml:"patruls"`     //list of patruls(glued with father chat)
	Banned         []*User   `toml:"banned"`      //list of people that cant use durbelik commands in the chat
	BadwordsFilter *BadWords `toml:"badwords_filter"`
}

//User to get user from main db
type User struct {
	UserID    int    `toml:"userid"`
	FirstName string `toml:"firstname"`
	LastName  string `toml:"lastname"`
	UserName  string `toml:"username"`
}

func (u *User) String() string {
	var text string
	u.avoidCrashFirstName()
	text += fmt.Sprintf("*USER:* [%s](tg://user?id=%d)\nID: `%d`", u.FirstName, u.UserID, u.UserID)
	return text
}

func (u *User) avoidCrashFirstName() {
	var firstnameCrash = [...]byte{
		byte('['),
		byte(']'),
	}
	var ret []byte
	firstname := []byte(u.FirstName)
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
	u.FirstName = string(ret)

}

//BadWords contains information if badwords filter is enabled in chat
type BadWords struct {
	Enabled bool
	//only one of this three can be enabled
	Replace bool
	React   bool
	Delete  bool
}

//BufferedChat keeps in local memory information about chat
type BufferedChat struct {
	Mu   *sync.Mutex
	Chat map[int64]*ChatConfig
}

var (
	//BuffChat instance of BufferedChat to keep information in local memory
	BuffChat = &BufferedChat{
		Mu:   &sync.Mutex{},
		Chat: make(map[int64]*ChatConfig),
	}
)

//UserInChat Get User from UsersInChat table
type UserInChat struct {
	ID             int
	UserID         int
	ChatID         int64
	Data           int64
	AddedBy        int
	MessageAmount  int64
	CoupleOfTheDay bool
	CallList       bool
}

//MongoResult is a user in mongo database
type MongoResult struct {
	Firstname string
	Lastname  string
	Username  string
	Date      string
}

//NewMongoResult new instance of MongoResult
func NewMongoResult() *MongoResult {
	return &MongoResult{}
}

//BannedSticker are stickers that was banned in channel
type BannedSticker struct {
	ID     int
	ChatID int64
	Name   string
}

//ProcessStickerSpam is for sticker processing
type ProcessStickerSpam struct {
	Mu              *sync.Mutex
	UserPrivateChan map[int]chan *tgbotapi.Message
}

//TimeRestriction is a time processing thing in command handler
type TimeRestriction struct {
	TimeMagnitude string //Hours, minutes, etc
	TimeNumber    int64  //Time that entered user
	TimeUnix      int64  //Time in unix
}

//Violation type for getting warns and afk
type Violation struct {
	ID        int64
	UserID    int
	ChatID    int64
	GivenByID int
	Reason    string
	Date      int64
	Expdate   int64
	Status    bool
	IsWarn    bool
}

func (v *Violation) String() string {
	if v.Reason != "" {
		if v.IsWarn {
			return fmt.Sprintf("*Warn*\nID: %d\nReason: %s\nDate: %s", v.ID, v.Reason, time.Unix(v.Date, 0).Format("02.01 15:04"))
		}
		return fmt.Sprintf("*AFK*\nID: %d\nReason: %s\nDate: %s", v.ID, v.Reason, time.Unix(v.Date, 0).Format("02.01 15:04"))
	}
	if v.IsWarn {
		return fmt.Sprintf("*Warn*\nID: %d\nDate: %s", v.ID, time.Unix(v.Date, 0).Format("02.01 15:04"))
	}
	return fmt.Sprintf("*AFK*\nID: %d\nDate: %s", v.ID, time.Unix(v.Date, 0).Format("02.01 15:04"))

}

//Banned is a struct for users that cant use durbelik commands in chat(and shall be removed from cotd and call lists)
type Banned struct {
	ID     int64
	UserID int
	ChatID int64
}
