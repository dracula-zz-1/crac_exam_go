package services

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"crac_exam_v3/backend/models"
	"crac_exam_v3/backend/store"
)

func setupDB(t *testing.T) *sql.DB {
	t.Helper()
	t.Setenv("CRAC_EXAM_DATA_DIR", t.TempDir())
	db, err := store.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func seedQuestion(t *testing.T, db *sql.DB, j string, la, lb, lc, qtype int, answer string) int64 {
	t.Helper()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	q := models.Question{
		J: j, P: "P" + j, I: "I" + j, Q: "题目 " + j, T: answer,
		A: "甲", B: "乙", C: "丙", D: "丁",
		LA: la, LB: lb, LC: lc, Type: qtype,
	}
	if err := store.InsertQuestions(tx, []models.Question{q}); err != nil {
		t.Fatalf("insert question: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit: %v", err)
	}
	var id int64
	if err := db.QueryRow(`SELECT id FROM questions WHERE J = ?`, j).Scan(&id); err != nil {
		t.Fatalf("get id: %v", err)
	}
	return id
}

func seedExamBank(t *testing.T, db *sql.DB, category string, singles, multiples int) {
	t.Helper()
	for i := 0; i < singles; i++ {
		la, lb, lc := flagsFor(category)
		seedQuestion(t, db, fmt.Sprintf("%sS%03d", category, i), la, lb, lc, 1, "A")
	}
	for i := 0; i < multiples; i++ {
		la, lb, lc := flagsFor(category)
		seedQuestion(t, db, fmt.Sprintf("%sM%03d", category, i), la, lb, lc, 2, "AC")
	}
}

func flagsFor(category string) (int, int, int) {
	switch category {
	case "A":
		return 1, 0, 0
	case "B":
		return 0, 1, 0
	default:
		return 0, 0, 1
	}
}

func TestLoginCreatesProfileAndIsStable(t *testing.T) {
	db := setupDB(t)
	svc := NewUserService(db)

	u1, err := svc.Login("测试用户", "110101199001011234")
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	u2, err := svc.Login("测试用户", "110101199001011234")
	if err != nil {
		t.Fatalf("login again: %v", err)
	}
	if u1.ID != u2.ID {
		t.Fatalf("same id card should map to same user: %d vs %d", u1.ID, u2.ID)
	}
	if _, err := svc.Login("", "123"); err == nil {
		t.Fatalf("empty username should fail")
	}
}

func TestCreateExamSubmitAndStatistics(t *testing.T) {
	db := setupDB(t)
	user := login(t, db)
	seedExamBank(t, db, "A", 32, 8)
	svc := NewExamService(db)

	start, err := svc.CreateExam(user.ID, "A")
	if err != nil {
		t.Fatalf("create exam: %v", err)
	}
	if start.Total != 40 || len(start.Questions) != 40 {
		t.Fatalf("unexpected exam size: total=%d questions=%d", start.Total, len(start.Questions))
	}
	if start.TimeLimitSeconds != 40*60 || start.PassScore != 30 {
		t.Fatalf("unexpected config: %+v", start)
	}

	// 服务端计时：把开始时间往前拨 5 秒
	if _, err := db.Exec(`UPDATE exam_records SET started_at = started_at - 5000 WHERE id = ?`, start.ExamID); err != nil {
		t.Fatalf("shift start: %v", err)
	}

	details, err := store.GetExamDetails(db, start.ExamID)
	if err != nil {
		t.Fatalf("get details: %v", err)
	}
	answers := map[string]string{}
	for _, d := range details {
		answers[fmt.Sprintf("%d", d.QuestionID)] = d.CorrectAnswer
	}

	result, err := svc.SubmitExam(start.ExamID, answers)
	if err != nil {
		t.Fatalf("submit: %v", err)
	}
	if result.Score != 40 || !result.Passed {
		t.Fatalf("expected full score and pass, got %+v", result)
	}
	if result.DurationSeconds < 5 {
		t.Fatalf("duration should be computed server-side (>=5s), got %d", result.DurationSeconds)
	}

	// 幂等：再次提交返回同样结果
	again, err := svc.SubmitExam(start.ExamID, nil)
	if err != nil {
		t.Fatalf("resubmit: %v", err)
	}
	if again.Score != result.Score {
		t.Fatalf("resubmit changed score: %d -> %d", result.Score, again.Score)
	}

	dash, err := NewStatisticsService(db).GetDashboard(user.ID)
	if err != nil {
		t.Fatalf("dashboard: %v", err)
	}
	if dash.TotalExams != 1 || dash.PassedExams != 1 || dash.PassRate != 100 {
		t.Fatalf("unexpected dashboard: %+v", dash)
	}
	if dash.BestScores["A"] != 40 {
		t.Fatalf("unexpected best score: %+v", dash.BestScores)
	}
	if dash.ErrorCount != 0 {
		t.Fatalf("all-correct exam should not add errors, got %d", dash.ErrorCount)
	}
}

func TestSubmitWrongAnswersAddsErrorsAndIsIdempotent(t *testing.T) {
	db := setupDB(t)
	user := login(t, db)
	seedExamBank(t, db, "A", 32, 8)
	svc := NewExamService(db)

	start, err := svc.CreateExam(user.ID, "A")
	if err != nil {
		t.Fatalf("create exam: %v", err)
	}
	details, _ := store.GetExamDetails(db, start.ExamID)
	answers := map[string]string{}
	for _, d := range details {
		if d.CorrectAnswer != "A" {
			answers[fmt.Sprintf("%d", d.QuestionID)] = "A"
		} else {
			answers[fmt.Sprintf("%d", d.QuestionID)] = "B"
		}
	}
	result, err := svc.SubmitExam(start.ExamID, answers)
	if err != nil {
		t.Fatalf("submit: %v", err)
	}
	if result.Score != 0 || result.Passed {
		t.Fatalf("expected 0 score, got %+v", result)
	}
	errors, _ := store.CountErrors(db, user.ID)
	if errors != 40 {
		t.Fatalf("expected 40 error questions, got %d", errors)
	}
	// 重复提交不应产生重复错题
	if _, err := svc.SubmitExam(start.ExamID, answers); err != nil {
		t.Fatalf("resubmit: %v", err)
	}
	errors, _ = store.CountErrors(db, user.ID)
	if errors != 40 {
		t.Fatalf("resubmit duplicated errors: %d", errors)
	}
}

func TestInsufficientBankRefusesExam(t *testing.T) {
	db := setupDB(t)
	user := login(t, db)
	seedExamBank(t, db, "A", 5, 2)
	svc := NewExamService(db)
	_, err := svc.CreateExam(user.ID, "A")
	if err == nil || !strings.Contains(err.Error(), "不足") {
		t.Fatalf("expected insufficient bank error, got %v", err)
	}
}

func TestAbandonedExamCleanup(t *testing.T) {
	db := setupDB(t)
	user := login(t, db)
	seedExamBank(t, db, "A", 32, 8)
	svc := NewExamService(db)

	if _, err := svc.CreateExam(user.ID, "A"); err != nil {
		t.Fatalf("create exam: %v", err)
	}
	if _, err := svc.CreateExam(user.ID, "A"); err != nil {
		t.Fatalf("create exam 2: %v", err)
	}
	store.CleanupAbandonedExams(db)

	total, _, _, err := store.ExamTotals(db, user.ID)
	if err != nil {
		t.Fatalf("totals: %v", err)
	}
	if total != 0 {
		t.Fatalf("abandoned exams should be removed, got %d", total)
	}
	var details int
	if err := db.QueryRow(`SELECT COUNT(*) FROM exam_question_details`).Scan(&details); err != nil {
		t.Fatalf("count details: %v", err)
	}
	if details != 0 {
		t.Fatalf("exam details should cascade-delete, got %d", details)
	}
}

func TestPracticeFavoritesErrorsAndProgress(t *testing.T) {
	db := setupDB(t)
	user := login(t, db)
	id1 := seedQuestion(t, db, "LY0001", 1, 0, 0, 1, "A")
	id2 := seedQuestion(t, db, "LY0002", 1, 0, 0, 1, "B")
	svc := NewPracticeService(db)

	data, err := svc.GetPracticeData(user.ID, "A", "practice")
	if err != nil {
		t.Fatalf("practice data: %v", err)
	}
	if len(data.Questions) != 2 || data.Total != 2 {
		t.Fatalf("unexpected practice data: %+v", data)
	}

	if err := svc.SaveProgress(user.ID, "A", 1); err != nil {
		t.Fatalf("save progress: %v", err)
	}
	if err := svc.SaveProgress(user.ID, "A", 1); err != nil {
		t.Fatalf("save progress 2: %v", err)
	}
	data, _ = svc.GetPracticeData(user.ID, "A", "practice")
	if data.Progress != 1 {
		t.Fatalf("progress = %d, want 1", data.Progress)
	}
	var progressRows int
	db.QueryRow(`SELECT COUNT(*) FROM practice_progress WHERE user_id = ?`, user.ID).Scan(&progressRows)
	if progressRows != 1 {
		t.Fatalf("progress should upsert, got %d rows", progressRows)
	}

	if ok, err := svc.AddFavorite(user.ID, id1, "A"); err != nil || !ok {
		t.Fatalf("add favorite: %v %v", ok, err)
	}
	if ok, err := svc.AddFavorite(user.ID, id1, "A"); err != nil || ok {
		t.Fatalf("duplicate favorite should be ignored: %v %v", ok, err)
	}
	favIDs, _ := svc.GetFavoriteIDs(user.ID)
	if len(favIDs) != 1 {
		t.Fatalf("favorite ids = %v", favIDs)
	}
	favQuestions, _ := svc.GetFavoriteQuestions(user.ID, "A")
	if len(favQuestions) != 1 {
		t.Fatalf("favorite questions = %d", len(favQuestions))
	}

	if ok, err := svc.AddError(user.ID, id2, "A"); err != nil || !ok {
		t.Fatalf("add error: %v %v", ok, err)
	}
	if ok, err := svc.AddError(user.ID, id2, "A"); err != nil || ok {
		t.Fatalf("duplicate error should be ignored: %v %v", ok, err)
	}
	errs, _ := store.CountErrors(db, user.ID)
	if errs != 1 {
		t.Fatalf("error count = %d", errs)
	}
	errorQuestions, _ := svc.GetErrorQuestions(user.ID, "A")
	if len(errorQuestions) != 1 || errorQuestions[0].ID != id2 {
		t.Fatalf("unexpected error questions: %+v", errorQuestions)
	}

	if err := svc.RemoveError(user.ID, id2); err != nil {
		t.Fatalf("remove error: %v", err)
	}
	errs, _ = store.CountErrors(db, user.ID)
	if errs != 0 {
		t.Fatalf("error count after remove = %d", errs)
	}

	if err := svc.RemoveFavorite(user.ID, id1); err != nil {
		t.Fatalf("remove favorite: %v", err)
	}
	favIDs, _ = svc.GetFavoriteIDs(user.ID)
	if len(favIDs) != 0 {
		t.Fatalf("favorite ids after remove = %v", favIDs)
	}
}

func TestRandomPracticeCoverageAndWeights(t *testing.T) {
	db := setupDB(t)
	user := login(t, db)
	var ids []int64
	for i := 0; i < 10; i++ {
		ids = append(ids, seedQuestion(t, db, fmt.Sprintf("R%02d", i), 1, 0, 0, 1, "A"))
	}
	psvc := NewPracticeService(db)

	// 第 1 题错 3 次，第 2 题错 1 次
	for i := 0; i < 3; i++ {
		if _, err := psvc.AddError(user.ID, ids[0], "A"); err != nil {
			t.Fatalf("add error: %v", err)
		}
	}
	if _, err := psvc.AddError(user.ID, ids[1], "A"); err != nil {
		t.Fatalf("add error: %v", err)
	}

	data, err := psvc.GetPracticeData(user.ID, "A", "random")
	if err != nil {
		t.Fatalf("random data: %v", err)
	}
	counts := map[int64]int{}
	for _, q := range data.Questions {
		counts[q.ID]++
	}
	if len(counts) != 10 {
		t.Fatalf("cycle should cover all 10 questions, got %d", len(counts))
	}
	for _, id := range ids {
		if counts[id] < 1 {
			t.Fatalf("question %d missing from random cycle", id)
		}
	}
	if counts[ids[0]] != 4 { // 基础 1 + 3 次错误
		t.Fatalf("wrong question 1 count = %d, want 4", counts[ids[0]])
	}
	if counts[ids[1]] != 2 {
		t.Fatalf("wrong question 2 count = %d, want 2", counts[ids[1]])
	}
	if len(data.Questions) != 14 {
		t.Fatalf("cycle length = %d, want 14", len(data.Questions))
	}
	if data.Progress != 0 {
		t.Fatalf("initial progress = %d", data.Progress)
	}

	// 位置持久化：同一轮顺序不变
	firstOrder := make([]int64, len(data.Questions))
	for i, q := range data.Questions {
		firstOrder[i] = q.ID
	}
	if err := psvc.SaveRandomProgress(user.ID, "A", 5); err != nil {
		t.Fatalf("save random progress: %v", err)
	}
	again, _ := psvc.GetPracticeData(user.ID, "A", "random")
	if again.Progress != 5 {
		t.Fatalf("progress = %d, want 5", again.Progress)
	}
	for i, q := range again.Questions {
		if q.ID != firstOrder[i] {
			t.Fatalf("cycle order changed at %d", i)
		}
	}

	// 完成本轮后应生成新的一轮，并继续保证全覆盖
	if err := psvc.SaveRandomProgress(user.ID, "A", len(again.Questions)); err != nil {
		t.Fatalf("complete cycle: %v", err)
	}
	next, err := psvc.GetPracticeData(user.ID, "A", "random")
	if err != nil {
		t.Fatalf("next cycle: %v", err)
	}
	nextCounts := map[int64]int{}
	for _, q := range next.Questions {
		nextCounts[q.ID]++
	}
	if len(nextCounts) != 10 || len(next.Questions) != 14 {
		t.Fatalf("next cycle coverage/length = %d/%d", len(nextCounts), len(next.Questions))
	}
}

func TestWrongCountIncrements(t *testing.T) {
	db := setupDB(t)
	user := login(t, db)
	id := seedQuestion(t, db, "WC1", 1, 0, 0, 1, "A")
	psvc := NewPracticeService(db)
	psvc.AddError(user.ID, id, "A")
	psvc.AddError(user.ID, id, "A")
	psvc.AddError(user.ID, id, "A")

	var n int
	if err := db.QueryRow(`SELECT wrong_count FROM error_questions WHERE user_id = ? AND question_id = ? AND category = 'A'`,
		user.ID, id).Scan(&n); err != nil {
		t.Fatalf("query wrong_count: %v", err)
	}
	if n != 3 {
		t.Fatalf("wrong_count = %d, want 3", n)
	}
	list, _ := psvc.GetErrorQuestions(user.ID, "A")
	if len(list) != 1 || list[0].WrongCount != 3 {
		t.Fatalf("error question wrong count not exposed: %+v", list)
	}
}

func TestQuestionPageFilterUsesORSemantics(t *testing.T) {
	db := setupDB(t)
	seedQuestion(t, db, "AONLY", 1, 0, 0, 1, "A")
	seedQuestion(t, db, "BONLY", 0, 1, 0, 1, "A")
	seedQuestion(t, db, "AB", 1, 1, 0, 1, "A")

	page, err := store.PageQuestions(db, 1, 20, "", true, true, false)
	if err != nil {
		t.Fatalf("page: %v", err)
	}
	if page.Total != 3 || len(page.Data) != 3 {
		t.Fatalf("A+B should return 3 via OR, got total=%d data=%d", page.Total, len(page.Data))
	}

	page, err = store.PageQuestions(db, 1, 2, "", true, false, false)
	if err != nil {
		t.Fatalf("page A: %v", err)
	}
	if page.Total != 2 || len(page.Data) != 2 || page.TotalPages != 1 {
		t.Fatalf("unexpected A page: %+v", page)
	}

	page, _ = store.PageQuestions(db, 1, 20, "BONLY", false, false, false)
	if page.Total != 1 {
		t.Fatalf("search should match J, got %d", page.Total)
	}
}

func TestUpdateQuestionValidationAndDeleteCleanup(t *testing.T) {
	db := setupDB(t)
	user := login(t, db)
	id := seedQuestion(t, db, "LY0003", 1, 0, 0, 1, "A")
	qsvc := NewQuestionService(db)

	q, _ := qsvc.GetQuestion(id)
	q.T = "dca"
	q.Q = "更新后的题干"
	if err := qsvc.UpdateQuestion(*q); err != nil {
		t.Fatalf("update: %v", err)
	}
	q2, _ := qsvc.GetQuestion(id)
	if q2.T != "ACD" || q2.Type != 2 {
		t.Fatalf("answer should be normalized: %+v", q2)
	}

	psvc := NewPracticeService(db)
	psvc.AddFavorite(user.ID, id, "A")
	psvc.AddError(user.ID, id, "A")
	if err := qsvc.DeleteQuestion(id); err != nil {
		t.Fatalf("delete: %v", err)
	}
	q3, _ := qsvc.GetQuestion(id)
	if q3 != nil {
		t.Fatalf("question should be deleted")
	}
	if n, _ := store.CountErrors(db, user.ID); n != 0 {
		t.Fatalf("errors should be cleaned, got %d", n)
	}
	if n, _ := store.CountFavorites(db, user.ID); n != 0 {
		t.Fatalf("favorites should be cleaned, got %d", n)
	}

	bad := models.Question{ID: id, J: "X", Q: "Y", T: "Z"}
	if err := qsvc.UpdateQuestion(bad); err == nil {
		t.Fatalf("invalid answer should be rejected")
	}
}

func TestUpdateQuestionRejectsDuplicateJPI(t *testing.T) {
	db := setupDB(t)
	id1 := seedQuestion(t, db, "LY1000", 1, 0, 0, 1, "A")
	id2 := seedQuestion(t, db, "LY1001", 1, 0, 0, 1, "A")
	q1, _ := store.GetQuestion(db, id1)
	q2, _ := store.GetQuestion(db, id2)
	q2.J, q2.P, q2.I = q1.J, q1.P, q1.I

	err := NewQuestionService(db).UpdateQuestion(*q2)
	if err == nil || !strings.Contains(err.Error(), "联合") {
		t.Fatalf("expected duplicate J/P/I error, got %v", err)
	}
}

func TestClearQuestionBankKeepsSubmittedHistory(t *testing.T) {
	db := setupDB(t)
	user := login(t, db)
	seedExamBank(t, db, "A", 32, 8)
	examSvc := NewExamService(db)
	start, err := examSvc.CreateExam(user.ID, "A")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	details, _ := store.GetExamDetails(db, start.ExamID)
	answers := map[string]string{}
	for _, d := range details {
		answers[fmt.Sprintf("%d", d.QuestionID)] = d.CorrectAnswer
	}
	if _, err := examSvc.SubmitExam(start.ExamID, answers); err != nil {
		t.Fatalf("submit: %v", err)
	}

	if err := NewSettingsService(db).ClearQuestionBank(); err != nil {
		t.Fatalf("clear bank: %v", err)
	}
	if n, _ := store.CountQuestions(db); n != 0 {
		t.Fatalf("questions should be cleared, got %d", n)
	}
	if n, _ := store.CountExamRecords(db); n != 1 {
		t.Fatalf("submitted history should be kept, got %d", n)
	}
	// 历史快照仍可复盘
	review, err := examSvc.GetDetails(start.ExamID)
	if err != nil || len(review) != 40 {
		t.Fatalf("history review should remain, got %d, err=%v", len(review), err)
	}
}

func TestClearUserDataTransAction(t *testing.T) {
	db := setupDB(t)
	user := login(t, db)
	seedExamBank(t, db, "A", 32, 8)
	examSvc := NewExamService(db)
	start, err := examSvc.CreateExam(user.ID, "A")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	details, _ := store.GetExamDetails(db, start.ExamID)
	answers := map[string]string{}
	for _, d := range details {
		answers[fmt.Sprintf("%d", d.QuestionID)] = "B"
	}
	examSvc.SubmitExam(start.ExamID, answers)
	psvc := NewPracticeService(db)
	psvc.AddFavorite(user.ID, details[0].QuestionID, "A")
	psvc.AddError(user.ID, details[0].QuestionID, "A")
	psvc.SaveProgress(user.ID, "A", 3)

	if err := NewSettingsService(db).ClearUserData(user.ID); err != nil {
		t.Fatalf("clear user: %v", err)
	}
	if n, _ := store.CountExamRecords(db); n != 0 {
		t.Fatalf("exam records should be cleared, got %d", n)
	}
	if n, _ := store.CountErrors(db, user.ID); n != 0 {
		t.Fatalf("errors should be cleared, got %d", n)
	}
	if n, _ := store.CountFavorites(db, user.ID); n != 0 {
		t.Fatalf("favorites should be cleared, got %d", n)
	}
	if n, _ := store.CountQuestions(db); n != 40 {
		t.Fatalf("question bank should be untouched, got %d", n)
	}
}

func TestImportReplaceSemantics(t *testing.T) {
	db := setupDB(t)
	user := login(t, db)
	id := seedQuestion(t, db, "OLD", 1, 0, 0, 1, "A")
	psvc := NewPracticeService(db)
	psvc.AddFavorite(user.ID, id, "A")
	psvc.AddError(user.ID, id, "A")
	psvc.SaveProgress(user.ID, "A", 1)

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	if err := store.RemoveAllQuestionsForImport(tx); err != nil {
		t.Fatalf("clear: %v", err)
	}
	newQs := []models.Question{
		{J: "N1", P: "P", I: "I1", Q: "新题1", T: "A", A: "a", B: "b", C: "c", D: "d", LA: 1, Type: 1},
		{J: "N2", P: "P", I: "I2", Q: "新题2", T: "AB", A: "a", B: "b", C: "c", D: "d", LA: 1, Type: 2},
	}
	if err := store.InsertQuestions(tx, newQs); err != nil {
		t.Fatalf("insert: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit: %v", err)
	}

	if n, _ := store.CountQuestions(db); n != 2 {
		t.Fatalf("questions = %d", n)
	}
	if n, _ := store.CountErrors(db, user.ID); n != 0 {
		t.Fatalf("errors should be cleared on import, got %d", n)
	}
	if n, _ := store.CountFavorites(db, user.ID); n != 0 {
		t.Fatalf("favorites should be cleared on import, got %d", n)
	}
	var progress int
	db.QueryRow(`SELECT COUNT(*) FROM practice_progress`).Scan(&progress)
	if progress != 0 {
		t.Fatalf("progress should be cleared on import, got %d", progress)
	}
}

func TestShuffleOptionsKeepsCorrectMapping(t *testing.T) {
	cases := []models.Question{
		{A: "甲", B: "乙", C: "丙", D: "丁", T: "A", Type: 1},
		{A: "one", B: "two", C: "three", D: "four", T: "C", Type: 1},
		{A: "甲", B: "乙", C: "丙", D: "丁", T: "AC", Type: 2},
		{A: "甲", B: "乙", C: "丙", D: "丁", T: "ABCD", Type: 2},
	}
	for _, q := range cases {
		origCorrect := map[string]bool{}
		for idx, ch := range "ABCD" {
			if strings.ContainsRune(q.T, ch) {
				orig := []string{q.A, q.B, q.C, q.D}[idx]
				origCorrect[orig] = true
			}
		}
		for i := 0; i < 300; i++ {
			opts, ans := shuffleQuestionOptions(q)
			if len(ans) != len(q.T) {
				t.Fatalf("answer length changed: %q -> %q", q.T, ans)
			}
			if ans != sortAnswer(ans) {
				t.Fatalf("answer not sorted: %q", ans)
			}
			seenOpts := map[string]bool{}
			for _, v := range opts {
				if seenOpts[v] {
					t.Fatalf("option duplicated after shuffle: %v", opts)
				}
				seenOpts[v] = true
			}
			gotCorrect := map[string]bool{}
			for idx, ch := range "ABCD" {
				if strings.ContainsRune(ans, ch) {
					gotCorrect[opts[idx]] = true
				}
			}
			if len(gotCorrect) != len(origCorrect) {
				t.Fatalf("correct option count changed: %v -> %v", origCorrect, gotCorrect)
			}
			for text := range origCorrect {
				if !gotCorrect[text] {
					t.Fatalf("correct option %q lost after shuffle: opts=%v ans=%q", text, opts, ans)
				}
			}
		}
	}
}

func sortAnswer(s string) string {
	b := []byte(s)
	for i := 0; i < len(b); i++ {
		for j := i + 1; j < len(b); j++ {
			if b[j] < b[i] {
				b[i], b[j] = b[j], b[i]
			}
		}
	}
	return string(b)
}

func TestNormalizeAnswerLetters(t *testing.T) {
	cases := map[string]string{
		"a":       "A",
		" d,b a ": "ABD",
		"DDCA":    "ACD",
		"Z":       "",
		"":        "",
	}
	for in, want := range cases {
		if got := normalizeAnswerLetters(in); got != want {
			t.Fatalf("normalizeAnswerLetters(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestImportRollbackKeepsOldBank(t *testing.T) {
	db := setupDB(t)
	login(t, db)
	seedQuestion(t, db, "KEEP", 1, 0, 0, 1, "A")

	// 让题目插入必然失败，验证导入事务回滚后旧题库仍在
	if _, err := db.Exec(`CREATE TRIGGER fail_questions BEFORE INSERT ON questions
		BEGIN SELECT RAISE(ABORT, 'boom'); END;`); err != nil {
		t.Fatalf("create trigger: %v", err)
	}
	zipPath := makeTestBankZip(t, t.TempDir(), 2)
	if _, err := NewImportService(db).ImportZip(zipPath); err == nil {
		t.Fatalf("import should fail because of trigger")
	}
	count, _ := store.CountQuestions(db)
	if count != 1 {
		t.Fatalf("old bank should survive rollback, got %d questions", count)
	}
	q, _ := store.GetQuestion(db, 1)
	if q == nil || q.J != "KEEP" {
		t.Fatalf("old question missing after rollback: %+v", q)
	}
}

// TestRealBankEndToEndIfProvided 用真实题库跑完整学习/考试流程。
// 设置 CRAC_TEST_ZIP 后启用。
func TestRealBankEndToEndIfProvided(t *testing.T) {
	zipPath := os.Getenv("CRAC_TEST_ZIP")
	if zipPath == "" {
		t.Skip("set CRAC_TEST_ZIP to run real bank end-to-end")
	}
	db := setupDB(t)
	user := login(t, db)

	result, err := NewImportService(db).ImportZip(zipPath)
	if err != nil {
		t.Fatalf("import: %v", err)
	}
	if result.Imported != 1375 || result.DuplicateJPI != 0 {
		t.Fatalf("unexpected import result: imported=%d duplicateJPI=%d",
			result.Imported, result.DuplicateJPI)
	}

	dash, err := NewStatisticsService(db).GetDashboard(user.ID)
	if err != nil {
		t.Fatalf("dashboard: %v", err)
	}
	if dash.TotalQuestions != 1375 || dash.WithImages != 53 {
		t.Fatalf("unexpected dashboard: total=%d images=%d", dash.TotalQuestions, dash.WithImages)
	}

	practice, err := NewPracticeService(db).GetPracticeData(user.ID, "A", "practice")
	if err != nil {
		t.Fatalf("practice: %v", err)
	}
	if len(practice.Questions) != 683 {
		t.Fatalf("practice A questions = %d, want 683", len(practice.Questions))
	}

	wantSizes := map[string]int{"A": 40, "B": 60, "C": 90}
	for _, cat := range []string{"A", "B", "C"} {
		start, err := NewExamService(db).CreateExam(user.ID, cat)
		if err != nil {
			t.Fatalf("create %s exam: %v", cat, err)
		}
		if start.Total != wantSizes[cat] {
			t.Fatalf("%s exam total = %d, want %d", cat, start.Total, wantSizes[cat])
		}
		details, err := store.GetExamDetails(db, start.ExamID)
		if err != nil {
			t.Fatalf("details: %v", err)
		}
		answers := map[string]string{}
		for _, d := range details {
			answers[fmt.Sprintf("%d", d.QuestionID)] = d.CorrectAnswer
		}
		res, err := NewExamService(db).SubmitExam(start.ExamID, answers)
		if err != nil {
			t.Fatalf("submit %s exam: %v", cat, err)
		}
		if !res.Passed || res.Score != res.TotalCount {
			t.Fatalf("%s exam not full pass: %+v", cat, res)
		}
	}
}

func login(t *testing.T, db *sql.DB) *models.User {
	t.Helper()
	u, err := NewUserService(db).Login("测试用户", "ID-0001")
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	return u
}
