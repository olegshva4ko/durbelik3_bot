package sqldatabase

import (
	"Durbelik3/internal/text"
	"Durbelik3/pkg/models"
	"Durbelik3/pkg/mongodatabase"
	"Durbelik3/pkg/tools"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf16"

	_ "github.com/go-sql-driver/mysql" //mysql driver

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

//Sqldatabase thing that allows you to communicate via bot, with database, with buffer
type Sqldatabase struct {
	Bot      *tgbotapi.BotAPI
	SQLtoken string
	Buff     *models.BufferedChat
	MB       *mongodatabase.MongoDatabase
}

//CheckDB checks if connection can be established
func (mydb *Sqldatabase) CheckDB() error {
	db, err := sql.Open("mysql", mydb.SQLtoken)
	if err != nil {
		fmt.Println("line 33", err.Error())
		return err
	}
	if err := db.Ping(); err != nil {
		return err
	}
	db.Close()
	return nil
}

//MessageProcess adds user to db and returns if it is a first message
func (mydb *Sqldatabase) MessageProcess(message *tgbotapi.Message) bool {
	var firstUserMessage bool
	//Check for new users, add to ChatDB if not exists
	if message.NewChatMembers != nil {
		for _, v := range *message.NewChatMembers {
			go mydb.AddUserToMainDB(&v)
			mydb.AddUserToChatDB(&v, message.From.ID, message.Chat.ID)
		}
	}
	//Check for user, add to UserDB if not exists
	go mydb.AddUserToMainDB(message.From)
	mydb.AddUserToChatDB(message.From, message.From.ID, message.Chat.ID)
	//Check for updates in UserDB
	firstUserMessage = mydb.CheckForUserUpdates(message.From, message)
	return firstUserMessage
}

//AddUserToMainDB adds user to Users db
func (mydb *Sqldatabase) AddUserToMainDB(user *tgbotapi.User) {
	db, err := sql.Open("mysql", mydb.SQLtoken)
	if err != nil {
		fmt.Println("line 65", err.Error())
		return
	}
	defer db.Close()
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS Users(userid INT NOT NULL UNIQUE, firstname text NOT NULL, lastname text NOT NULL, username text NOT NULL, PRIMARY KEY(userid))")
	if err != nil {
		fmt.Println("line 71", err.Error())
		return
	}

	_, err = db.Exec("INSERT INTO Users(userid, firstname, lastname, username) VALUES (?, ?, ?, ?)", //Userid is primary key, so user wont be added 2 or more times
		user.ID, user.FirstName, user.LastName, user.UserName)
	if err != nil {
		fmt.Println("line 78", err.Error())
		return
	}
	return
}

//AddUserToChatDB adds user to Chat DB
func (mydb *Sqldatabase) AddUserToChatDB(user *tgbotapi.User, addedByInt int, chatid int64) {
	db, err := sql.Open("mysql", mydb.SQLtoken)
	if err != nil {
		fmt.Println("line 88", err.Error())
		return
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS UsersInChat(id BIGINT NOT NULL AUTO_INCREMENT, userid INT NOT NULL, chatid BIGINT NOT NULL, data INT NOT NULL, addedby INT NOT NULL, messageamount BIGINT NOT NULL, coupleoftheday BOOLEAN NOT NULL, calllist BOOLEAN NOT NULL,PRIMARY KEY(id))")
	if err != nil {
		fmt.Println("line 95", err.Error())
		return
	}
	//check if user is not in database
	err = db.QueryRow("SELECT userid FROM UsersInChat where userid = ? and chatid = ?", user.ID, chatid).Scan(&user.ID)
	if err != sql.ErrNoRows {
		return
	}
	_, err = db.Exec("INSERT INTO UsersInChat(userid, chatid, data, addedby, messageamount, coupleoftheday, calllist) VALUES (?, ?, ?, ?, ?, ?, ?)", //Userid is primary key, so user wont be added 2 or more times
		user.ID, chatid, time.Now().Unix(), addedByInt, 0, false, false)
	if err != nil {
		fmt.Println("line 106", err.Error())
		return
	}
	return
}

//CheckForUserUpdates adds user to Users db
func (mydb *Sqldatabase) CheckForUserUpdates(user *tgbotapi.User, message *tgbotapi.Message) bool {
	db, err := sql.Open("mysql", mydb.SQLtoken)
	if err != nil {
		fmt.Println("line 116", err.Error())
		return false
	}
	defer db.Close()

	//u := new(models.UserInChat)
	messageAmount := 0
	//check if user is not in database
	err = db.QueryRow("SELECT messageamount FROM UsersInChat where userid = ? and chatid = ?", user.ID, message.Chat.ID).Scan(
		&messageAmount,
	)
	if err != nil {
		fmt.Println("line 134", err.Error())
		return false
	}
	//u.MessageAmount++

	_, err = db.Exec("UPDATE UsersInChat SET messageamount = messageamount + 1 where userid = ? AND chatid = ?", //Userid isnt primary key, so user can be added 2 or more times
		user.ID, message.Chat.ID)
	if err != nil {
		fmt.Println("line 142", err.Error())
		return false
	}

	go mydb.CheckForUserUpdatesMain(user, message.Chat.ID)

	if messageAmount <= 4 {
		return true
	}
	return false

}

//CheckForUserUpdatesMain checks for updates in main database and send updates to chat if needed
func (mydb *Sqldatabase) CheckForUserUpdatesMain(user *tgbotapi.User, chatID int64) error {
	db, err := sql.Open("mysql", mydb.SQLtoken)
	if err != nil {
		fmt.Println("line 159", err.Error())
		return err
	}
	defer db.Close()

	u := new(models.User)
	err = db.QueryRow("SELECT * FROM Users where userid = ?", user.ID).Scan(
		&u.UserID,
		&u.FirstName,
		&u.LastName,
		&u.UserName,
	)
	if err != nil {
		fmt.Println("line 172", err.Error())
		return err
	}

	var msgText []string = append([]string{}, fmt.Sprintf("Updated: %d", user.ID))
	if u.FirstName != user.FirstName {
		msgText = append(msgText, fmt.Sprintf("FirstName: %s", user.FirstName))
	}
	if u.LastName != user.LastName {
		msgText = append(msgText, fmt.Sprintf("LastName: %s", user.LastName))
	}
	if u.UserName != user.UserName {
		msgText = append(msgText, fmt.Sprintf("UserName: %s", user.UserName))
	}
	if len(msgText) > 1 {
		_, err := db.Exec("UPDATE Users SET firstname = ?, lastname = ?, username = ? where userid = ?",
			user.FirstName, user.LastName, user.UserName, user.ID)
		if err != nil {
			fmt.Println("line 190", err.Error())
			return err
		}
		//send message to admin chats
		if mydb.Buff.Chat[chatID] != nil {
			if mydb.Buff.Chat[chatID].IsMain {
				for _, v := range mydb.Buff.Chat[chatID].AdminChats {
					mydb.Bot.Send(tgbotapi.NewMessage(v, strings.Join(msgText, "\n")))
				}
			} else {
				for _, v := range mydb.Buff.Chat[mydb.Buff.Chat[chatID].FatherChat].AdminChats {
					mydb.Bot.Send(tgbotapi.NewMessage(v, strings.Join(msgText, "\n")))
				}
			}
		}
		mydb.MB.MongoDbAddUpdate(user)
	}

	return nil
}

//GetUserFromUsersInChatByID ...
func (mydb *Sqldatabase) GetUserFromUsersInChatByID(userID int, chatID int64) (*models.UserInChat, error) {
	db, err := sql.Open("mysql", mydb.SQLtoken)
	if err != nil {
		fmt.Println("line 211", err.Error())
		return nil, err
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS UsersInChat(id BIGINT NOT NULL AUTO_INCREMENT, userid INT NOT NULL, chatid BIGINT NOT NULL, data INT NOT NULL, addedby INT NOT NULL, messageamount BIGINT NOT NULL, coupleoftheday BOOLEAN NOT NULL, calllist BOOLEAN NOT NULL,PRIMARY KEY(id))")
	if err != nil {
		fmt.Println("line 223", err.Error())
		return nil, err
	}

	u := &models.UserInChat{}
	err = db.QueryRow("SELECT * FROM UsersInChat WHERE userid = ? AND chatid = ?", userID, chatID).Scan(
		&u.ID,
		&u.UserID,
		&u.ChatID,
		&u.Data,
		&u.AddedBy,
		&u.MessageAmount,
		&u.CoupleOfTheDay,
		&u.CallList,
	)
	if err != nil {
		fmt.Println("line 229", err.Error())
		return nil, err
	}

	return u, nil
}

//GetAllUsersFromUsersInChat ...
func (mydb *Sqldatabase) GetAllUsersFromUsersInChat() ([]*models.UserInChat, error) {
	var users []*models.UserInChat
	db, err := sql.Open("mysql", mydb.SQLtoken)
	if err != nil {
		fmt.Println("line 216", err.Error())
		return nil, err
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS UsersInChat(id BIGINT NOT NULL AUTO_INCREMENT, userid INT NOT NULL, chatid BIGINT NOT NULL, data INT NOT NULL, addedby INT NOT NULL, messageamount BIGINT NOT NULL, coupleoftheday BOOLEAN NOT NULL, calllist BOOLEAN NOT NULL,PRIMARY KEY(id))")
	if err != nil {
		fmt.Println("line 223", err.Error())
		return nil, err
	}

	rows, err := db.Query("SELECT * FROM UsersInChat")
	if err != nil {
		fmt.Println("line 229", err.Error())
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		u := &models.UserInChat{}
		err := rows.Scan(&u.ID,
			&u.UserID,
			&u.ChatID,
			&u.Data,
			&u.AddedBy,
			&u.MessageAmount,
			&u.CoupleOfTheDay,
			&u.CallList)
		if err != nil {
			fmt.Println("line 245", err.Error())
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}
//GetBannedStickers returns list of names of banned stickers
func (mydb *Sqldatabase) GetBannedStickers(chatid int64) ([]*models.BannedSticker, error) {
	db, err := sql.Open("mysql", mydb.SQLtoken)
	if err != nil {
		fmt.Println("line 257", err.Error())
		return nil, err
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS BannedStickers(id INT NOT NULL AUTO_INCREMENT, chatid BIGINT NOT NULL, name char(100),PRIMARY KEY(id))")
	if err != nil {
		fmt.Println("line 264", err.Error())
		return nil, err
	}

	rows, err := db.Query("SELECT * FROM BannedStickers where chatid = ?", chatid)
	if err != nil {
		fmt.Println("line 270", err.Error())
		return nil, err
	}
	defer rows.Close()

	var StickersList []*models.BannedSticker
	for rows.Next() {
		bannedSticker := new(models.BannedSticker)
		if err := rows.Scan(&bannedSticker.ID, &bannedSticker.ChatID, &bannedSticker.Name); err != nil {
			return nil, err
		}
		StickersList = append(StickersList, bannedSticker)
	}
	return StickersList, nil

}

//GetUserByUserID gets from main DB user and bool if not found
func (mydb *Sqldatabase) GetUserByUserID(id int) (*models.User, bool) {
	db, err := sql.Open("mysql", mydb.SQLtoken)
	if err != nil {
		fmt.Println("line 291", err.Error())
		return nil, false
	}
	defer db.Close()

	user := new(models.User)

	err = db.QueryRow("SELECT * FROM Users where userid = ?", id).Scan(
		&user.UserID,
		&user.FirstName,
		&user.LastName,
		&user.UserName,
	)
	if err != nil {
		fmt.Println("line 305", err.Error())
		return nil, false
	}

	return user, true
}

//GetBannedUsers ....
func (mydb *Sqldatabase) GetBannedUsers(chatID int64) ([]*models.User, error) {
	db, err := sql.Open("mysql", mydb.SQLtoken)
	if err != nil {
		fmt.Println("line 376", err.Error())
		return nil, err
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS BannedUsers(id BIGINT NOT NULL AUTO_INCREMENT, userid INT NOT NULL, chatid BIGINT NOT NULL,PRIMARY KEY(id))")
	if err != nil {
		fmt.Println("line 264", err.Error())
		return nil, err
	}

	rows, err := db.Query("SELECT userid FROM BannedUsers WHERE chatid = ?", chatID)
	if err != nil {
		fmt.Println("line 329", err.Error())
		return nil, err
	}

	var userIDs []int
	for rows.Next() {
		userID := 0
		if err := rows.Scan(
			&userID,
		); err != nil {
			return nil, err
		}
		userIDs = append(userIDs, userID)
	}
	rows.Close()

	users, err := mydb.getUsersByIDs(userIDs)
	if err != nil {
		return nil, err
	}

	return users, nil
}

//getUsersByIDs returns list of users by chatIDs
func (mydb *Sqldatabase) getUsersByIDs(userIDs []int) ([]*models.User, error) {
	db, err := sql.Open("mysql", mydb.SQLtoken)
	if err != nil {
		fmt.Println("line 376", err.Error())
		return nil, err
	}
	defer db.Close()

	var users []*models.User
	for _, userID := range userIDs {
		u := new(models.User)
		err = db.QueryRow("SELECT * FROM Users where userid = ?", userID).Scan(
			&u.UserID,
			&u.FirstName,
			&u.LastName,
			&u.UserName,
		)
		if err != nil {
			fmt.Println("line 373", err.Error())
			return nil, err
		}
		users = append(users, u)
	}

	return users, nil
}

//GetUserByUserName gets from main DB user and bool if not found
func (mydb *Sqldatabase) GetUserByUserName(username string) (*models.User, bool) {
	db, err := sql.Open("mysql", mydb.SQLtoken)
	if err != nil {
		fmt.Println("line 316", err.Error())
		return nil, false
	}
	defer db.Close()

	user := new(models.User)

	err = db.QueryRow("SELECT * FROM Users where username = ?", username).Scan(
		&user.UserID,
		&user.FirstName,
		&user.LastName,
		&user.UserName,
	)
	if err != nil {
		fmt.Println("line 330", err.Error())
		return nil, false
	}

	return user, true
}

/*WarnUser works with the next structure
Violations: id
Violations: userid
Violations: chatid
Violations: givenbyid
Violations: reason
Violations: date
Violations: expdate
Violations: status bool (active / passive)
Violations: iswarn bool if it is warn, not afk(isWarn? afk:warn)
sets
*/
func (mydb *Sqldatabase) WarnUsers(message *tgbotapi.Message, users []*models.User, reason string) (map[*models.User][]*models.Violation, error) {
	db, err := sql.Open("mysql", mydb.SQLtoken)
	if err != nil {
		fmt.Println("352 mysql:", err.Error())
		return nil, err
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS Violations(id BIGINT NOT NULL AUTO_INCREMENT, userid INT NOT NULL, chatid BIGINT NOT NULL, givenbyid INT NOT NULL, reason TEXT NOT NULL, date BIGINT NOT NULL, expdate BIGINT NOT NULL, status BOOLEAN NOT NULL, iswarn BOOLEAN NOT NULL, PRIMARY KEY(id))")
	if err != nil {
		fmt.Println("line 359", err.Error())
		return nil, err
	}

	var (
		chatID        int64
		expdate       int64                                = time.Unix(int64(message.Date), 0).AddDate(0, 3, 0).Unix()
		userWithWarns map[*models.User][]*models.Violation = make(map[*models.User][]*models.Violation)
	)
	if mydb.Buff.Chat[message.Chat.ID].IsMain { //if you want to give warn from admin chat
		chatID = message.Chat.ID
	} else {
		chatID = mydb.Buff.Chat[message.Chat.ID].FatherChat
	}
	for _, user := range users {
		_, err = db.Exec("INSERT INTO Violations(userid, chatid, givenbyid, reason, date, expdate, status, iswarn) VALUES(?, ?, ?, ?, ?, ?, true, true)",
			user.UserID, chatID, message.From.ID, reason, int64(message.Date), expdate)
		if err != nil {
			fmt.Println("line 337", err.Error())
			return nil, err
		}

		warns, err := mydb.GetViolationsByUserIDinChat(user.UserID, chatID, true)
		if err != nil {
			fmt.Println("line 383", err.Error())
			return nil, err
		}
		var activeWarns []*models.Violation
		for _, v := range warns { //remove passive warns
			if !v.Status {
				continue
			}
			activeWarns = append(activeWarns, v)
		}

		userWithWarns[user] = activeWarns
	}
	//change expdate for users
	//go mydb.ChangeExpdate(users, chatID, true)

	return userWithWarns, nil
}

//RemoveFromLists ...
func (mydb *Sqldatabase) RemoveFromLists(chatID int64, users []*models.User) {
	db, err := sql.Open("mysql", mydb.SQLtoken)
	if err != nil {
		fmt.Println("line 404", err.Error())
		return
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS UsersInChat(id BIGINT NOT NULL AUTO_INCREMENT, userid INT NOT NULL, chatid BIGINT NOT NULL, data INT NOT NULL, addedby INT NOT NULL, messageamount BIGINT NOT NULL, coupleoftheday BOOLEAN NOT NULL, calllist BOOLEAN NOT NULL,PRIMARY KEY(id))")
	if err != nil {
		fmt.Println("line 411", err.Error())
		return
	}

	for _, user := range users {
		_, err := db.Exec("UPDATE UsersInChat SET calllist = false, coupleoftheday = false WHERE userid = ? AND chatid = ?", user.UserID, chatID)
		if err != nil {
			fmt.Println("line 418", err.Error())
			continue
		}
	}
}

//ChangeExpdate ...
func (mydb *Sqldatabase) ChangeExpdate(users []*models.User, chatid int64, isWarn bool) {
	db, err := sql.Open("mysql", mydb.SQLtoken)
	if err != nil {
		fmt.Println("line 428", err.Error())
		return
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS Violations(id BIGINT NOT NULL AUTO_INCREMENT, userid INT NOT NULL, chatid BIGINT NOT NULL, givenbyid INT NOT NULL, reason TEXT NOT NULL, date BIGINT NOT NULL, expdate BIGINT NOT NULL, status BOOLEAN NOT NULL, iswarn BOOLEAN NOT NULL, PRIMARY KEY(id))")
	if err != nil {
		fmt.Println("line 435", err.Error())
		return
	}
	changesFlag := false
	for _, user := range users {
		var lastExpdate int64
		//find lastrow
		err := db.QueryRow("SELECT expdate FROM Violations WHERE id = (SELECT MAX(id) FROM Violations WHERE userid = ? AND chatid = ? AND iswarn = ? AND status = true)", user.UserID, chatid, isWarn).Scan(
			&lastExpdate,
		)
		if err != nil {
			fmt.Println("line 446", err.Error())
			continue
		}
		result, err := db.Exec("UPDATE Violations SET expdate = ? WHERE userid = ? AND chatid = ? AND iswarn = ? AND status = true", lastExpdate, user.UserID, chatid, isWarn)
		if err != nil {
			fmt.Println("line 451", err.Error())
			continue
		}
		affectedRows, err := result.RowsAffected()
		if err != nil {
			fmt.Println("line 456", err.Error())
			continue
		}
		if affectedRows > 0 {
			changesFlag = true
		}
	}
	if changesFlag {
		mydb.Bot.Send(tgbotapi.NewMessage(chatid, text.ExpDateUpdated[mydb.Buff.Chat[chatid].Language]))
	}
	return
}

//GetViolationsByUserIDinChat ...
func (mydb *Sqldatabase) GetViolationsByUserIDinChat(userid int, chatid int64, isWarn bool) ([]*models.Violation, error) {
	db, err := sql.Open("mysql", mydb.SQLtoken)
	if err != nil {
		fmt.Println("line 477", err.Error())
		return nil, err
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS Violations(id BIGINT NOT NULL AUTO_INCREMENT, userid INT NOT NULL, chatid BIGINT NOT NULL, givenbyid INT NOT NULL, reason TEXT NOT NULL, date BIGINT NOT NULL, expdate BIGINT NOT NULL, status BOOLEAN NOT NULL, iswarn BOOLEAN NOT NULL, PRIMARY KEY(id))")
	if err != nil {
		fmt.Println("line 484", err.Error())
		return nil, err
	}

	rows, err := db.Query("SELECT * FROM Violations WHERE userid = ? AND chatid = ? AND iswarn = ?", userid, chatid, isWarn)
	if err != nil {
		fmt.Println("line 490", err.Error())
		return nil, err
	}

	var violations []*models.Violation
	for rows.Next() {
		violation := new(models.Violation)
		err = rows.Scan(
			&violation.ID,
			&violation.UserID,
			&violation.ChatID,
			&violation.GivenByID,
			&violation.Reason,
			&violation.Date,
			&violation.Expdate,
			&violation.Status,
			&violation.IsWarn,
		)
		if err != nil {
			fmt.Println("line 509", err.Error())
			continue
		}
		violations = append(violations, violation)
	}
	rows.Close()
	return violations, nil
}

/*
//UnwarnUsers changes warn status from active to passive
func (mydb *Sqldatabase) UnwarnUsers(chatID int64, users []*models.User, warnNumber int64) {
	db, err := sql.Open("mysql", mydb.SQLtoken)
	if err != nil {
		fmt.Println("line 522", err.Error())
		return
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS Violations(id BIGINT NOT NULL AUTO_INCREMENT, userid INT NOT NULL, chatid BIGINT NOT NULL, givenbyid INT NOT NULL, reason TEXT NOT NULL, date BIGINT NOT NULL, expdate BIGINT NOT NULL, status BOOLEAN NOT NULL, iswarn BOOLEAN NOT NULL, PRIMARY KEY(id))")
	if err != nil {
		fmt.Println("line 529", err.Error())
		return
	}

	var (
		amountOfWarns int
	)

U:
	for _, user := range users {
		err := db.QueryRow("SELECT COUNT(*) FROM Violations WHERE userid = ? AND chatid = ? AND iswarn = true AND status = true", user.UserID, chatID).Scan(
			&amountOfWarns,
		)
		if err != nil || amountOfWarns == 0 {
			continue U
		}

		if amountOfWarns < int(warnNumber) || int(warnNumber) == 0 {
			//change status of last warn
			_, err = db.Exec("UPDATE Violations SET status=false WHERE id = (SELECT MAX(id) FROM(SELECT * FROM Violations) AS A WHERE A.userid = ? AND A.chatid = ? AND A.iswarn = true AND A.status = true)", user.UserID, chatID)
			if err != nil {
				fmt.Println("line 550", err.Error())
				continue U
			}
		} else {
			//remove status of concrete warn

			rows, err := db.Query("SELECT id FROM Violations WHERE userid = ? AND chatid = ? AND iswarn = true AND status = true", user.UserID, chatID)
			if err != nil {
				fmt.Println("line 558", err.Error())
				continue U
			}

			var (
				counter int
				ID      int64
			)
			for rows.Next() {
				counter++
				if int(warnNumber) == counter {
					err = rows.Scan(&ID)
					if err != nil {
						fmt.Println("line 571", err.Error())
						continue
					}
					_, err = db.Exec("UPDATE Violations SET status = false WHERE id = ?", ID)
					if err != nil {
						fmt.Println("line 576", err.Error())
						continue U
					}
					rows.Close()
					continue U
				}
			}
			rows.Close()

		}

	}
	//go mydb.ChangeExpdate(users, chatID, true)

}
*/
//UnwarnUsers deletes warn and if user doesnt have rights for it it will return error
func (mydb *Sqldatabase) UnwarnUsers(chatID int64, warnID int64, isWarn bool) (*models.Violation, error) {
	db, err := sql.Open("mysql", mydb.SQLtoken)
	if err != nil {
		fmt.Println("line 649", err.Error())
		return nil, err
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS Violations(id BIGINT NOT NULL AUTO_INCREMENT, userid INT NOT NULL, chatid BIGINT NOT NULL, givenbyid INT NOT NULL, reason TEXT NOT NULL, date BIGINT NOT NULL, expdate BIGINT NOT NULL, status BOOLEAN NOT NULL, iswarn BOOLEAN NOT NULL, PRIMARY KEY(id))")
	if err != nil {
		fmt.Println("line 656", err.Error())
		return nil, err
	}

	var (
		warn *models.Violation = &models.Violation{}
	)

	err = db.QueryRow("SELECT * FROM Violations WHERE id = ? AND iswarn = ?", warnID, isWarn).Scan(
		&warn.ID,
		&warn.UserID,
		&warn.ChatID,
		&warn.GivenByID,
		&warn.Reason,
		&warn.Date,
		&warn.Expdate,
		&warn.Status,
		&warn.IsWarn,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf(text.NoWarnWithSuchID[mydb.Buff.Chat[chatID].Language], warnID)
		}
	}
	if warn.ChatID != chatID {
		return nil, fmt.Errorf(text.YouCantDeleteThisWarn[mydb.Buff.Chat[chatID].Language], warnID)
	}

	_, err = db.Exec("UPDATE Violations SET status = false WHERE id = ?", warnID)
	if err != nil {
		fmt.Println("Line 701", err.Error())
	}

	return warn, nil
}

//GetUsersWithViolations returns map with User:{Violation1, Violation2, Violation3{}}
func (mydb *Sqldatabase) GetUsersWithViolations(chatID int64, users []*models.User, isWarn bool) (map[*models.User][]*models.Violation, error) {
	db, err := sql.Open("mysql", mydb.SQLtoken)
	if err != nil {
		fmt.Println("line 596", err.Error())
		return nil, err
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS Violations(id BIGINT NOT NULL AUTO_INCREMENT, userid INT NOT NULL, chatid BIGINT NOT NULL, givenbyid INT NOT NULL, reason TEXT NOT NULL, date BIGINT NOT NULL, expdate BIGINT NOT NULL, status BOOLEAN NOT NULL, iswarn BOOLEAN NOT NULL, PRIMARY KEY(id))")
	if err != nil {
		fmt.Println("line 603", err.Error())
		return nil, err
	}

	var (
		userWithViolations map[*models.User][]*models.Violation = make(map[*models.User][]*models.Violation)
	)

	for _, user := range users {
		rows, err := db.Query("SELECT * FROM Violations WHERE userid = ? AND chatid = ? AND iswarn = ?",
			user.UserID, chatID, isWarn)
		if err != nil {
			fmt.Println("line 617", err.Error())
			return nil, err
		}
		var violations []*models.Violation 
		for rows.Next() {
			violation := new(models.Violation)
			err = rows.Scan(
				&violation.ID,
				&violation.UserID,
				&violation.ChatID,
				&violation.GivenByID,
				&violation.Reason,
				&violation.Date,
				&violation.Expdate,
				&violation.Status,
				&violation.IsWarn,
			)
			if err != nil {
				fmt.Println("line 634", err.Error())
				continue
			}
			violations = append(violations, violation)
		}
		rows.Close()
		userWithViolations[user] = append(userWithViolations[user], violations...)

	}
	return userWithViolations, nil
}

//DeleteViolation deletes warn and if user doesnt have rights for it it will return error
func (mydb *Sqldatabase) DeleteViolation(chatID int64, warnID int64, isWarn bool) (*models.Violation, error) {
	db, err := sql.Open("mysql", mydb.SQLtoken)
	if err != nil {
		fmt.Println("line 649", err.Error())
		return nil, err
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS Violations(id BIGINT NOT NULL AUTO_INCREMENT, userid INT NOT NULL, chatid BIGINT NOT NULL, givenbyid INT NOT NULL, reason TEXT NOT NULL, date BIGINT NOT NULL, expdate BIGINT NOT NULL, status BOOLEAN NOT NULL, iswarn BOOLEAN NOT NULL, PRIMARY KEY(id))")
	if err != nil {
		fmt.Println("line 656", err.Error())
		return nil, err
	}

	var (
		warn *models.Violation = &models.Violation{}
	)

	err = db.QueryRow("SELECT * FROM Violations WHERE id = ? AND iswarn = ?", warnID, isWarn).Scan(
		&warn.ID,
		&warn.UserID,
		&warn.ChatID,
		&warn.GivenByID,
		&warn.Reason,
		&warn.Date,
		&warn.Expdate,
		&warn.Status,
		&warn.IsWarn,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf(text.NoWarnWithSuchID[mydb.Buff.Chat[chatID].Language], warnID)
		}
	}
	if warn.ChatID != chatID {
		return nil, fmt.Errorf(text.YouCantDeleteThisWarn[mydb.Buff.Chat[chatID].Language], warnID)
	}

	_, err = db.Exec("DELETE FROM Violations WHERE id = ?", warnID)
	if err != nil {
		fmt.Println("Line 684", err.Error())
	}
	// if isWarn {
	// 	users := []*models.User{&models.User{UserID: warn.UserID}}
	// 	go mydb.ChangeExpdate(users, chatID, true)
	// }

	return warn, nil
}

//UpdateViolation deletes warn and if user doesnt have rights for it it will return error
func (mydb *Sqldatabase) UpdateViolation(chatID int64, warnID int64, newReason string, isWarn bool) (*models.Violation, error) {
	db, err := sql.Open("mysql", mydb.SQLtoken)
	if err != nil {
		fmt.Println("line 694", err.Error())
		return nil, err
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS Violations(id BIGINT NOT NULL AUTO_INCREMENT, userid INT NOT NULL, chatid BIGINT NOT NULL, givenbyid INT NOT NULL, reason TEXT NOT NULL, date BIGINT NOT NULL, expdate BIGINT NOT NULL, status BOOLEAN NOT NULL, iswarn BOOLEAN NOT NULL, PRIMARY KEY(id))")
	if err != nil {
		fmt.Println("line 701", err.Error())
		return nil, err
	}

	var (
		warn *models.Violation = &models.Violation{}
	)

	err = db.QueryRow("SELECT * FROM Violations WHERE id = ? AND iswarn = ?", warnID, isWarn).Scan(
		&warn.ID,
		&warn.UserID,
		&warn.ChatID,
		&warn.GivenByID,
		&warn.Reason,
		&warn.Date,
		&warn.Expdate,
		&warn.Status,
		&warn.IsWarn,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf(text.NoWarnWithSuchID[mydb.Buff.Chat[chatID].Language], warnID)
		}
	}
	if warn.ChatID != chatID {
		return nil, fmt.Errorf(text.YouCantDeleteThisWarn[mydb.Buff.Chat[chatID].Language], warnID)
	}
	warn.Reason = newReason
	_, err = db.Exec("UPDATE Violations SET reason = ? WHERE id = ?", newReason, warnID)
	if err != nil {
		fmt.Println("Line 728", err.Error())
	}

	return warn, nil
}

//ViolationAutoRemove ....
func (mydb *Sqldatabase) ViolationAutoRemove() (map[*models.User][]*models.Violation, error) {
	db, err := sql.Open("mysql", mydb.SQLtoken)
	if err != nil {
		fmt.Println("line 738", err.Error())
		return nil, err
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS Violations(id BIGINT NOT NULL AUTO_INCREMENT, userid INT NOT NULL, chatid BIGINT NOT NULL, givenbyid INT NOT NULL, reason TEXT NOT NULL, date BIGINT NOT NULL, expdate BIGINT NOT NULL, status BOOLEAN NOT NULL, iswarn BOOLEAN NOT NULL, PRIMARY KEY(id))")
	if err != nil {
		fmt.Println("line 745", err.Error())
		return nil, err
	}

	rows, err := db.Query("SELECT * FROM Violations")
	if err != nil {
		fmt.Println("line 751", err.Error())
		return nil, err
	}
	defer rows.Close()
	var (
		userWithViolations   map[*models.User][]*models.Violation = make(map[*models.User][]*models.Violation)
		userIDWithViolations map[int][]*models.Violation          = make(map[int][]*models.Violation)
		now                                                       = time.Now().Format("2006.01.02")
	)

	for rows.Next() {
		violation := new(models.Violation)
		rows.Scan(
			&violation.ID,
			&violation.UserID,
			&violation.ChatID,
			&violation.GivenByID,
			&violation.Reason,
			&violation.Date,
			&violation.Expdate,
			&violation.Status,
			&violation.IsWarn,
		)
		expDate := time.Unix(violation.Expdate, 0).Format("2006.01.02")
		if now == expDate {
			userIDWithViolations[violation.UserID] = append(userIDWithViolations[violation.UserID], violation)
			_, err = db.Exec("UPDATE Violations SET status = false WHERE id = ?", &violation.ID)
			if err != nil {
				fmt.Println("Line 775", err.Error())
			}
		}

	}

	for userID, violations := range userIDWithViolations {
		u, found := mydb.GetUserByUserID(userID)
		if !found {
			continue
		}
		userWithViolations[u] = append(userWithViolations[u], violations...)
	}

	return userWithViolations, nil
}

//AFK
// Violations: id
// Violations: userid
// Violations: chatid
// Violations: givenbyid
// Violations: reason
// Violations: date
// Violations: expdate
// Violations: status bool (active / passive)
// Violations: iswarn bool if it is warn, not afk(isWarn? afk:warn)

//AFKUser writes afk to database
func (mydb *Sqldatabase) AFKUser(message *tgbotapi.Message, users []*models.User, reason string) {
	db, err := sql.Open("mysql", mydb.SQLtoken)
	if err != nil {
		fmt.Println("352 mysql:", err.Error())
		return
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS Violations(id BIGINT NOT NULL AUTO_INCREMENT, userid INT NOT NULL, chatid BIGINT NOT NULL, givenbyid INT NOT NULL, reason TEXT NOT NULL, date BIGINT NOT NULL, expdate BIGINT NOT NULL, status BOOLEAN NOT NULL, iswarn BOOLEAN NOT NULL, PRIMARY KEY(id))")
	if err != nil {
		fmt.Println("line 359", err.Error())
		return
	}

	var (
		chatID  int64
		expdate int64 = time.Unix(int64(message.Date), 0).AddDate(0, 1, 0).Unix()
	)
	if mydb.Buff.Chat[message.Chat.ID].IsMain { //if you want to give warn from admin chat
		chatID = message.Chat.ID
	} else {
		chatID = mydb.Buff.Chat[message.Chat.ID].FatherChat
	}

	for _, user := range users {
		_, err = db.Exec("INSERT INTO Violations(userid, chatid, givenbyid, reason, date, expdate, status, iswarn) VALUES(?, ?, ?, ?, ?, ?, true, false)",
			user.UserID, chatID, message.From.ID, reason, int64(message.Date), expdate)
		if err != nil {
			fmt.Println("line 904", err.Error())
			return
		}
	}

}

// //UnAFKUsers changes warn status from active to passive
// func (mydb *Sqldatabase) UnAFKUsers(chatID int64, users []*models.User, afkNumber int64) {
// 	db, err := sql.Open("mysql", mydb.SQLtoken)
// 	if err != nil {
// 		fmt.Println("line 522", err.Error())
// 		return
// 	}
// 	defer db.Close()

// 	_, err = db.Exec("CREATE TABLE IF NOT EXISTS Violations(id BIGINT NOT NULL AUTO_INCREMENT, userid INT NOT NULL, chatid BIGINT NOT NULL, givenbyid INT NOT NULL, reason TEXT NOT NULL, date BIGINT NOT NULL, expdate BIGINT NOT NULL, status BOOLEAN NOT NULL, iswarn BOOLEAN NOT NULL, PRIMARY KEY(id))")
// 	if err != nil {
// 		fmt.Println("line 529", err.Error())
// 		return
// 	}

// 	var (
// 		amountOfAFKs int
// 	)

// U:
// 	for _, user := range users {
// 		err := db.QueryRow("SELECT COUNT(*) FROM Violations WHERE userid = ? AND chatid = ? AND iswarn = false AND status = true", user.UserID, chatID).Scan(
// 			&amountOfAFKs,
// 		)
// 		if err != nil || amountOfAFKs == 0 {
// 			continue U
// 		}

// 		if amountOfAFKs < int(afkNumber) || int(afkNumber) == 0 {
// 			//change status of last warn
// 			_, err = db.Exec("UPDATE Violations SET status=false WHERE id = (SELECT MAX(id) FROM(SELECT * FROM Violations) AS A WHERE A.userid = ? AND A.chatid = ? AND A.iswarn = false AND A.status = true)", user.UserID, chatID)
// 			if err != nil {
// 				fmt.Println("line 550", err.Error())
// 				continue U
// 			}
// 		} else {
// 			//remove status of concrete warn

// 			rows, err := db.Query("SELECT id FROM Violations WHERE userid = ? AND chatid = ? AND iswarn = false AND status = true", user.UserID, chatID)
// 			if err != nil {
// 				fmt.Println("line 558", err.Error())
// 				continue U
// 			}

// 			var (
// 				counter int
// 				ID      int64
// 			)
// 			for rows.Next() {
// 				counter++
// 				if int(afkNumber) == counter {
// 					err = rows.Scan(&ID)
// 					if err != nil {
// 						fmt.Println("line 571", err.Error())
// 						continue
// 					}
// 					_, err = db.Exec("UPDATE Violations SET status = false WHERE id = ?", ID)
// 					if err != nil {
// 						fmt.Println("line 576", err.Error())
// 						continue U
// 					}
// 					rows.Close()
// 					continue U
// 				}
// 			}
// 			rows.Close()

// 		}
// 	}
// }

//UnAFKUsers changes warn status from active to passive
func (mydb *Sqldatabase) UnAFKUsers(chatID int64, afkID int64, isWarn bool) (*models.Violation, error) {
	db, err := sql.Open("mysql", mydb.SQLtoken)
	if err != nil {
		fmt.Println("line 649", err.Error())
		return nil, err
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS Violations(id BIGINT NOT NULL AUTO_INCREMENT, userid INT NOT NULL, chatid BIGINT NOT NULL, givenbyid INT NOT NULL, reason TEXT NOT NULL, date BIGINT NOT NULL, expdate BIGINT NOT NULL, status BOOLEAN NOT NULL, iswarn BOOLEAN NOT NULL, PRIMARY KEY(id))")
	if err != nil {
		fmt.Println("line 656", err.Error())
		return nil, err
	}

	var (
		AFK *models.Violation = &models.Violation{}
	)

	err = db.QueryRow("SELECT * FROM Violations WHERE id = ? AND iswarn = ?", afkID, isWarn).Scan(
		&AFK.ID,
		&AFK.UserID,
		&AFK.ChatID,
		&AFK.GivenByID,
		&AFK.Reason,
		&AFK.Date,
		&AFK.Expdate,
		&AFK.Status,
		&AFK.IsWarn,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf(text.NoWarnWithSuchID[mydb.Buff.Chat[chatID].Language], afkID)
		}
	}
	if AFK.ChatID != chatID {
		return nil, fmt.Errorf(text.YouCantDeleteThisWarn[mydb.Buff.Chat[chatID].Language], afkID)
	}

	_, err = db.Exec("UPDATE Violations SET status = false WHERE id = ?", afkID)
	if err != nil {
		fmt.Println("Line 1076", err.Error())
	}

	return AFK, nil
}

//GetAFKInChat returns list of all afk in the chat
func (mydb *Sqldatabase) GetAFKInChat(chatID int64) ([]*models.Violation, error) {
	db, err := sql.Open("mysql", mydb.SQLtoken)
	if err != nil {
		fmt.Println("line 477", err.Error())
		return nil, err
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS Violations(id BIGINT NOT NULL AUTO_INCREMENT, userid INT NOT NULL, chatid BIGINT NOT NULL, givenbyid INT NOT NULL, reason TEXT NOT NULL, date BIGINT NOT NULL, expdate BIGINT NOT NULL, status BOOLEAN NOT NULL, iswarn BOOLEAN NOT NULL, PRIMARY KEY(id))")
	if err != nil {
		fmt.Println("line 484", err.Error())
		return nil, err
	}

	rows, err := db.Query("SELECT * FROM Violations WHERE chatid = ? AND iswarn = false", chatID)
	if err != nil {
		fmt.Println("line 1000", err.Error())
		return nil, err
	}

	var violations []*models.Violation
	for rows.Next() {
		violation := new(models.Violation)
		err = rows.Scan(
			&violation.ID,
			&violation.UserID,
			&violation.ChatID,
			&violation.GivenByID,
			&violation.Reason,
			&violation.Date,
			&violation.Expdate,
			&violation.Status,
			&violation.IsWarn,
		)
		if err != nil {
			fmt.Println("line 1019", err.Error())
			continue
		}
		violations = append(violations, violation)
	}
	rows.Close()
	return violations, nil
}

//CommandHandler main command proccessor
func (mydb *Sqldatabase) CommandHandler(message *tgbotapi.Message) ([]*models.User, string, *models.TimeRestriction, error) {
	var (
		err    error
		users  []*models.User
		reason string
		time   *models.TimeRestriction = new(models.TimeRestriction)
	)
	var msgSplitted []string
	var word []string

	utf16array := utf16.Encode([]rune(message.Text)) //Encode with unf16

	for i, v := range message.Text + " " { //split text by whitespaces
		if !unicode.IsSpace(v) {
			word = append(word, string(v))
		}
		if (unicode.IsSpace(v) || i == len(message.Text)-1) && word != nil { //Find space

			msgSplitted = append(msgSplitted, strings.Join(word, ""))
			word = nil
		}
	}

	if message.ReplyToMessage != nil {
		user := &models.User{
			UserID:    message.ReplyToMessage.From.ID,
			FirstName: message.ReplyToMessage.From.FirstName,
			LastName:  message.ReplyToMessage.From.LastName,
			UserName:  message.ReplyToMessage.From.UserName,
		}
		users = append(users, user)

		time.TimeMagnitude, time.TimeNumber, time.TimeUnix, err = TimeProc(msgSplitted[len(msgSplitted)-1])
		if err != nil {
			fmt.Println(err)
			fmt.Println("User might be restricted forever")
		}

		if len(msgSplitted) == 1 { // if only command - ban forever
			return users, "", time, nil
		}

		if time.TimeUnix == 0 {
			reason = strings.Join(msgSplitted[1:], " ")
		} else {
			reason = strings.Join(msgSplitted[1:len(msgSplitted)-1], " ")
		}

		reason = tools.AvoidMarkdownCrashReason([]byte(reason))

		return users, reason, time, nil
	} else if len(msgSplitted) == 1 {

		time.TimeMagnitude = "nil"
		time.TimeNumber = 0
		time.TimeUnix = 0
		reason = ""
		user := &models.User{
			UserID:    message.From.ID,
			FirstName: message.From.FirstName,
			LastName:  message.From.LastName,
		}
		users = append(users, user)
		return users, reason, time, nil
	} else if findMessageEntity(message) {
		var (
			positionOfName int
			lengthOfName   int
			countUsers     int
		)
		// for _, v := range msgSplitted[1:] {
		// 	if string(v[0]) == "@" {
		// 		user, found := mydb.GetUserByUserName(v[1:])
		// 		if found {
		// 			users = append(users, user)
		// 			countUsers++
		// 		}
		// 	}
		// }
		//MB 1
		for i := len(*message.Entities) - 1; i != -1; i-- {
			if (*message.Entities)[i].Type == "text_mention" || (*message.Entities)[i].Type == "mention" {
				positionOfName = (*message.Entities)[i].Offset
				lengthOfName = (*message.Entities)[i].Length
				break
			}
		}

		for _, v := range *message.Entities {
			if v.Type == "text_mention" {
				user := &models.User{
					UserID:    v.User.ID,
					FirstName: v.User.FirstName,
					LastName:  v.User.LastName,
					UserName:  v.User.UserName,
				}
				users = append(users, user)
				countUsers++
				continue
			}
			if v.Type == "mention" {
				user, found := mydb.GetUserByUserName(string(utf16.Decode(utf16array[v.Offset+1 : v.Offset+v.Length])))
				if found {
					users = append(users, user)
					countUsers++
					continue
				}
			}
		}

		time.TimeMagnitude, time.TimeNumber, time.TimeUnix, err = TimeProc(msgSplitted[len(msgSplitted)-1])
		if err != nil {
			fmt.Println(err)
			fmt.Println("User might be restricted forever")
		}
		if len(users) == 0 {
			return nil, "", time, errors.New("Users not found")
		}

		if len(msgSplitted) == countUsers+1 { //ban forever
			time.TimeMagnitude = "nil"
			time.TimeNumber = 0
			time.TimeUnix = 0
			return users, "", time, nil
		}

		if time.TimeNumber == 0 {
			reason = string(utf16.Decode(utf16array[positionOfName+lengthOfName:])) //Get it back

		} else {
			ReasonWithoutTime := string(utf16.Decode(utf16array[positionOfName+lengthOfName:])) //Get it back

			var DeleteTime []string
			var word []string
			for i, v := range ReasonWithoutTime + " " {
				if !unicode.IsSpace(v) { //if not space add word
					word = append(word, string(v))

				} // to avoid last symbol ignoring
				if (unicode.IsSpace(v) || i == len(ReasonWithoutTime)-1) && word != nil { //Find space
					DeleteTime = append(DeleteTime, strings.Join(word, "")) //if space append word
					word = nil                                              //to delete two or more spaces in the row
				}
			}
			reason = strings.Join(DeleteTime[:len(DeleteTime)-1], " ")
		}
		reason = tools.AvoidMarkdownCrashReason([]byte(reason))

		reason = strings.TrimLeft(reason, " ")

		return users, reason, time, nil
	} else if _, err := strconv.Atoi(msgSplitted[1]); err == nil { //if number - go
		var countUsers int

		for _, v := range msgSplitted[1:] {
			id, err := strconv.Atoi(v)
			if err == nil {
				user, found := mydb.GetUserByUserID(id)
				if found {
					countUsers++
					users = append(users, user)
				}
			} else {
				break
			}
		}

		time.TimeMagnitude, time.TimeNumber, time.TimeUnix, err = TimeProc(msgSplitted[len(msgSplitted)-1])
		if err != nil {
			fmt.Println(err)
			fmt.Println("User might be restricted forever")
		}
		if len(users) == 0 {
			return nil, "", time, errors.New("Users not found")
		}

		if len(msgSplitted) == countUsers+1 { //ban forever
			time.TimeMagnitude = "nil"
			time.TimeNumber = 0
			time.TimeUnix = 0
			return users, "", time, nil
		}

		if time.TimeUnix == 0 {
			reason = strings.Join(msgSplitted[countUsers+1:], " ")
		} else {
			reason = strings.Join(msgSplitted[countUsers+1:len(msgSplitted)-1], " ")
		}

		reason = tools.AvoidMarkdownCrashReason([]byte(reason))

		return users, reason, time, nil
	} else if _, err := strconv.Atoi(msgSplitted[len(msgSplitted)-1]); err == nil { //!command <text> number
		time.TimeMagnitude, time.TimeNumber, time.TimeUnix, err = TimeProc(msgSplitted[len(msgSplitted)-1])
		if err != nil {
			fmt.Println(err)
			fmt.Println("User might be restricted forever")
		}
		if time.TimeUnix == 0 {
			reason = strings.Join(msgSplitted[1:], " ")
		} else {
			reason = strings.Join(msgSplitted[1:len(msgSplitted)-1], " ")
		}
		reason = tools.AvoidMarkdownCrashReason([]byte(reason))

		return nil, reason, time, nil
	}

	return nil, "", nil, errors.New("Not valid command/nothing found")

}

func findMessageEntity(message *tgbotapi.Message) bool {

	if message.Entities != nil {
		for _, v := range *message.Entities {
			if v.Type == "mention" || v.Type == "text_mention" {
				return true
			}
		}
	}
	return false
}

//TimeProc Parses time for ban
func TimeProc(timeInMessage string) (string, int64, int64, error) { //first is a time(m,h,d) second is a real time, third is a unix time
	var (
		timeParam       string
		restrictionTime int64
	)

	splittedTime := strings.Split(timeInMessage, "")

	t, err := strconv.ParseInt(splittedTime[len(splittedTime)-1], 10, 64)
	if err == nil {
		timeParam = "h"
		t, err = strconv.ParseInt(timeInMessage, 10, 64)
		if err != nil {
			fmt.Println(err.Error())
			return "nil", 0, 0, errors.New("Wrong Time")
		}
	} else {
		t, err = strconv.ParseInt(strings.Join(splittedTime[:len(splittedTime)-1], ""), 10, 64)
		if err != nil {
			fmt.Println(err.Error())
			return "nil", 0, 0, errors.New("Wrong Time")
		}
		timeParam = splittedTime[len(splittedTime)-1]
	}

	switch timeParam {
	case "m":
		restrictionTime = time.Now().Add(time.Minute * time.Duration(t)).Unix()
		if t == 1 {
			timeParam = "minute"
		} else {
			timeParam = "minutes"
		}
	case "h":
		restrictionTime = time.Now().Add(time.Hour * time.Duration(t)).Unix()
		if t == 1 {
			timeParam = "hour"
		} else {
			timeParam = "hours"
		}
	case "d":
		restrictionTime = time.Now().Add(time.Hour * 24 * time.Duration(t)).Unix()
		if t == 1 {
			timeParam = "day"
		} else {
			timeParam = "days"
		}
	default:
		return "nil", 0, 0, errors.New("Wrong Time")
	}
	return timeParam, t, restrictionTime, nil

}
