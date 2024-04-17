package rule

import (
	"Alarm/internal/pkg/mail"
	"Alarm/internal/web/models"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

var PortCodeStatus = map[int]string{-1: "错误", 0: "开启", 1: "超时", 2: "关闭"}

// Tcp
// 功能：用于TCP监控任务的结构体，包含监控状态、工具集、资产信息、规则设定、邮件接收人等字段
//
// 字段：
//
//	State State  // 当前监控状态
//	tools tools  // 工具集，包含各种辅助方法和接口
//	asset_id int  // 资产ID
//	asset_name string // 资产名称
//	address string  // 监控对象地址
//	rule_id int    // 规则ID
//	rule string    // 规则描述
//	health_limit int  // 健康阈值，用于判断监控状态
//	wrong_limit int   // 异常阈值，用于判断监控状态
//	alarm_id int     // 告警ID
//	interval int     // 间隔时间，用于控制监控频率
//	mailto []string  // 邮件接收人列表
//	enable_ports string // 可用端口列表（字符串形式）
//	disable_ports string // 不可用端口列表（字符串形式）
//	portsetting map[string]int // 端口设置，键为端口名，值为端口号
//	hisopenerr map[string]bool // 历史打开错误，键为端口名，值为是否发生过错误
//	hiscloseerr map[string]bool // 历史关闭错误，键为端口名，值为是否发生过错误
//	nowopenerr []string  // 当前打开错误端口列表
//	nowcloseerr []string // 当前关闭错误端口列表
//	nowopentimeout []string // 当前打开超时端口列表
//	nowclosetimeout []string // 当前关闭超时端口列表
type Tcp struct {
	State           State           // 当前监控状态
	tools           tools           // 工具集
	asset_id        int             // 资产ID
	asset_name      string          // 资产名称
	address         string          // 监控对象地址
	rule_id         int             // 规则ID
	rule            string          // 规则描述
	health_limit    int             // 健康阈值，用于判断监控状态
	wrong_limit     int             // 异常阈值，用于判断监控状态
	alarm_id        int             // 告警ID
	interval        int             // 间隔时间，用于控制监控频率
	mailto          []string        // 邮件接收人列表
	enable_ports    string          // 可用端口列表（字符串形式）
	disable_ports   string          // 不可用端口列表（字符串形式）
	portsetting     map[string]int  // 端口设置，键为端口名，值为端口号
	hisopenerr      map[string]bool // 历史打开错误，键为端口名，值为是否发生过错误
	hiscloseerr     map[string]bool // 历史关闭错误，键为端口名，值为是否发生过错误
	nowopenerr      []string        // 当前打开错误端口列表
	nowcloseerr     []string        // 当前关闭错误端口列表
	nowopentimeout  []string        // 当前打开超时端口列表
	nowclosetimeout []string        // 当前关闭超时端口列表
}

// NewTcp
// 功能：创建一个新的TCP监控任务实例，初始化其属性值并返回指针
//
// 参数：
//   id int          // TCP监控任务的唯一标识，用于查询相关配置信息
//   Rcp *models.Cache  // RCP缓存对象，提供与RCP服务交互的能力
//   mail *mail.MailPool // 邮件池对象，提供邮件发送功能
//   db *models.Database // 数据库连接对象，提供查询和存储监控任务相关配置信息的能力

// 返回值：
//
//	*Tcp // 新创建的TCP监控任务实例指针
func NewTcp(id int, Rcp *models.Cache, mail *mail.MailPool, db *models.Database) *Tcp {
	// 创建一个空的TCP监控任务结构体
	var p Tcp

	// 初始化工具集属性
	p.tools.db = db
	p.tools.Rcp = Rcp
	p.tools.mail = mail

	// 初始化监控状态属性
	p.State.correlation_id = id

	// 从数据库中获取与监控任务关联的资产规则信息
	var ar models.AssetRule
	p.tools.db.Engine.Where("id = ?", id).Get(&ar)

	// 将资产规则中的资产ID和规则ID赋值给监控任务
	p.asset_id = ar.AssetID
	p.rule_id = ar.RuleID

	// 从数据库中获取与监控任务关联的规则信息
	var r models.Rule
	p.tools.db.Engine.Where("id=?", p.rule_id).Get(&r)

	// 将规则名称、异常阈值、健康阈值、告警ID等信息赋值给监控任务
	p.rule = r.Name
	p.wrong_limit = r.DeclineLimit
	p.health_limit = r.RecoverLimit
	p.alarm_id = r.AlarmID

	// 从数据库中获取与监控任务关联的TCP信息
	var ti models.TCPInfo
	p.tools.db.Engine.Where("id=?", p.rule_id).Get(&ti)

	// 将启用端口列表、禁用端口列表信息赋值给监控任务
	p.enable_ports = ti.EnablePorts
	p.disable_ports = ti.DisablePorts

	// 从数据库中获取与监控任务关联的资产信息
	var a models.Asset
	p.tools.db.Engine.Where("id=?", p.asset_id).Get(&a)

	// 将资产名称、资产地址信息赋值给监控任务
	p.asset_name = a.Name
	p.address = a.Address

	// 构建端口设置映射表，键为端口号，值为初始设置值（0表示开启，2表示禁用）
	m := map[string]int{}
	for _, i := range strings.Split(p.enable_ports, ",") {
		m[i] = 0
	}
	for _, i := range strings.Split(p.disable_ports, ",") {
		m[i] = 2
	}
	p.portsetting = m

	// 从数据库中获取与监控任务关联的告警模板信息
	if p.alarm_id != 0 { // 如果绑定了告警规则，即告警规则id不为0，则从数据库中获取告警模板信息
		// 从数据库中获取与告警ID对应的告警模板信息
		var at models.AlarmTemplate
		p.tools.db.Engine.Where("id=?", p.alarm_id).Get(&at)

		// 将告警模板信息中的检查间隔、邮件接收人列表赋值给Tcp监控任务
		p.interval = at.Interval
		p.mailto = at.Mails
	} else { // 如果告警ID为0，则表示该监控任务没有关联告警模板，则将检查间隔、邮件接收人列表置为0
		p.interval = 0
		p.mailto = []string{}
	}

	// 初始化监控状态的其他属性
	p.State.abn = 0
	p.State.nor = 0
	p.State.Status = 3
	p.State.reason = ""
	p.State.time = time.Now()

	// 初始化历史错误映射表
	p.hiscloseerr = make(map[string]bool)
	p.hisopenerr = make(map[string]bool)

	// 返回新创建的TCP监控任务实例指针
	return &p
}

// Scan
// 功能：执行TCP监控任务的扫描操作，根据扫描结果更新监控状态，并在必要时触发告警
//
// 参数：无
//
// 返回值：
//
//	error // 执行过程中发生的任何错误；若无错误，则返回nil
func (p *Tcp) Scan() error {
	// 更新当前扫描时间
	//log.Println(p.State.nor, p.State.abn, p.State.Status, p.nowcloseerr)
	p.State.time = time.Now()

	// 执行状态判断，返回布尔值表示当前监控状态是否符合预期
	s, err := p.Jude()

	if err != nil {
		// 若状态判断过程中发生错误，直接返回错误信息

		return err
	}

	// 根据当前总体状态（Status）和状态判断结果（s），更新监控状态
	switch p.State.Status {
	case 3: //当前处于检测中状态
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
				p.Alarm() // 触发告警
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
				//根据历史端口错误数据编写恢复的原因
				if len(p.hiscloseerr) != 0 {
					sc := []string{}
					for i := range p.hiscloseerr {
						sc = append(sc, i)
					}
					p.hiscloseerr = map[string]bool{}
					p.State.reason += fmt.Sprintf("预期关闭的端口%s处于关闭状态 ", strings.Join(sc, ","))

				}
				if len(p.hisopenerr) != 0 {
					so := []string{}
					for i := range p.hisopenerr {
						so = append(so, i)
					}
					p.hisopenerr = map[string]bool{}
					p.State.reason += fmt.Sprintf("预期开启的端口%s处于开启状态 ", strings.Join(so, ","))

				}

				p.Alarm() // 触发告警

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

	//log.Println(p.State.nor, p.State.abn, p.State.Status, p.interval)
	// 扫描操作完成，返回nil表示无错误
	return nil
}

// Jude
// 功能：判断TCP监控任务中资产端口的实际状态是否符合预期，并返回判断结果及可能发生的错误
//
// 参数：无
//
// 返回值：
//
//	bool // 判断结果，true表示所有端口状态符合预期，false表示存在端口状态不符合预期
//	error // 执行过程中发生的任何错误；若无错误，则返回nil
func (p *Tcp) Jude() (bool, error) {
	// 重置状态信息
	p.State.reason = ""
	p.nowcloseerr = []string{}
	p.nowopenerr = []string{}
	p.nowopentimeout = []string{}
	p.nowclosetimeout = []string{}

	// 初始化成功标志，开始时假设无错误
	var flag = true

	// 发起远程调用获取资产端口状态数据
	res, err := p.tools.Rcp.Client.Get(fmt.Sprintf("%d", p.State.correlation_id)).Bytes()
	if err != nil {
		return false, err
	}

	// 解析响应数据为JSON对象
	var data map[string]interface{}
	json.Unmarshal(res, &data)

	// 检查响应状态，若成功则继续处理端口状态数据
	if data["status"] == "success" {
		result := data["result"].(map[string]interface{})
		targets := data["target"].([]interface{})
		portstatus := result["portstatus"].([]interface{})

		// 遍历目标端口范围，检查每个端口的实际状态是否符合预期
		for i := 0; i < len(targets); i++ {
			target := strings.Split(targets[i].(string), p.address+":")[1]
			portstate := portstatus[i].([]interface{})

			// 获取目标端口范围的最小端口号
			minport, err := strconv.Atoi(strings.Split(target, "-")[0])
			if err != nil {
				return false, err
			}

			// 遍历端口状态列表，对比实际状态与预期状态
			for j := 0; j < len(portstate); j++ {
				if int(portstate[j].(float64)) != p.portsetting[target] {
					// 实际状态与预期状态不符，记录错误信息并更新成功标志
					if p.portsetting[target] == 0 {
						if int(portstate[0].(float64)) == 1 { //预期开启端口超时
							p.nowopentimeout = append(p.nowopentimeout, fmt.Sprintf("%d", minport+j))
							p.hisopenerr[fmt.Sprintf("%d", minport+j)] = true
						} else { //预期开启端口超时
							p.nowopenerr = append(p.nowopenerr, fmt.Sprintf("%d", minport+j))
							p.hisopenerr[fmt.Sprintf("%d", minport+j)] = true
						}
					} else { //预期关闭端口超时
						if int(portstate[0].(float64)) == 1 {
							p.nowclosetimeout = append(p.nowclosetimeout, fmt.Sprintf("%d", minport+j))
							p.hiscloseerr[fmt.Sprintf("%d", minport+j)] = true
						} else { //预期关闭端口开启
							p.nowcloseerr = append(p.nowcloseerr, fmt.Sprintf("%d", minport+j))
							p.hiscloseerr[fmt.Sprintf("%d", minport+j)] = true
						}
					}
					flag = false
				}
			}
		}

		// 如果存在端口状态不符合预期，构建错误原因字符串
		if !flag {
			if len(p.nowopenerr) != 0 {
				p.State.reason += fmt.Sprintf("预期开启的端口%s处于关闭状态 ", strings.Join(p.nowopenerr, ","))
			}
			if len(p.nowopentimeout) != 0 {
				p.State.reason += fmt.Sprintf("预期开启的端口%s处于超时状态 ", strings.Join(p.nowopentimeout, ","))
			}
			if len(p.nowcloseerr) != 0 {
				p.State.reason += fmt.Sprintf("预期关闭的端口%s处于开启状态 ", strings.Join(p.nowcloseerr, ","))
			}
			if len(p.nowclosetimeout) != 0 {
				p.State.reason += fmt.Sprintf("预期关闭的端口%s处于超时状态 ", strings.Join(p.nowclosetimeout, ","))
			}
		}

		// 返回判断结果和错误（此处无错误，返回nil）
		return flag, nil
	} else {
		// 响应状态非成功，返回错误信息
		return false, errors.New("response error")
	}
}

// Alarm
// 功能：根据当前TCP监控任务状态触发相应的告警通知，并保存状态信息
//
// 参数：无
//
// 返回值：无
func (p *Tcp) Alarm() {
	//log.Println("要发邮件",p.alarm_id)
	// 如果已设置告警ID（表示已配置告警），则进行告警通知
	if p.alarm_id > 0 {
		// 根据当前状态构建告警主题、消息及接收人列表
		var subject, message string
		var to []string
		if p.State.Status == 3 { // 异常结束
			subject = fmt.Sprintf("【告警】%s资产-【规则】-异常结束", p.asset_name)
			message = fmt.Sprintf(`告警类型：TCP检测<br>
		告警节点：异常结束<br>
		告警资产：%s<br>
		资产地址：%s<br>
		检测规则：%s<br>
		告警内容：<br>
		&nbsp&nbsp&nbsp&nbsp该资产监控出现变更，本次告警中止并解除<br>
		告警时间：%s<br><br>`, p.asset_name,
				p.address, p.rule, p.State.time.Format("2006-01-02 15:04:05"))
		} else if p.State.abn == p.wrong_limit { //告警次数达到告警限制,异常触发
			subject = fmt.Sprintf("【告警】%s资产-【规则】-异常触发", p.asset_name)
			message = fmt.Sprintf(`告警类型：TCP端口探测<br>
		告警节点：异常触发<br>
		告警资产：%s<br>
		资产地址：%s<br>
		检测规则：%s<br>
		告警内容：<br>
		&nbsp&nbsp&nbsp&nbsp%s<br>
		告警时间：%s<br><br>
		该资产在此规则监控下触发异常，请尽快处理！`, p.asset_name,
				p.address, p.rule, p.State.reason, p.State.time.Format("2006-01-02 15:04:05"))
			to = p.mailto
		} else if p.State.abn > p.wrong_limit { //告警持续,异常持续
			subject = fmt.Sprintf("【告警】%s资产-【规则】-异常持续", p.asset_name)
			message = fmt.Sprintf(`告警类型：TCP端口探测<br>
		告警节点：异常持续<br>
		告警资产：%s<br>
		资产地址：%s<br>
		检测规则：%s<br>
		告警内容：<br>
		&nbsp&nbsp&nbsp&nbsp%s<br>
		告警时间：%s<br><br>
		该资产在此规则监控下处于异常持续中，请尽快处理！`, p.asset_name,
				p.address, p.rule, p.State.reason, p.State.time.Format("2006-01-02 15:04:05"))
			to = p.mailto
		} else if p.State.nor > 0 { //告警解除,异常恢复
			subject = fmt.Sprintf("【告警解除】%s资产-【规则】-异常恢复", p.asset_name)
			message = fmt.Sprintf(`告警类型：TCP端口探测<br>
		告警节点：异常恢复<br>
		告警资产：%s<br>
		资产地址：%s<br>
		检测规则：%s<br>
		告警内容：<br>
		&nbsp&nbsp&nbsp&nbsp%s<br>
		告警时间：%s<br><br>
		该资产在此规则监控下解除异常，告警结束`, p.asset_name,
				p.address, p.rule, p.State.reason, p.State.time.Format("2006-01-02 15:04:05"))
			to = p.mailto
		}

		// 发送告警邮件
		err := p.tools.mail.SendMail(subject, to, []string{}, []string{}, message, []string{})
		if err != nil {
			log.Println(err)

		}
		//fmt.Println("邮件发送")
	}

	// 保存当前监控状态
	p.Save()
}

// Save
// 功能：保存TCP监控任务的当前状态到数据库，包括更新资产状态和告警日志
//
// 参数：无
//
// 返回值：无
func (p *Tcp) Save() {
	// 更新资产状态
	_, err := p.tools.db.Engine.Where("id=?", p.asset_id).Cols("state").Update(&models.Asset{
		State: p.State.Status,
	})
	if err != nil {
		log.Println(err)
	}

	// 更新告警日志
	var m []models.Mail
	for i := 0; i < len(p.mailto); i++ {
		m = append(m, models.Mail{
			State:   true,
			Address: p.mailto[i],
		})
	}
	// 更新资产状态
	var alarmstate int
	if p.State.Status == 3 {
		alarmstate = 3
	} else if p.State.abn == p.wrong_limit {
		alarmstate = 1
	} else if p.State.abn > p.wrong_limit {
		alarmstate = 2
	} else if p.State.nor > 0 {
		alarmstate = 3
	}
	//保存告警日志

	_, err = p.tools.db.Engine.InsertOne(models.AlarmLog{
		AssetID:   p.asset_id,
		RuleID:    p.rule_id,
		State:     alarmstate,
		Mails:     m,
		Messages:  []string{p.State.reason},
		CreatedAt: p.State.time,
	})
	if err != nil {
		log.Println(err)
	}
}

// Update
// 功能：将TCP监控任务的状态更新为“异常检测中”，并触发告警通知
//
// 参数：无
//
// 返回值：无
func (p *Tcp) Update() {
	// 将状态更新为“检测中”
	p.State.Status = 3

	// 触发告警通知
	p.Alarm()
}
