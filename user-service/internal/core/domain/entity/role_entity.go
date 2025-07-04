package entity

type RoleEntity struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Users []UserEntity
}