package commands

import "Durbelik3/pkg/models"

//GetFatherChat returns fatherchatID
func (command *CommandAPI) GetFatherChat(chatID int64) *models.ChatConfig {
	return command.Buff.Chat[command.Buff.Chat[chatID].FatherChat]
}

//GetChatLanguage returns chat language
func (command *CommandAPI) GetChatLanguage(chatID int64) string {
	return command.Buff.Chat[chatID].Language
}

//ChatNotExists checks if chat is in config
func (command *CommandAPI) ChatNotExists(chatID int64) bool {
	if command.Buff.Chat[chatID] == nil {
		return true
	}
	return false
}