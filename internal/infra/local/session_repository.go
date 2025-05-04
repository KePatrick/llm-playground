package local

import (
	"context"
	"encoding/json"
	"kepatrick/llm-playground/internal/domain/entity"
	"os"
	"path/filepath"
)

type FileSessionRepo struct {
	BaseDir string // Base directory where session files will be stored
}

// Constructor for FileSessionRepo
func NewFileSessionRepo(baseDir string) *FileSessionRepo {
	return &FileSessionRepo{BaseDir: baseDir}
}

// AppendMessage appends a new message to the session's local JSON file.
// If the file does not exist, it will be created.
func (r *FileSessionRepo) AppendMessage(ctx context.Context, sessionID string, msg entity.Message) error {
	path := filepath.Join(r.BaseDir, sessionID+".json")

	var messages []entity.Message

	// Check if file exists
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		// File doesn't exist, create new slice with the message
		messages = []entity.Message{msg}
	} else {
		// File exists, read and unmarshal messages
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(data, &messages); err != nil {
			return err
		}

		// Append new message
		messages = append(messages, msg)
	}

	// Marshal and write to file (overwrite or create)
	data, err := json.MarshalIndent(messages, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// FetchPrevMessage retrieves all messages from the session's local JSON file
func (r *FileSessionRepo) FetchPrevMessage(ctx context.Context, sessionID string) ([]entity.Message, error) {
	path := filepath.Join(r.BaseDir, sessionID+".json")

	// Read file content
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// If file doesn't exist, return empty message list
			return []entity.Message{}, nil
		}
		return nil, err
	}

	// Deserialize the JSON data into message structs
	var messages []entity.Message
	if err := json.Unmarshal(data, &messages); err != nil {
		return nil, err
	}

	return messages, nil
}

func (r *FileSessionRepo) ExistKey(ctx context.Context, sessionID string) bool {
	path := filepath.Join(r.BaseDir, sessionID+".json")

	// Read file content
	_, err := os.ReadFile(path)

	return err == nil
}
