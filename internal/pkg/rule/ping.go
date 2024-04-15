package rule

import (
	"Alarm/internal/pkg/mail"
	"Alarm/internal/web/models"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"
)

// Ping
// 功能：表示一个Ping监控任务，包含监控状态、工具集、资产信息、规则设定、告警设置以及间隔时间和邮件通知列表
type Ping struct {
	// 监控状态，包含正常计数、异常计数、总体状态、原因描述、关联ID及记录时间
	State State

	// 工具集，包含缓存、邮件服务和数据库连接对象
	tools tools

	// 资产信息
	asset_id   int    // 资产ID，整数类型，唯一标识被监控资产
	asset_name string // 资产名称，字符串类型，描述被监控资产的名称或别称
	address    string // 监控地址，字符串类型，提供被监控资产的网络可达地址（如IP地址）

	// 规则设定
	rule_id       int    // 规则ID，整数类型，唯一标识监控规则
	rule          string // 规则描述，字符串类型，说明监控规则的具体内容或条件
	health_limit  int    // 健康阈值，整数类型，连续健康检查次数达到此值时认为资产状态恢复
	wrong_limit   int    // 异常阈值，整数类型，连续异常检查次数达到此值时触发告警
	mode          int    // 检查模式，整数类型，定义Ping检查的执行方式（如：连续、间隔等）
	latency_limit int    // 延迟阈值，整数类型，最大允许的Ping响应延迟（单位：毫秒）
	lost_limit    int    // 丢包阈值，整数类型，最大允许的Ping请求丢包率（百分比）

	// 告警设置

	alarm_id int // 告警ID，整数类型，唯一标识告警策略
	interval int // 检查间隔，整数类型，定义两次异常持续之间的间隔时间（单位：秒）

	// 邮件通知
	mailto []string // 邮件接收人列表，字符串切片类型，存储接收告警通知的邮箱地址
}

// NewPing
// 功能：创建一个新的Ping监控任务实例，通过传入的资产规则ID获取相关配置信息，并初始化监控任务状态及工具集
//
// 参数：
//
//	id int // 资产规则ID，整数类型，用于查询所需配置信息
//	Rcp *models.Cache // 缓存对象，指向Cache实例，用于缓存数据
//	mail *mail.MailPool // 邮件服务对象，指向MailPool实例，用于发送邮件通知
//	db *models.Database // 数据库连接对象，指向Database实例，用于与数据库交互
//
// 返回值：
//
//	*Ping // 新创建的Ping监控任务实例，指针类型，指向Ping结构体
func NewPing(id int, Rcp *models.Cache, mail *mail.MailPool, db *models.Database) *Ping {
	// 初始化Ping监控任务结构体
	var p Ping

	// 设置工具集
	p.tools.db = db
	p.tools.Rcp = Rcp
	p.tools.mail = mail

	// 初始化监控状态，设置关联ID为传入的资产规则ID
	p.State.correlation_id = id

	// 从数据库中获取与资产规则ID对应的资产规则信息
	var ar models.AssetRule
	p.tools.db.Engine.Where("id = ?", id).Get(&ar)

	// 将资产规则信息中的资产ID、规则ID赋值给Ping监控任务
	p.asset_id = ar.AssetID
	p.rule_id = ar.RuleID

	// 从数据库中获取与规则ID对应的规则信息
	var r models.Rule
	p.tools.db.Engine.Where("id=?", p.rule_id).Get(&r)

	// 将规则信息中的规则名称、异常阈值、健康阈值、告警ID赋值给Ping监控任务
	p.rule = r.Name
	p.wrong_limit = r.DeclineLimit
	p.health_limit = r.RecoverLimit
	p.alarm_id = r.AlarmID

	// 从数据库中获取与规则ID对应的Ping规则信息
	var pi models.PingInfo
	p.tools.db.Engine.Where("id=?", p.rule_id).Get(&pi)

	// 将Ping规则信息中的延迟阈值、检查模式、丢包阈值赋值给Ping监控任务
	p.latency_limit = pi.LatencyLimit
	p.mode = pi.Mode
	p.lost_limit = pi.LostLimit

	if p.alarm_id != 0 { // 如果绑定了告警规则，即告警规则id不为0，则从数据库中获取告警模板信息
		// 从数据库中获取与告警ID对应的告警模板信息
		var at models.AlarmTemplate
		p.tools.db.Engine.Where("id=?", p.alarm_id).Get(&at)

		// 将告警模板信息中的检查间隔、邮件接收人列表赋值给Ping监控任务
		p.interval = at.Interval
		p.mailto = at.Mails
	} else { // 如果告警ID为0，则表示该监控任务没有关联告警模板，则将检查间隔、邮件接收人列表置为0
		p.interval = 0
		p.mailto = []string{}
	}

	// 从数据库中获取与资产ID对应的资产信息
	var a models.Asset
	p.tools.db.Engine.Where("id=?", p.asset_id).Get(&a)

	// 将资产信息中的资产名称、监控地址赋值给Ping监控任务
	p.asset_name = a.Name
	p.address = a.Address

	// 初始化监控状态的其他字段
	p.State.abn = 0           // 异常计数初始化为0
	p.State.nor = 0           // 正常计数初始化为0
	p.State.Status = 3        // 总体状态初始化为未知状态（假设3表示未知）
	p.State.reason = ""       // 原因描述初始化为空字符串
	p.State.time = time.Now() // 记录时间设置为当前时间

	// 返回新创建的Ping监控任务实例
	return &p
}

// Scan
// 功能：执行Ping监控任务的扫描操作，根据扫描结果更新监控状态，并在必要时发送告警邮件
//
// 参数：无
//
// 返回值：
//   error // 错误信息，若扫描过程中发生错误则返回相应的错误信息，否则返回nil

func (p *Ping) Scan() error {
	// 更新当前扫描时间
	p.State.time = time.Now()

	// 执行状态判断，返回布尔值表示当前监控状态是否符合预期
	s, err := p.Jude()

	if err != nil {
		// 若状态判断过程中发生错误，直接返回错误信息
		return err
	}

	// 根据当前总体状态（Status）和状态判断结果（s），更新监控状态
	switch p.State.Status {
	case 3: // 当前处于检测中状态
		if s {
			// 符合预期，正常计数递增
			p.State.nor++

			// 若正常计数达到健康阈值，重置异常计数并更新总体状态为正常
			if p.State.nor >= p.health_limit {
				p.State.abn = 0
				p.State.Status = 1
			}
		} else {
			// 不符合预期，异常计数递增
			p.State.abn++

			// 若异常计数达到异常阈值，重置正常计数并更新总体状态为异常，发送告警邮件
			if p.State.abn >= p.wrong_limit {
				p.State.nor = 0
				p.State.Status = 2
				p.Alarm()
			}
		}
	case 1: // 正常状态
		if s {
			// 符合预期，正常计数递增，重置异常计数
			p.State.nor++
			p.State.abn = 0
		} else {
			// 不符合预期，异常计数递增
			p.State.abn++

			// 若异常计数达到异常阈值，重置正常计数并更新总体状态为异常，发送告警邮件
			if p.State.abn >= p.wrong_limit {
				p.State.nor = 0
				p.State.Status = 2
				p.Alarm() // 触发告警
			}
		}
	case 2: // 异常状态
		if s {
			// 符合预期，正常计数递增
			p.State.nor++

			// 若正常计数达到健康阈值，重置异常计数并更新总体状态为正常，发送告警邮件
			if p.State.nor >= p.health_limit {
				p.State.abn = 0
				p.State.Status = 1
				p.Alarm()
			}
		} else {
			// 不符合预期，重置正常计数，异常计数递增

			p.State.nor = 0
			p.State.abn++

			// 每隔一定间隔（interval）发送一次告警邮件
			if p.interval > 0 && (p.State.abn-p.wrong_limit)%p.interval == 0 {
				p.Alarm() // 触发告警
			}
		}
	}

	// 打印当前正常计数、异常计数和总体状态
	log.Println(p.State.nor, p.State.abn, p.State.Status)

	// 扫描操作完成，返回nil表示无错误
	return nil
}

// Jude
// 功能：对Ping监控任务的监控对象进行状态判断，根据返回结果更新监控状态的reason字段，并返回判断结果及可能的错误信息
//
// 参数：无
//
// 返回值：
//
//	bool // 判断结果，若监控对象状态符合预期，则返回true，否则返回false
//	error // 错误信息，若判断过程中发生错误则返回相应的错误信息，否则返回nil
func (p *Ping) Jude() (bool, error) {
	// 打印开始进行ping操作的信息，包括监控对象地址
	log.Println("ping", p.address)

	// 通过RCP客户端发送GET请求，获取与correlation_id关联的数据
	res, err := p.tools.Rcp.Client.Get(fmt.Sprintf("%d", p.State.correlation_id)).Bytes()

	if err != nil {
		// 若请求过程中发生错误，返回false及错误信息
		return false, err
	}

	// 将请求结果解析为map类型数据
	var data map[string]interface{}
	json.Unmarshal(res, &data)

	// 若请求成功（status为"success"），进一步处理结果
	if data["status"] == "success" {
		// 提取结果数据中的latency（响应时间）和package_loss_rate（丢包率）
		result := data["result"].(map[string]interface{})
		rl := result["latency"].([]interface{})[0].(float64)
		rp := result["package_loss_rate"].([]interface{})[0].(float64)

		// 根据监控任务的检查模式进行状态判断
		if p.mode == 1 { // 同时错误

			// 若响应时间小于等于延迟阈值且丢包率小于等于丢包阈值，返回true
			if rl < float64(p.latency_limit) && rp < float64(p.lost_limit) {
				return true, nil
			} else {
				// 否则更新监控状态的reason字段，记录不符合预期的原因，并返回false
				p.State.reason = fmt.Sprintf("响应时间大于等于%d ms（%.2f），丢包率大于等于%d %%（%.2f%%）", p.latency_limit, rl, p.lost_limit, rp)
				return false, nil
			}

		} else { // 任一错误

			// 若响应时间小于延迟阈值或丢包率小于丢包阈值，返回true
			if rl < float64(p.latency_limit) || rp < float64(p.lost_limit) {
				return true, nil
			} else {
				// 否则根据具体情况更新监控状态的reason字段，记录不符合预期的原因，并返回false
				if rl >= float64(p.latency_limit) {
					p.State.reason = fmt.Sprintf("响应时间大于等于%d ms（%.2f）", p.latency_limit, rl)
				} else {
					p.State.reason = fmt.Sprintf("丢包率大于等于%d %%（%.2f%%）", p.lost_limit, rp)
				}
				return false, nil
			}
		}

	} else { // 请求失败

		// 返回false及错误信息（"respone error"）
		return false, errors.New("respone error")

	}
}

// Alarm
// 功能：根据当前监控状态发送告警邮件，邮件内容根据状态变化情况动态生成
//
// 参数：无
//
// 返回值：无
func (p *Ping) Alarm() {
	// 根据当前监控状态确定邮件主题和内容
	if p.alarm_id != 0 {
		var subject, message string
		if p.State.Status == 3 { //异常结束
			subject = fmt.Sprintf("【告警】%s资产-【规则】-异常结束", p.asset_name)
			message = fmt.Sprintf(`告警类型：PING检测<br>
		告警节点：异常结束<br>
		告警资产：%s<br>
		资产地址：%s<br>
		检测规则：%s<br>
		告警内容：<br>
		&nbsp&nbsp&nbsp&nbsp该资产监控出现变更，本次告警中止并解除<br>
		告警时间：%s<br><br>`, p.asset_name,
				p.address, p.rule, p.State.time.Format("2006-01-02 15:04:05"))

		} else if p.State.abn == p.wrong_limit { //异常触发
			subject = fmt.Sprintf("【告警】%s资产-【规则】-异常触发", p.asset_name)
			message = fmt.Sprintf(`告警类型：PING检测<br>
		告警节点：异常触发<br>
		告警资产：%s<br>
		资产地址：%s<br>
		检测规则：%s<br>
		告警内容：<br>
		&nbsp&nbsp&nbsp&nbsp%s<br>
		告警时间：%s<br><br>
		该资产在此规则监控下触发异常，请尽快处理！`, p.asset_name,
				p.address, p.rule, p.State.reason, p.State.time.Format("2006-01-02 15:04:05"))

		} else if p.State.abn > p.wrong_limit { //异常持续
			subject = fmt.Sprintf("【告警】%s资产-【规则】-异常持续", p.asset_name)
			message = fmt.Sprintf(`告警类型：PING检测<br>
		告警节点：异常持续<br>
		告警资产：%s<br>
		资产地址：%s<br>
		检测规则：%s<br>
		告警内容：<br>
		&nbsp&nbsp&nbsp&nbsp%s<br>
		告警时间：%s<br><br>
		该资产在此规则监控下处于异常持续中，请尽快处理！`, p.asset_name,
				p.address, p.rule, p.State.reason, p.State.time.Format("2006-01-02 15:04:05"))

		} else if p.State.nor > 0 { //异常结束
			subject = fmt.Sprintf("【告警解除】%s资产-【规则】-异常恢复", p.asset_name)
			message = fmt.Sprintf(`告警类型：PING检测<br>
		告警节点：异常终止<br>
		告警资产：%s<br>
		资产地址：%s<br>
		检测规则：%s<br>
		告警内容：<br>
		&nbsp&nbsp&nbsp&nbsp%s<br>
		告警时间：%s<br><br>
		该资产在此规则监控下解除异常，告警结束`, p.asset_name,
				p.address, p.rule, p.State.reason, p.State.time.Format("2006-01-02 15:04:05"))
		}
		// 发送邮件，使用邮件服务对象和邮件接收人列表
		err := p.tools.mail.SendMail(subject, p.mailto, []string{}, []string{}, message, []string{})
		// 若发送邮件过程中发生错误，打印错误信息
		if err != nil {
			log.Println(err)
		}
	}

	// 保存告警日志和更新资产状态
	p.Save()

}

// Save
// 功能：保存告警日志和更新资产状态
//
// 参数：无
//
// 返回值：无
func (p *Ping) Save() {
	// 更新资产状态，将当前监控状态的Status字段写入数据库中对应资产的State字段
	_, err := p.tools.db.Engine.Where("id=?", p.asset_id).Cols("state").Update(&models.Asset{
		State: p.State.Status,
	})

	// 若更新过程中发生错误，打印错误信息
	if err != nil {
		log.Println(err)
	}

	// 存储告警日志，构建Mail结构体数组和AlarmLog结构体
	var m []models.Mail
	for i := 0; i < len(p.mailto); i++ {
		m = append(m, models.Mail{
			State:   true,
			Address: p.mailto[i],
		})
	}
	var alarmstate int
	//计算异常状态
	if p.State.Status == 3 {
		alarmstate = 3
	} else if p.State.abn == p.wrong_limit {
		alarmstate = 1
	} else if p.State.abn > p.wrong_limit {
		alarmstate = 2
	} else if p.State.nor > 0 {
		alarmstate = 3
	}
	// 存储告警日志
	_, err = p.tools.db.Engine.InsertOne(models.AlarmLog{
		AssetID:   p.asset_id,
		RuleID:    p.rule_id,
		State:     alarmstate,
		Mails:     m,
		Messages:  []string{p.State.reason},
		CreatedAt: p.State.time,
	})

	// 若插入过程中发生错误，打印错误信息
	if err != nil {
		log.Println(err)
	}
}

// Update
// 功能：更新监控状态为未知状态（Status=3），并发送告警终止邮件
//
// 参数：无
//
// 返回值：无
func (p *Ping) Update() {
	// 更新监控状态为未知状态
	p.State.Status = 3
	// 发送告警邮件
	if p.alarm_id != 0 { // 如果告警模板ID不为0，则发送告警邮件
		p.Alarm()
	}

}
