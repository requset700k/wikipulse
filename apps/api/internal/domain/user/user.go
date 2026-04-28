// Package user는 사용자 도메인 타입을 정의한다.
// Role은 Keycloak realm role과 대응: student / instructor / admin.
// ADR docs/ADR/keycloak-rbac.md 참조.
package user

type Role string

const (
	RoleStudent    Role = "student"
	RoleInstructor Role = "instructor"
	RoleAdmin      Role = "admin"
)

type User struct {
	ID     string `json:"id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
	Role   Role   `json:"role"`
	Points int    `json:"points"`
}
