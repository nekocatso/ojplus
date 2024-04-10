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

type Tcp struct {
	State State
	tools tools
	// 参数asset_name,address,rule,health_limit,wrong_limit,interval,mode,latency_limit,lost_limit,email
	asset_id        int
	asset_name      string
	address         string
	rule_id         int
	rule            string
	health_limit    int
	wrong_limit     int
	alarm_id        int
	interval        int
	mailto          []string
	enable_ports    string
	disable_ports   string
	portsetting     map[string]int
	hisopenerr      map[string]bool
	hiscloseerr     map[string]bool
	nowopenerr      []string
	nowcloseerr     []string
	nowopentimeout  []string
	nowclosetimeout []string
}

func NewTcp(id int, Rcp *models.Cache, mail *mail.MailBox, db *models.Database) *Tcp {
	var p Tcp
	p.tools.db = db
	p.tools.Rcp = Rcp
	p.tools.mail = mail
	p.State.correlation_id = id
	var ar models.AssetRule
	p.tools.db.Engine.Where("id = ?", id).Get(&ar)
	p.asset_id = ar.AssetID
	p.rule_id = ar.RuleID

	var r models.Rule
	p.tools.db.Engine.Where("id=?", p.rule_id).Get(&r)
	p.rule = r.Name
	p.wrong_limit = r.DeclineLimit
	p.health_limit = r.RecoverLimit
	p.alarm_id = r.AlarmID

	var ti models.TCPInfo
	p.tools.db.Engine.Where("id=?", p.rule_id).Get(&ti)
	p.enable_ports = ti.EnablePorts
	p.disable_ports = ti.DisablePorts

	var a models.Asset
	p.tools.db.Engine.Where("id=?", p.asset_id).Get(&a)
	p.asset_name = a.Name
	p.address = a.Address

	m := map[string]int{}
	for _, i := range strings.Split(p.enable_ports, ",") {
		m[i] = 0
	}
	for _, i := range strings.Split(p.disable_ports, ",") {
		m[i] = 2
	}
	p.portsetting = m
	//fmt.Println(m)
	var at models.AlarmTemplate
	p.tools.db.Engine.Where("id=?", r.AlarmID).Get(&at)
	p.interval = at.Interval
	p.mailto = at.Mails

	p.State.abn = 0
	p.State.nor = 0
	p.State.Status = 3
	p.State.reason = ""
	p.State.time = time.Now()
	p.hiscloseerr = make(map[string]bool)
	p.hisopenerr = make(map[string]bool)
	return &p
}

func (p *Tcp) Scan() error {
	p.State.time = time.Now()
	s, err := p.Jude()

	if err != nil {
		return err
	}
	switch p.State.Status {
	case 3:
		if s {
			p.State.nor++
			if p.State.nor >= p.health_limit {
				p.State.abn = 0
				p.State.Status = 1
			}
		} else {
			p.State.abn++
			if p.State.abn >= p.wrong_limit {
				p.State.nor = 0
				p.State.Status = 2
				p.Sendmail()
			}
		}
	case 1:
		if s {
			p.State.nor++
			p.State.abn = 0
		} else {
			p.State.abn++
			if p.State.abn >= p.wrong_limit {
				p.State.nor = 0
				p.State.Status = 2
				p.Sendmail()

			}
		}
	case 2:
		if s {
			p.State.nor++
			if p.State.nor >= p.health_limit {
				p.State.abn = 0
				p.State.Status = 1
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

				p.Sendmail()

			}
		} else {
			p.State.nor = 0
			p.State.abn++
			if (p.State.abn-p.wrong_limit)%p.interval == 0 {

				p.Sendmail()

			}

		}

	}
	fmt.Println(p.State.nor, p.State.abn, p.State.Status, p.nowcloseerr)
	return nil
}

func (p *Tcp) Jude() (bool, error) { //返回true是无错误，返回false是出错
	p.State.reason = ""
	p.nowcloseerr = []string{}
	p.nowopenerr = []string{}
	p.nowopentimeout = []string{}
	p.nowclosetimeout = []string{}
	var flag = true
	res, err := p.tools.Rcp.Client.Get(fmt.Sprintf("%d", p.State.correlation_id)).Bytes()
	if err != nil {
		return false, err
	}
	var data map[string]interface{}
	json.Unmarshal(res, &data)
	if data["status"] == "success" {
		result := data["result"].(map[string]interface{})
		targets := data["target"].([]interface{})
		portstatus := result["portstatus"].([]interface{})
		for i := 0; i < len(targets); i++ {

			target := strings.Split(targets[i].(string), p.address+":")[1]

			portstate := portstatus[i].([]interface{})

			minport, err := strconv.Atoi(strings.Split(target, "-")[0])
			if err != nil {
				return false, err
			}
			for j := 0; j < len(portstate); j++ {
				if int(portstate[j].(float64)) != p.portsetting[target] {

					if p.portsetting[target] == 0 {
						if int(portstate[0].(float64)) == 1 {
							p.nowopentimeout = append(p.nowopentimeout, fmt.Sprintf("%d", minport+j))
							p.hisopenerr[fmt.Sprintf("%d", minport+j)] = true
						} else {
							p.nowopenerr = append(p.nowopenerr, fmt.Sprintf("%d", minport+j))
							p.hisopenerr[fmt.Sprintf("%d", minport+j)] = true
						}
					} else {
						if int(portstate[0].(float64)) == 1 {
							p.nowclosetimeout = append(p.nowclosetimeout, fmt.Sprintf("%d", minport+j))
							p.hiscloseerr[fmt.Sprintf("%d", minport+j)] = true
						} else {
							p.nowcloseerr = append(p.nowcloseerr, fmt.Sprintf("%d", minport+j))

							p.hiscloseerr[fmt.Sprintf("%d", minport+j)] = true
						}

					}
					flag = false
				}
			}

		}
		if !flag {
			//
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
		return flag, nil
	} else {
		return false, errors.New("respone error")
	}

}
func (p *Tcp) Sendmail() {
	var subject, message string
	var to []string
	if p.State.Status == 3 {
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
	} else if p.State.abn == p.wrong_limit {
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
	} else if p.State.abn > p.wrong_limit {
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
	} else if p.State.nor > 0 {
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
	p.tools.mail.SendMail(subject, to, []string{}, []string{}, message, []string{})
	p.Save()
	fmt.Println("邮件发送")
}
func (p *Tcp) Save() {
	//更新资产状态
	_, err := p.tools.db.Engine.Where("id=?", p.asset_id).Cols("state").Update(&models.Asset{
		State: p.State.Status,
	})
	if err != nil {
		log.Println(err)
	}
	//更新告警日志
	var m []models.Mail
	for i := 0; i < len(p.mailto); i++ {
		m = append(m, models.Mail{
			State:   true,
			Address: p.mailto[i],
		})
	}
	var alarmstate int
	if p.State.abn == p.wrong_limit {
		alarmstate = 1
	} else if p.State.abn > p.wrong_limit {
		alarmstate = 2
	} else if p.State.nor > 0 {
		alarmstate = 3
	}
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
func (p *Tcp) Update() {
	p.State.Status = 3
	p.Sendmail()
}
