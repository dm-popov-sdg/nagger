package storage

// GetDescription returns the task description
func (t Task) GetDescription() string {
	return t.Description
}

// GetID returns the task ID as a hex string
func (t Task) GetID() string {
	return t.ID.Hex()
}

// GetStatus returns the task status as a string
func (t Task) GetStatus() string {
	return string(t.Status)
}
