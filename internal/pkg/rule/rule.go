package rule

import (
	"Alarm/internal/pkg/mail"
	"Alarm/internal/web/models"
	"time"
)

// Rule
// 功能：定义一个接口，规定了Scan和Update两个方法，供实现该接口的具体规则类使用
type Rule interface {
	Scan() error // 执行扫描操作，返回错误信息（如有）
	Update()     // 更新规则状态
}

// State
// 功能：表示一个监控状态，包括正常计数、异常计数、总体状态、原因描述、关联ID以及记录时间
type State struct {
	nor            int       // 正常计数，整数类型，记录符合预期的检查次数
	abn            int       // 异常计数，整数类型，记录不符合预期的检查次数
	Status         int       // 总体状态，整数类型，表示当前监控状态（如：正常、异常等）
	reason         string    // 原因描述，字符串类型，说明当前状态的原因或详细信息
	correlation_id int       // 关联ID，整数类型，用于关联具体的监控任务或规则
	time           time.Time // 记录时间，Time类型，记录最后一次状态更新的时间点
}

// tools
// 功能：封装一些通用工具，如缓存、邮件服务和数据库连接
type tools struct {
	Rcp  *models.Cache    // 缓存对象，指向Cache实例，用于缓存数据
	mail *mail.MailPool   // 邮件服务对象，指向MailPool实例，用于发送邮件通知
	db   *models.Database // 数据库连接对象，指向Database实例，用于与数据库交互
}
