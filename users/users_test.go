package users

import (
	"fmt"
	"testing"
)

func TestCreateUser(t *testing.T) {
	manager := NewManager()
	
	user := manager.CreateUser("Alice")
	
	if user.ID != 1 {
		t.Errorf("Expected user ID 1, got %d", user.ID)
	}
	
	if user.Name != "Alice" {
		t.Errorf("Expected user name 'Alice', got '%s'", user.Name)
	}
	
	if user.Color == "" {
		t.Error("Expected user to have a color assigned")
	}
}

func TestCreateMultipleUsers(t *testing.T) {
	manager := NewManager()
	
	alice := manager.CreateUser("Alice")
	bob := manager.CreateUser("Bob")
	charlie := manager.CreateUser("Charlie")
	
	if alice.ID != 1 {
		t.Errorf("Expected Alice ID 1, got %d", alice.ID)
	}
	
	if bob.ID != 2 {
		t.Errorf("Expected Bob ID 2, got %d", bob.ID)
	}
	
	if charlie.ID != 3 {
		t.Errorf("Expected Charlie ID 3, got %d", charlie.ID)
	}
	
	// Ensure each user has a different color
	if alice.Color == bob.Color {
		t.Error("Alice and Bob should have different colors")
	}
	
	if bob.Color == charlie.Color {
		t.Error("Bob and Charlie should have different colors")
	}
}

func TestGetUser(t *testing.T) {
	manager := NewManager()
	
	originalUser := manager.CreateUser("Alice")
	retrievedUser := manager.GetUser(originalUser.ID)
	
	if retrievedUser == nil {
		t.Fatal("Expected to retrieve user, got nil")
	}
	
	if retrievedUser.ID != originalUser.ID {
		t.Errorf("Expected retrieved user ID %d, got %d", originalUser.ID, retrievedUser.ID)
	}
	
	if retrievedUser.Name != originalUser.Name {
		t.Errorf("Expected retrieved user name '%s', got '%s'", originalUser.Name, retrievedUser.Name)
	}
}

func TestGetNonExistentUser(t *testing.T) {
	manager := NewManager()
	
	user := manager.GetUser(999)
	
	if user != nil {
		t.Errorf("Expected nil for non-existent user, got %v", user)
	}
}

func TestRemoveUser(t *testing.T) {
	manager := NewManager()
	
	user := manager.CreateUser("Alice")
	
	// Verify user exists
	if !manager.UserExists(user.ID) {
		t.Error("User should exist before removal")
	}
	
	manager.RemoveUser(user.ID)
	
	// Verify user no longer exists
	if manager.UserExists(user.ID) {
		t.Error("User should not exist after removal")
	}
	
	retrievedUser := manager.GetUser(user.ID)
	if retrievedUser != nil {
		t.Errorf("Expected nil after user removal, got %v", retrievedUser)
	}
}

func TestGetAllUsers(t *testing.T) {
	manager := NewManager()
	
	alice := manager.CreateUser("Alice")
	bob := manager.CreateUser("Bob")
	
	users := manager.GetAllUsers()
	
	if len(users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(users))
	}
	
	// Check that both users are in the list
	foundAlice := false
	foundBob := false
	
	for _, user := range users {
		if user.ID == alice.ID && user.Name == alice.Name {
			foundAlice = true
		}
		if user.ID == bob.ID && user.Name == bob.Name {
			foundBob = true
		}
	}
	
	if !foundAlice {
		t.Error("Alice not found in user list")
	}
	
	if !foundBob {
		t.Error("Bob not found in user list")
	}
}

func TestGetUserCount(t *testing.T) {
	manager := NewManager()
	
	if manager.GetUserCount() != 0 {
		t.Errorf("Expected 0 users initially, got %d", manager.GetUserCount())
	}
	
	manager.CreateUser("Alice")
	
	if manager.GetUserCount() != 1 {
		t.Errorf("Expected 1 user after creation, got %d", manager.GetUserCount())
	}
	
	user2 := manager.CreateUser("Bob")
	
	if manager.GetUserCount() != 2 {
		t.Errorf("Expected 2 users after second creation, got %d", manager.GetUserCount())
	}
	
	manager.RemoveUser(user2.ID)
	
	if manager.GetUserCount() != 1 {
		t.Errorf("Expected 1 user after removal, got %d", manager.GetUserCount())
	}
}

func TestUserExists(t *testing.T) {
	manager := NewManager()
	
	if manager.UserExists(1) {
		t.Error("User 1 should not exist initially")
	}
	
	user := manager.CreateUser("Alice")
	
	if !manager.UserExists(user.ID) {
		t.Error("User should exist after creation")
	}
	
	if manager.UserExists(999) {
		t.Error("Non-existent user ID should return false")
	}
}

func TestUpdateUserName(t *testing.T) {
	manager := NewManager()
	
	user := manager.CreateUser("Alice")
	
	err := manager.UpdateUserName(user.ID, "Alice Smith")
	if err != nil {
		t.Fatalf("Failed to update user name: %v", err)
	}
	
	updatedUser := manager.GetUser(user.ID)
	if updatedUser.Name != "Alice Smith" {
		t.Errorf("Expected updated name 'Alice Smith', got '%s'", updatedUser.Name)
	}
}

func TestUpdateNonExistentUserName(t *testing.T) {
	manager := NewManager()
	
	err := manager.UpdateUserName(999, "Non-existent")
	if err == nil {
		t.Error("Expected error when updating non-existent user")
	}
}

func TestUpdateUserColor(t *testing.T) {
	manager := NewManager()
	
	user := manager.CreateUser("Alice")
	originalColor := user.Color
	
	err := manager.UpdateUserColor(user.ID, "#123456")
	if err != nil {
		t.Fatalf("Failed to update user color: %v", err)
	}
	
	updatedUser := manager.GetUser(user.ID)
	if updatedUser.Color != "#123456" {
		t.Errorf("Expected updated color '#123456', got '%s'", updatedUser.Color)
	}
	
	if updatedUser.Color == originalColor {
		t.Error("Color should have changed")
	}
}

func TestUpdateNonExistentUserColor(t *testing.T) {
	manager := NewManager()
	
	err := manager.UpdateUserColor(999, "#123456")
	if err == nil {
		t.Error("Expected error when updating non-existent user color")
	}
}

func TestGetNextAvailableID(t *testing.T) {
	manager := NewManager()
	
	if manager.GetNextAvailableID() != 1 {
		t.Errorf("Expected next ID to be 1, got %d", manager.GetNextAvailableID())
	}
	
	user1 := manager.CreateUser("Alice")
	
	if manager.GetNextAvailableID() != 2 {
		t.Errorf("Expected next ID to be 2, got %d", manager.GetNextAvailableID())
	}
	
	user2 := manager.CreateUser("Bob")
	
	if manager.GetNextAvailableID() != 3 {
		t.Errorf("Expected next ID to be 3, got %d", manager.GetNextAvailableID())
	}
	
	// Remove a user - next ID should still be 3
	manager.RemoveUser(user1.ID)
	
	if manager.GetNextAvailableID() != 3 {
		t.Errorf("Expected next ID to still be 3 after removal, got %d", manager.GetNextAvailableID())
	}
	
	// Verify user2 still exists
	if !manager.UserExists(user2.ID) {
		t.Error("User2 should still exist")
	}
}

func TestColorGeneration(t *testing.T) {
	manager := NewManager()
	
	// Create enough users to test color cycling
	colors := make(map[string]bool)
	
	for i := 0; i < 25; i++ {
		user := manager.CreateUser(fmt.Sprintf("User%d", i))
		colors[user.Color] = true
	}
	
	// Should have multiple different colors
	if len(colors) < 5 {
		t.Errorf("Expected at least 5 different colors, got %d", len(colors))
	}
}