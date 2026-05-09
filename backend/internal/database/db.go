package database

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func InitDB(dbPath string) error {
	var err error
	DB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	if err = DB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Database connected successfully")

	if err = runMigrations(); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

func runMigrations() error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username VARCHAR(50) UNIQUE NOT NULL,
			email VARCHAR(100) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			avatar_url VARCHAR(255),
			credits INTEGER DEFAULT 50,
			quota_limit INTEGER DEFAULT 50,
			role VARCHAR(20) DEFAULT 'user',
			status VARCHAR(20) DEFAULT 'active',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)`,
		`CREATE INDEX IF NOT EXISTS idx_users_username ON users(username)`,

		`CREATE TABLE IF NOT EXISTS sessions (
			id VARCHAR(100) PRIMARY KEY,
			user_id INTEGER NOT NULL,
			title VARCHAR(255),
			last_message_at DATETIME,
			message_count INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_updated_at ON sessions(updated_at)`,

		`CREATE TABLE IF NOT EXISTS image_records (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			session_id VARCHAR(100),
			prompt TEXT NOT NULL,
			revised_prompt TEXT,
			image_url VARCHAR(500),
			local_path VARCHAR(500),
			size VARCHAR(20) DEFAULT '1024x1024',
			quality VARCHAR(20) DEFAULT 'standard',
			style VARCHAR(20),
			model VARCHAR(50) DEFAULT 'gpt-image-2',
			credits_used INTEGER DEFAULT 1,
			generation_time INTEGER,
			status VARCHAR(20) DEFAULT 'pending',
			error_message TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_image_records_user_id ON image_records(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_image_records_session_id ON image_records(session_id)`,
		`CREATE INDEX IF NOT EXISTS idx_image_records_created_at ON image_records(created_at)`,
		`CREATE INDEX IF NOT EXISTS idx_image_records_status ON image_records(status)`,

		`CREATE TABLE IF NOT EXISTS usage_stats (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			date DATE NOT NULL,
			total_generations INTEGER DEFAULT 0,
			successful_generations INTEGER DEFAULT 0,
			failed_generations INTEGER DEFAULT 0,
			total_credits_used INTEGER DEFAULT 0,
			total_generation_time INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			UNIQUE(user_id, date)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_usage_stats_user_date ON usage_stats(user_id, date)`,

		`CREATE TABLE IF NOT EXISTS api_configs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			config_name VARCHAR(100) NOT NULL,
			api_key_encrypted TEXT NOT NULL,
			api_base_url VARCHAR(255),
			model VARCHAR(50) DEFAULT 'gpt-image-2',
			is_active BOOLEAN DEFAULT 1,
			max_requests_per_minute INTEGER DEFAULT 5,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_api_configs_user_id ON api_configs(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_api_configs_is_active ON api_configs(is_active)`,

		`CREATE TABLE IF NOT EXISTS llm_configs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			config_name VARCHAR(100) NOT NULL,
			api_key_encrypted TEXT NOT NULL,
			api_base_url VARCHAR(255),
			model VARCHAR(80) DEFAULT 'deepseek-chat',
			is_active BOOLEAN DEFAULT 1,
			max_requests_per_minute INTEGER DEFAULT 5,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_llm_configs_user_id ON llm_configs(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_llm_configs_is_active ON llm_configs(is_active)`,
	}

	for _, migration := range migrations {
		if _, err := DB.Exec(migration); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	if err := ensureColumnExists(
		"users",
		"quota_limit",
		`ALTER TABLE users ADD COLUMN quota_limit INTEGER DEFAULT 50`,
	); err != nil {
		return err
	}

	if _, err := DB.Exec(`UPDATE users SET quota_limit = 50 WHERE quota_limit IS NULL OR quota_limit = 0`); err != nil {
		return fmt.Errorf("failed to normalize quota_limit: %w", err)
	}
	if _, err := DB.Exec(`UPDATE users SET credits = CASE WHEN credits > COALESCE(quota_limit, 50) THEN COALESCE(quota_limit, 50) ELSE credits END`); err != nil {
		return fmt.Errorf("failed to normalize credits: %w", err)
	}

	log.Println("Database migrations completed successfully")
	return nil
}

func ensureColumnExists(tableName, columnName, alterSQL string) error {
	rows, err := DB.Query(fmt.Sprintf("PRAGMA table_info(%s)", tableName))
	if err != nil {
		return fmt.Errorf("failed to inspect table %s: %w", tableName, err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			cid       int
			name      string
			dataType  string
			notNull   int
			defaultV  sql.NullString
			primary   int
		)
		if err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultV, &primary); err != nil {
			return fmt.Errorf("failed to scan table info: %w", err)
		}
		if strings.EqualFold(name, columnName) {
			return nil
		}
	}

	if _, err := DB.Exec(alterSQL); err != nil {
		return fmt.Errorf("failed to add column %s.%s: %w", tableName, columnName, err)
	}

	return nil
}

func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}
