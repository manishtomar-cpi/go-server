package storage

type Storage interface {
	CreateStudent(name string, email string, age int) (int64, error) // will return new added id and error also
}
