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
	state state
	tools tools
	// 参数asset_name,address,rule,health_limit,wrong_limit,interval,mode,latency_limit,lost_limit,email
	asset_id      int
	asset_name    string
	address       string
	rule_id       int
	rule          string
	health_limit  int
	wrong_limit   int
	mode          int
	alarm_id      int
	interval      int
	mailto        []string
	enable_ports  string
	disable_ports string
	portsetting   map[string]int
	hisopenerr    map[string]bool
	hiscloseerr   map[string]bool
	nowopenerr    []string
	nowcloseerr   []string
}

func NewTcp(id int, Rcp *models.Cache, mail *mail.MailBox, db *models.Database) *Tcp {
	var p Tcp
	p.tools.db = db
	p.tools.Rcp = Rcp
	p.tools.mail = mail
	p.state.correlation_id = id
	var ar models.AssetRule
	p.tools.db.Engine.Where("id = ?", id).Get(&ar)
	p.asset_id = ar.AssetID
	p.rule_id = ar.RuleID


	var r models.Rule
	p.tools.db.Engine.Where("id=?", p.rule_id).Get(&r)
	p.rule = r.Name
	p.wrong_limit = r.WrongLimit
	p.health_limit = r.HealthLimit
	p.alarm_id = r.AlarmID

	var ti models.TcpInfo
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

	p.state.abn = 0
	p.state.nor = 0
	p.state.status = 1
	p.state.reason = ""
	p.state.time = time.Now()
	p.hiscloseerr = make(map[string]bool)
	p.hisopenerr = make(map[string]bool)
	return &p
}

func (p *Tcp) State() error {
	p.state.time = time.Now()
	s, err := p.Jude()

	if err != nil {
		return err
	}
	switch p.state.status {
	case 1:
		if s {
			p.state.nor++
			if p.state.nor >= p.health_limit {
				p.state.abn = 0
				p.state.status = 2
			}
		} else {
			p.state.abn++
			if p.state.abn >= p.wrong_limit {
				p.state.nor = 0
				p.state.status = 3
				p.Sendmail()
			}
		}
	case 2:
		if s {
			p.state.nor++
			p.state.abn = 0
		} else {
			p.state.abn++
			if p.state.abn >= p.wrong_limit {
				p.state.nor = 0
				p.state.status = 3
				p.Sendmail()

			}
		}
	case 3:
		if s {
			p.state.nor++
			if p.state.nor >= p.health_limit {
				p.state.abn = 0
				p.state.status = 2
				if len(p.hiscloseerr) != 0 {
					sc := []string{}
					for i, _ := range p.hiscloseerr {
						sc = append(sc, i)
					}
					p.hiscloseerr = map[string]bool{}
					p.state.reason += fmt.Sprintf("预期关闭的端口%s处于关闭状态 ", strings.Join(sc, ","))

				}
				if len(p.hisopenerr) != 0 {
					so := []string{}
					for i, _ := range p.hisopenerr {
						so = append(so, i)
					}
					p.hisopenerr = map[string]bool{}
					p.state.reason += fmt.Sprintf("预期开启的端口%s处于开启状态 ", strings.Join(so, ","))

				}

				p.Sendmail()

			}
		} else {
			p.state.nor = 0
			p.state.abn++
			if (p.state.abn-p.wrong_limit)%p.interval == 0 {

				p.Sendmail()

			}

		}
	}
	fmt.Println(p.state.nor, p.state.abn, p.state.status, p.nowcloseerr)
	return nil
}

func (p *Tcp) Jude() (bool, error) { //返回true是无错误，返回false是出错
	p.state.reason = ""
	p.nowcloseerr = []string{}
	p.nowopenerr = []string{}

	var flag = true
	res, err := p.tools.Rcp.Client.Get(fmt.Sprintf("%d", p.state.correlation_id)).Bytes()
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
			if len(portstate) == 1 {
				//t, err := strconv.Atoi(target)
				if err != nil {
					return false, err
				}

				if int(portstate[0].(float64)) != p.portsetting[target] {
					if p.portsetting[target] == 0 {
						p.nowopenerr = append(p.nowopenerr, target)
						p.hisopenerr[target] = true
					} else {
						p.nowcloseerr = append(p.nowcloseerr, target)
						p.hiscloseerr[target] = true
					}
					flag = false
				}
			} else {
				minport, err := strconv.Atoi(strings.Split(target, "-")[0])
				if err != nil {
					return false, err
				}
				for j := 0; j < len(portstate); j++ {
					if int(portstate[j].(float64)) != p.portsetting[target] {

						if p.portsetting[target] == 0 {

							p.nowopenerr = append(p.nowopenerr, fmt.Sprintf("%d", minport+j))
							p.hisopenerr[fmt.Sprintf("%d", minport+j)] = true
						} else {

							p.nowcloseerr = append(p.nowcloseerr, fmt.Sprintf("%d", minport+j))

							p.hiscloseerr[fmt.Sprintf("%d", minport+j)] = true

						}
						flag = false
					}
				}
			}

		}
		if !flag {
			//
			if len(p.nowopenerr) != 0 {

				p.state.reason += fmt.Sprintf("预期开启的端口%s处于关闭状态 ", strings.Join(p.nowopenerr, ","))
			}
			if len(p.nowcloseerr) != 0 {

				p.state.reason += fmt.Sprintf("预期关闭的端口%s处于开启状态 ", strings.Join(p.nowcloseerr, ","))
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

	if p.state.abn == p.wrong_limit {
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
			p.address, p.rule, p.state.reason, p.state.time.String())
		to = p.mailto
	} else if p.state.abn > p.wrong_limit {
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
			p.address, p.rule, p.state.reason, p.state.time.String())
		to = p.mailto
	} else if p.state.nor > 0 {
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
			p.address, p.rule, p.state.reason, p.state.time.String())
		to = p.mailto
	}
	p.tools.mail.SendMail(subject, to, []string{}, []string{}, message, []string{})
	p.Save()
	fmt.Println("邮件发送")
}
func (p *Tcp) Save() {
	var m []models.Mail
	for i := 0; i < len(p.mailto); i++ {
		m = append(m, models.Mail{
			State:   true,
			Address: p.mailto[i],
		})
	}
	var alarmstate int
	if p.state.abn == p.wrong_limit {
		alarmstate = 1
	} else if p.state.abn > p.wrong_limit {
		alarmstate = 2
	} else if p.state.nor > 0 {
		alarmstate = 3
	}
	_, err := p.tools.db.Engine.InsertOne(models.AlarmLog{
		AssetID:   p.asset_id,
		RuleID:    p.rule_id,
		State:     alarmstate,
		Mails:     m,
		Message:   []string{p.state.reason},
		CreatedAt: p.state.time,
	})
	if err != nil {
		log.Println(err)
	}
}
