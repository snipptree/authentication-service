package data

import "go.mongodb.org/mongo-driver/bson/primitive"

type Repository interface {
	Insert(user User) (primitive.ObjectID, error)
	GetAll() ([]*User, error)
	GetByEmail(email string) (*User, error)
	GetOne(userEmail string) (*User, error)
	Update(user User) error
	Delete(id string) error
}
