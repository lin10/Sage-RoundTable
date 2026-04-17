package session

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/sage-roundtable/core/pkg/errors"
	"github.com/sage-roundtable/core/pkg/logger"
)

// SessionState 会话状态
type SessionState string

const (
	SessionStateInitialized SessionState = "initialized"
	SessionStateQuestioning SessionState = "questioning"
	SessionStateAnswering   SessionState = "answering"
	SessionStateDebating    SessionState = "debating"
	SessionStateResolving   SessionState = "resolving"
	SessionStateCompleted   SessionState = "completed"
	SessionStateCancelled   SessionState = "cancelled"
)

// DebateType 辩论类型
type DebateType string

const (
	DebateTypeIndividual       DebateType = "individual"
	DebateTypeCrossExamination DebateType = "cross_examination"
)

// Statement 智者发言
type Statement struct {
	ID        string    `json:"id"`
	DebateID  string    `json:"debate_id"`
	AgentID   string    `json:"agent_id"`
	AgentName string    `json:"agent_name"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// DebateRound 辩论轮次
type DebateRound struct {
	ID          string      `json:"id"`
	SessionID   string      `json:"session_id"`
	RoundNumber int         `json:"round_number"`
	Type        DebateType  `json:"type"`
	Statements  []Statement `json:"statements"`
	StartedAt   time.Time   `json:"started_at"`
	CompletedAt *time.Time  `json:"completed_at,omitempty"`
}

// Question 问题
type Question struct {
	ID         string     `json:"id"`
	SessionID  string     `json:"session_id"`
	Content    string     `json:"content"`
	Answer     *string    `json:"answer,omitempty"`
	AskedAt    time.Time  `json:"asked_at"`
	AnsweredAt *time.Time `json:"answered_at,omitempty"`
}

// Resolution 决议书
type Resolution struct {
	ID            string    `json:"id"`
	SessionID     string    `json:"session_id"`
	Consensus     []string  `json:"consensus"`
	Disagreements []string  `json:"disagreements"`
	Strategies    string    `json:"strategies"` // JSON object
	Risks         []string  `json:"risks"`
	Actions       []string  `json:"actions"`
	Quotes        []string  `json:"quotes"`
	GeneratedAt   time.Time `json:"generated_at"`
}

// Session 会话
type Session struct {
	ID           string                 `json:"id"`
	UserID       string                 `json:"user_id"`
	State        SessionState           `json:"state"`
	InitialQuery string                 `json:"initial_query"`
	ContextInfo  map[string]string      `json:"context_info,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	CompletedAt  *time.Time             `json:"completed_at,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`

	// 关联数据（不存储在 sessions 表中）
	Questions  []Question    `json:"questions,omitempty"`
	Debates    []DebateRound `json:"debates,omitempty"`
	Resolution *Resolution   `json:"resolution,omitempty"`
}

// Store 会话存储接口
type Store interface {
	// Session CRUD
	CreateSession(session *Session) error
	GetSession(id string) (*Session, error)
	UpdateSession(session *Session) error
	DeleteSession(id string) error
	ListSessions(userID string, limit, offset int) ([]*Session, error)

	// Questions
	AddQuestion(question *Question) error
	UpdateQuestionAnswer(questionID string, answer string) error
	GetQuestionsBySession(sessionID string) ([]Question, error)

	// Debates
	CreateDebateRound(debate *DebateRound) error
	AddStatement(statement *Statement) error
	GetDebatesBySession(sessionID string) ([]DebateRound, error)

	// Resolution
	SaveResolution(resolution *Resolution) error
	GetResolution(sessionID string) (*Resolution, error)

	// Close 关闭数据库连接
	Close() error
}

// SQLiteStore SQLite 存储实现
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore 创建 SQLite 存储实例
func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, errors.WrapError(errors.ErrDatabaseError, err, "failed to open database")
	}

	// 配置连接池
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	store := &SQLiteStore{db: db}

	// 初始化数据库表
	if err := store.initializeSchema(); err != nil {
		return nil, errors.WrapError(errors.ErrDatabaseError, err, "failed to initialize schema")
	}

	logger.Log.Info("SQLite store initialized", "path", dbPath)
	return store, nil
}

// initializeSchema 初始化数据库表结构
func (s *SQLiteStore) initializeSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS sessions (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		state TEXT NOT NULL DEFAULT 'initialized',
		initial_query TEXT NOT NULL,
		context_info TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		completed_at TIMESTAMP,
		metadata TEXT
	);
	
	CREATE TABLE IF NOT EXISTS questions (
		id TEXT PRIMARY KEY,
		session_id TEXT NOT NULL,
		content TEXT NOT NULL,
		answer TEXT,
		asked_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		answered_at TIMESTAMP,
		FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
	);
	
	CREATE TABLE IF NOT EXISTS debates (
		id TEXT PRIMARY KEY,
		session_id TEXT NOT NULL,
		round_number INTEGER NOT NULL,
		type TEXT NOT NULL,
		started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		completed_at TIMESTAMP,
		FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
	);
	
	CREATE TABLE IF NOT EXISTS statements (
		id TEXT PRIMARY KEY,
		debate_id TEXT NOT NULL,
		agent_id TEXT NOT NULL,
		agent_name TEXT NOT NULL,
		content TEXT NOT NULL,
		timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (debate_id) REFERENCES debates(id) ON DELETE CASCADE
	);
	
	CREATE TABLE IF NOT EXISTS resolutions (
		id TEXT PRIMARY KEY,
		session_id TEXT NOT NULL UNIQUE,
		consensus TEXT,
		disagreements TEXT,
		strategies TEXT,
		risks TEXT,
		actions TEXT,
		quotes TEXT,
		generated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
	);
	
	CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
	CREATE INDEX IF NOT EXISTS idx_sessions_state ON sessions(state);
	CREATE INDEX IF NOT EXISTS idx_questions_session_id ON questions(session_id);
	CREATE INDEX IF NOT EXISTS idx_debates_session_id ON debates(session_id);
	CREATE INDEX IF NOT EXISTS idx_statements_debate_id ON statements(debate_id);
	`

	_, err := s.db.Exec(schema)
	return err
}

// CreateSession 创建会话
func (s *SQLiteStore) CreateSession(session *Session) error {
	contextJSON, _ := json.Marshal(session.ContextInfo)
	metadataJSON, _ := json.Marshal(session.Metadata)

	query := `INSERT INTO sessions (id, user_id, state, initial_query, context_info, created_at, updated_at, metadata) 
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	now := time.Now()
	_, err := s.db.Exec(query,
		session.ID,
		session.UserID,
		session.State,
		session.InitialQuery,
		string(contextJSON),
		now,
		now,
		string(metadataJSON),
	)

	if err != nil {
		return errors.WrapError(errors.ErrDatabaseError, err, "failed to create session")
	}

	session.CreatedAt = now
	session.UpdatedAt = now

	logger.Log.Debug("Session created", "id", session.ID)
	return nil
}

// GetSession 获取会话
func (s *SQLiteStore) GetSession(id string) (*Session, error) {
	query := `SELECT id, user_id, state, initial_query, context_info, created_at, updated_at, completed_at, metadata 
	          FROM sessions WHERE id = ?`

	var session Session
	var contextJSON, metadataJSON sql.NullString
	var completedAt sql.NullTime

	err := s.db.QueryRow(query, id).Scan(
		&session.ID,
		&session.UserID,
		&session.State,
		&session.InitialQuery,
		&contextJSON,
		&session.CreatedAt,
		&session.UpdatedAt,
		&completedAt,
		&metadataJSON,
	)

	if err == sql.ErrNoRows {
		return nil, errors.NewAppError(errors.ErrSessionNotFound, "session not found", fmt.Sprintf("id: %s", id))
	}
	if err != nil {
		return nil, errors.WrapError(errors.ErrDatabaseError, err, "failed to get session")
	}

	if completedAt.Valid {
		session.CompletedAt = &completedAt.Time
	}

	// 解析 JSON 字段
	if contextJSON.Valid {
		json.Unmarshal([]byte(contextJSON.String), &session.ContextInfo)
	}
	if metadataJSON.Valid {
		json.Unmarshal([]byte(metadataJSON.String), &session.Metadata)
	}

	// 加载关联数据
	session.Questions, _ = s.GetQuestionsBySession(id)
	session.Debates, _ = s.GetDebatesBySession(id)
	session.Resolution, _ = s.GetResolution(id)

	return &session, nil
}

// UpdateSession 更新会话
func (s *SQLiteStore) UpdateSession(session *Session) error {
	contextJSON, _ := json.Marshal(session.ContextInfo)
	metadataJSON, _ := json.Marshal(session.Metadata)

	query := `UPDATE sessions SET state = ?, context_info = ?, updated_at = ?, completed_at = ?, metadata = ? WHERE id = ?`

	var completedAt *time.Time
	if session.CompletedAt != nil {
		completedAt = session.CompletedAt
	}

	result, err := s.db.Exec(query,
		session.State,
		string(contextJSON),
		time.Now(),
		completedAt,
		string(metadataJSON),
		session.ID,
	)

	if err != nil {
		return errors.WrapError(errors.ErrDatabaseError, err, "failed to update session")
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.NewAppError(errors.ErrSessionNotFound, "session not found", fmt.Sprintf("id: %s", session.ID))
	}

	session.UpdatedAt = time.Now()
	return nil
}

// DeleteSession 删除会话
func (s *SQLiteStore) DeleteSession(id string) error {
	result, err := s.db.Exec("DELETE FROM sessions WHERE id = ?", id)
	if err != nil {
		return errors.WrapError(errors.ErrDatabaseError, err, "failed to delete session")
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.NewAppError(errors.ErrSessionNotFound, "session not found", fmt.Sprintf("id: %s", id))
	}

	logger.Log.Debug("Session deleted", "id", id)
	return nil
}

// ListSessions 列出会话
func (s *SQLiteStore) ListSessions(userID string, limit, offset int) ([]*Session, error) {
	query := `SELECT id, user_id, state, initial_query, created_at, updated_at, completed_at 
	          FROM sessions WHERE user_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`

	rows, err := s.db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, errors.WrapError(errors.ErrDatabaseError, err, "failed to list sessions")
	}
	defer rows.Close()

	var sessions []*Session
	for rows.Next() {
		var session Session
		var completedAt sql.NullTime

		err := rows.Scan(
			&session.ID,
			&session.UserID,
			&session.State,
			&session.InitialQuery,
			&session.CreatedAt,
			&session.UpdatedAt,
			&completedAt,
		)
		if err != nil {
			continue
		}

		if completedAt.Valid {
			session.CompletedAt = &completedAt.Time
		}

		sessions = append(sessions, &session)
	}

	return sessions, nil
}

// AddQuestion 添加问题
func (s *SQLiteStore) AddQuestion(question *Question) error {
	query := `INSERT INTO questions (id, session_id, content, asked_at) VALUES (?, ?, ?, ?)`

	_, err := s.db.Exec(query,
		question.ID,
		question.SessionID,
		question.Content,
		question.AskedAt,
	)

	if err != nil {
		return errors.WrapError(errors.ErrDatabaseError, err, "failed to add question")
	}

	return nil
}

// UpdateQuestionAnswer 更新问题答案
func (s *SQLiteStore) UpdateQuestionAnswer(questionID string, answer string) error {
	query := `UPDATE questions SET answer = ?, answered_at = ? WHERE id = ?`

	now := time.Now()
	result, err := s.db.Exec(query, answer, now, questionID)
	if err != nil {
		return errors.WrapError(errors.ErrDatabaseError, err, "failed to update question answer")
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.NewAppError(errors.ErrNotFound, "question not found", fmt.Sprintf("id: %s", questionID))
	}

	return nil
}

// GetQuestionsBySession 获取会话的问题列表
func (s *SQLiteStore) GetQuestionsBySession(sessionID string) ([]Question, error) {
	query := `SELECT id, session_id, content, answer, asked_at, answered_at 
	          FROM questions WHERE session_id = ? ORDER BY asked_at ASC`

	rows, err := s.db.Query(query, sessionID)
	if err != nil {
		return nil, errors.WrapError(errors.ErrDatabaseError, err, "failed to get questions")
	}
	defer rows.Close()

	var questions []Question
	for rows.Next() {
		var q Question
		var answer sql.NullString
		var answeredAt sql.NullTime

		err := rows.Scan(&q.ID, &q.SessionID, &q.Content, &answer, &q.AskedAt, &answeredAt)
		if err != nil {
			continue
		}

		if answer.Valid {
			q.Answer = &answer.String
		}
		if answeredAt.Valid {
			q.AnsweredAt = &answeredAt.Time
		}

		questions = append(questions, q)
	}

	return questions, nil
}

// CreateDebateRound 创建辩论轮次
func (s *SQLiteStore) CreateDebateRound(debate *DebateRound) error {
	query := `INSERT INTO debates (id, session_id, round_number, type, started_at) VALUES (?, ?, ?, ?, ?)`

	_, err := s.db.Exec(query,
		debate.ID,
		debate.SessionID,
		debate.RoundNumber,
		debate.Type,
		debate.StartedAt,
	)

	if err != nil {
		return errors.WrapError(errors.ErrDatabaseError, err, "failed to create debate round")
	}

	return nil
}

// AddStatement 添加发言
func (s *SQLiteStore) AddStatement(statement *Statement) error {
	query := `INSERT INTO statements (id, debate_id, agent_id, agent_name, content, timestamp) 
	          VALUES (?, ?, ?, ?, ?, ?)`

	_, err := s.db.Exec(query,
		statement.ID,
		statement.DebateID,
		statement.AgentID,
		statement.AgentName,
		statement.Content,
		statement.Timestamp,
	)

	if err != nil {
		return errors.WrapError(errors.ErrDatabaseError, err, "failed to add statement")
	}

	return nil
}

// GetDebatesBySession 获取会话的辩论记录
func (s *SQLiteStore) GetDebatesBySession(sessionID string) ([]DebateRound, error) {
	query := `SELECT id, session_id, round_number, type, started_at, completed_at 
	          FROM debates WHERE session_id = ? ORDER BY round_number ASC`

	rows, err := s.db.Query(query, sessionID)
	if err != nil {
		return nil, errors.WrapError(errors.ErrDatabaseError, err, "failed to get debates")
	}
	defer rows.Close()

	var debates []DebateRound
	for rows.Next() {
		var debate DebateRound
		var completedAt sql.NullTime

		err := rows.Scan(&debate.ID, &debate.SessionID, &debate.RoundNumber, &debate.Type, &debate.StartedAt, &completedAt)
		if err != nil {
			continue
		}

		if completedAt.Valid {
			debate.CompletedAt = &completedAt.Time
		}

		// 加载该轮次的发言
		debate.Statements, _ = s.getStatementsByDebate(debate.ID)

		debates = append(debates, debate)
	}

	return debates, nil
}

// getStatementsByDebate 获取辩论的发言列表
func (s *SQLiteStore) getStatementsByDebate(debateID string) ([]Statement, error) {
	query := `SELECT id, debate_id, agent_id, agent_name, content, timestamp 
	          FROM statements WHERE debate_id = ? ORDER BY timestamp ASC`

	rows, err := s.db.Query(query, debateID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var statements []Statement
	for rows.Next() {
		var stmt Statement
		err := rows.Scan(&stmt.ID, &stmt.DebateID, &stmt.AgentID, &stmt.AgentName, &stmt.Content, &stmt.Timestamp)
		if err != nil {
			continue
		}
		statements = append(statements, stmt)
	}

	return statements, nil
}

// SaveResolution 保存决议书
func (s *SQLiteStore) SaveResolution(resolution *Resolution) error {
	consensusJSON, _ := json.Marshal(resolution.Consensus)
	disagreementsJSON, _ := json.Marshal(resolution.Disagreements)
	risksJSON, _ := json.Marshal(resolution.Risks)
	actionsJSON, _ := json.Marshal(resolution.Actions)
	quotesJSON, _ := json.Marshal(resolution.Quotes)

	query := `INSERT OR REPLACE INTO resolutions 
	          (id, session_id, consensus, disagreements, strategies, risks, actions, quotes, generated_at) 
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := s.db.Exec(query,
		resolution.ID,
		resolution.SessionID,
		string(consensusJSON),
		string(disagreementsJSON),
		resolution.Strategies,
		string(risksJSON),
		string(actionsJSON),
		string(quotesJSON),
		resolution.GeneratedAt,
	)

	if err != nil {
		return errors.WrapError(errors.ErrDatabaseError, err, "failed to save resolution")
	}

	return nil
}

// GetResolution 获取决议书
func (s *SQLiteStore) GetResolution(sessionID string) (*Resolution, error) {
	query := `SELECT id, session_id, consensus, disagreements, strategies, risks, actions, quotes, generated_at 
	          FROM resolutions WHERE session_id = ?`

	var resolution Resolution
	var consensusJSON, disagreementsJSON, risksJSON, actionsJSON, quotesJSON sql.NullString

	err := s.db.QueryRow(query, sessionID).Scan(
		&resolution.ID,
		&resolution.SessionID,
		&consensusJSON,
		&disagreementsJSON,
		&resolution.Strategies,
		&risksJSON,
		&actionsJSON,
		&quotesJSON,
		&resolution.GeneratedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.NewAppError(errors.ErrNotFound, "resolution not found", fmt.Sprintf("session_id: %s", sessionID))
	}
	if err != nil {
		return nil, errors.WrapError(errors.ErrDatabaseError, err, "failed to get resolution")
	}

	// 解析 JSON 字段
	if consensusJSON.Valid {
		json.Unmarshal([]byte(consensusJSON.String), &resolution.Consensus)
	}
	if disagreementsJSON.Valid {
		json.Unmarshal([]byte(disagreementsJSON.String), &resolution.Disagreements)
	}
	if risksJSON.Valid {
		json.Unmarshal([]byte(risksJSON.String), &resolution.Risks)
	}
	if actionsJSON.Valid {
		json.Unmarshal([]byte(actionsJSON.String), &resolution.Actions)
	}
	if quotesJSON.Valid {
		json.Unmarshal([]byte(quotesJSON.String), &resolution.Quotes)
	}

	return &resolution, nil
}

// Close 关闭数据库连接
func (s *SQLiteStore) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// MemoryStore 内存存储实现（用于开发和测试）
type MemoryStore struct {
	sessions    map[string]*Session
	questions   map[string][]Question
	debates     map[string][]DebateRound
	resolutions map[string]*Resolution
}

// NewMemoryStore 创建内存存储实例
func NewMemoryStore() *MemoryStore {
	logger.Log.Info("Memory store initialized")
	return &MemoryStore{
		sessions:    make(map[string]*Session),
		questions:   make(map[string][]Question),
		debates:     make(map[string][]DebateRound),
		resolutions: make(map[string]*Resolution),
	}
}

// CreateSession 创建会话
func (m *MemoryStore) CreateSession(session *Session) error {
	now := time.Now()
	session.CreatedAt = now
	session.UpdatedAt = now
	m.sessions[session.ID] = session
	logger.Log.Debug("Session created in memory", "id", session.ID)
	return nil
}

// GetSession 获取会话
func (m *MemoryStore) GetSession(id string) (*Session, error) {
	session, exists := m.sessions[id]
	if !exists {
		return nil, errors.NewAppError(errors.ErrSessionNotFound, "session not found", fmt.Sprintf("id: %s", id))
	}

	// 返回副本
	sessionCopy := *session
	sessionCopy.Questions = m.questions[id]
	sessionCopy.Debates = m.debates[id]
	if resolution, exists := m.resolutions[id]; exists {
		sessionCopy.Resolution = resolution
	}

	return &sessionCopy, nil
}

// UpdateSession 更新会话
func (m *MemoryStore) UpdateSession(session *Session) error {
	if _, exists := m.sessions[session.ID]; !exists {
		return errors.NewAppError(errors.ErrSessionNotFound, "session not found", fmt.Sprintf("id: %s", session.ID))
	}
	session.UpdatedAt = time.Now()
	m.sessions[session.ID] = session
	return nil
}

// DeleteSession 删除会话
func (m *MemoryStore) DeleteSession(id string) error {
	if _, exists := m.sessions[id]; !exists {
		return errors.NewAppError(errors.ErrSessionNotFound, "session not found", fmt.Sprintf("id: %s", id))
	}
	delete(m.sessions, id)
	delete(m.questions, id)
	delete(m.debates, id)
	delete(m.resolutions, id)
	return nil
}

// ListSessions 列出会话
func (m *MemoryStore) ListSessions(userID string, limit, offset int) ([]*Session, error) {
	var result []*Session
	for _, session := range m.sessions {
		if session.UserID == userID {
			result = append(result, session)
		}
	}

	// 简单的分页
	start := offset
	if start >= len(result) {
		return []*Session{}, nil
	}
	end := start + limit
	if end > len(result) {
		end = len(result)
	}

	return result[start:end], nil
}

// AddQuestion 添加问题
func (m *MemoryStore) AddQuestion(question *Question) error {
	m.questions[question.SessionID] = append(m.questions[question.SessionID], *question)
	return nil
}

// UpdateQuestionAnswer 更新问题答案
func (m *MemoryStore) UpdateQuestionAnswer(questionID string, answer string) error {
	for sessionID, questions := range m.questions {
		for i, q := range questions {
			if q.ID == questionID {
				m.questions[sessionID][i].Answer = &answer
				now := time.Now()
				m.questions[sessionID][i].AnsweredAt = &now
				return nil
			}
		}
	}
	return errors.NewAppError(errors.ErrValidationFailed, "question not found", fmt.Sprintf("id: %s", questionID))
}

// GetQuestionsBySession 获取会话的问题
func (m *MemoryStore) GetQuestionsBySession(sessionID string) ([]Question, error) {
	if questions, exists := m.questions[sessionID]; exists {
		return questions, nil
	}
	return []Question{}, nil
}

// CreateDebateRound 创建辩论轮次
func (m *MemoryStore) CreateDebateRound(debate *DebateRound) error {
	m.debates[debate.SessionID] = append(m.debates[debate.SessionID], *debate)
	return nil
}

// AddStatement 添加发言
func (m *MemoryStore) AddStatement(statement *Statement) error {
	for sessionID, debates := range m.debates {
		for i, d := range debates {
			if d.ID == statement.DebateID {
				m.debates[sessionID][i].Statements = append(m.debates[sessionID][i].Statements, *statement)
				return nil
			}
		}
	}
	return errors.NewAppError(errors.ErrValidationFailed, "debate not found", fmt.Sprintf("id: %s", statement.DebateID))
}

// GetDebatesBySession 获取会话的辩论
func (m *MemoryStore) GetDebatesBySession(sessionID string) ([]DebateRound, error) {
	if debates, exists := m.debates[sessionID]; exists {
		return debates, nil
	}
	return []DebateRound{}, nil
}

// SaveResolution 保存决议
func (m *MemoryStore) SaveResolution(resolution *Resolution) error {
	m.resolutions[resolution.SessionID] = resolution
	return nil
}

// GetResolution 获取决议
func (m *MemoryStore) GetResolution(sessionID string) (*Resolution, error) {
	if resolution, exists := m.resolutions[sessionID]; exists {
		return resolution, nil
	}
	return nil, errors.NewAppError(errors.ErrValidationFailed, "resolution not found", fmt.Sprintf("session_id: %s", sessionID))
}

// Close 关闭存储（内存存储无需操作）
func (m *MemoryStore) Close() error {
	return nil
}
