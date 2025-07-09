package database

type SupabaseClient interface {
	Get(table string) ([]byte, error)
	Insert(table string, data interface{}) ([]byte, error)
	Delete(table string, filter string) ([]byte, error)
}
