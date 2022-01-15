package user

import (
	"ohmnyom/domain/user"
	"ohmnyom/internal/time"
)

var users = []*user.User{
	{
		Id:        "id-testuser1",
		Email:     "email-testuser1@test.com",
		Name:      "name-testuser1",
		Password:  "password-testuser1",
		Photourl:  "photourl-testuser1",
		OAuthType: user.OAuthType_NONE,
		OAuthId:   "",
		SignedUp:  time.Date(2006, 1, 2, 15, 4, 5, 987, time.UTC),
		Pets:      []string{"pet1-1", "pet1-2", "pet1-3"},
	},
	{
		Id:        "id-testuser2",
		Email:     "email-testuser2@test.com",
		Name:      "name-testuser2",
		Photourl:  "photourl-testuser2",
		OAuthType: user.OAuthType_GOOGLE,
		OAuthId:   "oauthid-testuser2",
		SignedUp:  time.Date(2007, 2, 3, 16, 5, 6, 876, time.UTC),
		Pets:      []string{"pet2-1", "pet2-2"},
	},
	{
		Id:        "id-testuser3",
		Email:     "email-testuser3@test.com",
		Name:      "name-testuser3",
		Password:  "password-testuser3",
		Photourl:  "photourl-testuser3",
		OAuthType: user.OAuthTYpe_KAKAO,
		OAuthId:   "oauthid-testuser3",
		SignedUp:  time.Date(2008, 3, 4, 17, 6, 7, 765, time.UTC),
		Pets:      nil,
	},
}
