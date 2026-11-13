package store

import (
	"database/sql"
	"fmt"
	"time"

	"crac_exam_v3/backend/models"
)

const (
	ExamInProgress = 0
	ExamSubmitted  = 1
)

// ExamRecord 考试记录（内部结构，时间使用毫秒时间戳）。
type ExamRecord struct {
	ID              int64
	Category        string
	ExamDate        time.Time
	StartedAt       time.Time
	SubmittedAt     time.Time
	DurationSeconds int
	UserID          int64
	Score           int
	TotalQuestions  int
	CorrectCount    int
	PassScore       int
	Status          int
}

func scanExamRecord(s rowScanner) (*ExamRecord, error) {
	var r ExamRecord
	var examDate, startedAt, submittedAt int64
	err := s.Scan(&r.ID, &r.Category, &examDate, &startedAt, &submittedAt,
		&r.DurationSeconds, &r.UserID, &r.Score, &r.TotalQuestions, &r.CorrectCount, &r.PassScore, &r.Status)
	if err != nil {
		return nil, err
	}
	r.ExamDate = toTime(examDate)
	r.StartedAt = toTime(startedAt)
	r.SubmittedAt = toTime(submittedAt)
	return &r, nil
}

const examRecordColumns = `id, category, exam_date, started_at, submitted_at, duration_seconds,
	user_id, score, total_questions, correct_count, pass_score, status`

func GetExamRecord(db *sql.DB, id int64) (*ExamRecord, error) {
	row := db.QueryRow(fmt.Sprintf(`SELECT %s FROM exam_records WHERE id = ?`, examRecordColumns), id)
	r, err := scanExamRecord(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return r, err
}

func GetExamRecordTx(tx *sql.Tx, id int64) (*ExamRecord, error) {
	row := tx.QueryRow(fmt.Sprintf(`SELECT %s FROM exam_records WHERE id = ?`, examRecordColumns), id)
	r, err := scanExamRecord(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return r, err
}

// CreateExam 在事务中创建考试记录与题目快照。
func CreateExam(db *sql.DB, category string, userID int64, total, passScore int, startedAt time.Time, details []models.ExamDetail) (int64, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	res, err := tx.Exec(`INSERT INTO exam_records
		(category, exam_date, started_at, user_id, total_questions, pass_score, status)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		category, startedAt.UnixMilli(), startedAt.UnixMilli(), userID, total, passScore, ExamInProgress)
	if err != nil {
		return 0, err
	}
	examID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	stmt, err := tx.Prepare(`INSERT INTO exam_question_details
		(exam_id, question_id, seq, question_text, option_a, option_b, option_c, option_d,
		 correct_answer, user_answer, is_correct, type, image_data)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, '', 0, ?, ?)`)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	for i := range details {
		d := &details[i]
		if _, err := stmt.Exec(examID, d.QuestionID, i+1, d.QuestionText,
			d.OptionA, d.OptionB, d.OptionC, d.OptionD, d.CorrectAnswer, d.Type, d.ImageData); err != nil {
			return 0, err
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return examID, nil
}

func GetExamDetails(db *sql.DB, examID int64) ([]models.ExamDetail, error) {
	rows, err := db.Query(`SELECT id, exam_id, question_id, seq, question_text,
		option_a, option_b, option_c, option_d, correct_answer, user_answer, is_correct, type, image_data
		FROM exam_question_details WHERE exam_id = ? ORDER BY seq`, examID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanDetails(rows)
}

func GetExamDetailsTx(tx *sql.Tx, examID int64) ([]models.ExamDetail, error) {
	rows, err := tx.Query(`SELECT id, exam_id, question_id, seq, question_text,
		option_a, option_b, option_c, option_d, correct_answer, user_answer, is_correct, type, image_data
		FROM exam_question_details WHERE exam_id = ? ORDER BY seq`, examID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanDetails(rows)
}

func scanDetails(rows *sql.Rows) ([]models.ExamDetail, error) {
	out := make([]models.ExamDetail, 0, 64)
	for rows.Next() {
		var d models.ExamDetail
		var isCorrect int
		if err := rows.Scan(&d.ID, &d.ExamID, &d.QuestionID, &d.Seq, &d.QuestionText,
			&d.OptionA, &d.OptionB, &d.OptionC, &d.OptionD, &d.CorrectAnswer, &d.UserAnswer,
			&isCorrect, &d.Type, &d.ImageData); err != nil {
			return nil, err
		}
		d.IsCorrect = isCorrect == 1
		out = append(out, d)
	}
	return out, rows.Err()
}

func UpdateDetailAnswer(tx *sql.Tx, detailID int64, answer string, isCorrect bool) error {
	_, err := tx.Exec(`UPDATE exam_question_details SET user_answer = ?, is_correct = ? WHERE id = ?`,
		answer, boolToInt(isCorrect), detailID)
	return err
}

func FinishExam(tx *sql.Tx, examID int64, score, correct, total, duration int, submittedAt time.Time) error {
	_, err := tx.Exec(`UPDATE exam_records SET score = ?, correct_count = ?, total_questions = ?,
		duration_seconds = ?, submitted_at = ?, status = ? WHERE id = ?`,
		score, correct, total, duration, submittedAt.UnixMilli(), ExamSubmitted, examID)
	return err
}

// RecordWrongTx 在考试交卷事务内累计错题次数。
func RecordWrongTx(tx *sql.Tx, userID, questionID int64, category string) error {
	_, err := tx.Exec(`INSERT INTO error_questions
		(user_id, question_id, category, created_at, wrong_count)
		VALUES (?, ?, ?, ?, 1)
		ON CONFLICT(user_id, question_id, category) DO UPDATE SET
			wrong_count = wrong_count + 1,
			created_at = excluded.created_at`,
		userID, questionID, category, nowMS())
	return err
}

// DeleteExam 删除考试记录（题目详情由外键级联删除）。
func DeleteExam(db *sql.DB, examID int64) error {
	_, err := db.Exec(`DELETE FROM exam_records WHERE id = ?`, examID)
	return err
}

func ListExamSummaries(db *sql.DB, userID int64, limit int) ([]models.ExamSummary, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := db.Query(`SELECT id, category, exam_date, score, total_questions, correct_count,
		pass_score, duration_seconds FROM exam_records
		WHERE user_id = ? AND status = ? ORDER BY exam_date DESC LIMIT ?`,
		userID, ExamSubmitted, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.ExamSummary, 0, limit)
	for rows.Next() {
		var s models.ExamSummary
		var examDate int64
		if err := rows.Scan(&s.ID, &s.Category, &examDate, &s.Score, &s.Total,
			&s.CorrectCount, &s.PassScore, &s.DurationSeconds); err != nil {
			return nil, err
		}
		s.ExamDate = toTime(examDate)
		s.Passed = s.Score >= s.PassScore
		out = append(out, s)
	}
	return out, rows.Err()
}

// ExamTotals 已提交考试的次数、通过次数与平均分。
func ExamTotals(db *sql.DB, userID int64) (total, passed int, avg float64, err error) {
	err = db.QueryRow(`SELECT COUNT(*),
		COALESCE(SUM(CASE WHEN score >= pass_score THEN 1 ELSE 0 END), 0),
		COALESCE(AVG(score), 0)
		FROM exam_records WHERE user_id = ? AND status = ?`, userID, ExamSubmitted).
		Scan(&total, &passed, &avg)
	return
}

// BestScores 每个类别的最佳成绩。
func BestScores(db *sql.DB, userID int64) (map[string]int, error) {
	rows, err := db.Query(`SELECT category, MAX(score) FROM exam_records
		WHERE user_id = ? AND status = ? GROUP BY category`, userID, ExamSubmitted)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]int{}
	for rows.Next() {
		var cat string
		var score int
		if err := rows.Scan(&cat, &score); err != nil {
			return nil, err
		}
		out[cat] = score
	}
	return out, rows.Err()
}

// ScorePoints 成绩趋势（按时间升序）。
func ScorePoints(db *sql.DB, userID int64, category string, since time.Time) ([]models.ScorePoint, error) {
	query := `SELECT category, exam_date, score, total_questions, pass_score FROM exam_records
		WHERE user_id = ? AND status = ?`
	args := []any{userID, ExamSubmitted}
	if category != "" {
		query += ` AND category = ?`
		args = append(args, category)
	}
	if !since.IsZero() {
		query += ` AND exam_date >= ?`
		args = append(args, since.UnixMilli())
	}
	query += ` ORDER BY exam_date ASC`

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.ScorePoint, 0, 32)
	for rows.Next() {
		var p models.ScorePoint
		var examDate int64
		var passScore int
		if err := rows.Scan(&p.Category, &examDate, &p.Score, &p.Total, &passScore); err != nil {
			return nil, err
		}
		p.Date = toTime(examDate)
		p.Passed = p.Score >= passScore
		out = append(out, p)
	}
	return out, rows.Err()
}

func CountExamRecords(db *sql.DB) (int, error) {
	var n int
	err := db.QueryRow(`SELECT COUNT(*) FROM exam_records`).Scan(&n)
	return n, err
}

// ExamTotalDuration 已提交考试的总时长（秒）。
func ExamTotalDuration(db *sql.DB, userID int64) (int64, error) {
	var n int64
	err := db.QueryRow(`SELECT COALESCE(SUM(duration_seconds), 0) FROM exam_records
		WHERE user_id = ? AND status = ?`, userID, ExamSubmitted).Scan(&n)
	return n, err
}
