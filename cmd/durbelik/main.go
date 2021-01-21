package main

import (
	"Durbelik3/internal/configs"
	"Durbelik3/internal/text"
	"Durbelik3/pkg/commands"
	"context"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func main() {
	commandAPI, err := configs.MakeConfig()
	if err != nil {
		panic(err)
	}
	defer commandAPI.MB.Client.Disconnect(context.TODO()) //close connection before error

	u := tgbotapi.UpdateConfig{
		Offset:  0,
		Timeout: 100,
	}

	botInit(commandAPI)

	updates, err := commandAPI.Bot.GetUpdatesChan(u)
	if err != nil {
		panic(err)
	}

	for update := range updates {
		if update.EditedMessage != nil ||
			update.EditedChannelPost != nil ||
			update.ChannelPost != nil {
			continue
		}
		//Chan for private updates
		commands.PrivateUpdates <- update
		if update.CallbackQuery != nil {
			//do some stuff
			//add clothing button
			callbackSplitted := strings.Split(update.CallbackQuery.Data, ";")
			switch callbackSplitted[0] {
			case "close":
				userID := strconv.Itoa(update.CallbackQuery.From.ID)
				if callbackSplitted[1] == userID || commandAPI.CheckAdmins(update.CallbackQuery.Message) {
					commandAPI.Bot.AnswerCallbackQuery(tgbotapi.NewCallback(update.CallbackQuery.ID, text.Hiding[commandAPI.Buff.Chat[update.CallbackQuery.Message.Chat.ID].Language]))
					commandAPI.Bot.DeleteMessage(tgbotapi.NewDeleteMessage(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID))
				} else {
					commandAPI.Bot.AnswerCallbackQuery(tgbotapi.NewCallback(update.CallbackQuery.ID, text.ThatButtonIsNotForYou[commandAPI.Buff.Chat[update.CallbackQuery.Message.Chat.ID].Language]))
				}
			}
		}
		if update.Message != nil {
			//means it is private message
			if update.Message.From != nil && update.Message.Chat.ID == int64(update.Message.From.ID) {
				continue
			}
			//add chat to config if not exists
			if commandAPI.ChatNotExists(update.Message.Chat.ID) {
				configs.AddChatToConfig(update.Message.Chat, commandAPI)
			}
			//remove pinmessages
			if update.Message.PinnedMessage != nil {
				go commandAPI.Bot.DeleteMessage(tgbotapi.NewDeleteMessage(update.Message.Chat.ID, update.Message.MessageID))
			}
			//add left chat member handler
			//add process true mafia reply
			//add triggers parsing(wiki, maybe stickers, maybe badwords)
			//add banned sticker parsing
			//Adds user to database. Checks new users. Checks firstmessage
			//mb voteban process
			go addCheckUser(update.Message)

			msgSplitted := strings.Split(update.Message.Text, " ")
			switch msgSplitted[0] {
			/*WARNS*/
			case "!warn":
				go commandAPI.Bot.DeleteMessage(tgbotapi.NewDeleteMessage(update.Message.Chat.ID, update.Message.MessageID))
				if commandAPI.CheckAdmins(update.Message) {
					go commandAPI.WarnUser(update.Message)
				}
			case "!unwarn":
				go commandAPI.Bot.DeleteMessage(tgbotapi.NewDeleteMessage(update.Message.Chat.ID, update.Message.MessageID))
				if commandAPI.CheckAdmins(update.Message) {
					go commandAPI.UnwarnUsers(update.Message)
				}
			case "!getwarns":
				go commandAPI.Bot.DeleteMessage(tgbotapi.NewDeleteMessage(update.Message.Chat.ID, update.Message.MessageID))
				if commandAPI.CheckAdmins(update.Message) {
					go commandAPI.GetWarns(update.Message)
				}
			case "!deletewarn":
				go commandAPI.Bot.DeleteMessage(tgbotapi.NewDeleteMessage(update.Message.Chat.ID, update.Message.MessageID))
				if commandAPI.CheckAdmins(update.Message) {
					go commandAPI.DeleteViolation(update.Message, true)
				}
			case "!updatewarn":
				go commandAPI.Bot.DeleteMessage(tgbotapi.NewDeleteMessage(update.Message.Chat.ID, update.Message.MessageID))
				if commandAPI.CheckAdmins(update.Message) {
					go commandAPI.UpdateViolation(update.Message, true)
				}
			/*FINISH WARNS*/
			/*START AFK*/
			case "!afk":
				go commandAPI.Bot.DeleteMessage(tgbotapi.NewDeleteMessage(update.Message.Chat.ID, update.Message.MessageID))
				if commandAPI.CheckAdminsAndPatruls(update.Message) {
					go commandAPI.AFKUser(update.Message)
				}
			case "!unafk":
				go commandAPI.Bot.DeleteMessage(tgbotapi.NewDeleteMessage(update.Message.Chat.ID, update.Message.MessageID))
				if commandAPI.CheckAdminsAndPatruls(update.Message) {
					go commandAPI.UnAFKUser(update.Message)
				}
			case "!getafk":
				go commandAPI.Bot.DeleteMessage(tgbotapi.NewDeleteMessage(update.Message.Chat.ID, update.Message.MessageID))
				if commandAPI.CheckAdminsAndPatruls(update.Message) {
					go commandAPI.GetAFK(update.Message)
				}
			case "!deleteafk":
				go commandAPI.Bot.DeleteMessage(tgbotapi.NewDeleteMessage(update.Message.Chat.ID, update.Message.MessageID))
				if commandAPI.CheckAdminsAndPatruls(update.Message) {
					go commandAPI.DeleteViolation(update.Message, false)
				}
			case "!updateafk":
				go commandAPI.Bot.DeleteMessage(tgbotapi.NewDeleteMessage(update.Message.Chat.ID, update.Message.MessageID))
				if commandAPI.CheckAdminsAndPatruls(update.Message) {
					go commandAPI.UpdateViolation(update.Message, false)
				}
			case "!afkphoto":
				go commandAPI.Bot.DeleteMessage(tgbotapi.NewDeleteMessage(update.Message.Chat.ID, update.Message.MessageID))
				if commandAPI.CheckAdminsAndPatruls(update.Message) {
					go commandAPI.PhotoAFK(update.Message)
				}
			/*FINISH AFK*/
			/*START MAIN COMMANDS*/
			case "!ban":
				go commandAPI.Bot.DeleteMessage(tgbotapi.NewDeleteMessage(update.Message.Chat.ID, update.Message.MessageID))
				if commandAPI.CheckAdmins(update.Message) {
					go commandAPI.BanUsers(update.Message)
				}
			case "!unban":
				go commandAPI.Bot.DeleteMessage(tgbotapi.NewDeleteMessage(update.Message.Chat.ID, update.Message.MessageID))
				if commandAPI.CheckAdmins(update.Message) {
					go commandAPI.UnbanUsers(update.Message)
				}
			case "!mute":
				go commandAPI.Bot.DeleteMessage(tgbotapi.NewDeleteMessage(update.Message.Chat.ID, update.Message.MessageID))
				if commandAPI.CheckAdmins(update.Message) {
					go commandAPI.MuteUsers(update.Message)
				}
			case "!unmute":
				go commandAPI.Bot.DeleteMessage(tgbotapi.NewDeleteMessage(update.Message.Chat.ID, update.Message.MessageID))
				if commandAPI.CheckAdmins(update.Message) {
					go commandAPI.UnmuteUsers(update.Message)
				}
			/*FINISH MAIN COMMANDS*/
			/*START INFO COMMANDS*/
			case "!info":
				go commandAPI.Bot.DeleteMessage(tgbotapi.NewDeleteMessage(update.Message.Chat.ID, update.Message.MessageID))
				if commandAPI.CheckAdmins(update.Message) {
					go commandAPI.SendInfo(update.Message)
				}
			case "!infof":
				go commandAPI.Bot.DeleteMessage(tgbotapi.NewDeleteMessage(update.Message.Chat.ID, update.Message.MessageID))
				if commandAPI.CheckAdmins(update.Message) {
					go commandAPI.SendInfoFull(update.Message)
				}
			case "!status":
			case "!commands":
			case "!points":
			case "!mystats":
				//show additional commands
			case "!commandsa":
				//show only in private mode or not in main chat
			case "!commandsp":
				//show only in private mode or not in main chat
			case "!rules":
				//think about personification (SAME FOR WELCOME MESSAGE)
			/*FINISH INFO COMMANDS*/
			/*START MAIN ADDITIONAL COMMANDS*/
			case "!bansticker":
			case "!report":
			case "!callme": //add to call list
			case "!dontcallme": //remove from call list
			case "!couple": //add to couple of the day list
			case "!dontcouple": //remove from couple of the day list
			case "!call":
			case "!coupleoftheday":
			case "!predict":
			case "!def":
			/*FINISH MAIN ADDITIONAL COMMANDS*/
			/*START SUBCOMMANDS*/
			case "@durbelik_bot": //mb hentai gif
			case "!broadcast": //instead of key
			case "!horoscope":
			/*FINISH SUBCOMMANDS*/
			case "!test":
				go commandAPI.TestCommand(update.Message)

			}

			switch update.Message.Command() {
			case "ping":
				go func() {
					msgToDelete, err := commandAPI.Bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID,
						text.Ping[commandAPI.Buff.Chat[update.Message.Chat.ID].Language]))
					if err != nil {
						return
					}
					time.Sleep(5 * time.Second)
					commandAPI.Bot.DeleteMessage(tgbotapi.NewDeleteMessage(msgToDelete.Chat.ID, msgToDelete.MessageID))
				}()
			case "help": //send help message to private messages
			case "callme":
			case "dontcallme":
			case "couple":
			case "dontcouple":
			case "rules":
			case "call":
			case "commands":
			case "coupleoftheday":
			case "points":
			}

		}
	}
}

//closureFunctions
var (
	addCheckUser func(*tgbotapi.Message)
)

func botInit(commandAPI *commands.CommandAPI) {
	go commandAPI.MainPrivate()
	addCheckUser = commandAPI.AddCheckUser()
	//add once in 24h funcs (FIRST AUTO WARN REMOVE)
	go oncePerDay(commandAPI)
}

func oncePerDay(commandAPI *commands.CommandAPI) {
	loc, _ := time.LoadLocation("Local")
	for {
		currTime := time.Now().Add(time.Hour * 24) //go to next date
		startTime := time.Date(                    //truncate it to 00 00
			currTime.Year(),
			currTime.Month(),
			currTime.Day(),
			0, 0, 0, 0, loc)

		time.Sleep(startTime.Sub(time.Now())) //start at midnight

		time.Sleep(12 * time.Hour) //start at midday
		go commandAPI.ViolationAutoRemove()

	}
}
