package enum

// UserRole 用户角色枚举
const (
	UserRole  = "user"
	AdminRole = "admin"
)

var RoleTextMap = map[string]string{
	UserRole:  "用户",
	AdminRole: "管理员",
}
