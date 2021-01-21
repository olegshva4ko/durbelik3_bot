package commands

import (
	"Durbelik3/internal/graph"
	"Durbelik3/internal/text"
	"Durbelik3/pkg/models"
	"Durbelik3/pkg/mongodatabase"
	"Durbelik3/pkg/sqldatabase"
	"Durbelik3/pkg/tools"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

//CommandAPI thing that allows you to communicate via bot, with database, with buffer
type CommandAPI struct {
	Bot  *tgbotapi.BotAPI
	DB   *sqldatabase.Sqldatabase
	MB   *mongodatabase.MongoDatabase
	Buff *models.BufferedChat
}

/*
AddCheckUser adds user to database. Sends welcome message if needed. Processes sticker spam.
Checks user first message. If it is contains link then ban user
*/
func (command *CommandAPI) AddCheckUser() func(*tgbotapi.Message) {
	type safeDatabaseUsage struct {
		Mu             *sync.Mutex
		MessageProcess func(*tgbotapi.Message) bool
	}
	var (
	// ProcessStickerSpam *models.ProcessStickerSpam = &models.ProcessStickerSpam{
	// 	Mu:              &sync.Mutex{},
	// 	UserPrivateChan: make(map[int]chan *tgbotapi.Message),
	// }
	// SafeDB *safeDatabaseUsage = &safeDatabaseUsage{
	// 	Mu:             &sync.Mutex{},
	// 	MessageProcess: command.DB.MessageProcess,
	// }
	)
	//start channel for stickerspam routing
	stickerSpamRoute := make(chan *tgbotapi.Message)
	//start router
	go command.stickerSpamRouter(stickerSpamRoute)
	//make channel for multiprocessing
	safeDBchan := make(chan *tgbotapi.Message)
	for i := 0; i < 4; i++ { //start some goroutines
		go command.messageProcess(safeDBchan)
	}

	return func(message *tgbotapi.Message) {
		//Send welcome messages
		if message.NewChatMembers != nil {
			//Delete message "Liuda added some1"
			command.Bot.DeleteMessage(tgbotapi.NewDeleteMessage(message.Chat.ID, message.MessageID))
			for _, v := range *message.NewChatMembers {
				//Send welcome message
				go command.SendWelcomeMessage(&v, message.Chat.ID) //mb add go
			}
		}
		stickerSpamRoute <- message

		// //send messages to private chan
		// for id, ch := range ProcessStickerSpam.UserPrivateChan {
		// 	if id == message.From.ID {
		// 		ch <- message
		// 	}
		// }
		// if ProcessStickerSpam.UserPrivateChan[message.From.ID] != nil {
		// 	ProcessStickerSpam.UserPrivateChan[message.From.ID] <- message
		// }
		//process sticker spam
		// if message.Sticker != nil {
		// 	if ProcessStickerSpam.UserPrivateChan[message.From.ID] == nil { // if user chan already exists skipping
		// 		ProcessStickerSpam.Mu.Lock() //add to map and create chan
		// 		ProcessStickerSpam.UserPrivateChan[message.From.ID] = make(chan *tgbotapi.Message)
		// 		ProcessStickerSpam.Mu.Unlock()
		// 		go command.StickerSpamProcessor(ProcessStickerSpam, ProcessStickerSpam.UserPrivateChan[message.From.ID], message.From.ID)
		// 		ProcessStickerSpam.UserPrivateChan[message.From.ID] <- message
		// 	}
		// }

		safeDBchan <- message
		// //process first message and adding to database
		// SafeDB.Mu.Lock()
		// firstMessage := SafeDB.MessageProcess(message)
		// SafeDB.Mu.Unlock()
		// if firstMessage {

		// 	messageInAdminChat := tools.CheckAdminAndPatrulChats(message)
		// 	if !messageInAdminChat {
		// 		command.BanForLink(message)
		// 	}
		// }
	}

}

//messageProcess worker pool pattern
func (command *CommandAPI) messageProcess(messages chan *tgbotapi.Message) {
	for message := range messages {
		firstMessage := command.DB.MessageProcess(message)
		if firstMessage && command.Buff.Chat[message.Chat.ID].IsMain {
			command.BanForLink(message)
		}
	}
}

//SendWelcomeMessage Message to new Users
func (command *CommandAPI) SendWelcomeMessage(user *tgbotapi.User, chatID int64) {

	firstName := tools.AvoidMarkdownCrashFirstname([]byte(user.FirstName)) //avoid markdown crashing
	msg := tgbotapi.NewMessage(chatID,
		fmt.Sprintf(text.WelcomeMessage[command.Buff.Chat[chatID].Language],
			firstName, user.ID))
	msg.ParseMode = "markdown"
	msg.DisableWebPagePreview = true

	//delete welcome message
	deleteMessage, err := command.Bot.Send(msg)
	if err != nil {
		go command.splitMessage(msg, msg.Text, err)
		return
	}
	time.Sleep(45 * time.Second)
	command.Bot.DeleteMessage(tgbotapi.NewDeleteMessage(chatID, deleteMessage.MessageID))

}

//stickerSpamRouter will resend messages to StickerSpamProcessor
func (command *CommandAPI) stickerSpamRouter(messages chan *tgbotapi.Message) {
	var ProcessStickerAndGifSpam *models.ProcessStickerSpam = &models.ProcessStickerSpam{
		Mu:              &sync.Mutex{},
		UserPrivateChan: make(map[int]chan *tgbotapi.Message),
	}
	backchannel := make(chan int) //chan to know whether we should close it or no
	for {
		select {
		case message := <-messages:
			if ProcessStickerAndGifSpam.UserPrivateChan[message.From.ID] != nil {
				ProcessStickerAndGifSpam.UserPrivateChan[message.From.ID] <- message
			}
			//create if not exists
			if message.Sticker != nil || message.Animation != nil {
				if ProcessStickerAndGifSpam.UserPrivateChan[message.From.ID] == nil { // if user chan already exists skipping
					ProcessStickerAndGifSpam.Mu.Lock() //add to map and create chan
					ProcessStickerAndGifSpam.UserPrivateChan[message.From.ID] = make(chan *tgbotapi.Message)
					ProcessStickerAndGifSpam.Mu.Unlock()
					go command.StickerAndGifSpamProcessor(ProcessStickerAndGifSpam.UserPrivateChan[message.From.ID], message.From.ID, backchannel)
					ProcessStickerAndGifSpam.UserPrivateChan[message.From.ID] <- message
				}
			}
		case userID := <-backchannel:
			ProcessStickerAndGifSpam.Mu.Lock()
			close(ProcessStickerAndGifSpam.UserPrivateChan[userID])
			delete(ProcessStickerAndGifSpam.UserPrivateChan, userID)
			ProcessStickerAndGifSpam.Mu.Unlock()
		}
	}
}

//StickerAndGifSpamProcessor ...
func (command *CommandAPI) StickerAndGifSpamProcessor(messages chan *tgbotapi.Message, userID int, backchannel chan int) {
	numberOfStickers := 0
	numberOfAnimations := 0
	ticker := time.NewTicker(60 * time.Second)
	for {
		select {
		case message := <-messages:
			if message.Sticker != nil {
				numberOfStickers++
				ticker.Stop()
				ticker = time.NewTicker(5 * time.Second)

				if numberOfStickers == 4 {
					msg := tgbotapi.NewMessage(message.Chat.ID,
						text.StickerWarning[command.Buff.Chat[message.Chat.ID].Language])

					msg.ReplyToMessageID = message.MessageID
					_, err := command.Bot.Send(msg)
					if err != nil {
						go command.splitMessage(msg, msg.Text, err)
					}
					continue
				}
				if numberOfStickers == 5 {
					u := &models.User{
						FirstName: message.From.FirstName,
						LastName:  message.From.LastName,
						UserName:  message.From.UserName,
						UserID:    message.From.ID,
					}
					timeRestriction := &models.TimeRestriction{
						TimeUnix:      time.Now().Add(10 * time.Minute).Unix(),
						TimeNumber:    10,
						TimeMagnitude: "minutes",
					}
					command.muteUsers([]*models.User{u}, message, text.StickerMute[command.Buff.Chat[message.Chat.ID].Language], timeRestriction)

					ticker.Stop()
					backchannel <- userID
					return
				}
			} else if message.Animation != nil {
				numberOfAnimations++
				ticker.Stop()
				ticker = time.NewTicker(5 * time.Second)

				if numberOfAnimations == 4 {
					msg := tgbotapi.NewMessage(message.Chat.ID,
						text.GifWarning[command.Buff.Chat[message.Chat.ID].Language])

					msg.ReplyToMessageID = message.MessageID
					_, err := command.Bot.Send(msg)
					if err != nil {
						go command.splitMessage(msg, msg.Text, err)
					}
					continue
				}
				if numberOfAnimations == 5 {
					u := &models.User{
						FirstName: message.From.FirstName,
						LastName:  message.From.LastName,
						UserName:  message.From.UserName,
						UserID:    message.From.ID,
					}
					timeRestriction := &models.TimeRestriction{
						TimeUnix:      time.Now().Add(10 * time.Minute).Unix(),
						TimeNumber:    10,
						TimeMagnitude: "minutes",
					}
					command.muteUsers([]*models.User{u}, message, text.GifMute[command.Buff.Chat[message.Chat.ID].Language], timeRestriction)

					ticker.Stop()
					backchannel <- userID
					return
				}
			}
			//else { //if message is not sticker
			// 	ticker.Stop()
			// 	backchannel <- userID
			// 	return
			// }
		case <-ticker.C:
			ticker.Stop()
			// processStickerSpam.Mu.Lock()
			// close(processStickerSpam.UserPrivateChan[userID])
			// delete(processStickerSpam.UserPrivateChan, userID)
			// processStickerSpam.Mu.Unlock()
			backchannel <- userID
			return
		}
	}
}

//BanForLink ...
func (command *CommandAPI) BanForLink(message *tgbotapi.Message) {

	re := regexp.MustCompile(`(?i)(\b(https?|ftp|file):\/\/[-A-Z0-9+&@#\/%?=~_|!:,.;]*[-A-Z0-9+&@#\/%=~_|])`)
	if string(re.Find([]byte(message.Text))[:]) != "" {
		command.Bot.DeleteMessage(tgbotapi.NewDeleteMessage(message.Chat.ID, message.MessageID))

		//create KickChatMemberConfig
		chatMemberConfig := tgbotapi.ChatMemberConfig{
			ChatID: message.Chat.ID,
			UserID: message.From.ID,
		}
		kickMemberConfig := tgbotapi.KickChatMemberConfig{
			ChatMemberConfig: chatMemberConfig,
			UntilDate:        1,
		}
		command.Bot.KickChatMember(kickMemberConfig)

		firstName := tools.AvoidMarkdownCrashFirstname([]byte(message.From.FirstName))
		msg := tgbotapi.MessageConfig{
			Text: fmt.Sprintf(text.BanForLink[command.Buff.Chat[message.Chat.ID].Language],
				firstName, message.From.ID),
			ParseMode: "markdown",
			BaseChat:  tgbotapi.BaseChat{},
		}
		if command.Buff.Chat[message.Chat.ID].IsMain {
			for _, v := range command.Buff.Chat[message.Chat.ID].AdminChats {
				msg.BaseChat.ChatID = v
				_, err := command.Bot.Send(msg)
				if err != nil {
					go command.splitMessage(msg, msg.Text, err)
				}
			}
		} else {
			for _, v := range command.Buff.Chat[command.Buff.Chat[message.Chat.ID].FatherChat].AdminChats {
				msg.BaseChat.ChatID = v
				_, err := command.Bot.Send(msg)
				if err != nil {
					go command.splitMessage(msg, msg.Text, err)
				}
			}
		}

	}
}

//WarnUser warns user
func (command *CommandAPI) WarnUser(message *tgbotapi.Message) {

	users, reason, _, err := command.DB.CommandHandler(message)
	if err != nil {
		command.Bot.Send(tgbotapi.NewMessage(message.Chat.ID,
			text.SomethingWentWrong[command.Buff.Chat[message.Chat.ID].Language]))
		return
	}

	userWithWarns, err := command.DB.WarnUsers(message, users, reason)
	if err != nil {
		command.Bot.Send(tgbotapi.NewMessage(message.Chat.ID,
			text.SomethingWentWrong[command.Buff.Chat[message.Chat.ID].Language]))
		return
	}

	var (
		filterBanned []*models.User
		filterMuted  []*models.User
		filterWarned = make(map[*models.User]int)
	)

	for user, warns := range userWithWarns {

		if len(warns)%5 == 0 {
			filterBanned = append(filterBanned, user)
		} else {
			filterWarned[user] = len(warns)
			filterMuted = append(filterMuted, user)
		}
	}
	time := &models.TimeRestriction{
		TimeUnix:      time.Now().Add(10 * time.Minute).Unix(),
		TimeNumber:    10,
		TimeMagnitude: "minutes",
	}
	//MUTE USERS
	go command.muteUsers(filterMuted, message, reason, time)
	//ban filterBanned users
	go command.banUsers(filterBanned, message, reason)

	//...
	//send message about warn
	var (
		singular       bool = len(filterWarned) == 1
		msgText        []string
		firstnameAdmin = tools.AvoidMarkdownCrashFirstname([]byte(message.From.FirstName))
		chatID         int64
	)

	if command.Buff.Chat[message.Chat.ID].IsMain { //if you want to give warn from admin chat
		chatID = message.Chat.ID
	} else {
		chatID = command.Buff.Chat[message.Chat.ID].FatherChat
	}

	for user, numberOfWarns := range filterWarned {
		if singular {
			/*
				will send message on this moment. Dont go further
			*/
			firstname := tools.AvoidMarkdownCrashFirstname([]byte(user.FirstName))
			msgText = append(msgText, fmt.Sprintf("[%s](tg://user?id=%d) (*ID*: `%d`)", firstname, user.UserID, user.UserID))
			if reason != "" {
				msg := tgbotapi.NewMessage(chatID,
					fmt.Sprintf(text.ChatWarnWithReasonSingular[command.Buff.Chat[chatID].Language],
						strings.Join(msgText, ""), numberOfWarns%5, reason))
				msg.ParseMode = "markdown"
				_, err := command.Bot.Send(msg)
				if err != nil {
					go command.splitMessage(msg, msg.Text, err)
				}

				//send message to admin chat
				msg.Text = fmt.Sprintf(text.AdminChatWarnWithReasonSingular[command.Buff.Chat[chatID].Language],
					firstnameAdmin, message.From.ID, numberOfWarns%5, strings.Join(msgText, ""), reason)
				for _, chatID := range command.Buff.Chat[chatID].AdminChats {
					msg.BaseChat.ChatID = chatID
					_, err := command.Bot.Send(msg)
					if err != nil {
						go command.splitMessage(msg, msg.Text, err)
					}
				}

			} else {
				msg := tgbotapi.NewMessage(message.Chat.ID,
					fmt.Sprintf(text.ChatWarnWithoutReasonSingular[command.Buff.Chat[chatID].Language],
						strings.Join(msgText, ""), numberOfWarns%5))
				msg.ParseMode = "markdown"
				_, err := command.Bot.Send(msg)
				if err != nil {
					go command.splitMessage(msg, msg.Text, err)
				}

				//send message to admin chat
				msg.Text = fmt.Sprintf(text.AdminChatWarnWithoutReasonSingular[command.Buff.Chat[chatID].Language],
					firstnameAdmin, message.From.ID, numberOfWarns%5, strings.Join(msgText, ""))
				for _, adminChatID := range command.Buff.Chat[chatID].AdminChats {
					msg.BaseChat.ChatID = adminChatID
					_, err := command.Bot.Send(msg)
					if err != nil {
						go command.splitMessage(msg, msg.Text, err)
					}
				}

			}
			return
		}
		firstname := tools.AvoidMarkdownCrashFirstname([]byte(user.FirstName))
		msgText = append(msgText, fmt.Sprintf("[%s](tg://user?id=%d)", firstname, user.UserID))

	}

	if reason != "" {
		msg := tgbotapi.NewMessage(chatID,
			fmt.Sprintf(text.ChatWarnWithReasonPlural[command.Buff.Chat[chatID].Language],
				strings.Join(msgText, ", "), reason))
		msg.ParseMode = "markdown"
		command.Bot.Send(msg)
	} else {
		msg := tgbotapi.NewMessage(chatID,
			fmt.Sprintf(text.ChatWarnWithoutReasonPlural[command.Buff.Chat[chatID].Language],
				strings.Join(msgText, ", ")))
		msg.ParseMode = "markdown"
		command.Bot.Send(msg)
	}
	msg := tgbotapi.MessageConfig{
		ParseMode: "markdown",
		BaseChat:  tgbotapi.BaseChat{},
	}
	if reason != "" {
		msg.Text = fmt.Sprintf(text.AdminChatWarnWithReasonPlural[command.Buff.Chat[chatID].Language],
			firstnameAdmin, message.From.ID, strings.Join(msgText, ", "), reason)
	} else {
		msg.Text = fmt.Sprintf(text.AdminChatWarnWithoutReasonPlural[command.Buff.Chat[chatID].Language],
			firstnameAdmin, message.From.ID, strings.Join(msgText, ", "))
	}
	for _, adminChatID := range command.Buff.Chat[chatID].AdminChats {
		msg.BaseChat.ChatID = adminChatID
		_, err := command.Bot.Send(msg)
		if err != nil {
			go command.splitMessage(msg, msg.Text, err)
		}
	}
}

//banUsers is function for banning users after 5 warnings
func (command *CommandAPI) banUsers(users []*models.User, message *tgbotapi.Message, reason string) {
	if len(users) == 0 {
		return
	}
	var (
		singular       bool = len(users) == 1
		msgText        []string
		firstnameAdmin = tools.AvoidMarkdownCrashFirstname([]byte(message.From.FirstName))
		chatID         int64
	)

	if command.Buff.Chat[message.Chat.ID].IsMain {
		chatID = message.Chat.ID
	} else {
		chatID = command.Buff.Chat[message.Chat.ID].FatherChat
	}
	//REMOVE FROM CALLLIST AND OTHERS
	go command.DB.RemoveFromLists(chatID, users)
	for _, user := range users {
		//kick user
		chatMemberConfig := tgbotapi.ChatMemberConfig{
			ChatID: chatID,
			UserID: user.UserID,
		}

		kickMemberConfig := tgbotapi.KickChatMemberConfig{
			ChatMemberConfig: chatMemberConfig,
			UntilDate:        0,
		}
		command.Bot.KickChatMember(kickMemberConfig)
		if singular {
			firstname := tools.AvoidMarkdownCrashFirstname([]byte(user.FirstName))
			msgText = append(msgText, fmt.Sprintf("[%s](tg://user?id=%d) (*ID*: `%d`)", firstname, user.UserID, user.UserID))
			if reason != "" {
				msg := tgbotapi.NewMessage(chatID,
					fmt.Sprintf(text.ChatBanWithReasonSingularWarns[command.Buff.Chat[chatID].Language], strings.Join(msgText, ""), reason))
				msg.ParseMode = "markdown"
				_, err := command.Bot.Send(msg)
				if err != nil {
					go command.splitMessage(msg, msg.Text, err)
				}

				//send message to admin chat
				msg.Text = fmt.Sprintf(text.AdminBanWithReasonWarns[command.Buff.Chat[chatID].Language], firstnameAdmin, message.From.ID, strings.Join(msgText, ""), reason)
				for _, chatID := range command.Buff.Chat[chatID].AdminChats {
					msg.BaseChat.ChatID = chatID
					_, err := command.Bot.Send(msg)
					if err != nil {
						go command.splitMessage(msg, msg.Text, err)
					}
				}

			} else {
				msg := tgbotapi.NewMessage(chatID,
					fmt.Sprintf(text.ChatBanWithoutReasonSingularWarns[command.Buff.Chat[chatID].Language], strings.Join(msgText, "")))
				msg.ParseMode = "markdown"
				_, err := command.Bot.Send(msg)
				if err != nil {
					go command.splitMessage(msg, msg.Text, err)
				}

				//send message to admin chat
				msg.Text = fmt.Sprintf(text.AdminBanWithoutReasonWarns[command.Buff.Chat[chatID].Language], firstnameAdmin, message.From.ID, strings.Join(msgText, ""))
				for _, adminChatID := range command.Buff.Chat[chatID].AdminChats {
					msg.BaseChat.ChatID = adminChatID
					_, err := command.Bot.Send(msg)
					if err != nil {
						go command.splitMessage(msg, msg.Text, err)
					}
				}

			}
			return
		}
		firstname := tools.AvoidMarkdownCrashFirstname([]byte(user.FirstName))
		msgText = append(msgText, fmt.Sprintf("[%s](tg://user?id=%d)", firstname, user.UserID))
	}
	//send to chat plural message
	if reason != "" { //with reason
		msg := tgbotapi.NewMessage(chatID,
			fmt.Sprintf(text.ChatBanWithReasonPluralWarns[command.Buff.Chat[chatID].Language],
				strings.Join(msgText, ", "), reason))
		msg.ParseMode = "markdown"
		_, err := command.Bot.Send(msg)
		if err != nil {
			go command.splitMessage(msg, msg.Text, err)
		}
	} else { //without reason
		msg := tgbotapi.NewMessage(chatID,
			fmt.Sprintf(text.ChatBanWithoutReasonPluralWarns[command.Buff.Chat[chatID].Language],
				strings.Join(msgText, ", ")))
		msg.ParseMode = "markdown"
		_, err := command.Bot.Send(msg)
		if err != nil {
			go command.splitMessage(msg, msg.Text, err)
		}
	}
	msg := tgbotapi.MessageConfig{
		ParseMode: "markdown",
		BaseChat:  tgbotapi.BaseChat{},
	}

	if reason != "" {
		msg.Text = fmt.Sprintf(text.AdminBanWithReasonWarns[command.Buff.Chat[chatID].Language],
			firstnameAdmin, message.From.ID, strings.Join(msgText, ", "), reason)
	} else {
		msg.Text = fmt.Sprintf(text.AdminBanWithoutReasonWarns[command.Buff.Chat[chatID].Language],
			firstnameAdmin, message.From.ID, strings.Join(msgText, ", "))
	}
	for _, adminChatID := range command.Buff.Chat[chatID].AdminChats {
		msg.BaseChat.ChatID = adminChatID
		_, err := command.Bot.Send(msg)
		if err != nil {
			go command.splitMessage(msg, msg.Text, err)
		}
	}
}

//muteUsers is function for muting users
func (command *CommandAPI) muteUsers(users []*models.User, message *tgbotapi.Message, reason string, timeRestriction *models.TimeRestriction) {
	if len(users) == 0 {
		return
	}
	var (
		singular       bool = len(users) == 1
		msgText        []string
		firstnameAdmin = tools.AvoidMarkdownCrashFirstname([]byte(message.From.FirstName))
		chatID         int64
		timeText       string
	)
	hasToBeSent := strings.Split(message.Text, " ") //check if message about mute should be sent
	if command.Buff.Chat[message.Chat.ID].IsMain {
		chatID = message.Chat.ID
	} else {
		chatID = command.Buff.Chat[message.Chat.ID].FatherChat
	}

	if timeRestriction.TimeNumber == 0 {
		timeText = text.Forever[command.Buff.Chat[chatID].Language]
	} else {
		timeText = fmt.Sprintf("%d %s", timeRestriction.TimeNumber, timeRestriction.TimeMagnitude)
	}

	var f bool = false
	restrictChatMemberConfig := tgbotapi.RestrictChatMemberConfig{
		UntilDate:             timeRestriction.TimeUnix,
		CanSendMessages:       &f,
		CanSendMediaMessages:  &f,
		CanSendOtherMessages:  &f,
		CanAddWebPagePreviews: &f,
	}

	for _, user := range users {
		//kick user
		chatMemberConfig := tgbotapi.ChatMemberConfig{
			ChatID: chatID,
			UserID: user.UserID,
		}
		restrictChatMemberConfig.ChatMemberConfig = chatMemberConfig
		//check chatmember. If he is banned this function wont unban him
		chatMember, err := command.Bot.GetChatMember(tgbotapi.ChatConfigWithUser{
			ChatID: chatID,
			UserID: user.UserID,
		})
		if err != nil {
			fmt.Println(err.Error())
		}
		if chatMember.Status != "kicked" {
			command.Bot.RestrictChatMember(restrictChatMemberConfig)
		}
		if hasToBeSent[0] != "!warn" {
			if singular {
				firstname := tools.AvoidMarkdownCrashFirstname([]byte(user.FirstName))
				msgText = append(msgText, fmt.Sprintf("[%s](tg://user?id=%d) (*ID*: `%d`)", firstname, user.UserID, user.UserID))

				if reason != "" { //one user with reason to chat
					msg := tgbotapi.NewMessage(chatID,
						fmt.Sprintf(text.ChatMuteWithReasonSingularWarn[command.Buff.Chat[chatID].Language], strings.Join(msgText, ""), reason))
					msg.ParseMode = "markdown"
					_, err := command.Bot.Send(msg)
					if err != nil {
						go command.splitMessage(msg, msg.Text, err)
					}

					//send message to admin chat
					msg.Text = fmt.Sprintf(text.AdminMuteWithReasonWarn[command.Buff.Chat[chatID].Language],
						firstnameAdmin, message.From.ID, strings.Join(msgText, ""), timeText, reason)
					for _, adminChatID := range command.Buff.Chat[chatID].AdminChats {
						msg.BaseChat.ChatID = adminChatID
						_, err := command.Bot.Send(msg)
						if err != nil {
							go command.splitMessage(msg, msg.Text, err)
						}
					}

				} else { //one user without reason to chat
					msg := tgbotapi.NewMessage(chatID,
						fmt.Sprintf(text.ChatMuteWithoutReasonSingularWarn[command.Buff.Chat[chatID].Language], strings.Join(msgText, "")))
					msg.ParseMode = "markdown"
					_, err := command.Bot.Send(msg)
					if err != nil {
						go command.splitMessage(msg, msg.Text, err)
					}

					//send message to admin chat

					msg.Text = fmt.Sprintf(text.AdminMuteWithoutReasonWarn[command.Buff.Chat[chatID].Language], firstnameAdmin, message.From.ID, strings.Join(msgText, ""), timeText)
					for _, adminChatID := range command.Buff.Chat[chatID].AdminChats {
						msg.BaseChat.ChatID = adminChatID
						_, err := command.Bot.Send(msg)
						if err != nil {
							go command.splitMessage(msg, msg.Text, err)
						}
					}

				}
				return
			}
		}

		firstname := tools.AvoidMarkdownCrashFirstname([]byte(user.FirstName))
		msgText = append(msgText, fmt.Sprintf("[%s](tg://user?id=%d)", firstname, user.UserID))
	}
	if hasToBeSent[0] == "!warn" {
		return
	}
	//send to chat plural message
	if reason != "" { //two or more users with reason to chat
		msg := tgbotapi.NewMessage(chatID,
			fmt.Sprintf(text.ChatMuteWithReasonPluralWarn[command.Buff.Chat[chatID].Language], strings.Join(msgText, ", "), reason))
		msg.ParseMode = "markdown"
		_, err := command.Bot.Send(msg)
		if err != nil {
			go command.splitMessage(msg, msg.Text, err)
		}
	} else { //two or more users without reason to chat
		msg := tgbotapi.NewMessage(chatID,
			fmt.Sprintf(text.ChatMuteWithoutReasonPluralWarn[command.Buff.Chat[chatID].Language], strings.Join(msgText, ", ")))
		msg.ParseMode = "markdown"
		_, err := command.Bot.Send(msg)
		if err != nil {
			go command.splitMessage(msg, msg.Text, err)
		}
	}
	msg := tgbotapi.MessageConfig{
		ParseMode: "markdown",
		BaseChat:  tgbotapi.BaseChat{},
	}

	if reason != "" { //with reason to admin chat
		msg.Text = fmt.Sprintf(text.AdminMuteWithReasonWarn[command.Buff.Chat[chatID].Language], firstnameAdmin, message.From.ID, strings.Join(msgText, ", "), timeText, reason)
	} else { //without reason to admin chat
		msg.Text = fmt.Sprintf(text.AdminMuteWithoutReasonWarn[command.Buff.Chat[chatID].Language], firstnameAdmin, message.From.ID, strings.Join(msgText, ", "), timeText)
	}
	for _, adminChatID := range command.Buff.Chat[chatID].AdminChats {
		msg.BaseChat.ChatID = adminChatID
		_, err := command.Bot.Send(msg)
		if err != nil {
			go command.splitMessage(msg, msg.Text, err)
		}
	}
}

//UnwarnUsers allows you to unwarn user
func (command *CommandAPI) UnwarnUsers(message *tgbotapi.Message) {
	_, _, time, err := command.DB.CommandHandler(message)
	if err != nil {
		command.Bot.Send(tgbotapi.NewMessage(message.Chat.ID,
			text.SomethingWentWrong[command.Buff.Chat[message.Chat.ID].Language]))
		return
	}
	var (
		msgText []string
		chatID  int64
	)
	if command.Buff.Chat[message.Chat.ID].IsMain {
		chatID = message.Chat.ID
	} else {
		chatID = command.Buff.Chat[message.Chat.ID].FatherChat
	}

	violation, err := command.DB.UnwarnUsers(chatID, time.TimeNumber, true)
	if err != nil {
		command.Bot.Send(tgbotapi.NewMessage(message.Chat.ID, err.Error()))
		return
	}
	user, found := command.DB.GetUserByUserID(violation.UserID)
	if !found {
		command.Bot.Send(tgbotapi.NewMessage(message.Chat.ID,
			text.SomethingWentWrong[command.Buff.Chat[message.Chat.ID].Language]))
		return
	}

	firstname := tools.AvoidMarkdownCrashFirstname([]byte(user.FirstName))
	msgText = append(msgText, fmt.Sprintf("[%s](tg://user?id=%d)", firstname, user.UserID))
	msg := tgbotapi.NewMessage(chatID,
		fmt.Sprintf(text.Unwarn[command.Buff.Chat[chatID].Language], strings.Join(msgText, "")))
	msg.ParseMode = "markdown"

	_, err = command.Bot.Send(msg)
	if err != nil {
		go command.splitMessage(msg, msg.Text, err)
	}
	for _, adminchat := range command.Buff.Chat[chatID].AdminChats {
		msg.BaseChat.ChatID = adminchat
		_, err := command.Bot.Send(msg)
		if err != nil {
			go command.splitMessage(msg, msg.Text, err)
		}
	}

}

/*GetWarns returns list of user warns
Firstname *ID*:`32132132`

date time reason

Vsiogo warns: 9
*/
func (command *CommandAPI) GetWarns(message *tgbotapi.Message) {
	users, _, _, err := command.DB.CommandHandler(message)
	if err != nil {
		command.Bot.Send(tgbotapi.NewMessage(message.Chat.ID,
			text.SomethingWentWrong[command.Buff.Chat[message.Chat.ID].Language]))
		return
	}
	var (
		msgText    string
		activeWarn int
		chatID     int64
	)
	if command.Buff.Chat[message.Chat.ID].IsMain {
		chatID = message.Chat.ID
	} else {
		chatID = command.Buff.Chat[message.Chat.ID].FatherChat
	}
	userWithWarns, err := command.DB.GetUsersWithViolations(chatID, users, true)
	if err != nil {
		command.Bot.Send(tgbotapi.NewMessage(message.Chat.ID,
			text.SomethingWentWrong[command.Buff.Chat[message.Chat.ID].Language]))
		return
	}

	for user, warns := range userWithWarns {
		firstname := tools.AvoidMarkdownCrashFirstname([]byte(user.FirstName))
		msgText += fmt.Sprintf("[%s](tg://user?id=%d) *ID:*`%d`\n\n", firstname, user.UserID, user.UserID)
		for _, warn := range warns {
			if warn.Status {
				if warn.Reason != "" {
					msgText += fmt.Sprintf("%d. \\[%s] %s\n", warn.ID, time.Unix(warn.Date, 0).Format("02.01 15:04"), warn.Reason)
				} else {
					msgText += fmt.Sprintf("%d. \\[%s] %s\n", warn.ID, time.Unix(warn.Date, 0).Format("02.01 15:04"), text.AbsentReason[command.Buff.Chat[message.Chat.ID].Language])
				}
				activeWarn++
			}
		}
		msgText += fmt.Sprintf(text.GetWarns[command.GetChatLanguage(chatID)], len(warns), activeWarn)
		activeWarn = 0
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, msgText)
	msg.ParseMode = "markdown"
	_, err = command.Bot.Send(msg)
	if err != nil {
		go command.splitMessage(msg, msg.Text, err)
	}

}

//DeleteViolation ...
func (command *CommandAPI) DeleteViolation(message *tgbotapi.Message, isWarn bool) {
	_, _, time, err := command.DB.CommandHandler(message)
	if err != nil && time == nil {
		command.Bot.Send(tgbotapi.NewMessage(message.Chat.ID,
			text.SomethingWentWrong[command.Buff.Chat[message.Chat.ID].Language]))
		return
	}
	var chatID int64
	if command.Buff.Chat[message.Chat.ID].IsMain {
		chatID = message.Chat.ID
	} else {
		chatID = command.Buff.Chat[message.Chat.ID].FatherChat
	}
	violation, err := command.DB.DeleteViolation(chatID, time.TimeNumber, isWarn)
	if err != nil {
		command.Bot.Send(tgbotapi.NewMessage(message.Chat.ID, err.Error()))
		return
	}
	user, found := command.DB.GetUserByUserID(violation.UserID)
	if !found {
		command.Bot.Send(tgbotapi.NewMessage(message.Chat.ID,
			text.SomethingWentWrong[command.Buff.Chat[message.Chat.ID].Language]))
		return
	}
	msg := tgbotapi.NewMessage(message.Chat.ID,
		fmt.Sprintf(text.WarnHasBeenDeleted[command.Buff.Chat[message.Chat.ID].Language], violation, user))
	msg.ParseMode = "markdown"
	_, err = command.Bot.Send(msg)
	if err != nil {
		go command.splitMessage(msg, msg.Text, err)
	}
	for _, adminChatID := range command.Buff.Chat[chatID].AdminChats {
		msg.BaseChat.ChatID = adminChatID
		_, err = command.Bot.Send(msg)
		if err != nil {
			go command.splitMessage(msg, msg.Text, err)
		}
	}
	for _, patrulChatID := range command.Buff.Chat[chatID].PatrulChats {
		msg.BaseChat.ChatID = patrulChatID
		_, err = command.Bot.Send(msg)
		if err != nil {
			go command.splitMessage(msg, msg.Text, err)
		}
	}
}

//UpdateViolation ...
func (command *CommandAPI) UpdateViolation(message *tgbotapi.Message, isWarn bool) {
	_, reason, time, err := command.DB.CommandHandler(message)
	if err != nil && time == nil {
		command.Bot.Send(tgbotapi.NewMessage(message.Chat.ID,
			text.SomethingWentWrong[command.Buff.Chat[message.Chat.ID].Language]))
		return
	}

	var chatID int64
	if command.Buff.Chat[message.Chat.ID].IsMain {
		chatID = message.Chat.ID
	} else {
		chatID = command.Buff.Chat[message.Chat.ID].FatherChat
	}

	violation, err := command.DB.UpdateViolation(chatID, time.TimeNumber, reason, isWarn)
	if err != nil {
		command.Bot.Send(tgbotapi.NewMessage(message.Chat.ID, err.Error()))
		return
	}
	user, found := command.DB.GetUserByUserID(violation.UserID)
	if !found {
		command.Bot.Send(tgbotapi.NewMessage(message.Chat.ID,
			text.SomethingWentWrong[command.Buff.Chat[message.Chat.ID].Language]))
		return
	}

	msg := tgbotapi.NewMessage(message.Chat.ID,
		fmt.Sprintf(text.WarnHasBeenUpdated[command.Buff.Chat[message.Chat.ID].Language], violation, user))
	msg.ParseMode = "markdown"
	_, err = command.Bot.Send(msg)
	if err != nil {
		go command.splitMessage(msg, msg.Text, err)
	}

	for _, adminChatID := range command.Buff.Chat[chatID].AdminChats {
		msg.BaseChat.ChatID = adminChatID
		_, err = command.Bot.Send(msg)
		if err != nil {
			go command.splitMessage(msg, msg.Text, err)
		}
	}
	for _, patrulChatID := range command.Buff.Chat[chatID].PatrulChats {
		msg.BaseChat.ChatID = patrulChatID
		_, err = command.Bot.Send(msg)
		if err != nil {
			go command.splitMessage(msg, msg.Text, err)
		}
	}
}

//ViolationAutoRemove ...
func (command *CommandAPI) ViolationAutoRemove() {
	userWithViolations, err := command.DB.ViolationAutoRemove()
	if err != nil {
		return
	}
	var (
		msgText               string                                         = ""
		chatIDtoUserWithWarns map[int64]map[*models.User][]*models.Violation = make(map[int64]map[*models.User][]*models.Violation)
	)

	for user, violations := range userWithViolations { //send to users
		if len(violations) > 1 {
			msgText += "З вас було знято наступні порушення\nYour violations have expired\n---\n"
		} else {
			msgText += "З вас було зняте порушення\nYour violation has expired\n---\n"
		}
		for _, violation := range violations {
			msgText += fmt.Sprintf(text.WarnAutoRemove[command.Buff.Chat[violation.ChatID].Language],
				command.Buff.Chat[violation.ChatID].Title, violation.Reason)

			if chatIDtoUserWithWarns[violation.ChatID] == nil {
				chatIDtoUserWithWarns[violation.ChatID] = make(map[*models.User][]*models.Violation)
			}
			chatIDtoUserWithWarns[violation.ChatID][user] = append(chatIDtoUserWithWarns[violation.ChatID][user], violation)

		}
		msg := tgbotapi.NewMessage(int64(user.UserID), msgText)
		msg.ParseMode = "markdown"
		_, err = command.Bot.Send(msg)
		if err != nil {
			go command.splitMessage(msg, msg.Text, err)
		}
		msgText = ""
	}

	var msg tgbotapi.MessageConfig = tgbotapi.MessageConfig{
		ParseMode: "markdown",
		BaseChat:  tgbotapi.BaseChat{},
	}

	for chatID, userWithWarns := range chatIDtoUserWithWarns {
		for user, violations := range userWithWarns {
			firstname := tools.AvoidMarkdownCrashFirstname([]byte(user.FirstName))
			msg.Text += fmt.Sprintf(text.WarnHasBeenRemoved[command.Buff.Chat[chatID].Language],
				firstname, user.UserID, user.UserID)
			for _, violation := range violations {
				msg.Text += fmt.Sprintf("%s\n\n", violation)
			}
		}
		for _, adminChatID := range command.Buff.Chat[chatID].AdminChats {
			msg.BaseChat.ChatID = adminChatID
			_, err = command.Bot.Send(msg)
			if err != nil {
				go command.splitMessage(msg, msg.Text, err)
			}
		}

		msg.Text = ""
	}

}

//AFK

//AFKUser adds afk to the list
func (command *CommandAPI) AFKUser(message *tgbotapi.Message) {
	users, reason, _, err := command.DB.CommandHandler(message)
	if err != nil {
		command.Bot.Send(tgbotapi.NewMessage(message.Chat.ID,
			text.SomethingWentWrong[command.Buff.Chat[message.Chat.ID].Language]))
		return
	}
	//reply to true mafia
	if message.ReplyToMessage != nil &&
		(message.ReplyToMessage.From.ID == 468253535 || message.ReplyToMessage.From.ID == 761250017) &&
		message.ReplyToMessage.Entities != nil {
		users = nil
		for _, v := range *message.ReplyToMessage.Entities {
			if v.Type == "text_mention" {
				u := &models.User{
					UserID:    v.User.ID,
					FirstName: v.User.FirstName,
					LastName:  v.User.LastName,
					UserName:  v.User.UserName,
				}
				users = append(users, u)
			}
		}
	}

	go command.DB.AFKUser(message, users, reason)

	var (
		chatID     int64
		singular   bool = len(users) == 1
		usersNames string
	)
	if command.Buff.Chat[message.Chat.ID].IsMain { //if you want to give warn from admin chat
		chatID = message.Chat.ID
	} else {
		chatID = command.Buff.Chat[message.Chat.ID].FatherChat
	}

	for _, user := range users {
		firstname := tools.AvoidMarkdownCrashFirstname([]byte(user.FirstName))
		usersNames += fmt.Sprintf(" [%s](tg://user?id=%d),", firstname, user.UserID)
	}
	var msg tgbotapi.MessageConfig = tgbotapi.MessageConfig{
		BaseChat:  tgbotapi.BaseChat{},
		ParseMode: "markdown",
	}

	//to chat
	if singular {
		msg.Text = fmt.Sprintf(text.AddAFKSingular[command.Buff.Chat[chatID].Language], usersNames)
		msg.BaseChat.ChatID = chatID
		_, err = command.Bot.Send(msg)
		if err != nil {
			go command.splitMessage(msg, msg.Text, err)
		}
	} else {
		msg.Text = fmt.Sprintf(text.AddAFKPlural[command.Buff.Chat[chatID].Language], usersNames)
		msg.BaseChat.ChatID = chatID
		_, err = command.Bot.Send(msg)
		if err != nil {
			go command.splitMessage(msg, msg.Text, err)
		}
	}

	//to admins and patruls
	if singular {
		usersNames += fmt.Sprintf(" (*ID:* %d)", users[0].UserID)
		if reason != "" {
			msg.Text = fmt.Sprintf(text.AddAFKSingularWithReasonAdmin[command.Buff.Chat[chatID].Language], usersNames)
		} else {
			msg.Text = fmt.Sprintf(text.AddAFKSingularWithoutReasonAdmin[command.Buff.Chat[chatID].Language], usersNames)

		}
		for _, adminChatID := range command.Buff.Chat[chatID].AdminChats {
			msg.BaseChat.ChatID = adminChatID
			_, err = command.Bot.Send(msg)
			if err != nil {
				go command.splitMessage(msg, msg.Text, err)
			}
		}
		//to patruls
		if reason != "" {
			msg.Text = fmt.Sprintf(text.AddAFKSingularWithReasonPatrul[command.Buff.Chat[chatID].Language], time.Unix(int64(message.Date), 0).Format("02.01.2006 15:04:05"), usersNames)
		} else {
			msg.Text = fmt.Sprintf(text.AddAFKSingularWithoutReasonPatrul[command.Buff.Chat[chatID].Language], time.Unix(int64(message.Date), 0).Format("02.01.2006 15:04:05"), usersNames)

		}
		for _, patrulChatID := range command.Buff.Chat[chatID].PatrulChats {
			msg.BaseChat.ChatID = patrulChatID
			_, err = command.Bot.Send(msg)
			if err != nil {
				go command.splitMessage(msg, msg.Text, err)
			}
		}
	} else { //more than 1 person
		//to admins
		if reason != "" {
			msg.Text = fmt.Sprintf(text.AddAFKPluralWithReasonAdmin[command.Buff.Chat[chatID].Language], usersNames)
		} else {
			msg.Text = fmt.Sprintf(text.AddAFKPluralWithoutReasonAdmin[command.Buff.Chat[chatID].Language], usersNames)

		}
		for _, adminChatID := range command.Buff.Chat[chatID].AdminChats {
			msg.BaseChat.ChatID = adminChatID
			_, err = command.Bot.Send(msg)
			if err != nil {
				go command.splitMessage(msg, msg.Text, err)
			}
		}
		//to patruls
		if reason != "" {
			msg.Text = fmt.Sprintf(text.AddAFKPluralWithReasonPatrul[command.Buff.Chat[chatID].Language], time.Unix(int64(message.Date), 0).Format("02.01.2006 15:04:05"), usersNames)
		} else {
			msg.Text = fmt.Sprintf(text.AddAFKPluralWithoutReasonPatrul[command.Buff.Chat[chatID].Language], time.Unix(int64(message.Date), 0).Format("02.01.2006 15:04:05"), usersNames)
		}

		for _, patrulChatID := range command.Buff.Chat[chatID].PatrulChats {
			msg.BaseChat.ChatID = patrulChatID
			_, err = command.Bot.Send(msg)
			if err != nil {
				go command.splitMessage(msg, msg.Text, err)
			}
		}
	}
}

//UnAFKUser changes afk status to passive
func (command *CommandAPI) UnAFKUser(message *tgbotapi.Message) {
	_, _, time, err := command.DB.CommandHandler(message)
	if err != nil {
		command.Bot.Send(tgbotapi.NewMessage(message.Chat.ID,
			text.SomethingWentWrong[command.Buff.Chat[message.Chat.ID].Language]))
		return
	}
	var (
		msgText []string
		chatID  int64
	)
	if command.Buff.Chat[message.Chat.ID].IsMain {
		chatID = message.Chat.ID
	} else {
		chatID = command.Buff.Chat[message.Chat.ID].FatherChat
	}

	violation, err := command.DB.UnwarnUsers(chatID, time.TimeNumber, false)
	if err != nil {
		command.Bot.Send(tgbotapi.NewMessage(message.Chat.ID, err.Error()))
		return
	}
	user, found := command.DB.GetUserByUserID(violation.UserID)
	if !found {
		command.Bot.Send(tgbotapi.NewMessage(message.Chat.ID,
			text.SomethingWentWrong[command.Buff.Chat[message.Chat.ID].Language]))
		return
	}

	firstname := tools.AvoidMarkdownCrashFirstname([]byte(user.FirstName))
	msgText = append(msgText, fmt.Sprintf("[%s](tg://user?id=%d)", firstname, user.UserID))

	msg := tgbotapi.NewMessage(chatID,
		fmt.Sprintf(text.UnAFK[command.Buff.Chat[chatID].Language], strings.Join(msgText, "")))
	msg.ParseMode = "markdown"

	_, err = command.Bot.Send(msg)
	if err != nil {
		go command.splitMessage(msg, msg.Text, err)
	}

	// if len(users) == 1 {
	//for _, user := range users {
	//	firstname := tools.AvoidMarkdownCrashFirstname([]byte(user.FirstName))
	msgText = append([]string{}, fmt.Sprintf("[%s](tg://user?id=%d) (*ID:*`%d`)", firstname, user.UserID, user.UserID))
	//	}
	msg.Text = fmt.Sprintf(text.UnAFK[command.Buff.Chat[chatID].Language], strings.Join(msgText, ""))
	// }

	for _, adminchat := range command.Buff.Chat[chatID].AdminChats {
		msg.BaseChat.ChatID = adminchat
		_, err := command.Bot.Send(msg)
		if err != nil {
			go command.splitMessage(msg, msg.Text, err)
		}
	}
	for _, patrulchat := range command.Buff.Chat[chatID].PatrulChats {
		msg.BaseChat.ChatID = patrulchat
		_, err := command.Bot.Send(msg)
		if err != nil {
			go command.splitMessage(msg, msg.Text, err)
		}
	}
}

/*GetAFK returns list of user warns
Firstname *ID*:`32132132`

date time reason

Vsiogo warns: 9
*/
func (command *CommandAPI) GetAFK(message *tgbotapi.Message) {
	users, _, _, err := command.DB.CommandHandler(message)
	if err != nil {
		command.Bot.Send(tgbotapi.NewMessage(message.Chat.ID,
			text.SomethingWentWrong[command.Buff.Chat[message.Chat.ID].Language]))
		return
	}
	var (
		msgText    string
		activeWarn int
		chatID     int64
	)
	if command.Buff.Chat[message.Chat.ID].IsMain {
		chatID = message.Chat.ID
	} else {
		chatID = command.Buff.Chat[message.Chat.ID].FatherChat
	}
	userWithWarns, err := command.DB.GetUsersWithViolations(chatID, users, false)
	if err != nil {
		command.Bot.Send(tgbotapi.NewMessage(message.Chat.ID,
			text.SomethingWentWrong[command.Buff.Chat[message.Chat.ID].Language]))
		return
	}

	for user, warns := range userWithWarns {
		firstname := tools.AvoidMarkdownCrashFirstname([]byte(user.FirstName))
		msgText += fmt.Sprintf("[%s](tg://user?id=%d) *ID:* `%d`\n\n", firstname, user.UserID, user.UserID)
		for _, warn := range warns {
			if warn.Status {
				if warn.Reason != "" {
					msgText += fmt.Sprintf("%d. \\[%s] %s\n", warn.ID, time.Unix(warn.Date, 0).Format("02.01 15:04"), warn.Reason)
				} else {
					msgText += fmt.Sprintf("%d. \\[%s] %s\n", warn.ID, time.Unix(warn.Date, 0).Format("02.01 15:04"), text.AbsentReason[command.Buff.Chat[message.Chat.ID].Language])

				}
				activeWarn++
			}
		}
		msgText += fmt.Sprintf(text.GetAFK[command.GetChatLanguage(chatID)], len(warns), activeWarn)
		activeWarn = 0
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, msgText)
	msg.ParseMode = "markdown"
	_, err = command.Bot.Send(msg)
	if err != nil {
		go command.splitMessage(msg, msg.Text, err)
	}

}

//PhotoAFK sends photo of afk statistics
func (command *CommandAPI) PhotoAFK(message *tgbotapi.Message) {
	var chatID int64

	if command.Buff.Chat[message.Chat.ID].IsMain {
		chatID = message.Chat.ID
	} else {
		chatID = command.Buff.Chat[message.Chat.ID].FatherChat
	}

	afkStats, err := command.DB.GetAFKInChat(chatID)
	if err != nil {
		command.Bot.Send(tgbotapi.NewMessage(message.Chat.ID,
			text.SomethingWentWrong[command.Buff.Chat[message.Chat.ID].Language]))
		return
	}

	var daysAmount int = 20
	msgSplitted := strings.Split(message.Text, " ")
	if len(msgSplitted) > 1 {
		daysAmount, err = strconv.Atoi(msgSplitted[1])
		if err != nil {
			daysAmount = 20
		}
	}

	graph.CreateImage(afkStats, daysAmount, chatID)
	photo := tgbotapi.NewPhotoUpload(message.Chat.ID,
		fmt.Sprintf("./internal/graph/%d_graph.png", chatID))
	command.Bot.Send(photo)
}

//TestCommand ...
func (command *CommandAPI) TestCommand(message *tgbotapi.Message) {

}

//splitMessage cuts text into smaller peaces if error appeared
func (command *CommandAPI) splitMessage(msg tgbotapi.MessageConfig, text string, err error) {
	if err.Error() == "Bad Request: message is too long" {
		howManyPieces := len(text) / 4096
		for i := 0; i <= howManyPieces; i++ {
			if i == howManyPieces {
				msg.Text = text[i*4096:]
				command.Bot.Send(msg)
				return
			}
			msg.Text = text[i*4096 : (i+1)*4096]
			command.Bot.Send(msg)

		}
	} else if string(err.Error()[0:34]) == "Bad Request: can't parse entities:" {
		splittedError := strings.Split(err.Error(), " ")
		index, err := strconv.Atoi(splittedError[len(splittedError)-1]) //last symbol is offset of unclosed entity
		if err != nil {
			return
		}
		howManyPieces := len(text) / (index - 1)

		for i := 0; i <= howManyPieces; i++ {
			if i == howManyPieces {
				msg.Text = text[i*(index-1):]
				command.Bot.Send(msg)
				return
			}
			msg.Text = text[i*(index-1) : (i+1)*(index-1)]
			command.Bot.Send(msg)
		}
	} else if len(err.Error()) > 20 && string(err.Error()[0:18]) == "Too Many Requests:" {
		timeToSleep := strings.Split(err.Error(), " ")
		parseTime, _ := strconv.Atoi(timeToSleep[len(timeToSleep)-1])
		time.Sleep(time.Duration(parseTime) * time.Millisecond)
		command.Bot.Send(msg)
	}
}

//CheckAdmins checks if user is admin
func (command *CommandAPI) CheckAdmins(message *tgbotapi.Message) bool {
	if command.Buff.Chat[message.Chat.ID].IsMain {
		for _, v := range command.Buff.Chat[message.Chat.ID].Admins {
			if v.UserID == message.From.ID {
				return true
			}
		}
	} else {
		for _, v := range command.GetFatherChat(message.Chat.ID).Admins {
			if v.UserID == message.From.ID {
				return true
			}
		}
	}

	return false
}

//CheckAdminsAndPatruls returns if current user is admin or patruls
func (command *CommandAPI) CheckAdminsAndPatruls(message *tgbotapi.Message) bool {

	var chatID int64
	if command.Buff.Chat[message.Chat.ID].IsMain {
		chatID = message.Chat.ID
	} else {
		chatID = command.Buff.Chat[message.Chat.ID].FatherChat
	}

	if len(command.Buff.Chat[chatID].Admins) >= len(command.Buff.Chat[chatID].Patruls) {
		for i := 0; i < len(command.Buff.Chat[chatID].Admins)-1; i++ {
			if command.Buff.Chat[chatID].Admins[i].UserID == message.From.ID {
				return true
			}
			if i < len(command.Buff.Chat[chatID].Patruls) {
				if command.Buff.Chat[chatID].Patruls[i].UserID == message.From.ID {
					return true
				}
			}
		}
	} else {
		for i := 0; i < len(command.Buff.Chat[chatID].Patruls)-1; i++ {
			if command.Buff.Chat[chatID].Patruls[i].UserID == message.From.ID {
				return true
			}
			if i < len(command.Buff.Chat[chatID].Admins) {
				if command.Buff.Chat[chatID].Admins[i].UserID == message.From.ID {
					return true
				}
			}
		}
	}

	return false
}

//BanUsers command for banning user for some time
func (command *CommandAPI) BanUsers(message *tgbotapi.Message) {
	users, reason, time, err := command.DB.CommandHandler(message)
	if err != nil {
		command.Bot.Send(tgbotapi.NewMessage(message.Chat.ID,
			text.SomethingWentWrong[command.Buff.Chat[message.Chat.ID].Language]))
		return
	}
	//send message about warn
	var (
		singular       bool = len(users) == 1
		msgText        []string
		firstnameAdmin = tools.AvoidMarkdownCrashFirstname([]byte(message.From.FirstName))
		chatID         int64
	)

	if command.Buff.Chat[message.Chat.ID].IsMain { //if you want to give warn from admin chat
		chatID = message.Chat.ID
	} else {
		chatID = command.Buff.Chat[message.Chat.ID].FatherChat
	}

	for _, user := range users {
		msgText = append(msgText, fmt.Sprintf("[%s](tg://user?id=%d)", user.FirstName, user.UserID))
		chatMemberConfig := tgbotapi.ChatMemberConfig{
			ChatID: message.Chat.ID,
			UserID: user.UserID,
		}
		kickMemberConfig := tgbotapi.KickChatMemberConfig{
			ChatMemberConfig: chatMemberConfig,
			UntilDate:        time.TimeUnix,
		}
		command.Bot.KickChatMember(kickMemberConfig)
		//all func for singular
		if singular {
			//with reason
			if reason != "" {
				msg := tgbotapi.MessageConfig{
					Text: fmt.Sprintf(
						text.ChatBanWithReasonSingular[command.Buff.Chat[chatID].Language],
						strings.Join(msgText, ""), reason),
					BaseChat: tgbotapi.BaseChat{
						ChatID: chatID,
					},
					ParseMode: "markdown",
				}
				_, err := command.Bot.Send(msg)
				if err != nil {
					go command.splitMessage(msg, msg.Text, err)
				}

				//message about forever ban
				if time.TimeUnix == 0 {
					msg.Text = fmt.Sprintf(
						text.AdminBanWithReasonForeverSingular[command.Buff.Chat[chatID].Language],
						firstnameAdmin, message.From.ID, strings.Join(msgText, ""), user.UserID, reason)
				} else { //show time
					msg.Text = fmt.Sprintf(
						text.AdminBanWithReasonSingular[command.Buff.Chat[chatID].Language],
						firstnameAdmin, message.From.ID, strings.Join(msgText, ""), user.UserID, reason, time.TimeNumber, time.TimeMagnitude)
				}

				for _, adminChatID := range command.Buff.Chat[chatID].AdminChats {
					msg.BaseChat.ChatID = adminChatID
					_, err := command.Bot.Send(msg)
					if err != nil {
						go command.splitMessage(msg, msg.Text, err)
					}
				}
			} else { //without reason
				msg := tgbotapi.MessageConfig{
					Text: fmt.Sprintf(
						text.ChatBanWithoutReasonSingular[command.Buff.Chat[chatID].Language],
						strings.Join(msgText, "")),
					BaseChat: tgbotapi.BaseChat{
						ChatID: chatID,
					},
					ParseMode: "markdown",
				}
				_, err := command.Bot.Send(msg)
				if err != nil {
					go command.splitMessage(msg, msg.Text, err)
				}

				//message about forever ban
				if time.TimeUnix == 0 {
					msg.Text = fmt.Sprintf(
						text.AdminBanWithoutReasonForeverSingular[command.Buff.Chat[chatID].Language],
						firstnameAdmin, message.From.ID, strings.Join(msgText, ""), user.UserID)
				} else { //show time
					msg.Text = fmt.Sprintf(
						text.AdminBanWithoutReasonSingular[command.Buff.Chat[chatID].Language],
						firstnameAdmin, message.From.ID, strings.Join(msgText, ""), user.UserID, time.TimeNumber, time.TimeMagnitude)
				}

				for _, adminChatID := range command.Buff.Chat[chatID].AdminChats {
					msg.BaseChat.ChatID = adminChatID
					_, err := command.Bot.Send(msg)
					if err != nil {
						go command.splitMessage(msg, msg.Text, err)
					}
				}
			}
			return
		}
	}

	if reason != "" {
		msg := tgbotapi.NewMessage(chatID,
			fmt.Sprintf(text.ChatBanWithReasonPlural[command.Buff.Chat[chatID].Language],
				strings.Join(msgText, ", "), reason))
		msg.ParseMode = "markdown"
		_, err := command.Bot.Send(msg)
		if err != nil {
			go command.splitMessage(msg, msg.Text, err)
		}
	} else {
		msg := tgbotapi.NewMessage(chatID,
			fmt.Sprintf(text.ChatBanWithoutReasonPlural[command.Buff.Chat[chatID].Language],
				strings.Join(msgText, ", ")))
		msg.ParseMode = "markdown"
		_, err := command.Bot.Send(msg)
		if err != nil {
			go command.splitMessage(msg, msg.Text, err)
		}
	}

	msg := tgbotapi.MessageConfig{
		ParseMode: "markdown",
		BaseChat:  tgbotapi.BaseChat{},
	}
	if reason != "" {
		if time.TimeUnix == 0 {
			msg.Text = fmt.Sprintf(
				text.AdminBanWithReasonForeverPlural[command.Buff.Chat[chatID].Language],
				firstnameAdmin, message.From.ID, strings.Join(msgText, ""), reason)
		} else { //show time
			msg.Text = fmt.Sprintf(
				text.AdminBanWithReasonPlural[command.Buff.Chat[chatID].Language],
				firstnameAdmin, message.From.ID, strings.Join(msgText, ""), reason, time.TimeNumber, time.TimeMagnitude)
		}
	} else {
		if time.TimeUnix == 0 {
			msg.Text = fmt.Sprintf(
				text.AdminBanWithoutReasonForeverPlural[command.Buff.Chat[chatID].Language],
				firstnameAdmin, message.From.ID, strings.Join(msgText, ""))
		} else { //show time
			msg.Text = fmt.Sprintf(
				text.AdminBanWithoutReasonPlural[command.Buff.Chat[chatID].Language],
				firstnameAdmin, message.From.ID, strings.Join(msgText, ""), time.TimeNumber, time.TimeMagnitude)
		}
	}

	for _, adminChatID := range command.Buff.Chat[chatID].AdminChats {
		msg.BaseChat.ChatID = adminChatID
		_, err := command.Bot.Send(msg)
		if err != nil {
			go command.splitMessage(msg, msg.Text, err)
		}
	}
}

//UnbanUsers command for unbanning user
func (command *CommandAPI) UnbanUsers(message *tgbotapi.Message) {
	users, _, _, err := command.DB.CommandHandler(message)
	if err != nil {
		command.Bot.Send(tgbotapi.NewMessage(message.Chat.ID,
			text.SomethingWentWrong[command.Buff.Chat[message.Chat.ID].Language]))
		return
	}
	//send message about warn
	var (
		chatID  int64
		msgText []string
	)

	if command.Buff.Chat[message.Chat.ID].IsMain { //if you want to give warn from admin chat
		chatID = message.Chat.ID
	} else {
		chatID = command.Buff.Chat[message.Chat.ID].FatherChat
	}
	for _, user := range users {
		chatMemberConfig := tgbotapi.ChatMemberConfig{
			ChatID: chatID,
			UserID: user.UserID,
		}
		command.Bot.UnbanChatMember(chatMemberConfig)
		firstName := tools.AvoidMarkdownCrashFirstname([]byte(user.FirstName))
		msgText = append(msgText, fmt.Sprintf("[%s](tg://user?id=%d)",firstName , user.UserID))
	}


	msg := tgbotapi.NewMessage(0, 
		fmt.Sprintf(text.Unban[command.GetChatLanguage(chatID)], 
		strings.Join(msgText, ", "), command.Buff.Chat[chatID].Title))
	msg.ParseMode = "markdown"
	
	for _, adminChat := range command.Buff.Chat[chatID].AdminChats {
		msg.BaseChat.ChatID = adminChat
		_, err := command.Bot.Send(msg)
		if err != nil {
			go command.splitMessage(msg, msg.Text, err)
		}
	}

}

//MuteUsers functions that mutes users
func (command *CommandAPI) MuteUsers(message *tgbotapi.Message) {
	users, reason, time, err := command.DB.CommandHandler(message)
	if err != nil {
		command.Bot.Send(tgbotapi.NewMessage(message.Chat.ID,
			text.SomethingWentWrong[command.Buff.Chat[message.Chat.ID].Language]))
		return
	}

	command.muteUsers(users, message, reason, time)
}


//UnmuteUsers command for unmuting users
func (command *CommandAPI) UnmuteUsers(message *tgbotapi.Message) {
	users, _, _, err := command.DB.CommandHandler(message)
	if err != nil {
		command.Bot.Send(tgbotapi.NewMessage(message.Chat.ID,
			text.SomethingWentWrong[command.Buff.Chat[message.Chat.ID].Language]))
		return
	}
	//send message about warn
	var (
		chatID  int64
		msgText []string
		t bool = true
		firstnameAdmin = tools.AvoidMarkdownCrashFirstname([]byte(message.From.FirstName))
		restrictChatMemberConfig = tgbotapi.RestrictChatMemberConfig{
			UntilDate:             1,
			CanSendMessages:       &t,
			CanSendMediaMessages:  &t,
			CanSendOtherMessages:  &t,
			CanAddWebPagePreviews: &t,
		}
	)

	if command.Buff.Chat[message.Chat.ID].IsMain { //if you want to give warn from admin chat
		chatID = message.Chat.ID
	} else {
		chatID = command.Buff.Chat[message.Chat.ID].FatherChat
	}
	for _, user := range users {
		chatMemberConfig := tgbotapi.ChatMemberConfig{
			ChatID: chatID,
			UserID: user.UserID,
		}
		restrictChatMemberConfig.ChatMemberConfig = chatMemberConfig
		chatMember, _ := command.Bot.GetChatMember(tgbotapi.ChatConfigWithUser{
			ChatID: chatID,
			UserID: user.UserID,
		})

		if chatMember.Status != "kicked" {
		//unmuting
		command.Bot.RestrictChatMember(restrictChatMemberConfig)			
		}
		firstName := tools.AvoidMarkdownCrashFirstname([]byte(user.FirstName))
		msgText = append(msgText, fmt.Sprintf("[%s](tg://user?id=%d)",firstName , user.UserID))
	}


	msg := tgbotapi.NewMessage(0, 
		fmt.Sprintf(text.UnmuteAdmin[command.GetChatLanguage(chatID)], 
		firstnameAdmin, strings.Join(msgText, ", "), command.Buff.Chat[chatID].Title))
	msg.ParseMode = "markdown"
	
	for _, adminChat := range command.Buff.Chat[chatID].AdminChats {
		msg.BaseChat.ChatID = adminChat
		_, err := command.Bot.Send(msg)
		if err != nil {
			go command.splitMessage(msg, msg.Text, err)
		}
	}
	if len(users) == 1 {
		msg.Text = fmt.Sprintf(text.UnmuteSingular[command.GetChatLanguage(chatID)], 
		strings.Join(msgText, ", "))
	} else {
		msg.Text = fmt.Sprintf(text.UnmutePlural[command.GetChatLanguage(chatID)], 
		strings.Join(msgText, ", "))
	}
	msg.BaseChat.ChatID = chatID
	_, err = command.Bot.Send(msg)
	if err != nil {
		go command.splitMessage(msg, msg.Text, err)
	}
}

//SendInfo sends main user info in chat; cannot send more that 1 info per time
func (command *CommandAPI) SendInfo(message *tgbotapi.Message) {
	users, _, _, err := command.DB.CommandHandler(message)
	if err != nil {
		command.Bot.Send(tgbotapi.NewMessage(message.Chat.ID,
			text.SomethingWentWrong[command.Buff.Chat[message.Chat.ID].Language]))
		return
	}
	var (
		chatID  int64
		user *models.User = users[0]
	)

	if command.Buff.Chat[message.Chat.ID].IsMain { //if you want to give warn from admin chat
		chatID = message.Chat.ID
	} else {
		chatID = command.Buff.Chat[message.Chat.ID].FatherChat
	}

	userUsersInChat, err := command.DB.GetUserFromUsersInChatByID(user.UserID, chatID)
	if err != nil {
		command.Bot.Send(tgbotapi.NewMessage(message.Chat.ID,
			text.SomethingWentWrong[command.Buff.Chat[message.Chat.ID].Language]))
		return
	}
	userWarns, err := command.DB.GetViolationsByUserIDinChat(user.UserID, chatID, true)
	if err != nil {
		command.Bot.Send(tgbotapi.NewMessage(message.Chat.ID,
			text.SomethingWentWrong[command.Buff.Chat[message.Chat.ID].Language]))
		return
	}
	userAfks, err := command.DB.GetViolationsByUserIDinChat(user.UserID, chatID, false)
	if err != nil {
		command.Bot.Send(tgbotapi.NewMessage(message.Chat.ID,
			text.SomethingWentWrong[command.Buff.Chat[message.Chat.ID].Language]))
		return
	}

	text := command.getInfoText(user, userUsersInChat, userWarns, userAfks, command.Buff.Chat[chatID].Language)
	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	msg.ParseMode = "markdown"
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Close", "close;"+strconv.Itoa(message.From.ID)),
		),
	)

	_, err = command.Bot.Send(msg)
	if err != nil {
		go command.splitMessage(msg, msg.Text, err)
	}
}


//SendInfoFull sends full user info
func (command *CommandAPI) SendInfoFull(message *tgbotapi.Message) {
	users, _, _, err := command.DB.CommandHandler(message)
	if err != nil {
		command.Bot.Send(tgbotapi.NewMessage(message.Chat.ID,
			text.SomethingWentWrong[command.Buff.Chat[message.Chat.ID].Language]))
		return
	}
	var (
		chatID  int64
		user *models.User = users[0]
	)

	if command.Buff.Chat[message.Chat.ID].IsMain { //if you want to give warn from admin chat
		chatID = message.Chat.ID
	} else {
		chatID = command.Buff.Chat[message.Chat.ID].FatherChat
	}

	userUsersInChat, err := command.DB.GetUserFromUsersInChatByID(user.UserID, chatID)
	if err != nil {
		command.Bot.Send(tgbotapi.NewMessage(message.Chat.ID,
			text.SomethingWentWrong[command.Buff.Chat[message.Chat.ID].Language]))
		return
	}
	userWarns, err := command.DB.GetViolationsByUserIDinChat(user.UserID, chatID, true)
	if err != nil {
		command.Bot.Send(tgbotapi.NewMessage(message.Chat.ID,
			text.SomethingWentWrong[command.Buff.Chat[message.Chat.ID].Language]))
		return
	}
	userAfks, err := command.DB.GetViolationsByUserIDinChat(user.UserID, chatID, false)
	if err != nil {
		command.Bot.Send(tgbotapi.NewMessage(message.Chat.ID,
			text.SomethingWentWrong[command.Buff.Chat[message.Chat.ID].Language]))
		return
	}
	addedBy, _ := command.DB.GetUserByUserID(userUsersInChat.AddedBy)
	text := command.getFullInfoText(user, addedBy, userUsersInChat, userWarns, userAfks, command.Buff.Chat[chatID].Language)
	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	msg.ParseMode = "markdown"
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Close", "close;"+strconv.Itoa(message.From.ID)),
		),
	)

	_, err = command.Bot.Send(msg)
	if err != nil {
		go command.splitMessage(msg, msg.Text, err)
	}
}

func (command *CommandAPI) getInfoText(user *models.User, userInChat *models.UserInChat, warns, afk []*models.Violation, language string) string {
	var msgText []string

	firstName := tools.AvoidMarkdownCrashFirstname([]byte(user.FirstName))
	msgText = append(msgText, fmt.Sprintf(text.Participant[language], firstName, user.UserID))//Participant: name
	msgText = append(msgText, fmt.Sprintf("*ID:* `%d`", user.UserID))//ID: 123
	msgText = append(msgText, fmt.Sprintf(text.FirstMessage[language], time.Unix(userInChat.Data, 0).Format("2006.01.02")))//First message: 2020.20.20
	msgText = append(msgText, fmt.Sprintf(text.MessageAmount[language], userInChat.MessageAmount))//Amount of messages: 2313
	countDays, _ := tools.TimeInGroup(userInChat.Data, language)
	msgText = append(msgText, fmt.Sprintf(text.TimeInGroup[language], countDays))//Time in group: 123 days 4142 hours 13213 minutes
	if len(afk) != 0 {
		countActiveAfks := 0
		for i := range afk {
			if afk[i].Status {
				countActiveAfks++
			}
		}
		if countActiveAfks == 1 {
			msgText = append(msgText, fmt.Sprintf(text.AFKSingular[language], len(afk)))//Warns: 3, 
		} else {
			msgText = append(msgText, fmt.Sprintf(text.AFKPlural[language], len(afk), countActiveAfks))//Warns: 3, 
		}
	}
	if len(warns) != 0 {
		countActiveWarns := 0
		comments := text.CommentsToWarns[language]
		for i := range warns {
			if warns[i].Status {
				countActiveWarns++
				comments += fmt.Sprintf("%d. %s\n", countActiveWarns, warns[i].Reason)
			}
		}
		if countActiveWarns == 1 {
			msgText = append(msgText, fmt.Sprintf(text.WarnsSingular[language], len(warns)))//Warns: 3, 
		} else {
			msgText = append(msgText, fmt.Sprintf(text.WarnsPlural[language], len(warns), countActiveWarns))//Warns: 3, 
		}
		msgText = append(msgText, comments)
	}
	
	return strings.Join(msgText, "\n")
}

func (command *CommandAPI) getFullInfoText(user, addedBy *models.User, userInChat *models.UserInChat, warns, afk []*models.Violation, language string) string {
	var msgText []string

	firstName := tools.AvoidMarkdownCrashFirstname([]byte(user.FirstName))
	msgText = append(msgText, fmt.Sprintf("[%s](tg://user?id=%d) (ID: `%d`)", firstName, user.UserID, user.UserID))
	if user.UserName != "" {
		userName := tools.AvoidMarkdownCrashReason([]byte())
		msgText = append(msgText, fmt.Sprintf("*ID:* `%d`", user.UserID))//ID: 123

	}
	msgText = append(msgText, fmt.Sprintf(text.FirstMessage[language], time.Unix(userInChat.Data, 0).Format("2006.01.02")))//First message: 2020.20.20
	msgText = append(msgText, fmt.Sprintf(text.MessageAmount[language], userInChat.MessageAmount))//Amount of messages: 2313
	countDays, _ := tools.TimeInGroup(userInChat.Data, language)
	msgText = append(msgText, fmt.Sprintf(text.TimeInGroup[language], countDays))//Time in group: 123 days 4142 hours 13213 minutes
	if len(afk) != 0 {
		countActiveAfks := 0
		for i := range afk {
			if afk[i].Status {
				countActiveAfks++
			}
		}
		if countActiveAfks == 1 {
			msgText = append(msgText, fmt.Sprintf(text.AFKSingular[language], len(afk)))//Warns: 3, 
		} else {
			msgText = append(msgText, fmt.Sprintf(text.AFKPlural[language], len(afk), countActiveAfks))//Warns: 3, 
		}
	}
	if len(warns) != 0 {
		countActiveWarns := 0
		comments := text.CommentsToWarns[language]
		for i := range warns {
			if warns[i].Status {
				countActiveWarns++
				comments += fmt.Sprintf("%d. %s\n", countActiveWarns, warns[i].Reason)
			}
		}
		if countActiveWarns == 1 {
			msgText = append(msgText, fmt.Sprintf(text.WarnsSingular[language], len(warns)))//Warns: 3, 
		} else {
			msgText = append(msgText, fmt.Sprintf(text.WarnsPlural[language], len(warns), countActiveWarns))//Warns: 3, 
		}
		msgText = append(msgText, comments)
	}
	
	return strings.Join(msgText, "\n")
}