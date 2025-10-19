package sqlite

import (
	"database/sql"

	"github.com/manishtomar-cpi/go-server/internal/config"
	_ "github.com/mattn/go-sqlite3" // _ because we are using this behind the seen
)

type Sqlite struct {
	Db *sql.DB
}

func New(cfg *config.Config) (*Sqlite, error) {
	db, err := sql.Open("sqlite3", cfg.Storage_path)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS students(
	       id INTEGER PRIMARY KEY AUTOINCREMENT,
		   name TEXT,
		   age INTEGER,
		   email TEXT
	   )`)

	if err != nil {
		return nil, err
	}

	return &Sqlite{
		Db: db,
	}, nil
}

func (s *Sqlite) CreateStudent(name string, email string, age int) (int64, error) {
	stmt, err := s.Db.Prepare("INSERT INTO students (name,email,age) VALUES(?,?,?)") //preparing the data
	if err != nil {
		return 0, err
	}
	defer stmt.Close()
	res, err := stmt.Exec(name, email, age) // inserting the data
	if err != nil {
		return 0, err
	}
	id, dbErr := res.LastInsertId()

	if dbErr != nil {
		return 0, dbErr
	}
	return id, nil
}
