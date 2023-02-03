package tdb

import (
	"fmt"
	"log"
	"strings"

	"github.com/jackc/pgx"
)

type Answer struct {
	Id       int    `db:"id"`
	EventTex string `db:"event"`
	First    string `db:"fir"`
	Second   string `db:"sec"`
}

func GetEvent(con *pgx.Conn, event string) (*Answer, error) {
	ans := &Answer{}
	points := strings.Split(event, ".")
	row := con.QueryRow("SELECT id, event, fir, sec FROM events WHERE event LIKE $1", fmt.Sprintf("%%%s%%", points[0]))
	err := row.Scan(&ans.Id, &ans.EventTex, &ans.First, &ans.Second)
	if err != nil {
		log.Println("EMPTY", err)
		return &Answer{}, err
	}

	return ans, nil
}
