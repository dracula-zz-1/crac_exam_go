package store

import (
	"database/sql"
	"strings"
)

// 应用级键值元数据。目前用于题库来源/版本/自动更新设置。
const (
	MetaBankSourceURL    = "bank.source_url"
	MetaBankSourceName   = "bank.source_name"
	MetaBankSHA256       = "bank.sha256"
	MetaBankSize         = "bank.size"
	MetaBankETag         = "bank.etag"
	MetaBankLastModified = "bank.last_modified"
	MetaBankVersion      = "bank.version"
	MetaBankDiscoveryURL = "bank.discovery_url"
	MetaBankImportedAt   = "bank.imported_at"
	MetaBankAutoCheck    = "bank.auto_check"
	MetaBankAutoImport   = "bank.auto_import"

	// 应用自身（GitHub Release）在线更新源。
	MetaAppGitHubRepo = "app.github_repo"
)

func GetMeta(db *sql.DB, key string) (string, error) {
	var v string
	err := db.QueryRow(`SELECT value FROM app_meta WHERE key = ?`, key).Scan(&v)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return v, err
}

func GetMetaMap(db *sql.DB, prefix string) (map[string]string, error) {
	rows, err := db.Query(`SELECT key, value FROM app_meta WHERE key LIKE ?`, prefix+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]string{}
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, err
		}
		out[strings.TrimPrefix(k, prefix)] = v
	}
	return out, rows.Err()
}

func SetMeta(db *sql.DB, key, value string) error {
	_, err := db.Exec(`INSERT INTO app_meta (key, value) VALUES (?, ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value`, key, value)
	return err
}

func SetMetas(db *sql.DB, values map[string]string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	stmt, err := tx.Prepare(`INSERT INTO app_meta (key, value) VALUES (?, ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for k, v := range values {
		if _, err := stmt.Exec(k, v); err != nil {
			return err
		}
	}
	return tx.Commit()
}
