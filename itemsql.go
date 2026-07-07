package fynewidgets

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Fact represents a piece of information with metadata
type Fact struct {
	Title   string `json:"title"`
	Detail    string `json:"detail"`
	Category string `json:"category"`
}

// NewFact creates and returns a new Fact instance
func NewFact(title, detail, category string) Fact {
	return Fact{
		Title:   title,
		Detail:    detail,
		Category: category,
	}
}

// Database
var db *sql.DB

// CreateTable creates the fact table if it doesn't exist
func CreateTable(sqlDB *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS facts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		detail TEXT NOT NULL,
		category TEXT NOT NULL
	)`

	_, err := sqlDB.Exec(query)
	return err
}

// GetFacts retrieves all facts from the database
func GetFacts() ([]Fact, error) {
	var facts []Fact
	query := `SELECT id, title, detail, category FROM facts`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var f Fact
		var id int
		if err := rows.Scan(&id, &f.Title, &f.Detail, &f.Category); err != nil {
			return nil, err
		}
		f.ID = id
		facts = append(facts, f)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return facts, nil
}

// GetFactByID retrieves a single fact from the database by its ID
func GetFactByID(id int) (*Fact, error) {
	var f Fact
	var id int
	query := `SELECT id, title, detail, category FROM facts WHERE id = ?`
	row := db.QueryRow(query, id)

	if err := row.Scan(&id, &f.Title, &f.Detail, &f.Category); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("fact not found")
		}
		return nil, err
	}

	f.ID = id
	return &f, nil
}

// UpdateFact updates an existing fact by its ID
func UpdateFact(id int, title, detail, category string) error {
	query := `UPDATE facts SET title = ?, detail = ?, category = ? WHERE id = ?`
	_, err := db.Exec(query, title, detail, category, id)
	return err
}

// DeleteFact deletes a fact by its ID
func DeleteFact(id int) error {
	query := `DELETE FROM facts WHERE id = ?`
	result, err := db.Exec(query, id)
	if err != nil {
		return err
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("fact not found")
	}
	return nil
}

