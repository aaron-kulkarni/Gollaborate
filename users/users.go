package users

import (
	"fmt"
	"sync"
)

// User represents a user in the collaborative editor
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

// Manager handles user creation and management
type Manager struct {
	nextUserID int
	users      map[int]*User
	mutex      sync.RWMutex
}

// NewManager creates a new user manager
func NewManager() *Manager {
	return &Manager{
		nextUserID: 1,
		users:      make(map[int]*User),
	}
}

// CreateUser creates a new user with a unique ID
func (um *Manager) CreateUser(name string) *User {
	um.mutex.Lock()
	defer um.mutex.Unlock()

	user := &User{
		ID:    um.nextUserID,
		Name:  name,
		Color: generateUserColor(um.nextUserID),
	}
	um.users[user.ID] = user
	um.nextUserID++
	return user
}

// GetUser retrieves a user by ID
func (um *Manager) GetUser(userID int) *User {
	um.mutex.RLock()
	defer um.mutex.RUnlock()
	return um.users[userID]
}

// RemoveUser removes a user by ID
func (um *Manager) RemoveUser(userID int) {
	um.mutex.Lock()
	defer um.mutex.Unlock()
	delete(um.users, userID)
}

// GetAllUsers returns all active users
func (um *Manager) GetAllUsers() []*User {
	um.mutex.RLock()
	defer um.mutex.RUnlock()

	users := make([]*User, 0, len(um.users))
	for _, user := range um.users {
		users = append(users, user)
	}
	return users
}

// GetUserCount returns the number of active users
func (um *Manager) GetUserCount() int {
	um.mutex.RLock()
	defer um.mutex.RUnlock()
	return len(um.users)
}

// UserExists checks if a user with the given ID exists
func (um *Manager) UserExists(userID int) bool {
	um.mutex.RLock()
	defer um.mutex.RUnlock()
	_, exists := um.users[userID]
	return exists
}

// UpdateUserName updates a user's display name
func (um *Manager) UpdateUserName(userID int, newName string) error {
	um.mutex.Lock()
	defer um.mutex.Unlock()

	user, exists := um.users[userID]
	if !exists {
		return fmt.Errorf("user with ID %d not found", userID)
	}

	user.Name = newName
	return nil
}

// UpdateUserColor updates a user's color
func (um *Manager) UpdateUserColor(userID int, newColor string) error {
	um.mutex.Lock()
	defer um.mutex.Unlock()

	user, exists := um.users[userID]
	if !exists {
		return fmt.Errorf("user with ID %d not found", userID)
	}

	user.Color = newColor
	return nil
}

// generateUserColor generates a color for a user based on their ID
func generateUserColor(userID int) string {
	colors := []string{
		"#FF5733", "#33FF57", "#3357FF", "#FF33F1",
		"#F1FF33", "#33FFF1", "#FF8C33", "#8C33FF",
		"#33FF8C", "#FF3333", "#33FFFF", "#FFFF33",
		"#FF5733", "#8B4513", "#FF1493", "#00CED1",
		"#FFD700", "#32CD32", "#FF4500", "#9370DB",
		"#00FA9A", "#FF6347", "#4169E1", "#FF69B4",
	}
	return colors[(userID-1)%len(colors)]
}

// GetNextAvailableID returns what the next user ID would be (without creating a user)
func (um *Manager) GetNextAvailableID() int {
	um.mutex.RLock()
	defer um.mutex.RUnlock()
	return um.nextUserID
}