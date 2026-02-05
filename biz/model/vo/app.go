package vo

import "time"

type AppVo struct {
	ID           int64     `json:"id"`
	AppName      string    `json:"appName"`
	Cover        string    `json:"cover"`
	InitPrompt   string    `json:"initPrompt"`
	CodeGenType  string    `json:"codeGenType"`
	DeployKey    string    `json:"deployKey"`
	DeployedTime time.Time `json:"deployedTime"`
	Priority     int32     `json:"priority"`
	UserID       int64     `json:"userId"`
	CreateTime   time.Time `json:"createTime"`
	UpdateTime   time.Time `json:"updateTime"`
}
