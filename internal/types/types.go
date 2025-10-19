package types

type Student struct {
	Id    int64
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
	Age   int    `json:"age" validate:"required,gte=1,lte=100"`
}
