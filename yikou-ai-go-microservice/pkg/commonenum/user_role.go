package commonenum

// UserRoleEnum 用户角色枚举
type UserRoleEnum string

const (
	UserRole  UserRoleEnum = "user"
	AdminRole UserRoleEnum = "admin"
)

var RoleTextMap = map[UserRoleEnum]string{
	UserRole:  "用户",
	AdminRole: "管理员",
}
