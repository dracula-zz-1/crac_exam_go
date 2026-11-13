package store

import (
	"database/sql"
	"fmt"
)

// ImportRefs 是题库替换前快照的用户引用数据。
// 导入时按题目的自然键 (J,P,I) 重新映射到新题库，保留收藏、错题与练习进度。
type ImportRefs struct {
	Favorites []FavoriteRef
	Errors    []ErrorRef
	Progress  []ProgressRef
}

type FavoriteRef struct {
	UserID    int64
	Category  string
	CreatedAt int64
	J, P, I   string
}

type ErrorRef struct {
	UserID     int64
	Category   string
	CreatedAt  int64
	WrongCount int
	J, P, I    string
}

type ProgressRef struct {
	UserID       int64
	Category     string
	CurrentIndex int
	LastAccessed int64
	J, P, I      string
}

func jpiKey(j, p, i string) string { return j + "\x1f" + p + "\x1f" + i }

// SnapshotImportRefs 在删除旧题库前，把收藏/错题/练习进度转换成 J/P/I 引用。
func SnapshotImportRefs(tx *sql.Tx) (*ImportRefs, error) {
	refs := &ImportRefs{}

	rows, err := tx.Query(`SELECT f.user_id, f.category, f.created_at, q.J, q.P, q.I
		FROM favorite_questions f JOIN questions q ON q.id = f.question_id`)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var r FavoriteRef
		if err := rows.Scan(&r.UserID, &r.Category, &r.CreatedAt, &r.J, &r.P, &r.I); err != nil {
			rows.Close()
			return nil, err
		}
		refs.Favorites = append(refs.Favorites, r)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()

	rows, err = tx.Query(`SELECT e.user_id, e.category, e.created_at, e.wrong_count, q.J, q.P, q.I
		FROM error_questions e JOIN questions q ON q.id = e.question_id`)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var r ErrorRef
		if err := rows.Scan(&r.UserID, &r.Category, &r.CreatedAt, &r.WrongCount, &r.J, &r.P, &r.I); err != nil {
			rows.Close()
			return nil, err
		}
		refs.Errors = append(refs.Errors, r)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()

	// 先读完进度行，再逐条查询旧题目，避免同一事务内嵌套查询。
	type progressRow struct {
		userID, index, last int64
		category            string
	}
	var progressRows []progressRow
	rows, err = tx.Query(`SELECT user_id, category, current_index, last_accessed FROM practice_progress`)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var p progressRow
		if err := rows.Scan(&p.userID, &p.category, &p.index, &p.last); err != nil {
			rows.Close()
			return nil, err
		}
		progressRows = append(progressRows, p)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()

	for _, p := range progressRows {
		col, err := categoryColumn(p.category)
		if err != nil {
			continue
		}
		var j, pp, i string
		err = tx.QueryRow(fmt.Sprintf(`SELECT J, P, I FROM questions WHERE %s = 1
			ORDER BY id LIMIT 1 OFFSET ?`, col), p.index).Scan(&j, &pp, &i)
		if err == sql.ErrNoRows {
			continue
		}
		if err != nil {
			return nil, err
		}
		refs.Progress = append(refs.Progress, ProgressRef{
			UserID: p.userID, Category: p.category, CurrentIndex: int(p.index),
			LastAccessed: p.last, J: j, P: pp, I: i,
		})
	}

	return refs, nil
}

// RestoreImportRefs 在新题库插入后恢复用户数据，返回实际恢复的条数。
func RestoreImportRefs(tx *sql.Tx, refs *ImportRefs) (favorites, errors, progress int, err error) {
	if refs == nil {
		return 0, 0, 0, nil
	}

	type qinfo struct {
		id         int64
		jpi        string
		la, lb, lc int
	}
	var list []qinfo
	rows, err := tx.Query(`SELECT id, J, P, I, LA, LB, LC FROM questions ORDER BY id`)
	if err != nil {
		return 0, 0, 0, err
	}
	for rows.Next() {
		var q qinfo
		var j, p, i string
		if err := rows.Scan(&q.id, &j, &p, &i, &q.la, &q.lb, &q.lc); err != nil {
			rows.Close()
			return 0, 0, 0, err
		}
		q.jpi = jpiKey(j, p, i)
		list = append(list, q)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return 0, 0, 0, err
	}
	rows.Close()

	jpiID := make(map[string]int64, len(list))
	catIndex := map[string]map[string]int{"A": {}, "B": {}, "C": {}}
	catPos := map[string]int{"A": 0, "B": 0, "C": 0}
	for _, q := range list {
		jpiID[q.jpi] = q.id
		flags := map[string]int{"A": q.la, "B": q.lb, "C": q.lc}
		for cat, flag := range flags {
			if flag != 0 {
				catIndex[cat][q.jpi] = catPos[cat]
				catPos[cat]++
			}
		}
	}

	for _, f := range refs.Favorites {
		id, ok := jpiID[jpiKey(f.J, f.P, f.I)]
		if !ok {
			continue
		}
		res, err := tx.Exec(`INSERT OR IGNORE INTO favorite_questions
			(user_id, question_id, category, created_at) VALUES (?, ?, ?, ?)`,
			f.UserID, id, f.Category, f.CreatedAt)
		if err != nil {
			return 0, 0, 0, err
		}
		if n, _ := res.RowsAffected(); n > 0 {
			favorites++
		}
	}

	for _, e := range refs.Errors {
		id, ok := jpiID[jpiKey(e.J, e.P, e.I)]
		if !ok {
			continue
		}
		wrongCount := e.WrongCount
		if wrongCount < 1 {
			wrongCount = 1
		}
		res, err := tx.Exec(`INSERT OR IGNORE INTO error_questions
			(user_id, question_id, category, created_at, wrong_count) VALUES (?, ?, ?, ?, ?)`,
			e.UserID, id, e.Category, e.CreatedAt, wrongCount)
		if err != nil {
			return 0, 0, 0, err
		}
		if n, _ := res.RowsAffected(); n > 0 {
			errors++
		}
	}

	for _, p := range refs.Progress {
		idx, ok := catIndex[p.Category][jpiKey(p.J, p.P, p.I)]
		if !ok {
			continue
		}
		_, err := tx.Exec(`INSERT INTO practice_progress
			(user_id, category, current_index, last_accessed) VALUES (?, ?, ?, ?)
			ON CONFLICT(user_id, category) DO UPDATE SET
				current_index = excluded.current_index,
				last_accessed = excluded.last_accessed`,
			p.UserID, p.Category, idx, p.LastAccessed)
		if err != nil {
			return 0, 0, 0, err
		}
		progress++
	}

	return favorites, errors, progress, nil
}
