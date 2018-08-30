package models

import (
	"fmt"
	"github.com/qiniu/xlog.v1"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"qiniu.com/db"
	"time"
)

type UserModel struct {
}

var (
	User *UserModel
)

type UserInfo struct {
	Uid         string `bson:"uid"        json:"uid"`
	Password    string `bson:"password"   json:"password"`
	Status      bool   `bson:"status"      json:"status"`
	CreateAt    int64  `bson:"createdAt"   json:"createdAt"`
	UpdateAt    int64  `bson:"updatedAt"   json:"updatedAt"`
	IsSuperuser bool   `bson:"issuperuser"  json:"issuperuser"`
}

// -----------------------------------------------------------------------------------------------------------

func ValidateLogin(xl *xlog.Logger, uid, pwd string) error {

	/*
	   db.user.update( { "uid" : uid, "password" : pwd }, { "$set": bson.M{ "status" : true} })
	*/
	return db.WithCollection(
		USER_COL,
		func(c *mgo.Collection) error {
			err := c.Update(
				bson.M{
					USER_UUID:     uid,
					USER_PASSWORD: pwd,
				},
				bson.M{
					"$set": bson.M{
						USER_STATUS: true,
					},
				},
			)
			if err != nil {
				return fmt.Errorf("uid or password not correct")
			}
			return nil
		},
	)
}

func ValidateUid(xl *xlog.Logger, uid string) error {

	/*
	   db.user.Find({ "uid" : uid})
	*/
	return db.WithCollection(
		USER_COL,
		func(c *mgo.Collection) error {
			count, err := c.Find(
				bson.M{
					USER_UUID: uid,
				},
			).Count()
			if err != nil {
				return err
			}
			if count == 0 {
				return fmt.Errorf("uid not correct")
			}
			return nil
		},
	)
}

func ResetPassword(xl *xlog.Logger, uid, opwd, pwd string) error {

	/*
	   db.user.update( { "uid" : uid, "password" : opwd }, { "$set": bson.M{ "password" : pwd} })
	*/
	return db.WithCollection(
		USER_COL,
		func(c *mgo.Collection) error {
			query := bson.M{
				USER_UUID:     uid,
				USER_PASSWORD: opwd,
			}
			update := bson.M{
				"$set": bson.M{
					USER_PASSWORD:    pwd,
					ITEM_UPDATA_TIME: time.Now().Unix(),
				},
			}
			return c.Update(query, update)
		},
	)
}

func GetPwdByUID(xl *xlog.Logger, uid string) (string, error) {

	/*
	   db.user.find({ "uid" : uid})
	*/
	r := UserInfo{}
	err := db.WithCollection(
		USER_COL,
		func(c *mgo.Collection) error {
			return c.Find(
				bson.M{
					"uid": uid,
				},
			).One(&r)
		},
	)
	if err != nil {
		return "", fmt.Errorf("pwd get error: %v", err)
	}
	return r.Password, nil
}

func Logout(xl *xlog.Logger, uid string) error {

	/*
	   db.user.update({ "uid" : uid, "password" : pwd }, { "$set": bson.M{ "status" : false} })
	*/

	return db.WithCollection(
		USER_COL,
		func(c *mgo.Collection) error {
			return c.Update(
				bson.M{
					USER_UUID: uid,
				},
				bson.M{
					"$set": bson.M{
						USER_STATUS: false,
					},
				},
			)
		},
	)
}

func AddUser(xl *xlog.Logger, info UserInfo, uid, pwd string) error {

	/*
	   db.user.find( { "uid" : info.uid, "password" : info.pwd, "issuperuser" : true })
	   db.user.find( { "uid" : uid} )
	   db.user.update( { "uid" : uid}, bson.M{ "$set": bson.M{xxx}}, upsert:true)
	*/

	return db.WithCollection(
		USER_COL,
		func(c *mgo.Collection) error {
			count, err := c.Find(
				bson.M{
					USER_UUID:         uid,
					USER_PASSWORD:     pwd,
					USER_IS_SUPERUSER: true,
				},
			).Count()
			if err != nil {
				return err
			}
			if count == 0 {
				return fmt.Errorf("No access. uid is not superuser")
			}
			count, err = c.Find(
				bson.M{
					USER_UUID: info.Uid,
				},
			).Count()
			if count != 0 {
				return fmt.Errorf("uid is exit")
			}

			_, err = c.Upsert(
				bson.M{
					USER_UUID: info.Uid,
				},
				bson.M{
					"$set": bson.M{
						USER_UUID:         info.Uid,
						USER_PASSWORD:     info.Password,
						USER_IS_SUPERUSER: false,
						USER_STATUS:       false,
						ITEM_CREATE_TIME:  time.Now().Unix(),
						ITEM_UPDATA_TIME:  time.Now().Unix(),
					},
				},
			)
			return err
		},
	)
}

func DelUser(xl *xlog.Logger, info UserInfo, uid, pwd string) error {

	/*
	   db.user.find({ "uid" : info.uid, "password" : info.pwd, "issuperuser" : true })
	   db.user.remove({ "uid" : uid})
	*/

	return db.WithCollection(
		USER_COL,
		func(c *mgo.Collection) error {
			count, err := c.Find(
				bson.M{
					USER_UUID:         uid,
					USER_PASSWORD:     pwd,
					USER_IS_SUPERUSER: true,
				},
			).Count()
			if err != nil {
				return err
			}
			if count == 0 {
				return fmt.Errorf("No access. uid is not superuser")
			}

			return c.Remove(
				bson.M{
					USER_UUID: info.Uid,
				},
			)
		},
	)
}

func GetUserInfo(xl *xlog.Logger, index, rows int, uid, pwd string, category, like string) ([]UserInfo, error) {

	/*
	   db.user.find({ "uid" : info.uid, "password" : info.pwd, "issuperuser" : true })
	   db.user.find({category : like}).sort("uid").skip(index * row).limit(rows)
	*/

	// query by keywords
	query := bson.M{}
	if like != "" {
		query[category] = bson.M{
			"$regex": ".*" + like + ".*",
		}
	}

	skip := rows * index
	limit := rows
	if limit > 100 {
		limit = 100
	}

	// query
	r := []UserInfo{}
	err := db.WithCollection(
		USER_COL,
		func(c *mgo.Collection) error {
			count, err := c.Find(
				bson.M{
					USER_UUID:         uid,
					USER_PASSWORD:     pwd,
					USER_IS_SUPERUSER: true,
				},
			).Count()
			if err != nil {
				return err
			}
			if count == 0 {
				return fmt.Errorf("No access. uid is not superuser")
			}
			if err = c.Find(query).Sort(USER_UUID).Skip(skip).Limit(limit).All(&r); err != nil {
				return fmt.Errorf("query failed")
			}
			if count, err = c.Find(query).Count(); err != nil {
				return fmt.Errorf("query count failed")
			}
			return nil
		},
	)
	return r, err
}
