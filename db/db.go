package db

type SimpleDatabase struct {
	store *SimpleInMemoryDatabase[string, struct{}]
}

func NewSimpleDatabase() *SimpleDatabase {
	return &SimpleDatabase{
		store: NewSimpleInMemoryDatabase[string, struct{}](),
	}
}

func (db *SimpleDatabase) Add(value string) {
	db.store.Add(value, struct{}{})
}

func (db *SimpleDatabase) Exists(value string) bool {
	return db.store.Exists(value)
}
