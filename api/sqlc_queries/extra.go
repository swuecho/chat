package sqlc_queries

func (user *AuthUser) Role() string {
	role := "user"
	if user.IsSuperuser {
		role = "admin"
	}
	return role
}
