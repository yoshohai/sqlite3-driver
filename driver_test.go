package driver

import (
	"fmt"
	"os"
	"testing"

	backend "github.com/yoshohai/sqlite3-backend"
)

const testDbDir = "test_space"

func init() {
	PanicOnError()
	if err := os.MkdirAll(testDbDir, 0755); err != nil {
		panic(err)
	}
}

func TestCoverage(t *testing.T) {
	dbname := "test.db"
	dbPath := fmt.Sprintf("file:%s/%s?mode=rwc", testDbDir, dbname)

	db, _ := Open(dbPath, backend.NewBackend())
	defer db.Close()

	db.Exec(`DROP TABLE IF  EXISTS users;`)
	db.Exec(`CREATE TABLE IF NOT EXISTS users (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					name TEXT NOT NULL,
					age INTEGER
				 );`)

	insertUser(db, "Alice", 30)
	insertUser(db, "Bob", 25)
	updateUserAge(db, "Alice", 35)
	selectUser(db, "Alice")
	selectAllUsers(db)
	deleteUser(db, "Bob")
	selectAllUsers(db)
	bulkLoadUsers(db)
	selectAllUsers(db)
}

func insertUser(db *Connection, name string, age int64) {
	stmt, _ := db.PrepareStatement(`INSERT INTO users (name, age) VALUES (?, ?);`)
	defer stmt.Close()
	stmt.SetText(1, name)
	stmt.SetInt(2, age)
	stmt.Exec()
}

func updateUserAge(db *Connection, name string, age int64) {
	stmt, _ := db.PrepareStatement(`UPDATE users SET age = ? WHERE name = ?;`)
	defer stmt.Close()
	stmt.SetInt(1, age)
	stmt.SetText(2, name)
	stmt.Exec()
}

func selectUser(db *Connection, name string) {
	stmt, _ := db.PrepareStatement(`SELECT id, name, age FROM users WHERE name = ?;`)
	defer stmt.Close()

	stmt.SetText(1, name)
	rs := stmt.Query()
	defer rs.Close()
	for {
		if ok, _ := rs.Next(); !ok {
			break
		}
		id := rs.GetInt64(0)
		nam := rs.GetText(1)
		age := rs.GetInt64(2)
		fmt.Printf("User: id=%d, name=%s, age=%d\n", id, nam, age)
	}
}

func selectAllUsers(db *Connection) {
	stmt, _ := db.PrepareStatement(`SELECT id, name, age FROM users;`)
	defer stmt.Close()

	rs := stmt.Query()
	defer rs.Close()

	fmt.Println("All users:")
	for {
		if ok, _ := rs.Next(); !ok {
			break
		}
		id := rs.GetInt64(0)
		name := rs.GetText(1)
		age := rs.GetInt64(2)
		fmt.Printf("  id=%d, name=%s, age=%d\n", id, name, age)
	}
}

func deleteUser(db *Connection, name string) {
	stmt, _ := db.PrepareStatement(`DELETE FROM users WHERE name = ?;`)
	defer stmt.Close()
	stmt.SetText(1, name)
	stmt.Exec()
}

func bulkLoadUsers(db *Connection) {
	sql := `
	INSERT INTO users (name, age) VALUES ('ml1', 1000);
	INSERT INTO users (name, age) VALUES ('ml2', 1000);
	INSERT INTO users (name, age) VALUES ('ml3', 1000);`
	db.Exec(sql)
}
