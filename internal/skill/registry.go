package skill

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sage-roundtable/core/pkg/utils"
)

// Repository 管理所有的技能配置
type Repository struct {
	skills map[string]*SkillProfile
}

func NewRepository() *Repository {
	return &Repository{
		skills: make(map[string]*SkillProfile),
	}
}

// LoadFromDir 加载指定目录下的所有 .md 技能配置
func (r *Repository) LoadFromDir(dirPath string) error {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("failed to read skills directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
			continue
		}

		fullPath := filepath.Join(dirPath, entry.Name())
		content, err := os.ReadFile(fullPath)
		if err != nil {
			return fmt.Errorf("failed to read skill file %s: %w", fullPath, err)
		}

		profile := &SkillProfile{}
		body, err := utils.ParseMarkdownConfig(content, profile)
		if err != nil {
			return fmt.Errorf("failed to parse skill config from %s: %w", fullPath, err)
		}

		profile.Instruction = body

		if profile.ID == "" {
			return fmt.Errorf("skill profile in %s is missing 'id'", fullPath)
		}

		r.skills[profile.ID] = profile
	}

	return nil
}

// Get 根据 ID 获取技能配置
func (r *Repository) Get(id string) (*SkillProfile, bool) {
	skill, ok := r.skills[id]
	return skill, ok
}

// GetAll 获取所有已加载的技能配置
func (r *Repository) GetAll() []*SkillProfile {
	var list []*SkillProfile
	for _, s := range r.skills {
		list = append(list, s)
	}
	return list
}
