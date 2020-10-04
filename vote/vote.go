// Package vote provides functionality around reading and writing votes.
package vote

import (
	"database/sql"
	"fmt"
	"log"
)

const (
	// DBName is the name of the database for the votes API.
	DBName = "votes"
)

// ResultCount is a pair of a vote result and the sum of votes for the result.
type ResultCount struct {
	Result string `json:"result"`
	Count  int    `json:"count"`
}

// DB is the interface for all the operations allowed on votes.
type DB interface {
	Store(voterID, vote string) error
	Result(voterID string) (vote string, err error)
	Results() (results []ResultCount, err error)
}

// NewSQLDB creates a sql database to read and store votes.
func NewSQLDB(db *sql.DB) DB {
	return &sqlDB{
		conn: db,
	}
}

type execQuerier interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Exec(query string, args ...interface{}) (sql.Result, error)
}

type sqlDB struct {
	conn execQuerier
}

// Store a vote in the database.
func (db *sqlDB) Store(voterID, vote string) error {
	_, err := db.conn.Exec(`INSERT INTO votes (id, vote) VALUES ($1, $2)`, voterID, vote)
	if err == nil {
		return nil
	}
	log.Printf("INFO: vote: update vote for voter id %s\n", voterID)
	_, err = db.conn.Exec(`UPDATE votes SET vote = $1 WHERE id = $2`, vote, voterID)
	if err != nil {
		return fmt.Errorf("vote: store vote %s for voter id %s: %v", vote, voterID, err)
	}
	return nil
}

// Result returns a voter's result.
// If there are no votes, returns a ErrNoVote error.
func (db *sqlDB) Result(voterID string) (string, error) {
	var result string
	row := db.conn.QueryRow(`SELECT vote FROM votes WHERE id=$1`, voterID)
	switch err := row.Scan(&result); err {
	case nil:
		return result, nil
	case sql.ErrNoRows:
		return "", ErrNoVote{voterID}
	default:
		return "", fmt.Errorf("vote: get result for voter id %s: %v", voterID, err)
	}
}

// Results returns the pair of results and counts.
func (db *sqlDB) Results() ([]ResultCount, error) {
	rows, err := db.conn.Query(`SELECT vote, COUNT(id) AS count FROM votes GROUP BY vote`)
	if err != nil {
		return nil, fmt.Errorf("vote: retrieve voting results: %v", err)
	}
	defer rows.Close()

	var results []ResultCount
	for rows.Next() {
		var rc ResultCount
		if err := rows.Scan(&rc.Result, &rc.Count); err != nil {
			return nil, fmt.Errorf("vote: scan row to result count pair: %v", err)
		}
		results = append(results, rc)
	}
	return results, nil
}

// CreateTableIfNotExist creates the "votes" table if it does not exist already.
func CreateTableIfNotExist(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS votes (id VARCHAR(255) NOT NULL UNIQUE, vote VARCHAR(255) NOT NULL)`)
	if err != nil {
		return fmt.Errorf(`vote: create "votes" table: %v\n`, err)
	}
	return nil
}
