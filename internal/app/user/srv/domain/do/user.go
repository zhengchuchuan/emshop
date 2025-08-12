package do

import (
	"emshop/pkg/db"
	"time"
)

/*
1. 密文 2. 密文不可反解
 1. 对称加密
 2. 非对称加密
 3. md5 信息摘要算法
    密码如果不可以反解，用户找回密码
*/
type UserDO struct {
	db.BaseModel
	Mobile   string     `gorm:"index:idx_mobile;unique;type:varchar(11);not null"`
	Password string     `gorm:"type:varchar(100);not null"`
	NickName string     `gorm:"type:varchar(20)"`
	Birthday *time.Time `gorm:"type:datetime"`
	Gender   string     `gorm:"column:gender;default:male;type:varchar(6) comment 'female表示女, male表示男'"`
	Role     int        `gorm:"column:role;default:1;type:int comment '1表示普通用户, 2表示管理员'"`
}

func (u *UserDO) TableName() string {
	return "user"
}

// 用于返回用户列表 
// 在分页查询时后端通常只返回当前页的数据（如 10 条），
// 但前端需要知道总共有多少条数据（如 1000 条），
// 以便正确显示分页控件（如总页数、跳转页码等）
type UserDOList struct {
	TotalCount int64      `json:"totalCount,omitempty"` //总数
	Items      []*UserDO `json:"data"`                 //数据
}



