package rule

import (
	"Alarm/internal/pkg/Cache"
	"Alarm/internal/pkg/mail"
)

type rule interface{
	Scan()
}
type Ping struct{
	Rcp *Cache.RedisClientPool
	param map[string]string
	mail *mail.MailBox
	times int 
	status bool

}
type Tcp struct{
	Rcp *Cache.RedisClientPool
	param map[string]string
	mail *mail.MailBox
	times int 
	status bool

}
func New(param map [string]string,mail *mail.MailBox,Rcp *Cache.RedisClientPool) *Ping{
	return &Ping{param:param,mail:mail,Rcp: Rcp,times:0,status:false}
}
func (p *Ping)Scan(){
	conn:=p.Rcp.GetConn()
	conn.Do("")
}

