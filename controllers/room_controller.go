package controllers

import (
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"github.com/gomodels"
	"encoding/json"
	"log"
	"bytes"
	"errors"
)

type RoomController struct {
}

func NewRoomController() RoomController {
	return RoomController{}
}

// Find a user by id
func (rc *RoomController) FindUser(token string, id bson.ObjectId) (models.User, error) {
	var req *http.Request
	var user models.User

	req, err := http.NewRequest(http.MethodGet, models.ID_MS_URL+"/"+id.Hex(), nil)
	req.Header.Set("Authorization", "Basic "+token)
	if err != nil {
		log.Println(err)
		return user, err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return user, err
	}

	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		log.Println(err)
		return user, err
	}
	if len(user.Id.Hex()) == 0 {
		err = errors.New("Usuário não encontrado")
		log.Println(err)
		return user, err
	}
	return user, nil
}

// Add a room to a user
func (rc *RoomController) AddRoomToUser(user models.User, room models.Room) (error) {
	var req *http.Request
	user.Rooms = append(user.Rooms, room.Id)
	body, err := json.Marshal(user)
	if err != nil {
		log.Println(err)
		return err
	}
	req, err = http.NewRequest(http.MethodPut, models.ID_MS_URL, bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Basic "+user.Token)
	if err != nil {
		log.Println(err)
		return err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		err = errors.New("Erro na requisição" + " status " + resp.Status)
		log.Println(err)
		return err
	}
	return nil
}

func (rc *RoomController) RemoveUsers(token string, roomm models.Room) {
	for _, id := range roomm.Users {
		var req *http.Request
		userDel, _ := rc.FindUser(token, id)
		for index, roomId := range userDel.Rooms {
			if roomId == roomm.Id {
				userDel.Rooms = userDel.Rooms[:index+copy(userDel.Rooms[index:], userDel.Rooms[index+1:])]
				break
			}
		}
		body, err := json.Marshal(userDel)
		if err != nil {
			log.Println(err)
		}
		req, err = http.NewRequest(http.MethodPut, models.ID_MS_URL, bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Basic "+token)
		if err != nil {
			log.Println(err)
		}
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Println(err)
		}
		if resp.StatusCode != http.StatusOK {
			err = errors.New("Erro na requisição" + " status " + resp.Status)
			log.Println(err)
		}
		resp.Body.Close()
	}

}

func (rc *RoomController) Validate(user *models.User) bool {
	req, err := http.NewRequest(http.MethodGet, models.ID_MS_URL+"/validate/"+user.Token, nil)
	if err != nil {
		log.Println(err)
		return false
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false
	}
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		log.Println(err)
		return false
	}
	if len(user.Id.Hex()) == 0 {
		err = errors.New("Usuário não encontrado")
		log.Println(err)
		return false
	}
	return true
}
