package models

type RequestData struct {
	Url string `json:"url"`
}

type Url struct {
	Hash string
	Url  string
}

type SupabaseResponse []Url

var Config ConfigStruct
