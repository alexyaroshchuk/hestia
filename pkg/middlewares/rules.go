package middlewares

func AccessibleRoles() map[string][]string {
	return map[string][]string{
		"/api/v1/users": {"admin"},
		"/api/v1/flats": {"admin", "user"},
	}
}
