package server

import (
	"github.com/bmizerany/pat"
	"net/http"
	"log"
	"fmt"
	"github.com/gomodels"
	"../controllers"
	"encoding/json"
	"../db"
	"gopkg.in/mgo.v2/bson"
	"github.com/rs/cors"
)

type ServerProperties struct {
	Port    string
	Address string
}

var rc = controllers.NewRoomController()

func validAuthHeader(req *http.Request) (bool,models.User) {
	auth := req.Header.Get("Authorization")
	var user models.User
	if len(auth) <= 6 {
		return false,user
	}
	user.Token = auth[6:]
	if rc.Validate(&user){
		return true, user
	}else{
		return false,user
	}
}

func unauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
}

func DeleteRoom(w http.ResponseWriter, req *http.Request) {
	validou, user := validAuthHeader(req)
	if !validou{
		unauthorized(w)
		return
	}
	var room models.Room
	id := req.URL.Query().Get(":id")
	err := room.FindById(db.GetCollection(), bson.ObjectIdHex(id))

	rc.RemoveUsers(user.Token,room)
	room.Remove(db.GetCollection())
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	ResponseWithJSON(w, nil, http.StatusNoContent)
}

func ResponseWithJSON(w http.ResponseWriter, json []byte, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	w.Write(json)
}

func InsertRoom(w http.ResponseWriter, req *http.Request) {
	validou, user := validAuthHeader(req)
	if !validou{
		unauthorized(w)
		return
	}
	var room models.Room
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&room)
	if err != nil || len(room.Id.Hex()) > 0 {
		badRequest(w, err)
		return
	}
	room.Users = append(room.Users, user.Id)
	room.CreatedBy = user.Id
	err = room.Persist(db.GetCollection())
	if err != nil {
		badRequest(w, err)
		return
	}
	log.Println("Adicionando usuário")
	user.Token = req.Header.Get("Authorization")[6:]
	err = rc.AddRoomToUser(user,room)
	if err != nil {
		room.Remove(db.GetCollection())
		badRequest(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Location", req.URL.Path+"/"+room.Id.Hex())
	w.WriteHeader(http.StatusCreated)
}

func UpdateRoom(w http.ResponseWriter, req *http.Request) {
	validou,_ := validAuthHeader(req)
	if !validou{
		unauthorized(w)
		return
	}
	var room models.Room
	var roomUp models.Room

	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&room)
	roomUp.FindById(db.GetCollection(), room.Id)
	if len(roomUp.Id.Hex()) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if err != nil {
		badRequest(w, err)
		return
	}
	err = room.Merge(db.GetCollection())
	if err != nil {
		badRequest(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}


func addUser(w http.ResponseWriter, req *http.Request) {
	validou,user := validAuthHeader(req)
	if !validou{
		unauthorized(w)
		return
	}
	var room models.Room
	tag := req.URL.Query().Get(":tag")
	room.FindByTag(db.GetCollection(), tag)
	if len(room.Id.Hex()) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	room.Users = append(room.Users, user.Id)
	err := room.Merge(db.GetCollection())
	if err != nil {
		badRequest(w, err)
		return
	}
	err = rc.AddRoomToUser(user, room)
	if err != nil {
		badRequest(w, err)
		return
	}
	resp, _ := json.Marshal(room)
	ResponseWithJSON(w, resp, http.StatusOK)
}

func FindAllRooms(w http.ResponseWriter, req *http.Request) {
	validou,_ := validAuthHeader(req)
	if !validou{
		unauthorized(w)
		return
	}
	var rooms models.Rooms
	rooms, err := rooms.FindAll(db.GetCollection())
	if err != nil {
		badRequest(w, err)
		return
	}
	resp, _ := json.Marshal(rooms)
	ResponseWithJSON(w, resp, http.StatusOK)
}

func FindById(w http.ResponseWriter, req *http.Request) {
	validou,_ := validAuthHeader(req)
	if !validou{
		unauthorized(w)
		return
	}
	var room models.Room
	id := req.URL.Query().Get(":id")
	err := room.FindById(db.GetCollection(), bson.ObjectIdHex(id))
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	resp, _ := json.Marshal(room)
	ResponseWithJSON(w, resp, http.StatusOK)
}

func FindByUserId(w http.ResponseWriter, req *http.Request) {
	validou,_ := validAuthHeader(req)
	if !validou{
		unauthorized(w)
		return
	}
	var rooms models.Rooms
	id := req.URL.Query().Get(":id")
	rooms, err := rooms.FindByUserId(db.GetCollection(), bson.ObjectIdHex(id))
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	resp, _ := json.Marshal(rooms)
	ResponseWithJSON(w, resp, http.StatusOK)
}

func badRequest(w http.ResponseWriter, err error) {
	log.Println(err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
}

func Start(properties ServerProperties) {
	startDb()
	m := pat.New()
	handler := cors.AllowAll().Handler(m)
	mapEndpoints(*m, properties)
	http.Handle("/", handler)
	fmt.Println("servidor iniciado no endereço localhost:" + properties.Port + properties.Address)
	err := http.ListenAndServe(":"+properties.Port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
func mapEndpoints(m pat.PatternServeMux, properties ServerProperties) {
	m.Post(properties.Address, http.HandlerFunc(InsertRoom))
	m.Put(properties.Address, http.HandlerFunc(UpdateRoom))
	m.Del(properties.Address+"/:id", http.HandlerFunc(DeleteRoom))
	m.Get(properties.Address, http.HandlerFunc(FindAllRooms))
	m.Get(properties.Address+"/:id", http.HandlerFunc(FindById))
	m.Get(properties.Address+"/user/:id", http.HandlerFunc(FindByUserId))
	m.Post(properties.Address+"/:tag/users", http.HandlerFunc(addUser))
}

func startDb() {
	db.Start()
}
