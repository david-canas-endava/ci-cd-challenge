package main

import (
	"database/sql"
	"os"
	"strconv"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

const dbPath = "/etc/todos/todo.db"

func TestSQLiteDatabase(t *testing.T) {
	t.Log("Checking if database file exists")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Fatalf("Database file does not exist at %s", dbPath)
	}

	t.Log("Opening database connection")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Insert 10 items
	t.Log("Preparing insert statement")
	stmt, err := db.Prepare("INSERT INTO todo_items (id, name, completed) VALUES (?, ?, ?)")
	if err != nil {
		t.Fatalf("Failed to prepare insert statement: %v", err)
	}
	defer stmt.Close()

	ids := []int{}
	for i := 1; i <= 10; i++ {
		_, err := stmt.Exec(i, "Task "+strconv.Itoa(i), false)
		if err != nil {
			t.Fatalf("Failed to insert item %d: %v", i, err)
		}
		ids = append(ids, i)
	}
	t.Log("Inserted 10 items successfully")

	// Query and count items
	t.Log("Counting inserted items")
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM todo_items").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count items: %v", err)
	}
	if count < 10 {
		t.Errorf("Expected at least 10 items, but found %d", count)
	}

	// Delete only the inserted items
	t.Log("Deleting inserted items")
	for _, id := range ids {
		_, err = db.Exec("DELETE FROM todo_items WHERE id = ?", id)
		if err != nil {
			t.Fatalf("Failed to delete item with id %d: %v", id, err)
		}
	}
	t.Log("Deleted inserted items successfully")

	// Verify deletion
	t.Log("Verifying deletion of inserted items")
	err = db.QueryRow("SELECT COUNT(*) FROM todo_items WHERE id BETWEEN 1 AND 10").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count items after deletion: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 items after deletion, but found %d", count)
	}
	t.Log("All inserted items successfully deleted")
}
