package store

import "sync"

type User struct {
	ChatId       int64
	IsConnecting bool
	IsConnected  bool
	Partner      int64

	// State for conversational commands
	State        string
	Gender       string
	Preference   string
	Interests    []string
	BlockedUsers map[int64]struct{}
}

type UserStore struct {
	Mu    sync.Mutex
	Users map[int64]*User
}

func (u *UserStore) AddUser(user *User) {
	u.Mu.Lock()
	defer u.Mu.Unlock()
	u.Users[user.ChatId] = user
}

func (u *UserStore) RemoveUser(chatId int64) {
	u.Mu.Lock()
	defer u.Mu.Unlock()
	delete(u.Users, chatId)
}

func (u *UserStore) GetUser(chatId int64) (*User, bool) {
	u.Mu.Lock()
	defer u.Mu.Unlock()
	user, exists := u.Users[chatId]
	return user, exists
}

// FindMatch iterates through waiting users to find the best compatible partner.
func (u *UserStore) FindMatch(currentUser *User) (*User, bool) {
	u.Mu.Lock()
	defer u.Mu.Unlock()

	var bestMatch *User
	maxSharedInterests := -1

	for _, potentialPartner := range u.Users {
		// Skip if it's the same user or if the other user is not looking for a chat
		if potentialPartner.ChatId == currentUser.ChatId || !potentialPartner.IsConnecting {
			continue
		}

		// Check if either user has blocked the other
		if _, blocked := currentUser.BlockedUsers[potentialPartner.ChatId]; blocked {
			continue
		}
		if _, blocked := potentialPartner.BlockedUsers[currentUser.ChatId]; blocked {
			continue
		}

		// First, check for compatible gender preferences. This is a strict requirement.
		currentUserLikesPartner := currentUser.Preference == PrefAny || currentUser.Preference == potentialPartner.Gender
		partnerLikesCurrentUser := potentialPartner.Preference == PrefAny || potentialPartner.Preference == currentUser.Gender

		if !currentUserLikesPartner || !partnerLikesCurrentUser {
			continue // Gender preferences don't match, so skip this user.
		}

		// Now, calculate the match score based on shared interests.
		sharedCount := 0
		if len(currentUser.Interests) > 0 && len(potentialPartner.Interests) > 0 {
			interestSet := make(map[string]struct{})
			for _, interest := range currentUser.Interests {
				interestSet[interest] = struct{}{}
			}
			for _, interest := range potentialPartner.Interests {
				if _, found := interestSet[interest]; found {
					sharedCount++
				}
			}
		}

		// If this partner has more shared interests than the best one so far, they become the new best match.
		if sharedCount > maxSharedInterests {
			maxSharedInterests = sharedCount
			bestMatch = potentialPartner
		}
	}

	if bestMatch != nil {
		return bestMatch, true
	}

	return nil, false
}
