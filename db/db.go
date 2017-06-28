package db

import (
	"gopkg.in/mgo.v2"
	"fmt"
)

var fSession mgo.Session

func Start() {
	session, err := mgo.Dial("localhost")
	if err != nil {
		panic(err)
	}
	fSession = *session
	fmt.Println("Sessão do banco criada")
}

func GetCollection() (c *mgo.Collection) {
	s := fSession.Copy()
	c = s.DB("school").C("rooms")
	return c
}

func closeSession(c *mgo.Collection) {
	c.Database.Session.Close()
}
