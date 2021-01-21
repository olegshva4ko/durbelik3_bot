package mongodatabase

import (
	"Durbelik3/pkg/models"
	"context"
	"fmt"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

/* Calling after message instance.
Connects to mongoclient
Chooses the collection
Checks to not add the same user(user with the same last parametres)
IF parametres differs - write down new data
*/

//MongoDatabase keeps Client open
type MongoDatabase struct {
	Client *mongo.Client
}

//NewClient func that connects to mongodatabase 
func NewClient() (*MongoDatabase) {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://127.0.0.1:27017")) //connecting to db
	if err != nil {
		panic(err)
	}

	err = client.Connect(context.TODO())
	if err != nil {
		panic(err)
	}

	return &MongoDatabase{client}
}

//CheckDB checks db
func (m *MongoDatabase) CheckDB() error {
	err := m.Client.Ping(context.TODO(), nil)
	if err != nil {
		return err
	}
	return nil
}

//MongoDbAddUpdate adds and updates user in database
func (m *MongoDatabase) MongoDbAddUpdate(u *tgbotapi.User) {
	
	collection := m.Client.Database("User_history").Collection("user_history")
	filter := bson.D{{Key: "userid", Value: strconv.Itoa(u.ID)}}        //search with filter(by userid)
	cur, err := collection.Find(context.TODO(), filter, options.Find()) //returns array
	if err != nil {
		fmt.Println(err)
	}

	results := []*models.MongoResult{}

	for cur.Next(context.TODO()) { //every value in array store in result and append to results
		result := models.MongoResult{}
		cur.Decode(&result)
		results = append(results, &result)

	}

	cur.Close(context.TODO()) //Close cur after checking

	if len(results) > 0 { // check last value of array. If something has changed - update db
		last := results[len(results)-1]
		if (last.Firstname != u.FirstName) || (last.Lastname != u.LastName) || (last.Username != u.UserName) { //if user made changes in parametres write it to database
			m.MongoCreate(m.Client, u)
		}
	} else if len(results) == 0 { //No user in db
		m.MongoCreate(m.Client, u)
	}
	return
}

//MongoCreate creates new user in database
func (m *MongoDatabase) MongoCreate(client *mongo.Client, u *tgbotapi.User) error {
	collection := m.Client.Database("User_history").Collection("user_history")
	insertresult, err := collection.InsertOne(context.TODO(), bson.D{
		{Key: "userid", Value: strconv.Itoa(u.ID)},
		{Key: "firstname", Value: u.FirstName},
		{Key: "lastname", Value: u.LastName},
		{Key: "username", Value: u.UserName},
		{Key: "date", Value: time.Now().Format("2006.02.01")}},
	)
	if err != nil {
		fmt.Println(err)
	}

	//write data down to mongodb
	fmt.Println(insertresult.InsertedID)
	return nil
}

// //GetUserNickStory ...
// func GetUserNickStory(user *models.User) (string, error) {
// 	client := MongoConnect()
// 	defer client.Disconnect(context.TODO())
// 	// sverhu musthavechatid
// 	collection := client.Database("User_history").Collection("user_history")

// 	filter := bson.D{{Key: "userid", Value: user.UserID}}               //search with filter(by userid)
// 	cur, err := collection.Find(context.TODO(), filter, options.Find()) //returns array
// 	if err != nil {
// 		fmt.Println(err.Error())
// 		return "", err
// 	}

// 	results := []string{}

// 	for cur.Next(context.TODO()) { //every value in array store in result and append to results
// 		result := models.MongoResult{}
// 		cur.Decode(&result)
// 		results = append(results,
// 			fmt.Sprintf("FirstName: %s\nLastName: %s\nUserName: %s\nDate: %s",
// 				result.Firstname, result.Lastname, result.Username, result.Date))
// 	}
// 	cur.Close(context.TODO()) //Close cur after checking

// 	return strings.Join(results, "\n--------------------\n"), nil
// }
