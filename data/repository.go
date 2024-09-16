package data

type Repository interface {
	Insert(user User) (string, error)
	GetAll() ([]*User, error)
	GetByEmail(email string) (*User, error)
	GetOne(userEmail string) (*User, error)
	Update(user User) error
	Delete(id string) error
}
