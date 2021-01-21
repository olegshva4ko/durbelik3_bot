package configs

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"sync"
	"time"

	"Durbelik3/internal/text"
	"Durbelik3/pkg/commands"
	"Durbelik3/pkg/models"
	"Durbelik3/pkg/sqldatabase"
	"Durbelik3/pkg/mongodatabase"

	"github.com/BurntSushi/toml"
	_ "github.com/go-sql-driver/mysql" //mysql driver
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type configuration struct {
	BotToken    string               `toml:"bot_token"`
	SQLtoken    string               `toml:"sql_token"`
	ChatConfigs []*models.ChatConfig `toml:"ChatConfigs"`
}
type safeConfig struct {
	mu     *sync.Mutex
	config *configuration
}
type safeWriteFile struct {
	mu *sync.Mutex
}

var (
	tomlParsed string
	c          = &safeConfig{ //if mutex is needed you can call it by this struct
		mu:     &sync.Mutex{},
		config: new(configuration),
	}
	w = &safeWriteFile{
		mu: &sync.Mutex{},
	}
)

func init() {
	flag.StringVar(&tomlParsed, "config-path", "./internal/configs/config.toml", "path to config file")
}

//MakeConfig makes all initial configurations before the bot starts
func MakeConfig() (*commands.CommandAPI, error) {
	start := time.Now()
	flag.Parse()

	_, err := toml.DecodeFile(tomlParsed, &c.config) //decode file
	if err != nil {
		panic("Broken config file " + err.Error())
	}
	//create first instance in toml if empty
	if len(c.config.ChatConfigs) == 0 {
		if err := firstChat(); err != nil {
			panic("Cant handle first chat" + err.Error())
		}
	}
	//create bot instance
	bot, err := tgbotapi.NewBotAPI(c.config.BotToken)
	if err != nil {
		panic("Couldnt create bot " + err.Error())
	}
	//set commandAPI and Sqldatabase basic values
	commandAPI := new(commands.CommandAPI)
	commandAPI.DB = new(sqldatabase.Sqldatabase)
	commandAPI.MB = mongodatabase.NewClient()
	commandAPI.DB.MB = commandAPI.MB
	commandAPI.Buff = models.BuffChat
	commandAPI.DB.Buff = models.BuffChat
	commandAPI.Bot = bot
	commandAPI.DB.Bot = bot
	commandAPI.DB.SQLtoken = c.config.SQLtoken
	//check mysql db
	if err := commandAPI.DB.CheckDB(); err != nil {
		panic("Cant open db " + err.Error())
	}
	//check mongodb
	if err := commandAPI.MB.CheckDB(); err != nil {
		panic("Cant open mongodb "+err.Error())
	}
	//update chats
	if err := updateChats(commandAPI); err != nil {
		panic("Cannot update chats " + err.Error())
	}
	//set buff
	for _, chat := range c.config.ChatConfigs {
		models.BuffChat.Chat[chat.ID] = chat
	}
	//update all users that are in database
	if err := updateUsers(commandAPI); err != nil {
		panic("Cannot update users " + err.Error())
	}
	fmt.Println("Settings are done in", time.Now().Sub(start))

	return commandAPI, nil
}

func firstChat() error {
	c.config.ChatConfigs = []*models.ChatConfig{
		&models.ChatConfig{
			ID:             -1001354500052,
			Title:          "Тест__тэст__test222",
			UserName:       "anhsjdjfjci",
			Type:           "supergroup",
			AdminChats:     []int64{},
			PatrulChats:    []int64{},
			Language:       "ua",
			BannedStickers: []string{},
			BroadCastKey:   "",
			IsMain:         true,
			FatherChat:     0,
			Admins:         []*models.User{},
			Patruls:        []*models.User{},
			Banned:         []*models.User{},
			BadwordsFilter: &models.BadWords{},
		},
	}

	w.writeFile()

	return nil
}

func updateUsers(commandAPI *commands.CommandAPI) error {
	users, err := commandAPI.DB.GetAllUsersFromUsersInChat()
	if err != nil {
		return err
	}
	for _, v := range users {
		chatConfigWithUser := tgbotapi.ChatConfigWithUser{
			UserID: v.UserID,
			ChatID: v.ChatID,
		}
		user, err := commandAPI.Bot.GetChatMember(chatConfigWithUser)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		commandAPI.DB.CheckForUserUpdatesMain(user.User, v.ChatID)
	}
	return nil
}

func updateChats(commandAPI *commands.CommandAPI) error {
	var BroadcastKeys map[int64]string = make(map[int64]string) //keeps key to synchronize them

	for i, v := range c.config.ChatConfigs { //find searched chat
		chatConfig := tgbotapi.ChatConfig{
			ChatID: v.ID,
		}
		chat, err := commandAPI.Bot.GetChat(chatConfig)
		if err != nil {
			fmt.Println("Couldnt update chat " + err.Error())
			continue
		}

		if v.Title != chat.Title { //update title
			c.config.ChatConfigs[i].Title = chat.Title
		}
		if v.UserName != chat.UserName { //update username
			c.config.ChatConfigs[i].UserName = chat.UserName
		}
		if v.Type != chat.Type {
			c.config.ChatConfigs[i].Type = chat.Type
		}

		chatAdmins, err := commandAPI.Bot.GetChatAdministrators(chat.ChatConfig())
		if err != nil {
			fmt.Println("Couldnt update admins " + err.Error())
			continue
		}

		v.Admins = nil                   //remove old admins
		for _, val := range chatAdmins { //update chatadmins
			admin := &models.User{
				UserID:    val.User.ID,
				FirstName: val.User.FirstName,
				LastName:  val.User.LastName,
				UserName:  val.User.UserName,
			}
			c.config.ChatConfigs[i].Admins = append(c.config.ChatConfigs[i].Admins, admin)
		}

		bannedStickers, err := commandAPI.DB.GetBannedStickers(v.ID)
		if err != nil {
			fmt.Println("Couldnt add stickers " + err.Error())
			continue
		}
		for _, val := range bannedStickers {
			c.config.ChatConfigs[i].BannedStickers = append(v.BannedStickers, val.Name)
		}

		//users in black list by chatid
		bannedUsers, err := commandAPI.DB.GetBannedUsers(v.ID)
		if err != nil {
			fmt.Println("Couldnt add bannedUSers " + err.Error())
			continue
		}
		for _, val := range bannedUsers {
			c.config.ChatConfigs[i].Banned = append(c.config.ChatConfigs[i].Banned, val)
		}

		if v.IsMain { //if chat is main then try to get or set key existing
			if BroadcastKeys[v.ID] != "" {
				c.config.ChatConfigs[i].BroadCastKey = BroadcastKeys[v.ID]
				continue
			}
			c.config.ChatConfigs[i].BroadCastKey = text.GenerateSecretKey()
			BroadcastKeys[v.ID] = v.BroadCastKey

		} else { //if chat is not main create key if not exists for father chat
			if BroadcastKeys[v.FatherChat] != "" {
				c.config.ChatConfigs[i].BroadCastKey = BroadcastKeys[v.FatherChat]
				continue
			}
			c.config.ChatConfigs[i].BroadCastKey = text.GenerateSecretKey()
			BroadcastKeys[v.FatherChat] = v.BroadCastKey
		}

	}
	w.writeFile()

	return nil
}

//AddChatToConfig adds chat to toml file
func AddChatToConfig(chatInstance *tgbotapi.Chat, commandAPI *commands.CommandAPI) {
	newChat := &models.ChatConfig{
		ID:             chatInstance.ID,
		Title:          chatInstance.Title,
		UserName:       chatInstance.UserName,
		Type:           chatInstance.Type,
		AdminChats:     []int64{},
		PatrulChats:    []int64{},
		Language:       "ua",
		BannedStickers: []string{},
		BroadCastKey:   text.GenerateSecretKey(),
		IsMain:         true,
		FatherChat:     0,
		Admins:         []*models.User{},
		Patruls:        []*models.User{},
		Banned:         []*models.User{},
		BadwordsFilter: &models.BadWords{},
	}

	admins, err := commandAPI.Bot.GetChatAdministrators(chatInstance.ChatConfig())
	if err != nil {
		return
	}

	for _, v := range admins {
		var admin *models.User = &models.User{
			UserID:    v.User.ID,
			UserName:  v.User.UserName,
			FirstName: v.User.FirstName,
			LastName:  v.User.LastName,
		}
		newChat.Admins = append(newChat.Admins, admin)
	}
	c.mu.Lock()
	c.config.ChatConfigs = append(c.config.ChatConfigs, newChat)
	c.mu.Unlock()

	commandAPI.Buff.Mu.Lock()
	commandAPI.Buff.Chat[chatInstance.ID] = newChat
	commandAPI.Buff.Mu.Unlock()

	go w.writeFile()
}

func (w *safeWriteFile) writeFile() {
	w.mu.Lock()
	b := &bytes.Buffer{}
	if err := toml.NewEncoder(b).Encode(&c.config); err != nil { //try to rewrite config toml file
		fmt.Println("Error while  adding new chat")
		return
	}

	ioutil.WriteFile(`./internal/configs/config.toml`, b.Bytes(), 0600)
	w.mu.Unlock()
}
