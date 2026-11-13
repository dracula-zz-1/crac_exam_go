package services

import (
	"path/filepath"
	"testing"

	"crac_exam_v3/backend/store"
)

func TestSetUserLevelPersists(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "users.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	user, err := store.EnsureUser(db, "张三", "110101199001011234")
	if err != nil {
		t.Fatal(err)
	}
	if user.Level != "" {
		t.Fatalf("新用户等级应为空，实际 %q", user.Level)
	}

	svc := NewUserService(db)
	updated, err := svc.SetLevel(user.ID, "b")
	if err != nil {
		t.Fatal(err)
	}
	if updated == nil || updated.Level != "B" {
		t.Fatalf("SetLevel 结果异常: %+v", updated)
	}

	if _, err := svc.SetLevel(user.ID, "D"); err == nil {
		t.Fatal("非法等级应当报错")
	}
	if _, err := svc.SetLevel(user.ID, ""); err == nil {
		t.Fatal("空等级应当报错")
	}

	// 再次登录不应清掉已保存的等级
	again, err := store.EnsureUser(db, "张三", "110101199001011234")
	if err != nil {
		t.Fatal(err)
	}
	if again.Level != "B" {
		t.Fatalf("登录后等级被清空: %q", again.Level)
	}

	// none 表示尚未持证
	none, err := svc.SetLevel(user.ID, "none")
	if err != nil {
		t.Fatal(err)
	}
	if none.Level != "none" {
		t.Fatalf("none 等级保存失败: %q", none.Level)
	}
}
