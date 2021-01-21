package text

//StickerWarning translate in ukraine and english message
var StickerWarning map[string]string = map[string]string{
	"ua": "Дозволено не більше 5 стікерів за хвилину. Ви надіслали вже 4",
	"en": "It is not allowed to send 5 stickers per minute. You have already sent 4",
}

//GifWarning translate in ukraine and english message
var GifWarning map[string]string = map[string]string{
	"ua": "Дозволено не більше 5 гіфок за хвилину. Ви надіслали вже 4",
	"en": "It is not allowed to send 5 gifs per minute. You have already sent 4",
}

//StickerMute translated
var StickerMute map[string]string = map[string]string{
	"ua": "спам стікерами",
	"en": "sticker spam",
}

//GifMute translated
var GifMute map[string]string = map[string]string{
	"ua": "спам гіфками",
	"en": "gif spam",
}

//WelcomeMessage is a welcome message for every user
var WelcomeMessage map[string]string = map[string]string{
	"ua": "Привіт, [%s](tg://user?id=%d)! Перед початком гри радимо тобі уважно прочитати правила, ролі та поради. Це тобі знадобиться для класної та цікавої гри 🙌🏼\n\n📌 *Правила:*\nhttps://telegra.ph/rules-08-09\n📌 *Ролі:*\nhttps://telegra.ph/roles-08-08\n📌 *Поради для кращої гри:*\nhttps://telegra.ph/advice-08-08\n\nНехай щастить!",
	"en": "Hello, [%s](tg://user?id=%d)! Before you enter the game we suggest you to read next rules, roles and advices. You need it for cool and interesting game 🙌🏼\n\n📌 *Rules:*\nhttps://telegra.ph/rules-08-09\n📌 *Roles:*\nhttps://telegra.ph/roles-08-08\n📌 *Advices for better game*\nhttps://telegra.ph/advice-08-08\n\nGood luck!",
}

//Ping message to show if bot is online
var Ping map[string]string = map[string]string{
	"ua": "Бот зараз в мережі",
	"en": "Bot is currently online",
}

//AbsentReason if reason is undefined
var AbsentReason map[string]string = map[string]string{
	"ua": "не вказано",
	"en": "undefined",
}

//SomethingWentWrong if handler error occured
var SomethingWentWrong map[string]string = map[string]string{
	"ua": "Щось пішло не так...",
	"en": "Something went wrong...",
}

//BanForLink translate in ukraine and english message
var BanForLink map[string]string = map[string]string{
	"ua": "[%s](tg://user?id=%d) було забанено\n\n*Коментар:* перше повідомлення містить посилання",
	"en": "[%s](tg://user?id=%d) was banned\n\n*Reason:* first message contains the link",
}

//ChatWarnWithReasonSingular ...
var ChatWarnWithReasonSingular map[string]string = map[string]string{
	"ua": "*Дур-попередження!*\n\n%s, Ви порушили правила або норми поведінки чату та отримали %d попередження.\nЯкщо зібрати всі 5, то є можливість вирушити у Дурпекло:)\n\n*Коментар:* %s",
	"en": "*Dur-warning!*\n\n%s, you have broken the rules or norms of the chat and already got %d warning.\nIf you collect all 5, you will possibly go to the Durhell:)\n\n*Comment:* %s",
}

//ChatWarnWithoutReasonSingular ...
var ChatWarnWithoutReasonSingular map[string]string = map[string]string{
	"ua": "*Дур-попередження!*\n\n%s, Ви порушили правила або норми поведінки чату та отримали %d попередження.\nЯкщо зібрати всі 5, то є можливість вирушити у Дурпекло:)",
	"en": "*Dur-warning!*\n\n%s, you have broken the rules or norms of the chat and already got %d warning.\nIf you collect all 5, you will possibly go to the Durhell:)",
}

//ChatWarnWithReasonPlural ...
var ChatWarnWithReasonPlural map[string]string = map[string]string{
	"ua": "*Дур-попередження!*\n\n%s, Ви порушили правила або норми поведінки чату та отримали попередження.\nЯкщо зібрати всі 5, то є можливість вирушити у Дурпекло:)\n\n*Коментар:* %s",
	"en": "*Dur-warning!*\n\n%s, you have broken the rules or norms of the chat and already got warning.\nIf you collect all 5, you will possibly go to the Durhell:)\n\n*Comment:* %s",
}

//ChatWarnWithoutReasonPlural ...
var ChatWarnWithoutReasonPlural map[string]string = map[string]string{
	"ua": "*Дур-попередження!*\n\n%s, Ви порушили правила або норми поведінки чату та отримали попередження.\nЯкщо зібрати всі 5, то є можливість вирушити у Дурпекло:)",
	"en": "*Dur-warning!*\n\n%s, you have broken the rules or norms of the chat and already got warning.\nIf you collect all 5, you will possibly go to the Durhell:)",
}


//AdminChatWarnWithReasonSingular ...
var AdminChatWarnWithReasonSingular map[string]string = map[string]string{
	"ua": "*Дур-попередження!*\n\n[%s](tg://user?id=%d) видав(-ла) %d з 5 попереджень користувачу %s.\n\n*Коментар:* %s",
	"en": "*Dur-warning!*\n\n[%s](tg://user?id=%d) has given %d of 5 warnings for %s.\n\n*Comment:* %s",
}

//AdminChatWarnWithoutReasonSingular ...
var AdminChatWarnWithoutReasonSingular map[string]string = map[string]string{
	"ua": "*Дур-попередження!*\n\n[%s](tg://user?id=%d) видав(-ла) %d з 5 попереджень користувачу %s.",
	"en": "*Dur-warning!*\n\n[%s](tg://user?id=%d) has given %d of 5 warnings for %s.",
}
//AdminChatWarnWithReasonPlural ...
var AdminChatWarnWithReasonPlural map[string]string = map[string]string{
	"ua": "*Дур-попередження!*\n\n[%s](tg://user?id=%d) видав(-ла) по попередженню %s.\n\n*Коментар:* %s",
	"en": "*Dur-warning!*\n\n[%s](tg://user?id=%d) has given one warning to %s.\n\n*Comment:* %s",
}

//AdminChatWarnWithoutReasonPlural ...
var AdminChatWarnWithoutReasonPlural map[string]string = map[string]string{
	"ua": "*Дур-попередження!*\n\n[%s](tg://user?id=%d) видав(-ла) по попередженню %s.",
	"en": "*Dur-warning!*\n\n[%s](tg://user?id=%d) has given one warning to %s.",
}

//ChatBanWithReasonSingularWarns ...
var ChatBanWithReasonSingularWarns map[string]string = map[string]string{
	"ua": "*The end* ☠️ \n\n%s, ти отримав 5 попереджень і прямуєш у Дурпекло:)\n\n*Коментар:* %s",
	"en": "*The end* ☠️ \n\n%s, you have received 5 warnings and now Durhell is waiting for you:)\n\n*Comment:* %s",
}

//ChatBanWithoutReasonSingularWarns ...
var ChatBanWithoutReasonSingularWarns map[string]string = map[string]string{
	"ua": "*The end* ☠️ \n\n%s, ти отримав 5 попереджень і прямуєш у Дурпекло:)",
	"en": "*The end* ☠️ \n\n%s, you have received 5 warnings and now Durhell is waiting for you:)",
}


//ChatBanWithReasonPluralWarns ...
var ChatBanWithReasonPluralWarns map[string]string = map[string]string{
	"ua": "*The end* ☠️ \n\n%s, ви отримали 5 попереджень і прямуєте у Дурпекло:)\n\n*Коментар:* %s",
	"en": "*The end* ☠️ \n\n%s, you have received 5 warnings and now Durhell is waiting for you:)\n\n*Comment:* %s",
}

//ChatBanWithoutReasonPluralWarns ...
var ChatBanWithoutReasonPluralWarns map[string]string = map[string]string{
	"ua": "*The end* ☠️ \n\n%s, ви отримали 5 попереджень і прямуєте у Дурпекло:)",
	"en": "*The end* ☠️ \n\n%s, you have received 5 warnings and now Durhell is waiting for you:)",
}

//AdminBanWithReasonWarns ...
var AdminBanWithReasonWarns map[string]string = map[string]string{
	"ua": "*The end* ☠️ \n\n[%s](tg://user?id=%d) відправляє п'ятим варном %s прямісінько у Дурпекло:)\n\n*Коментар:* %s",
	"en": "*The end* ☠️ \n\n[%s](tg://user?id=%d) send %s to Durhell by giving fifth warn:)\n\n*Comment:* %s",
}

//AdminBanWithoutReasonWarns ...
var AdminBanWithoutReasonWarns map[string]string = map[string]string{
	"ua": "*The end* ☠️ \n\n[%s](tg://user?id=%d) відправляє п'ятим варном %s прямісінько у Дурпекло:)\n\n",
	"en": "*The end* ☠️ \n\n[%s](tg://user?id=%d) send %s to Durhell by giving fifth warn:)\n\n",
}

//MUTES 

//ChatMuteWithReasonSingularWarn ...
var ChatMuteWithReasonSingularWarn map[string]string = map[string]string{
	"ua": "%s запхали кляп у рота.\n\n*Коментар:* %s",
	"en": "%s has something big in the mouth.\n\n*Comment:* %s",
}

//ChatMuteWithoutReasonSingularWarn ...
var ChatMuteWithoutReasonSingularWarn map[string]string = map[string]string{
	"ua": "%s запхали кляп у рота.",
	"en": "%s has something big in the mouth.",
}


//ChatMuteWithReasonPluralWarn ...
var ChatMuteWithReasonPluralWarn map[string]string = map[string]string{
	"ua": "%s вирішили пограти у циганку і застряли одне в одного в роті.\n\n*Коментар:* %s",
	"en": "%s party got out of control and wrong things have appeared in your mouthes.\n\n*Comment:* %s",
}

//ChatMuteWithoutReasonPluralWarn ...
var ChatMuteWithoutReasonPluralWarn map[string]string = map[string]string{
	"ua": "%s вирішили пограти у циганку і застряли одне в одного в роті.",
	"en": "%s party got out of control and wrong things have appeared in your mouthes.",
}

//AdminMuteWithReasonWarn ...
var AdminMuteWithReasonWarn map[string]string = map[string]string{
	"ua": "[%s](tg://user?id=%d) запхав(-ла) кляп у рота %s.\n\n*Час:* %s\n*Коментар:* %s",
	"en": "[%s](tg://user?id=%d) put something in the mouth(es) of %s.\n\n*Time:* %s\n*Comment:* %s",
}

//AdminMuteWithoutReasonWarn ...
var AdminMuteWithoutReasonWarn map[string]string = map[string]string{
	"ua": "[%s](tg://user?id=%d) запхав(-ла) кляп у рота %s.\n*Час:* %s",
	"en": "[%s](tg://user?id=%d) put something in the mouth(es) of %s.\n*Час:* %s",
}

//Forever for mute with 0 time
var Forever map[string]string = map[string]string{
	"ua": "Назавжди",
	"en": "Forever",
}
//EXP DATE WILL BE UPDATES

//ExpDateUpdated ...
var ExpDateUpdated map[string]string = map[string]string{
	"ua": "Час зняття варнів було змінено",
	"en": "Time of warn-removing has been changed",
}

//UNWARN USER

//Unwarn ...
var Unwarn map[string]string = map[string]string{
	"ua": "Варн було знято з %s",
	"en": "Warn has been removed from %s",
}

//GET WARNS

//GetWarns ...
var GetWarns map[string] string = map[string] string {
	"ua": "\nВсього варнів: %d. З них активні: %d\n\n",
	"en": "\nAmount of warns: %d. Active: %d\n\n",
}

//DELETE WARN ERRORS

//NoWarnWithSuchID ...
var NoWarnWithSuchID map[string] string = map[string] string {
	"ua": "Порушення з ID %d не існує",
	"en": "Violation with ID %d doesn't exist",
}

//YouCantDeleteThisWarn ...
var YouCantDeleteThisWarn map[string] string = map[string] string {
	"ua": "Ви не можете видалити порушення з ID %d.",
	"en": "You can't delete violation with ID %d",
}

//WarnHasBeenDeleted ....
var WarnHasBeenDeleted map[string] string = map[string] string {
	"ua": "Порушення було видалено\n%s\n%s",
	"en": "Violation has been deleted\n%s\n%s",
}

//WarnHasBeenUpdated ....
var WarnHasBeenUpdated map[string] string = map[string] string {
	"ua": "Порушення було оновлено\n%s\n%s",
	"en": "Violation has been updated\n%s\n%s",
}

//AUTO WARN REMOVE

//WarnAutoRemove ...
var WarnAutoRemove map[string] string = map[string] string {
	"ua": "Чат: %s\nКоментар: %s\n---\n",
	"en": "Chat: %s\nReason: %s\n---\n",
}

//WarnHasBeenRemoved ...
var WarnHasBeenRemoved map[string] string = map[string] string {
	"ua": "Порушення було знято з [%s](tg://user?id=%d) *ID*:`%d`\n",
	"en": "Порушення було знято з [%s](tg://user?id=%d) *ID*:`%d`\n",
}

//AFK

//AddAFKSingular when give afk
var AddAFKSingular map[string] string = map[string] string {
	"ua": "%s твоє афк під час гри було записане! Надалі будь активнішим(-ою)😉",
	"en": "%s your afk during the game was written! We expect more activity from you, darling😉",
}

//AddAFKPlural when give afk
var AddAFKPlural map[string] string = map[string] string {
	"ua": "%s ваше афк під час гри було записане! Надалі будьте активніші😉",
	"en": "%s your afk during the game was written! We expect more activity from you, cuties😉",
}

//AddAFKSingularWithReasonAdmin when give afk
var AddAFKSingularWithReasonAdmin map[string] string = map[string] string {
	"ua": "%s був(-ла) пасивним(-ою) у грі.\n*Коментар:* %s",
	"en": "%s was afk in the game.\n*Comment:* %s",
}

//AddAFKSingularWithoutReasonAdmin when give afk
var AddAFKSingularWithoutReasonAdmin map[string] string = map[string] string {
	"ua": "%s був(-ла) пасивним(-ою) у грі.",
	"en": "%s was afk in the game.",
}

//AddAFKSingularWithReasonPatrul when give afk
var AddAFKSingularWithReasonPatrul map[string] string = map[string] string {
	"ua": "%s\n\n%s був(-ла) пасивним(-ою) у грі.\n*Коментар:* %s",
	"en": "%s\n\n%s was afk in the game.\n*Comment:* %s",
}

//AddAFKSingularWithoutReasonPatrul when give afk
var AddAFKSingularWithoutReasonPatrul map[string] string = map[string] string {
	"ua": "%s\n\n%s був(-ла) пасивним(-ою) у грі.",
	"en": "%s\n\n%s was afk in the game.",
}

//to admin and patruls plural

//AddAFKPluralWithReasonAdmin when give afk
var AddAFKPluralWithReasonAdmin map[string] string = map[string] string {
	"ua": "%s були пасивні у грі.\n*Коментар:* %s",
	"en": "%s were afk in the game.\n*Comment:* %s",
}

//AddAFKPluralWithoutReasonAdmin when give afk
var AddAFKPluralWithoutReasonAdmin map[string] string = map[string] string {
	"ua": "%s були пасивні у грі.",
	"en": "%s were afk in the game.",
}

//AddAFKPluralWithReasonPatrul when give afk
var AddAFKPluralWithReasonPatrul map[string] string = map[string] string {
	"ua": "%s\n\n%s були пасивні у грі.\n*Коментар:* %s",
	"en": "%s\n\n%s were afk in the game.\n*Comment:* %s",
}

//AddAFKPluralWithoutReasonPatrul when give afk
var AddAFKPluralWithoutReasonPatrul map[string] string = map[string] string {
	"ua": "%s\n\n%s був(-ла) пасивним(-ою) у грі.",
	"en": "%s\n\n%s was afk in the game.",
}

//UnAFK ...
var UnAFK map[string]string = map[string]string{
	"ua": "Афк було знято з %s",
	"en": "AFK has been removed from %s",
}

//GetAFK ...
var GetAFK map[string] string = map[string] string {
	"ua": "\nВсього афк: %d. З них активні: %d\n\n",
	"en": "\nAmount of afk: %d. Active: %d\n\n",
}

//BANS 

//ChatBanWithReasonSingular ...
var ChatBanWithReasonSingular map[string]string = map[string]string{
	"ua": "*The end* ☠️ \n\n%s, на жаль, твій час у цьому чаті підійшов до кінця і ми змушені відправити тебе у Дурпекло:)\n\n*Коментар:* %s",
	"en": "*The end* ☠️ \n\n%s, we are sorry to inform that your time in this chat has come to the end and we have to send you to the Durhell:)\n\n*Comment:* %s",
}

//ChatBanWithoutReasonSingular ...
var ChatBanWithoutReasonSingular map[string]string = map[string]string{
	"ua": "*The end* ☠️ \n\n%s, на жаль, твій час у цьому чаті підійшов до кінця і ми змушені відправити тебе у Дурпекло:)",
	"en": "*The end* ☠️ \n\n%s, we are sorry to inform that your time in this chat has come to the end and we have to send you to the Durhell:)",
}


//ChatBanWithReasonPlural ...
var ChatBanWithReasonPlural map[string]string = map[string]string{
	"ua": "*The end* ☠️ \n\n%s, на жаль, ваш час у цьому чаті підійшов до кінця і ми змушені відправити вас у Дурпекло:)\n\n*Коментар:* %s",
	"en": "*The end* ☠️ \n\n%s, we are sorry to inform that your time in this chat has come to the end and we have to send you to the Durhell:)\n\n*Comment:* %s",
}

//ChatBanWithoutReasonPlural ...
var ChatBanWithoutReasonPlural map[string]string = map[string]string{
	"ua": "*The end* ☠️ \n\n%s, на жаль, ваш час у цьому чаті підійшов до кінця і ми змушені відправити вас у Дурпекло:)",
	"en": "*The end* ☠️ \n\n%s, we are sorry to inform that your time in this chat has come to the end and we have to send you to the Durhell:)",
}

//AdminBanWithReasonForeverSingular ...
var AdminBanWithReasonForeverSingular map[string]string = map[string]string{
	"ua": "[%s](tg://user?id=%d) відправляє %s (*ID*: `%d`) в Дурпекло назавжди\n\n*Коментар:* %s",
	"en": "[%s](tg://user?id=%d) sends %s (*ID*: `%d`) to the Durhell til the end of time\n\n*Comment:* %s",
}
//AdminBanWithReasonSingular ...
var AdminBanWithReasonSingular map[string]string = map[string]string{
	"ua": "[%s](tg://user?id=%d) відправляє %s (*ID*: `%d`) в Дурпекло назавжди\n\n*Коментар:* %s\n*Час*: %d %s",
	"en": "[%s](tg://user?id=%d) sends %s (*ID*: `%d`) to the Durhell til the end of time\n\n*Comment:* %s\n*Time*: %d %s",
}

//AdminBanWithReasonForeverPlural ...
var AdminBanWithReasonForeverPlural map[string]string = map[string]string{
	"ua": "[%s](tg://user?id=%d) відправляє %s в Дурпекло назавжди\n\n*Коментар:* %s",
	"en": "[%s](tg://user?id=%d) sends %s to the Durhell til the end of time\n\n*Comment:* %s",
}
//AdminBanWithReasonPlural ...
var AdminBanWithReasonPlural map[string]string = map[string]string{
	"ua": "[%s](tg://user?id=%d) відправляє %s в Дурпекло\n\n*Коментар:* %s\n*Час*: %d %s",
	"en": "[%s](tg://user?id=%d) sends %s to the Durhell\n\n*Comment:* %s\n*Time*: %d %s",
}

//AdminBanWithoutReasonForeverPlural ...
var AdminBanWithoutReasonForeverPlural map[string]string = map[string]string{
	"ua": "[%s](tg://user?id=%d) відправляє %s в Дурпекло назавжди",
	"en": "[%s](tg://user?id=%d) sends %s to the Durhell til the end of time",
}
//AdminBanWithoutReasonPlural ...
var AdminBanWithoutReasonPlural map[string]string = map[string]string{
	"ua": "[%s](tg://user?id=%d) відправляє %s в Дурпекло\n\n*Час*: %d %s",
	"en": "[%s](tg://user?id=%d) sends %s to the Durhell\n\n*Time*: %d %s",
}
//AdminBanWithoutReasonForeverSingular ...
var AdminBanWithoutReasonForeverSingular map[string]string = map[string]string{
	"ua": "[%s](tg://user?id=%d) відправляє %s (*ID*: `%d`) в Дурпекло назавжди",
	"en": "[%s](tg://user?id=%d) sends %s (*ID*: `%d`) to the Durhell til the end of time",
}
//AdminBanWithoutReasonSingular ...
var AdminBanWithoutReasonSingular map[string]string = map[string]string{
	"ua": "[%s](tg://user?id=%d) відправляє %s (*ID*: `%d`) в Дурпекло\n\n*Час*: %d %s",
	"en": "[%s](tg://user?id=%d) sends %s (*ID*: `%d`) to the Durhell\n\n*Time*: %d %s",
}

//UNBAN

//AdminBanWithoutReasonSingular ...
var Unban map[string]string = map[string]string{
	"ua": "%s більше не в бані у чаті %s",
	"en": "%s are unbanned in chat %s",
}

//UNMUTE

//AdminBanWithoutReasonSingular ...
var UnmuteAdmin map[string]string = map[string]string{
	"ua": "%s дістав(-ла) кляп з рота у %s",
	"en": "%s got something out from %s mouth(es)",
}

//UnmuteSingular ...
var UnmuteSingular map[string]string = map[string]string{
	"ua": "%s знову може говорити",
	"en": "%s can speak again",
}

//UnmutePlural ...
var UnmutePlural map[string]string = map[string]string{
	"ua": "%s знову можуть говорити",
	"en": "%s can speak again",
}

//INFO

//Participant ...
var Participant map[string]string = map[string]string{
	"ua": "*Учасник:* [%s](tg://user?id=%d)",
	"en": "*Participant:* [%s](tg://user?id=%d)",
}

//FirstMessage ...
var FirstMessage map[string]string = map[string]string{
	"ua": "*Перше повідомлення:* %s",
	"en": "*First message:* %s",
}

//MessageAmount ...
var MessageAmount map[string]string = map[string]string{
	"ua": "*Кількість повідомлень:* %d",
	"en": "*Amount of messages:* %d",
}

//TimeInGroup ...
var TimeInGroup map[string]string = map[string]string{
	"ua": "*Час в групі:* %s",
	"en": "*Time in group:* %s",
}

//WarnsSingular ...
var WarnsSingular map[string]string = map[string]string{
	"ua": "*Попереджень:* %d, з них активних 1\n",
	"en": "*Warns:* %d, 1 of them is active\n",
}

//WarnsPlural ...
var WarnsPlural map[string]string = map[string]string{
	"ua": "*Попереджень:* %d, з них активних %d\n",
	"en": "*Warns:* %d, %d of them are active\n",
}


//WarnsSingular ...
var AFKSingular map[string]string = map[string]string{
	"ua": "*AFK:* %d, з них активних 1",
	"en": "*AFK:* %d, 1 of them is active",
}

//WarnsPlural ...
var AFKPlural map[string]string = map[string]string{
	"ua": "*AFK:* %d, з них активних %d",
	"en": "*AFK:* %d, %d of them are active",
}

//CommentsToWarns ...
var CommentsToWarns map[string]string = map[string]string{
	"ua": "*Коментарі до попереджень:*\n",
	"en": "*Reasons for warns:*\n",
}

//Hiding ...
var Hiding map[string]string = map[string]string{
	"ua": "Приховуємо...",
	"en": "Hiding...",
}

//ThatButtonIsNotForYou ...
var ThatButtonIsNotForYou map[string]string = map[string]string{
	"ua": "Ця кнопка не для тебе",
	"en": "Than button is not for you",
}
