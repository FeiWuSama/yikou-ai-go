package vo

import "time"

type LoginUserVo struct {
	ID          int64     `json:"id"`
	UserAccount string    `json:"user_account"`
	UserName    string    `json:"user_name"`
	UserAvatar  string    `json:"user_avatar"`
	UserProfile string    `json:"user_profile"`
	UserRole    string    `json:"user_role"`
	CreateTime  time.Time `json:"create_time"`
	UpdateTime  time.Time `json:"update_time"`
}
