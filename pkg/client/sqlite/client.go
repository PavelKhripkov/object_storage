package sqlite

import (
	"context"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

func NewClient(ctx context.Context, username, password, filePath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", filePath)
	if err != nil {
		return nil, err
	}

	//st, err := db.Prepare("CREATE TABLE IF NOT EXISTS user (id TEXT PRIMARY KEY, name VARCHAR)")
	//if err != nil {
	//	return nil, err
	//}
	//
	//_, err = st.Exec()
	//if err != nil {
	//	return nil, err
	//}

	return db, nil
}
