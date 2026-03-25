package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"temporal-shared"
)

type Store struct {
	db *sql.DB
}

type DefinitionRow struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Version     int       `json:"version"`
	Definition  string    `json:"definition"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func New(databaseURL string) (*Store, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}
	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return s, nil
}

func (s *Store) migrate() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS workflow_definitions (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT DEFAULT '',
			version INT DEFAULT 1,
			definition JSONB NOT NULL,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW()
		)
	`)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(`
		CREATE TABLE IF NOT EXISTS activity_registry (
			activity_name   TEXT NOT NULL,
			task_queue      TEXT NOT NULL,
			display_name    TEXT NOT NULL,
			description     TEXT DEFAULT '',
			service_name    TEXT NOT NULL,
			category        TEXT DEFAULT '',
			input_schema    JSONB DEFAULT '[]',
			default_mapping JSONB DEFAULT '{}',
			registered_at   TIMESTAMPTZ DEFAULT NOW(),
			last_heartbeat  TIMESTAMPTZ DEFAULT NOW(),
			PRIMARY KEY (activity_name, task_queue)
		)
	`)
	return err
}

func (s *Store) List() ([]shared.WorkflowDefinition, error) {
	rows, err := s.db.Query("SELECT id, name, description, version, definition FROM workflow_definitions ORDER BY updated_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var defs []shared.WorkflowDefinition
	for rows.Next() {
		var d shared.WorkflowDefinition
		var defJSON string
		if err := rows.Scan(&d.ID, &d.Name, &d.Description, &d.Version, &defJSON); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(defJSON), &d.Steps); err != nil {
			return nil, err
		}
		defs = append(defs, d)
	}
	return defs, rows.Err()
}

func (s *Store) Get(id string) (*shared.WorkflowDefinition, error) {
	var d shared.WorkflowDefinition
	var defJSON string
	err := s.db.QueryRow(
		"SELECT id, name, description, version, definition FROM workflow_definitions WHERE id = $1", id,
	).Scan(&d.ID, &d.Name, &d.Description, &d.Version, &defJSON)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(defJSON), &d.Steps); err != nil {
		return nil, err
	}
	return &d, nil
}

func (s *Store) Create(d *shared.WorkflowDefinition) error {
	stepsJSON, err := json.Marshal(d.Steps)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(
		`INSERT INTO workflow_definitions (id, name, description, version, definition) VALUES ($1, $2, $3, $4, $5)`,
		d.ID, d.Name, d.Description, d.Version, stepsJSON,
	)
	return err
}

func (s *Store) Update(d *shared.WorkflowDefinition) error {
	stepsJSON, err := json.Marshal(d.Steps)
	if err != nil {
		return err
	}
	res, err := s.db.Exec(
		`UPDATE workflow_definitions SET name=$1, description=$2, version=$3, definition=$4, updated_at=NOW() WHERE id=$5`,
		d.Name, d.Description, d.Version, stepsJSON, d.ID,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("definition %s not found", d.ID)
	}
	return nil
}

func (s *Store) Delete(id string) error {
	res, err := s.db.Exec("DELETE FROM workflow_definitions WHERE id = $1", id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("definition %s not found", id)
	}
	return nil
}

func (s *Store) UpsertActivity(a *shared.ActivityRegistration) error {
	schemaJSON, err := json.Marshal(a.InputSchema)
	if err != nil {
		return err
	}
	mappingJSON, err := json.Marshal(a.DefaultMapping)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(`
		INSERT INTO activity_registry (activity_name, task_queue, display_name, description, service_name, category, input_schema, default_mapping)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (activity_name, task_queue)
		DO UPDATE SET display_name = $3, description = $4, service_name = $5, category = $6,
		              input_schema = $7, default_mapping = $8, last_heartbeat = NOW()
	`, a.ActivityName, a.TaskQueue, a.DisplayName, a.Description, a.ServiceName, a.Category, schemaJSON, mappingJSON)
	return err
}

func (s *Store) BulkUpsertActivities(activities []shared.ActivityRegistration) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO activity_registry (activity_name, task_queue, display_name, description, service_name, category, input_schema, default_mapping)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (activity_name, task_queue)
		DO UPDATE SET display_name = $3, description = $4, service_name = $5, category = $6,
		              input_schema = $7, default_mapping = $8, last_heartbeat = NOW()
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for i := range activities {
		a := &activities[i]
		schemaJSON, err := json.Marshal(a.InputSchema)
		if err != nil {
			return err
		}
		mappingJSON, err := json.Marshal(a.DefaultMapping)
		if err != nil {
			return err
		}
		if _, err := stmt.Exec(a.ActivityName, a.TaskQueue, a.DisplayName, a.Description, a.ServiceName, a.Category, schemaJSON, mappingJSON); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) ListActivities() ([]shared.ActivityRegistration, error) {
	rows, err := s.db.Query(`
		SELECT activity_name, task_queue, display_name, description, service_name, category, input_schema, default_mapping
		FROM activity_registry
		WHERE last_heartbeat > NOW() - INTERVAL '5 minutes'
		ORDER BY category, display_name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []shared.ActivityRegistration
	for rows.Next() {
		var a shared.ActivityRegistration
		var schemaJSON, mappingJSON []byte
		if err := rows.Scan(&a.ActivityName, &a.TaskQueue, &a.DisplayName, &a.Description, &a.ServiceName, &a.Category, &schemaJSON, &mappingJSON); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(schemaJSON, &a.InputSchema); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(mappingJSON, &a.DefaultMapping); err != nil {
			return nil, err
		}
		result = append(result, a)
	}
	return result, rows.Err()
}

func (s *Store) DeleteActivitiesByService(serviceName string) error {
	_, err := s.db.Exec("DELETE FROM activity_registry WHERE service_name = $1", serviceName)
	return err
}

func (s *Store) Close() error {
	return s.db.Close()
}
