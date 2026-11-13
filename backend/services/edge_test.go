package services

import (
	"fmt"
	"strings"
	"sync"
	"testing"

	"crac_exam_v3/backend/models"
	"crac_exam_v3/backend/store"
)

func TestServiceInputValidation(t *testing.T) {
	db := setupDB(t)
	user := login(t, db)
	psvc := NewPracticeService(db)
	esvc := NewExamService(db)
	qsvc := NewQuestionService(db)

	if _, err := psvc.GetPracticeData(user.ID, "Z", "practice"); err == nil {
		t.Fatal("invalid category should fail")
	}
	if _, err := psvc.GetPracticeData(user.ID, "A", "bogus"); err == nil {
		t.Fatal("invalid mode should fail")
	}
	if err := psvc.SaveProgress(user.ID, "Z", 0); err == nil {
		t.Fatal("invalid progress category should fail")
	}
	if _, err := esvc.CreateExam(user.ID, "Z"); err == nil {
		t.Fatal("invalid exam category should fail")
	}
	if _, err := esvc.SubmitExam(-1, nil); err == nil {
		t.Fatal("invalid exam id should fail")
	}
	if _, err := esvc.GetResult(99999); err == nil {
		t.Fatal("missing exam result should fail")
	}
	if err := qsvc.DeleteQuestion(99999); err == nil {
		t.Fatal("deleting missing question should fail")
	}
	bad := models.Question{ID: 1, J: "", Q: "", T: "Z"}
	if err := qsvc.UpdateQuestion(bad); err == nil {
		t.Fatal("invalid question should fail")
	}
	if _, err := NewUserService(db).Login("", "x"); err == nil {
		t.Fatal("empty username should fail")
	}
}

func TestSQLInjectionStringsAreTreatedAsData(t *testing.T) {
	db := setupDB(t)
	id := seedQuestion(t, db, "SAFE", 1, 0, 0, 1, "A")

	page, err := store.PageQuestions(db, 1, 10, "' OR 1=1 --", false, false, false)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if page.Total != 0 {
		t.Fatalf("SQL injection-like search should match nothing, got %d", page.Total)
	}

	q, err := store.GetQuestion(db, id)
	if err != nil || q == nil {
		t.Fatalf("get question: %v", err)
	}
	q.Q = "x'); DROP TABLE questions; --"
	if err := NewQuestionService(db).UpdateQuestion(*q); err != nil {
		t.Fatalf("update with SQL-like text: %v", err)
	}
	count, err := store.CountQuestions(db)
	if err != nil || count != 1 {
		t.Fatalf("questions table should be intact, count=%d err=%v", count, err)
	}
	page, err = store.PageQuestions(db, 1, 10, "DROP TABLE", false, false, false)
	if err != nil {
		t.Fatalf("literal search: %v", err)
	}
	if page.Total != 1 {
		t.Fatalf("literal search should find the updated question, got %d", page.Total)
	}
}

func TestConcurrentUserWrites(t *testing.T) {
	db := setupDB(t)
	user := login(t, db)
	var ids []int64
	for i := 0; i < 10; i++ {
		ids = append(ids, seedQuestion(t, db, fmt.Sprintf("C%02d", i), 1, 0, 0, 1, "A"))
	}
	psvc := NewPracticeService(db)

	var wg sync.WaitGroup
	errCh := make(chan error, 500)
	for g := 0; g < 10; g++ {
		g := g
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 10; i++ {
				if err := psvc.SaveProgress(user.ID, "A", (g+i)%50); err != nil {
					errCh <- err
					return
				}
				if _, err := psvc.AddFavorite(user.ID, ids[(g+i)%len(ids)], "A"); err != nil {
					errCh <- err
					return
				}
				if _, err := psvc.AddError(user.ID, ids[(g+i)%len(ids)], "A"); err != nil {
					errCh <- err
					return
				}
			}
		}()
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		t.Fatalf("concurrent write failed: %v", err)
	}
	if n, _ := store.CountFavorites(db, user.ID); n == 0 {
		t.Fatal("favorites should be written")
	}
	if n, _ := store.CountErrors(db, user.ID); n == 0 {
		t.Fatal("errors should be written")
	}
}

func TestPageBounds(t *testing.T) {
	db := setupDB(t)
	for i := 0; i < 5; i++ {
		seedQuestion(t, db, fmt.Sprintf("P%02d", i), 1, 0, 0, 1, "A")
	}
	page, err := store.PageQuestions(db, 0, 0, "", false, false, false)
	if err != nil {
		t.Fatalf("page: %v", err)
	}
	if page.Page != 1 || page.PageSize != 20 {
		t.Fatalf("bounds not normalized: page=%d size=%d", page.Page, page.PageSize)
	}
	page, err = store.PageQuestions(db, 1, 1000, "", false, false, false)
	if err != nil {
		t.Fatalf("page: %v", err)
	}
	if page.PageSize != 100 {
		t.Fatalf("page size not clamped: %d", page.PageSize)
	}
	page, err = store.PageQuestions(db, 99, 20, "", false, false, false)
	if err != nil {
		t.Fatalf("page: %v", err)
	}
	if len(page.Data) != 0 {
		t.Fatalf("out-of-range page should be empty, got %d", len(page.Data))
	}
}

func TestExamAnswerNormalization(t *testing.T) {
	db := setupDB(t)
	user := login(t, db)
	seedExamBank(t, db, "A", 32, 8)
	svc := NewExamService(db)

	start, err := svc.CreateExam(user.ID, "A")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	details, _ := store.GetExamDetails(db, start.ExamID)
	answers := map[string]string{}
	for _, d := range details {
		// 小写、空格、逆序：交卷时应被规范化后判对
		raw := " " + strings.ToLower(reverse(d.CorrectAnswer)) + " "
		answers[fmt.Sprintf("%d", d.QuestionID)] = raw
	}
	res, err := svc.SubmitExam(start.ExamID, answers)
	if err != nil {
		t.Fatalf("submit: %v", err)
	}
	if res.Score != res.TotalCount || !res.Passed {
		t.Fatalf("normalized answers should all be correct: %+v", res)
	}
}

func reverse(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}
