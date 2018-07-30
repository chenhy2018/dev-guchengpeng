package iam

import (
	"bytes"
	"encoding/gob"
	"time"

	"qbox.us/biz/component/sessions"
	"qbox.us/iam/entity"
)

func init() {
	gob.Register(&entity.User{})
	gob.Register(&entity.Keypair{})
}

func isTokenExpired(tokenExpiry int64) bool {
	// 减120s，提前换取新token
	return time.Now().Unix() >= tokenExpiry-120
}

const (
	sessKeyUser        = "iam_user"
	sessKeyUserExpires = "iam_user_expires"
	sessUserExpires    = time.Minute * 5
)

func loadUserFromSession(sess sessions.SessionStore) (user *entity.User) {
	unixSec, err := sess.Get(sessKeyUserExpires).Int64()
	if err != nil {
		return
	}

	if time.Unix(unixSec, 0).Before(time.Now()) {
		sess.Delete(sessKeyUser)
		return
	}

	if data := sess.Get(sessKeyUser).MustBytes(); len(data) > 0 {
		gob.NewDecoder(bytes.NewReader(data)).Decode(&user)
	}
	return
}

func saveUserToSession(sess sessions.SessionStore, user *entity.User) {
	buf := bytes.NewBuffer(nil)
	err := gob.NewEncoder(buf).Encode(user)
	if err == nil {
		sess.Set(sessKeyUser, buf.Bytes())
		sess.Set(sessKeyUserExpires, time.Now().Add(sessUserExpires).Unix())
	}
}

const (
	sessKeyKeypair        = "iam_user_keypair"
	sessKeyKeypairExpires = "iam_user_keypair_expires"
	sessKeypairExpires    = time.Minute * 5
)

func loadKeypairFromSession(sess sessions.SessionStore) (keypair *entity.Keypair) {
	unixSec, err := sess.Get(sessKeyKeypairExpires).Int64()
	if err != nil {
		return
	}

	if time.Unix(unixSec, 0).Before(time.Now()) {
		sess.Delete(sessKeyKeypair)
		return
	}

	if data := sess.Get(sessKeyKeypair).MustBytes(); len(data) > 0 {
		gob.NewDecoder(bytes.NewReader(data)).Decode(&keypair)
	}
	return
}

func saveKeypairToSession(sess sessions.SessionStore, keypair *entity.Keypair) {
	buf := bytes.NewBuffer(nil)
	err := gob.NewEncoder(buf).Encode(keypair)
	if err == nil {
		sess.Set(sessKeyKeypair, buf.Bytes())
		sess.Set(sessKeyKeypairExpires, time.Now().Add(sessKeypairExpires).Unix())
	}
}
