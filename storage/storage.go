package storage

import (
	"time"

	// This line is must for working MySQL database
	_ "github.com/go-sql-driver/mysql"

	"github.com/jmoiron/sqlx"

	"github.com/maddevsio/telecomedian/config"
	"github.com/maddevsio/telecomedian/model"
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
func (m *MySQL) LastStandupFor(username string) (model.Standup, error) {
	var standup model.Standup
	err := m.conn.Get(&standup, "SELECT * FROM `standup` WHERE username=? ORDER BY id DESC LIMIT 1", username)
	return standup, err
}

// CreateLive creates live for user
func (m *MySQL) CreateLive(s model.Live) (model.Live, error) {
	res, err := m.conn.Exec(
		"INSERT INTO `lives` (username, lives) VALUES (?, ?)",
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

// UpdateLive updates standup entry in database
func (m *MySQL) UpdateLive(s model.Live) (model.Live, error) {
	_, err := m.conn.Exec(
		"UPDATE `lives` SET username=?, lives=? WHERE id=?",
		s.Username, s.Lives, s.ID,
	)
	if err != nil {
		return s, err
	}
	var i model.Live
	err = m.conn.Get(&i, "SELECT * FROM `lives` WHERE id=?", s.ID)
	return i, err
}

// SelectLive selects standup entry from database
func (m *MySQL) SelectLive(id int64) (model.Live, error) {
	var s model.Live
	err := m.conn.Get(&s, "SELECT * FROM `lives` WHERE id=?", id)
	return s, err
}

// DeleteLive deletes standup entry from database
func (m *MySQL) DeleteLive(id int64) error {
	_, err := m.conn.Exec("DELETE FROM `lives` WHERE id=?", id)
	return err
}

// ListLives returns array of standup entries from database
func (m *MySQL) ListLives() ([]model.Live, error) {
	items := []model.Live{}
	err := m.conn.Select(&items, "SELECT * FROM `lives`")
	return items, err
}
