package vo

import "time"

type UserVo struct {
	ID          int64     `json:"id,string"`
	UserAccount string    `json:"user_account"`
	UserName    string    `json:"user_name"`
	UserAvatar  string    `json:"user_avatar"`
	UserProfile string    `json:"user_profile"`
	UserRole    string    `json:"user_role"`
	CreateTime  time.Time `json:"create_time"`
	UpdateTime  time.Time `json:"update_time"`
}
