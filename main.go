package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

type Order struct { //used to convert our given json and convert to something we can use
	ID                   primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Date                 string             `json:"date,omitempty" bson:"date,omitempty"`
	Order_number         string             `json:"order_number,omitempty" bson:"order_number,omitempty"`
	Chef_id              string             `json:"chef_id,omitempty" bson:"chef_id,omitempty"`
	Experience_name      string             `json:"Experience_name,omitempty" bson:"Experience_name,omitempty"`
	Experience_id        string             `json:"Experience_id,omitempty" bson:"Experience_id,omitempty"`
	Head_count           string             `json:"Head_count,omitempty" bson:"Head_count,omitempty"`
	Chef_name            string             `json:"Chef_name,omitempty" bson:"Chef_name,omitempty"`
	Chef_email           string             `json:"Chef_email,omitempty" bson:"Chef_email,omitempty"`
	Chef_experience_time string             `json:"Chef_experience_time,omitempty" bson:"Chef_experience_time,omitempty"`
}

// type Experience struct {
// 	Time        string `json:"Time,omitempty" bson:"Time,omitempty"`               //making at a number between 9 and 17, (9 AM to 5 PM)
// 	Experiences string `json:"Experiences,omitempty" bson:"Experiences,omitempty"` //making it a string seperated by commas
// }

func main() {

	loginClient := options.Client().ApplyURI("mongodb+srv://rgupta:C0lvinrun123@cluster0.74it6.mongodb.net/myFirstDatabase?retryWrites=true&w=majority")
	client, _ = mongo.Connect(context.TODO(), loginClient)
	if client == nil {
		fmt.Println("Client NUll in start")
		log.Fatal(client)
	}
	/*databases, err := client.ListDatabaseNames(context.TODO(), bson.M{})
	if err != nil {
		log.Fatal(err) Prints out database names to see if connection was successful
	}*/

	router := mux.NewRouter()
	router.HandleFunc("/api/order", PostOrderEndpoint).Methods("POST")
	router.HandleFunc("/api/remove/{id}", RemoveSingleOrder).Methods("DELETE")
	router.HandleFunc("/api/orders", GetAllOrders).Methods("GET")
	router.HandleFunc("/api/order/{id}", GetSingleOrder).Methods("GET")
	log.Fatal(http.ListenAndServe(":80", router))

	//myFirstDatabase := client.Database("myDataBase")
	//orderCollection := myFirstDatabase.Collection("Order")
	//orderResult, err := orderCollection.InsertOne(ctx, bson.D{
	//	{Key: "title", Value: "The Polyglot Developer Podcast"},
	//	{Key: "author", Value: "Nic Raboy"},
	//})
}

func PostOrderEndpoint(w http.ResponseWriter, r *http.Request) { //getting the data from json format into an Order, and inserting into database
	w.Header().Set("Content-Type", "application/json")                   //response type, the type that the person asking for data will expect most likley
	orderCollection := client.Database("myDataBase").Collection("Order") //getting our collection
	var order Order
	var orders []Order
	//var orders []Order
	json.NewDecoder(r.Body).Decode(&order)
	//Check json format
	if orderCollection == nil {
		fmt.Println("Client is NULL")
		log.Fatal(client)
	}

	time_string := order.Chef_experience_time
	time, _ := strconv.Atoi(time_string)
	//here we want to query our database so we can make sure no one else has the time that we want to insert
	cursor, err := orderCollection.Find(context.TODO(), bson.M{}) //gets our cursor, with all of our data
	if err != nil {
		fmt.Print("Error getting cursor")
		log.Fatal(err)
		return
	}
	defer cursor.Close(context.TODO()) //now that we have cursor we can sever connnection
	for cursor.Next(context.TODO()) {  //iterating through cursor (our data in collection)
		var ord Order //temp variable for each cursor that we store in []orders
		err := cursor.Decode(&ord)
		if err != nil {
			fmt.Println("Decode Cursor Error")
			log.Fatal(err)
			return
		}
		orders = append(orders, ord) //appending our current cursor object to []orders
	}
	//now that we have a list we can work with, we can do the bonus properly pretty simply newOptions
	for i := range orders {
		curr_time_string := orders[i].Chef_experience_time
		curr_time, _ := strconv.Atoi(curr_time_string)
		if time == curr_time || time < 9 || time > 17 {
			fmt.Println("Your time is invalid, or already reserved")
			return
		}

	}

	insertion, err := orderCollection.InsertOne(context.TODO(), order) // inserting our order into the collection

	// if err != nil {
	// 	fmt.Println("Insertion Error")
	// 	log.Fatal(err)
	// }

	//fmt.Println(orders[1].Chef_experience_time)
	json.NewEncoder(w).Encode(insertion) //responding to the http request

}

func GetAllOrders(w http.ResponseWriter, r *http.Request) { //getting the data from json format into an Order
	w.Header().Add("Content-Type", "application/json")              //response type
	var orders []Order                                              //creates our list, we will store the cursor objects in here
	collection := client.Database("myDataBase").Collection("Order") //is our collection
	cursor, err := collection.Find(context.TODO(), bson.M{})        //gets our cursor, with all of our data
	if err != nil {
		fmt.Print("Error getting cursor")
		log.Fatal(err)
		return
	}
	defer cursor.Close(context.TODO()) //now that we have cursor we can sever connnection
	for cursor.Next(context.TODO()) {  //iterating through cursor (our data in collection)
		var ord Order //temp variable for each cursor that we store in []orders
		err := cursor.Decode(&ord)
		if err != nil {
			fmt.Println("Decode Cursor Error")
			log.Fatal(err)
			return
		}
		orders = append(orders, ord) //appending our current cursor object to []orders

	}

	json.NewEncoder(w).Encode(orders)

}

func GetSingleOrder(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	collection := client.Database("myDataBase").Collection("Order") //is our collection

	if collection == nil {
		fmt.Println("Getting collection nil Error")
		log.Fatal(collection)
	}
	rout_info := mux.Vars(r)
	//need this for a parameter, route parameters, and also contains the request we have been given
	id, err := primitive.ObjectIDFromHex(rout_info["id"]) //this line is getting the passed in id, and we need to convert it into something we can use, normal strings wont work
	//getting our info by, got this line from the web, was getting an error for a very, very long time
	if err != nil {
		fmt.Println("Getting ID Error")
		log.Fatal(err)
	}
	//data := bson.M{"Order_number": order_number} //gets our data by the "id" variable, we now just need to encode it
	// if data == nil {
	// 	fmt.Println("Getting Data nil Error")
	// 	log.Fatal(data)
	// }
	var order Order
	collection.FindOne(context.TODO(), Order{ID: id}).Decode(&order)
	//err doesnt work for some reason
	//finds the actual document, or piece of data, we want the document by the id that was defined earlier, and is then decoding it into the memory address of order (&order)
	json.NewEncoder(w).Encode(order) //incodes it in json and writes it out
}

func RemoveSingleOrder(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	collection := client.Database("myDataBase").Collection("Order") //is our collection

	if collection == nil {
		fmt.Println("Getting collection nil Error")
		log.Fatal(collection)
	}
	rout_info := mux.Vars(r)
	//need this for a parameter, route parameters, and also contains the request we have been given
	id, err := primitive.ObjectIDFromHex(rout_info["id"]) //this line is getting the passed in id, and we need to convert it into something we can use, normal strings wont work
	//getting our info by, got this line from the web, was getting an error for a very, very long time
	if err != nil {
		fmt.Println("Getting ID Error")
		log.Fatal(err)
	}
	//data := bson.M{"Order_number": order_number} //gets our data by the "id" variable, we now just need to encode it
	// if data == nil {
	// 	fmt.Println("Getting Data nil Error")
	// 	log.Fatal(data)
	// }
	deleted, err := collection.DeleteOne(context.TODO(), Order{ID: id})
	if err != nil {
		fmt.Println("Getting Deletion Error")
		log.Fatal(err)
	}
	//err doesnt work for some reason
	//finds the actual document, or piece of data, we want the document by the id that was defined earlier, and is then decoding it into the memory address of order (&order)
	json.NewEncoder(w).Encode(deleted)
}
