package storage

import (
	"time"

	// This line is must for working MySQL database
	_ "github.com/go-sql-driver/mysql"

	"github.com/jmoiron/sqlx"

	"github.com/maddevsio/punisher/config"
	"github.com/maddevsio/punisher/model"
)

// MySQL provides api for work with mysql database
type MySQL struct {
	conn *sqlx.DB
}

// NewMySQL creates a new instance of database API
func NewMySQL(c *config.BotConfig) (*MySQL, error) {
	conn, err := sqlx.Open("mysql", c.DatabaseURL)
	if err != nil {
		return nil, err
	}
	m := &MySQL{}
	m.conn = conn
	return m, nil
}

// CreateStandup creates standup entry in database
func (m *MySQL) CreateStandup(s model.Standup) (model.Standup, error) {
	res, err := m.conn.Exec(
		"INSERT INTO `standup` (created, modified, username, comment) VALUES (?, ?, ?, ?)",
		time.Now().UTC(), time.Now().UTC(), s.Username, s.Comment,
	)
	if err != nil {
		return s, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return s, err
	}
	s.ID = id
	return s, nil
}

// UpdateStandup updates standup entry in database
func (m *MySQL) UpdateStandup(s model.Standup) (model.Standup, error) {
	_, err := m.conn.Exec(
		"UPDATE `standup` SET modified=?, username=?, comment=? WHERE id=?",
		time.Now().UTC(), s.Username, s.Comment, s.ID,
	)
	if err != nil {
		return s, err
	}
	var i model.Standup
	err = m.conn.Get(&i, "SELECT * FROM `standup` WHERE id=?", s.ID)
	return i, err
}

// SelectStandup selects standup entry from database
func (m *MySQL) SelectStandup(id int64) (model.Standup, error) {
	var s model.Standup
	err := m.conn.Get(&s, "SELECT * FROM `standup` WHERE id=?", id)
	return s, err
}

// SelectStandupByMessageTS selects standup entry from database
func (m *MySQL) SelectStandupByMessageTS(messageTS string) (model.Standup, error) {
	var s model.Standup
	err := m.conn.Get(&s, "SELECT * FROM `standup` WHERE message_ts=?", messageTS)
	return s, err
}

// DeleteStandup deletes standup entry from database
func (m *MySQL) DeleteStandup(id int64) error {
	_, err := m.conn.Exec("DELETE FROM `standup` WHERE id=?", id)
	return err
}

// ListStandups returns array of standup entries from database
func (m *MySQL) ListStandups() ([]model.Standup, error) {
	items := []model.Standup{}
	err := m.conn.Select(&items, "SELECT * FROM `standup`")
	return items, err
}

//LastStandupFor returns last standup for intern
func (m *MySQL) LastStandupFor(username string) (model.Standup, error) {
	var standup model.Standup
	err := m.conn.Get(&standup, "SELECT * FROM `standup` WHERE username=? ORDER BY id DESC LIMIT 1", username)
	return standup, err
}

// CreateIntern creates intern
func (m *MySQL) CreateIntern(s model.Intern) (model.Intern, error) {
	res, err := m.conn.Exec(
		"INSERT INTO `interns` (username, lives) VALUES (?, ?)",
		s.Username, s.Lives,
	)
	if err != nil {
		return s, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return s, err
	}
	s.ID = id
	return s, nil
}

// UpdateIntern updates intern entry in database
func (m *MySQL) UpdateIntern(s model.Intern) (model.Intern, error) {
	_, err := m.conn.Exec(
		"UPDATE `interns` SET username=?, lives=? WHERE id=?",
		s.Username, s.Lives, s.ID,
	)
	if err != nil {
		return s, err
	}
	var i model.Intern
	err = m.conn.Get(&i, "SELECT * FROM `interns` WHERE id=?", s.ID)
	return i, err
}

// SelectIntern selects intern entry from database
func (m *MySQL) SelectIntern(id int64) (model.Intern, error) {
	var s model.Intern
	err := m.conn.Get(&s, "SELECT * FROM `interns` WHERE id=?", id)
	return s, err
}

// DeleteIntern deletes intern entry from database
func (m *MySQL) DeleteIntern(id int64) error {
	_, err := m.conn.Exec("DELETE FROM `interns` WHERE id=?", id)
	return err
}

// ListInterns returns array of intern entries from database
func (m *MySQL) ListInterns() ([]model.Intern, error) {
	items := []model.Intern{}
	err := m.conn.Select(&items, "SELECT * FROM `interns`")
	return items, err
}
