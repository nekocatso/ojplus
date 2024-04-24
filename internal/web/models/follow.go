package models

type Follow struct {
	ID         int `json:"id" xorm:"'id' pk autoincr"`
	LeaderID   int `json:"leaderId" xorm:"'leader_id' notnull unique(follow)"`
	FollowerID int `json:"followerId" xorm:"'follower_id' notnull unique(follow)"`
}

func (manager *UserManager) GetLeaders(userID int) ([]User, error) {
	users := []User{}
	builder := manager.db.Engine.Table("follow").Where("follower_id = ?", userID).Cols("id")
	err := manager.db.Engine.Table("user").In("id", builder).Find(users)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (manager *UserManager) GetFollowers(userID int) ([]User, error) {
	users := []User{}
	builder := manager.db.Engine.Table("follow").Where("leader_id = ?", userID).Cols("id")
	err := manager.db.Engine.Table("user").In("id", builder).Find(users)
	if err != nil {
		return nil, err
	}
	return users, nil
}
