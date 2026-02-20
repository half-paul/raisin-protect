package models

import "time"

// Requirement represents a compliance requirement within a framework version.
type Requirement struct {
	ID                 string    `json:"id"`
	FrameworkVersionID string    `json:"framework_version_id,omitempty"`
	ParentID           *string   `json:"parent_id,omitempty"`
	Identifier         string    `json:"identifier"`
	Title              string    `json:"title"`
	Description        *string   `json:"description,omitempty"`
	Guidance           *string   `json:"guidance,omitempty"`
	SectionOrder       int       `json:"section_order"`
	Depth              int       `json:"depth"`
	IsAssessable       bool      `json:"is_assessable"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// RequirementTreeNode is a requirement with nested children for tree rendering.
type RequirementTreeNode struct {
	ID           string                 `json:"id"`
	Identifier   string                 `json:"identifier"`
	Title        string                 `json:"title"`
	Depth        int                    `json:"depth"`
	IsAssessable bool                   `json:"is_assessable"`
	Children     []*RequirementTreeNode `json:"children"`
}

// BuildRequirementTree converts a flat list of requirements into a tree structure.
func BuildRequirementTree(reqs []Requirement) []*RequirementTreeNode {
	nodeMap := make(map[string]*RequirementTreeNode)
	var roots []*RequirementTreeNode

	// Create nodes
	for _, r := range reqs {
		node := &RequirementTreeNode{
			ID:           r.ID,
			Identifier:   r.Identifier,
			Title:        r.Title,
			Depth:        r.Depth,
			IsAssessable: r.IsAssessable,
			Children:     []*RequirementTreeNode{},
		}
		nodeMap[r.ID] = node
	}

	// Build tree
	for _, r := range reqs {
		node := nodeMap[r.ID]
		if r.ParentID == nil {
			roots = append(roots, node)
		} else if parent, ok := nodeMap[*r.ParentID]; ok {
			parent.Children = append(parent.Children, node)
		} else {
			// Orphan — attach to root
			roots = append(roots, node)
		}
	}

	return roots
}
