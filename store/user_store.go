package store

import (
	"fmt"
	"sync"
)

type User struct {
	ChatId       int64
	IsConnecting bool
	IsConnected  bool
	Partner      int64
	SelectedInterests []string
}

type UserStore struct {
	Mu    sync.Mutex
	Users map[int64]*User
}

func (us *UserStore) AddUser(user *User) {
	us.Mu.Lock()
	defer us.Mu.Unlock()
	// Ensure SelectedInterests is initialized
	if user.SelectedInterests == nil {
		user.SelectedInterests = make([]string, 0)
	}
	us.Users[user.ChatId] = user
}

func (us *UserStore) RemoveUser(chatId int64) {
	us.Mu.Lock()
	defer us.Mu.Unlock()
	delete(us.Users, chatId)
}

func (us *UserStore) GetUser(chatId int64) (*User, bool) {
	us.Mu.Lock()
	defer us.Mu.Unlock()
	user, exists := us.Users[chatId]
	return user, exists
}

func (us *UserStore) FindMatch(excludeChatId int64) (*User, bool) {
	us.Mu.Lock()
	defer us.Mu.Unlock()
	for _, user := range us.Users {
		if user.IsConnecting && user.ChatId != excludeChatId {
			return user, true
		}
	}
	return nil, false
}

// SetUserInterests updates the selected interests for a user.
func (us *UserStore) SetUserInterests(chatId int64, interests []string) error {
	us.Mu.Lock()
	defer us.Mu.Unlock()
	user, exists := us.Users[chatId]
	if !exists {
		return fmt.Errorf("user with chatId %d not found", chatId)
	}
	user.SelectedInterests = interests
	return nil
}

// GetUserInterests retrieves the selected interests for a user.
func (us *UserStore) GetUserInterests(chatId int64) ([]string, error) {
	us.Mu.Lock()
	defer us.Mu.Unlock()
	user, exists := us.Users[chatId]
	if !exists {
		return nil, fmt.Errorf("user with chatId %d not found", chatId)
	}
	// Return a copy to prevent external modification of the slice
	interestsCopy := make([]string, len(user.SelectedInterests))
	copy(interestsCopy, user.SelectedInterests)
	return interestsCopy, nil
}

// ClearUserInterests clears the selected interests for a user.
func (us *UserStore) ClearUserInterests(chatId int64) error {
	us.Mu.Lock()
	defer us.Mu.Unlock()
	user, exists := us.Users[chatId]
	if !exists {
		return fmt.Errorf("user with chatId %d not found", chatId)
	}
	user.SelectedInterests = make([]string, 0) // Set to an empty slice
	return nil
}
