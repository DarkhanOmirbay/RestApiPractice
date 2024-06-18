package data

type User struct {
	ID       int64
	Email    string
	Password Password
}
type Password struct {
	PlainText *string
	Hash      []byte
}
