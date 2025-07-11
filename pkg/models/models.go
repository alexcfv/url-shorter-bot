package models

type RequestData struct {
	Url string `json:"url"`
}
type Respons struct {
	Url string `json:"short_url"`
}
type Url struct {
	Hash string
	Url  string
}

type SupabaseResponse []Url

var Config ConfigStruct
