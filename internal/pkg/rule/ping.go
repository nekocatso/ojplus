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

type Ping struct {
	State State
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
	latency_limit int
	lost_limit    int
	alarm_id      int
	interval      int
	mailto        []string
}

func NewPing(id int, Rcp *models.Cache, mail *mail.MailBox, db *models.Database) *Ping {
	var p Ping
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

	var pi models.PingInfo
	p.tools.db.Engine.Where("id=?", p.rule_id).Get(&pi)
	p.latency_limit = pi.LatencyLimit
	p.mode = pi.Mode
	p.lost_limit = pi.LostLimit

	var at models.AlarmTemplate
	p.tools.db.Engine.Where("id=?", p.alarm_id).Get(&at)
	p.interval = at.Interval
	p.mailto = at.Mails

	var a models.Asset
	p.tools.db.Engine.Where("id=?", p.asset_id).Get(&a)
	p.asset_name = a.Name
	p.address = a.Address

	p.State.abn = 0
	p.State.nor = 0
	p.State.Status = 3
	p.State.reason = ""
	p.State.time = time.Now()
	return &p
}

func (p *Ping) Scan() error {
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
	log.Println(p.State.nor, p.State.abn, p.State.Status)
	return nil
}
func (p *Ping) Jude() (bool, error) { //返回true是无错误，返回false是出错
	res, err := p.tools.Rcp.Client.Get(fmt.Sprintf("%d", p.State.correlation_id)).Bytes()
	if err != nil {
		return false, err
	}
	var data map[string]interface{}
	json.Unmarshal(res, &data)

	if data["status"] == "success" {
		result := data["result"].(map[string]interface{})
		fmt.Println(result)
		rl := result["latency"].([]interface{})[0].(float64)
		rp := result["package_loss_rate"].([]interface{})[0].(float64)
		if p.mode == 1 { //同时错误

			if rl < float64(p.latency_limit) || rp < float64(p.lost_limit) {
				return true, nil
			} else {
				p.State.reason = fmt.Sprintf("响应时间大于等于%d ms（%.2f），丢包率大于等于%d %%（%.2f%%）", p.latency_limit, rl, p.lost_limit, rp)
				return false, nil

			}

		} else {
			if rl < float64(p.latency_limit) && rp < float64(p.lost_limit) { //任一错误
				return true, nil
			} else if rl >= float64(p.latency_limit) {
				p.State.reason = fmt.Sprintf("响应时间大于等于%d ms（%.2f）", p.latency_limit, rl)
				return false, nil
			} else {
				p.State.reason = fmt.Sprintf("丢包率大于等于%d %%（%.2f%%）", p.lost_limit, rp)
				return false, nil
			}

		}

	} else {
		return false, errors.New("respone error")
	}

}
func (p *Ping) Sendmail() {
	fmt.Println("sendmail")
	var subject, message string
	if p.State.Status == 3 {
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

	} else if p.State.abn == p.wrong_limit {
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

	} else if p.State.abn > p.wrong_limit {
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

	} else if p.State.nor > 0 {
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
	} else {

	}
	err := p.tools.mail.SendMail(subject, p.mailto, []string{}, []string{}, message, []string{})
	p.Save()
	if err != nil {
		log.Println(err)
	}
}
func (p *Ping) Save() {
	//更新资产状态
	_, err := p.tools.db.Engine.Where("id=?", p.asset_id).Cols("state").Update(&models.Asset{
		State: p.State.Status,
	})
	if err != nil {
		log.Println(err)
	}
	//存储告警日志
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
func (p *Ping) Update() {
	p.State.Status = 3
	p.Sendmail()
}
