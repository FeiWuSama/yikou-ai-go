package vo

import "time"

type AppVo struct {
	ID           int64     `json:"id,string"`
	AppName      string    `json:"appName"`
	Cover        string    `json:"cover"`
	InitPrompt   string    `json:"initPrompt"`
	CodeGenType  string    `json:"codeGenType"`
	DeployKey    string    `json:"deployKey"`
	DeployedTime time.Time `json:"deployedTime"`
	Priority     int32     `json:"priority"`
	UserID       int64     `json:"userId,string"`
	User         UserVo    `json:"user"`
	CreateTime   time.Time `json:"createTime"`
	UpdateTime   time.Time `json:"updateTime"`
}

type UserVo struct {
	ID          int64     `json:"id,string"`
	UserAccount string    `json:"userAccount"`
	UserName    string    `json:"userName"`
	UserAvatar  string    `json:"userAvatar"`
	UserProfile string    `json:"userProfile"`
	UserRole    string    `json:"userRole"`
	CreateTime  time.Time `json:"createTime"`
	UpdateTime  time.Time `json:"updateTime"`
}
