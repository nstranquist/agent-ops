package agenttrace

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	_ "modernc.org/sqlite"
)

// SearchHit is one FTS match.
type SearchHit struct {
	Source string  `json:"source"`
	Role   string  `json:"role"`
	Text   string  `json:"text"`
	Rank   float64 `json:"rank"`
}

// IndexResult summarizes an index run.
type IndexResult struct {
	DB           string `json:"db"`
	Source       string `json:"source"`
	MessageCount int    `json:"message_count"`
}

// OpenDB opens (or creates) a SQLite DB with FTS5 messages table.
func OpenDB(path string) (*sql.DB, error) {
	if strings.TrimSpace(path) == "" {
		return nil, fmt.Errorf("db path empty")
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}
	schema := `
CREATE TABLE IF NOT EXISTS messages (
  id INTEGER PRIMARY KEY,
  source TEXT NOT NULL,
  role TEXT NOT NULL,
  text TEXT NOT NULL
);
CREATE VIRTUAL TABLE IF NOT EXISTS messages_fts USING fts5(
  source, role, text,
  content='messages',
  content_rowid='id'
);
CREATE TRIGGER IF NOT EXISTS messages_ai AFTER INSERT ON messages BEGIN
  INSERT INTO messages_fts(rowid, source, role, text) VALUES (new.id, new.source, new.role, new.text);
END;
`
	if _, err := db.Exec(schema); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("schema: %w", err)
	}
	return db, nil
}

// IndexSession parses a JSONL session file into the SQLite FTS index.
func IndexSession(dbPath, sessionPath string) (IndexResult, error) {
	f, err := os.Open(sessionPath)
	if err != nil {
		return IndexResult{}, err
	}
	defer f.Close()
	sum, err := ParseJSONL(f, sessionPath)
	if err != nil {
		return IndexResult{}, err
	}
	db, err := OpenDB(dbPath)
	if err != nil {
		return IndexResult{}, err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return IndexResult{}, err
	}
	// replace prior rows for this source so re-index is idempotent
	if _, err := tx.Exec(`DELETE FROM messages_fts WHERE source = ?`, sessionPath); err != nil {
		// fts5 external content delete via messages table
	}
	if _, err := tx.Exec(`DELETE FROM messages WHERE source = ?`, sessionPath); err != nil {
		_ = tx.Rollback()
		return IndexResult{}, err
	}
	// rebuild fts content for remaining rows is automatic via triggers on insert;
	// after delete from content table, sync fts:
	if _, err := tx.Exec(`INSERT INTO messages_fts(messages_fts) VALUES('rebuild')`); err != nil {
		// rebuild may fail on empty; continue with deletes only
		_, _ = tx.Exec(`DELETE FROM messages_fts`)
		// re-insert all remaining after we finish? simpler approach: wipe fts rebuild after all inserts
	}

	stmt, err := tx.Prepare(`INSERT INTO messages(source, role, text) VALUES (?, ?, ?)`)
	if err != nil {
		_ = tx.Rollback()
		return IndexResult{}, err
	}
	defer stmt.Close()
	for _, m := range sum.Messages {
		if _, err := stmt.Exec(sessionPath, m.Role, m.Text); err != nil {
			_ = tx.Rollback()
			return IndexResult{}, err
		}
	}
	// ensure fts is consistent
	if _, err := tx.Exec(`INSERT INTO messages_fts(messages_fts) VALUES('rebuild')`); err != nil {
		_ = tx.Rollback()
		return IndexResult{}, fmt.Errorf("fts rebuild: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return IndexResult{}, err
	}
	return IndexResult{DB: dbPath, Source: sessionPath, MessageCount: sum.MessageCount}, nil
}

// Search runs FTS5 MATCH and returns ranked hits.
func Search(dbPath, query string, limit int) ([]SearchHit, error) {
	q := strings.TrimSpace(query)
	if q == "" {
		return nil, fmt.Errorf("empty query")
	}
	if limit <= 0 {
		limit = 20
	}
	db, err := OpenDB(dbPath)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	// quote query for simple token search; escape double quotes
	safe := strings.ReplaceAll(q, `"`, `""`)
	rows, err := db.Query(`
SELECT m.source, m.role, m.text, bm25(messages_fts) AS rank
FROM messages_fts
JOIN messages m ON m.id = messages_fts.rowid
WHERE messages_fts MATCH ?
ORDER BY rank
LIMIT ?`, `"`+safe+`"`, limit)
	if err != nil {
		// fallback unquoted MATCH for multi-token
		rows, err = db.Query(`
SELECT m.source, m.role, m.text, bm25(messages_fts) AS rank
FROM messages_fts
JOIN messages m ON m.id = messages_fts.rowid
WHERE messages_fts MATCH ?
ORDER BY rank
LIMIT ?`, safe, limit)
		if err != nil {
			return nil, err
		}
	}
	defer rows.Close()
	var hits []SearchHit
	for rows.Next() {
		var h SearchHit
		if err := rows.Scan(&h.Source, &h.Role, &h.Text, &h.Rank); err != nil {
			return nil, err
		}
		hits = append(hits, h)
	}
	return hits, rows.Err()
}
