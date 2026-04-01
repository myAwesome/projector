package store

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"
)

type Project struct {
	Name      string    `json:"name"`
	Dir       string    `json:"dir"`
	Script    string    `json:"script"`
	CreatedAt time.Time `json:"created_at"`
}

var ErrNotFound = errors.New("project not found")
var ErrExists = errors.New("project already exists")

func configDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".config", "project")
	return dir, os.MkdirAll(dir, 0755)
}

func projectsFile() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "projects.json"), nil
}

func Load() ([]Project, error) {
	path, err := projectsFile()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return []Project{}, nil
	}
	if err != nil {
		return nil, err
	}
	var projects []Project
	return projects, json.Unmarshal(data, &projects)
}

func save(projects []Project) error {
	path, err := projectsFile()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(projects, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func Add(p Project) error {
	projects, err := Load()
	if err != nil {
		return err
	}
	for _, existing := range projects {
		if existing.Name == p.Name {
			return ErrExists
		}
	}
	p.CreatedAt = time.Now()
	return save(append(projects, p))
}

func Find(name string) (Project, error) {
	projects, err := Load()
	if err != nil {
		return Project{}, err
	}
	for _, p := range projects {
		if p.Name == name {
			return p, nil
		}
	}
	return Project{}, ErrNotFound
}
