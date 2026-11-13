package store

import (
	"database/sql"
	"encoding/json"
	"time"
)

// RandomCycle 随机练习的一轮队列。queue 中允许同一题目出现多次（错题加权），
// 但每个题目至少出现一次，保证一轮内覆盖完整题库。
type RandomCycle struct {
	CycleNo  int
	IDs      []int64
	Position int
}

func GetRandomCycle(db *sql.DB, userID int64, category string) (*RandomCycle, error) {
	var cycleNo, position int
	var queue string
	err := db.QueryRow(`SELECT cycle_no, queue, position FROM random_cycles
		WHERE user_id = ? AND category = ?`, userID, category).Scan(&cycleNo, &queue, &position)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var ids []int64
	if err := json.Unmarshal([]byte(queue), &ids); err != nil {
		return nil, err
	}
	return &RandomCycle{CycleNo: cycleNo, IDs: ids, Position: position}, nil
}

func SaveRandomCycle(db *sql.DB, userID int64, category string, cycle *RandomCycle) error {
	raw, err := json.Marshal(cycle.IDs)
	if err != nil {
		return err
	}
	_, err = db.Exec(`INSERT INTO random_cycles
		(user_id, category, cycle_no, queue, position, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(user_id, category) DO UPDATE SET
			cycle_no = excluded.cycle_no,
			queue = excluded.queue,
			position = excluded.position,
			updated_at = excluded.updated_at`,
		userID, category, cycle.CycleNo, string(raw), cycle.Position, time.Now().UnixMilli())
	return err
}

func DeleteRandomCycle(db *sql.DB, userID int64, category string) error {
	_, err := db.Exec(`DELETE FROM random_cycles WHERE user_id = ? AND category = ?`, userID, category)
	return err
}

func DeleteRandomCyclesByUser(db *sql.DB, userID int64) error {
	_, err := db.Exec(`DELETE FROM random_cycles WHERE user_id = ?`, userID)
	return err
}
