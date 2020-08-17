package user

import (
	"time"

	"github.com/authgear/authgear-server/pkg/lib/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
)

type User struct {
	ID          string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	LastLoginAt *time.Time
}

func newUserModel(
	user *User,
	identities []*identity.Info,
	isVerified bool,
) *model.User {
	isAnonymous := false
	for _, i := range identities {
		if i.Type == authn.IdentityTypeAnonymous {
			isAnonymous = true
			break
		}
	}

	return &model.User{
		ID:          user.ID,
		CreatedAt:   user.CreatedAt,
		LastLoginAt: user.LastLoginAt,
		IsAnonymous: isAnonymous,
		IsVerified:  isVerified,
	}
}