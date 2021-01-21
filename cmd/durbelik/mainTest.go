package main

import (
	"Durbelik3/internal/configs"
	"Durbelik3/internal/text"
	"Durbelik3/pkg/commands"
	"context"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func mainTest() {
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
	for i := 0; i < 3; i++ {
		go processMessages(commandAPI, updates)
	}
	for range updates {

	}
}

//closureFunctions
var (
	addCheckUserTest func(*tgbotapi.Message)
)

func botInitTest(commandAPI *commands.CommandAPI) {
	go commandAPI.MainPrivate()
	addCheckUser = commandAPI.AddCheckUser()
	//add once in 24h funcs (FIRST AUTO WARN REMOVE)
	go oncePerDay(commandAPI)
}

func oncePerDayTest(commandAPI *commands.CommandAPI) {
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

func processMessages(commandAPI *commands.CommandAPI, updates tgbotapi.UpdatesChannel) {
	for update := range updates {

		if update.EditedMessage != nil ||
			update.EditedChannelPost != nil ||
			update.ChannelPost != nil ||
			update.CallbackQuery != nil {
			continue
		}
		//Chan for private updates
		commands.PrivateUpdates <- update
		if update.CallbackQuery != nil {
			//do some stuff
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
				commandAPI.Bot.DeleteMessage(tgbotapi.NewDeleteMessage(update.Message.Chat.ID, update.Message.MessageID))
			}
			//add left chat member handler
			//add process true mafia reply
			//add triggers parsing(wiki, maybe stickers, maybe badwords)
			//add banned sticker parsing
			//Adds user to database. Checks new users. Checks firstmessage
			addCheckUser(update.Message)

			msgSplitted := strings.Split(update.Message.Text, " ")
			switch msgSplitted[0] {
			/*WARNS*/
			case "!warn":
				commandAPI.Bot.DeleteMessage(tgbotapi.NewDeleteMessage(update.Message.Chat.ID, update.Message.MessageID))
				if commandAPI.CheckAdmins(update.Message) {
					go commandAPI.WarnUser(update.Message)
					continue
				}
			case "!unwarn":
				commandAPI.Bot.DeleteMessage(tgbotapi.NewDeleteMessage(update.Message.Chat.ID, update.Message.MessageID))
				if commandAPI.CheckAdmins(update.Message) {
					commandAPI.UnwarnUsers(update.Message)
					continue
				}
			case "!getwarns":
				commandAPI.Bot.DeleteMessage(tgbotapi.NewDeleteMessage(update.Message.Chat.ID, update.Message.MessageID))
				if commandAPI.CheckAdmins(update.Message) {
					go commandAPI.GetWarns(update.Message)
				}
			case "!deletewarn":
				commandAPI.Bot.DeleteMessage(tgbotapi.NewDeleteMessage(update.Message.Chat.ID, update.Message.MessageID))
				if commandAPI.CheckAdmins(update.Message) {
					commandAPI.DeleteViolation(update.Message, true)
					continue
				}
			case "!updatewarn":
				commandAPI.Bot.DeleteMessage(tgbotapi.NewDeleteMessage(update.Message.Chat.ID, update.Message.MessageID))
				if commandAPI.CheckAdmins(update.Message) {
					commandAPI.UpdateViolation(update.Message, true)
					continue
				}
			/*FINISH WARNS*/
			/*START AFK*/
			case "!afk":
				commandAPI.Bot.DeleteMessage(tgbotapi.NewDeleteMessage(update.Message.Chat.ID, update.Message.MessageID))
				if commandAPI.CheckAdminsAndPatruls(update.Message) {
					commandAPI.AFKUser(update.Message)
				}
			case "!unafk":
				commandAPI.Bot.DeleteMessage(tgbotapi.NewDeleteMessage(update.Message.Chat.ID, update.Message.MessageID))
				if commandAPI.CheckAdminsAndPatruls(update.Message) {
					commandAPI.UnAFKUser(update.Message)
				}
			case "!getafk":
				commandAPI.Bot.DeleteMessage(tgbotapi.NewDeleteMessage(update.Message.Chat.ID, update.Message.MessageID))
				if commandAPI.CheckAdminsAndPatruls(update.Message) {
					commandAPI.GetAFK(update.Message)
				}
			case "!deleteafk":
				commandAPI.Bot.DeleteMessage(tgbotapi.NewDeleteMessage(update.Message.Chat.ID, update.Message.MessageID))
				if commandAPI.CheckAdminsAndPatruls(update.Message) {
					commandAPI.DeleteViolation(update.Message, false)
				}
			case "!updateafk":
				commandAPI.Bot.DeleteMessage(tgbotapi.NewDeleteMessage(update.Message.Chat.ID, update.Message.MessageID))
				if commandAPI.CheckAdminsAndPatruls(update.Message) {
					commandAPI.UpdateViolation(update.Message, false)
				}
			case "!afkphoto":
				commandAPI.Bot.DeleteMessage(tgbotapi.NewDeleteMessage(update.Message.Chat.ID, update.Message.MessageID))
				if commandAPI.CheckAdminsAndPatruls(update.Message) {
					go commandAPI.PhotoAFK(update.Message)
				}
			/*FINISH AFK*/
			/*START MAIN COMMANDS*/
			case "!ban":
				commandAPI.Bot.DeleteMessage(tgbotapi.NewDeleteMessage(update.Message.Chat.ID, update.Message.MessageID))
				if commandAPI.CheckAdmins(update.Message) {
					go commandAPI.BanUsers(update.Message)
				}
			case "!unban":
				commandAPI.Bot.DeleteMessage(tgbotapi.NewDeleteMessage(update.Message.Chat.ID, update.Message.MessageID))
				if commandAPI.CheckAdmins(update.Message) {
					commandAPI.UnbanUsers(update.Message)
				}
			case "!mute":
			case "!unmute":
			/*FINISH MAIN COMMANDS*/
			/*START INFO COMMANDS*/
			case "!info":
			case "!infof":
			case "!status":
			case "!commands":
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
			/*FINISH MAIN ADDITIONAL COMMANDS*/
			/*START SUBCOMMANDS*/
			case "@durbelik_bot":
			case "!broadcast": //instead of key
			case "!horoscope":
			case "!points": //count points got for day
			case "!def": //somehow get info from
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
				continue
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
