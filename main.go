package main

import (
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type Test struct {
	Data string `bson:"text"`
	ID   string `bson:"id"`
}

func main() {
	fmt.Println("Mongodb sharding test")
	session, _ := mgo.Dial("localhost:11000")
	fmt.Println(session.LiveServers())
	defer session.Close()
	s := session.Clone()
	s.SetMode(mgo.Nearest, false)
	var data Test
	s.DB("datasets").C("data").Find(bson.M{}).One(&data)
	fmt.Println(data.Data)
	s.DB("datasets").C("data").Find(bson.M{}).One(&data)
	fmt.Println(data.Data)
	s.DB("datasets").C("data").Find(bson.M{}).One(&data)
	fmt.Println(data.Data)
	s.DB("datasets").C("data").Find(bson.M{}).One(&data)
	fmt.Println(data.Data)
	s.DB("datasets").C("data").Find(bson.M{}).One(&data)
	fmt.Println(data.Data)
	s.DB("datasets").C("data").Find(bson.M{}).One(&data)
	fmt.Println(data.Data)
	s.DB("datasets").C("data").Find(bson.M{}).One(&data)
	fmt.Println(data.Data)
}
