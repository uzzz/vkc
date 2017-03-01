package vkc

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"
	"text/template"
)

type Country struct {
	Id    int    `json:"id"`
	Title string `json:"title"`
}

type City struct {
	Id    int    `json:"id"`
	Title string `json:"title"`
}

type User struct {
	Id          int      `json:"id"`
	FirstName   string   `json:"first_name"`
	LastName    string   `json:"last_name"`
	Deactivated string   `json:"deactivated"`
	City        *City    `json:"city"`
	Country     *Country `json:"country"`
	Counters    struct {
		Audios int `json:"audios"`
		Videos int `json:"videos"`
	} `json:"counters"`
	Bdate    string `json:"bdate"`
	LastSeen struct {
		Time int `json:"time"`
	} `json:"last_seen"`
	HomeTown    *string `json:"home_town"`
	Sex         *int    `json:"sex"`
	Relation    *int    `json:"relation"`
	CanSeeAudio int     `json:"can_see_audio"`
	Occupation  *struct {
		Type string `json:"type"`
		Name string `json:"name"`
	}
}

func (u *User) IdBytes() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, u.Id)
	return buf.Bytes()
}

func (u *User) IsDeleted() bool {
	return u.Deactivated == "deleted"
}

func (u *User) IsBanned() bool {
	return u.Deactivated == "banned"
}

func (u *User) IsUnavailable() bool {
	return u.IsDeleted() || u.IsBanned()
}

type usersService struct {
	client *VkClient
}

func (s *usersService) Get(id int) (*User, error) {
	params := map[string]string{
		"user_ids": strconv.Itoa(id),
		"fields":   "city,country,counters,bdate,last_seen,home_town,sex,relation,can_see_audio,occupation",
	}
	req, err := s.client.NewRequest("users.get", params)
	if err != nil {
		return nil, err
	}

	users := make([]User, 0)
	_, err = s.client.Do(req, &users)
	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, nil
	} else {
		return &users[0], nil
	}
}

var (
	getUsers = `
var users = [];
var ids = [{{ range $i, $e := $ -}}{{if $i}}, {{end}}{{ $e }}{{ end }}];
var i = 0;
while (i < ids.length) {
    var response = API.users.get({"user_ids": ids[i], "fields": "city,country,bdate,last_seen,counters,home_town,sex,relation,can_see_audio,occupation"});
    users.push(response[0]);
    i = i + 1;
}
return users;`
	getUsersTemplate = template.Must(template.New("get.users").Parse(getUsers))
)

func (s *usersService) GetBatch(ids []int) ([]*User, error) {
	if len(ids) > MaximumExecuteBatchSize {
		return nil, fmt.Errorf("batch size can't be more than %d", MaximumExecuteBatchSize)
	}

	users := make([]*User, 0)

	var buf bytes.Buffer
	err := getUsersTemplate.Execute(&buf, ids)
	if err != nil {
		return nil, err
	}

	err = s.client.Execute(buf.String(), &users)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (s *usersService) GetRange(from, to int) ([]*User, error) {
	if from > to {
		return nil, errors.New("from should be less than to")
	}

	ids := make([]int, to-from+1)

	for i := from; i <= to; i++ {
		ids[i-from] = i
	}

	return s.GetBatch(ids)
}
