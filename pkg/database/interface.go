package database

type SupabaseClient interface {
	Get(table string, data map[string]string) ([]byte, error)
	Insert(table string, data interface{}) ([]byte, error)
	Delete(table string, filter string) ([]byte, error)
}
