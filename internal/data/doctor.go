package data

type Doctor struct {
	ID         int64  `json:"ID"`
	Name       string `json:"Name"`
	Surname    string `json:"Surname"`
	Position   string `json:"Position"`
	Age        uint8  `json:"Age"`
	Experience int8   `json:"Experience"`
}
