package db

import (
	"fmt"
	"strings"
	"time"

	dlog "github.com/amoghe/distillog"
	sql "github.com/lazada/sqle"
	_ "github.com/mattn/go-sqlite3"
)

// Groupchat ...
type Groupchat struct {
	ID             int64     `sql:"id"`
	Title 		   string    `sql:"title"`
	TelegramID 	   int64     `sql:"telegram_id"`
	State 		   string    `sql:"state"`
	InviteLink	   string    `sql:"invite_link"`
	CreatedAt      time.Time `sql:"created_at"`
}

func (gc *Groupchat) String() string {
    return fmt.Sprintf("%s [id %d] %s", gc.Title, gc.TelegramID, gc.InviteLink)
}

// GetGroupchats ...
func GetGroupchats(db *sql.DB, states []string) (groupchats []*Groupchat, err error) {
	if len(states) == 0 {
		states = []string{"active"}
	}

	args := make([]interface{}, len(states))
	for i, state := range states {
	    args[i] = state
	}

	var returnModel Groupchat
	sql := `select
	*
FROM
	groupchats
WHERE
	state IN (?` + strings.Repeat(",?", len(args)-1) + `)
ORDER BY
	state, title;`

	result, err := QuerySQLList(db, returnModel, sql, args...)
	if err != nil {
		return groupchats, err
	}

	for _, item := range result {
		if returnModel, ok := item.Interface().(*Groupchat); ok {
			groupchats = append(groupchats, returnModel)
		}
	}

	return groupchats, err
}

// AddGroupChatIfNotExist ...
func AddGroupChatIfNotExist(db *sql.DB, groupchat *Groupchat) (*Groupchat, error) {
	var returnModel Groupchat

	result, err := QuerySQLObject(db, returnModel, `SELECT * FROM groupchats WHERE telegram_id = ?;`, groupchat.TelegramID)
	if err != nil {
		return nil, err
	}

	if returnModel, ok := result.Interface().(*Groupchat); ok && returnModel.State != "" {
		return returnModel, fmt.Errorf(GroupChatAlreadyExists)
	}

	res, err := db.Exec(
		"INSERT INTO groupchats (title, telegram_id, invite_link, state) VALUES (?, ?, ?, ?);",
		groupchat.Title,
		groupchat.TelegramID,
		groupchat.InviteLink,
		groupchat.State,
	)

	if err != nil {
		return nil, err
	}

	groupchat.ID, _ = res.LastInsertId()
	groupchat.CreatedAt = time.Now()

	dlog.Debugf("%s (%d) added at %s\n", groupchat.Title, groupchat.ID, groupchat.CreatedAt)

	return groupchat, nil
}

// UpdateGroupChatInviteLink ...
func UpdateGroupChatInviteLink(db *sql.DB, groupchat *Groupchat) (int64, error) {
	result, err := db.Exec(
		"UPDATE groupchats SET invite_link = ? WHERE telegram_id = ?;",
		groupchat.InviteLink,
		groupchat.TelegramID)

	if err != nil {
		return -1, err
	}

	rows, err := result.RowsAffected()

	if err != nil {
		return -1, err
	}

	return rows, nil
}

// UpdateGroupChatTitle ...
func UpdateGroupChatTitle(db *sql.DB, groupchat *Groupchat) (int64, error) {
	result, err := db.Exec(
		"UPDATE groupchats SET title = ? WHERE telegram_id = ?;",
		groupchat.Title,
		groupchat.TelegramID)

	if err != nil {
		return -1, err
	}

	rows, err := result.RowsAffected()

	if err != nil {
		return -1, err
	}

	return rows, nil
}