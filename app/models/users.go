package models

import (
	"encoding/json"
	"strconv"

	"github.com/jinzhu/gorm"
)

type User struct {
	Username string `gorm:"unique_index" binding:"required"`
	Email    string `gorm:"unique_index" binding:"required"`
	gorm.Model
}

func (u *User) UnmarshalJSON(data []byte) error {
	var aux struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Email    string `json:"email"`
		Username string `json:"username"`
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	id, err := strconv.ParseUint(aux.ID, 10, 32)
	if err != nil {
		return err
	}

	u.ID = uint(id)
	u.Email = aux.Email
	u.Username = aux.Username

	return nil
}
