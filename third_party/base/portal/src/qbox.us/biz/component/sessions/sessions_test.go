package sessions

// TODO: refactor to none dependences on models

// import (
// 	"net/http"
// 	"net/http/httptest"
// 	"strings"
// 	"testing"
// 	"time"

// 	"github.com/stretchr/testify/assert"

// 	"labix.org/v2/mgo"
// 	"labix.org/v2/mgo/bson"

// 	"qbox.us/mgo2"

// 	"qbox.us/biz/models"
// )

// var (
// 	manager *SessionManager

// 	config = Config{
// 		CookieName:     "PORTAL_SESSION",
// 		CookieSecure:   true,
// 		SessionExpire:  3600,
// 		CookieExpire:   3600,
// 		RememberExpire: 3600,
// 	}
// )

// const (
// 	TEST_SECRET_KEY = "TEST_SECRET_KEY"
// )

// func initDatabase(t *testing.T) {
// 	var err error
// 	models.DefaultBizMongo, err = mgo2.NewDatabase("mongodb://localhost/qbox_biz_test", "strong")
// 	if !assert.NoError(t, err) {
// 		t.Fatal()
// 	}
// 	models.DefaultBizMongo.C(models.SessionCollection).RemoveAll(nil)
// }

// func initSessionManager(t *testing.T) *SessionManager {
// 	initDatabase(t)

// 	config.SecretKey = TEST_SECRET_KEY
// 	provider := NewMgoProvider(config, models.SessionStore.ConnectorV2())

// 	manager = NewSessionManager(provider)
// 	return manager
// }

// func Test_EmptySecretKey(t *testing.T) {
// 	initDatabase(t)

// 	// test empty secret key
// 	config.SecretKey = ""
// 	provider := NewMgoProvider(config, models.SessionStore.ConnectorV2())

// 	assert.Panics(t, func() {
// 		NewSessionManager(provider)
// 	})
// }

// func Test_ValidSid(t *testing.T) {
// 	manager := initSessionManager(t)

// 	sid := strings.Repeat("0", SESSION_ID_LENGTH)
// 	_, ok := manager.validSid(sid)
// 	assert.True(t, ok)

// 	sid = ""
// 	_, ok = manager.validSid(sid)
// 	assert.False(t, ok)

// 	sid = strings.Repeat("0", SESSION_ID_LENGTH+10)
// 	_, ok = manager.validSid(sid)
// 	assert.False(t, ok)
// }

// func Test_CreateSid(t *testing.T) {
// 	manager := initSessionManager(t)

// 	sid := CreateSid()
// 	_, ok := manager.validSid(sid)
// 	if !ok {
// 		t.Fatal()
// 	}
// }

// func Test_SessionManager(t *testing.T) {
// 	manager := initSessionManager(t)

// 	sid := CreateSid()
// 	r := createNewSessionRequest(sid, time.Now())
// 	w := httptest.NewRecorder()

// 	sess, _, err := manager.Start(w, r)
// 	if !assert.NoError(t, err) {
// 		t.Fatal()
// 	}

// 	assert.NotEqual(t, sid, sess.Sid())

// 	sid = sess.Sid()
// 	r = createNewSessionRequest(sid, time.Now())

// 	// 重新 Start 获取 session 数据
// 	sess, _, err = manager.Start(w, r)
// 	if !assert.NoError(t, err) {
// 		t.Fatal()
// 	}

// 	// 确保重获取的 sid 相同
// 	assert.Equal(t, sid, sess.Sid())

// 	sess, err = manager.Regenerate(w, r)
// 	if !assert.NoError(t, err) {
// 		t.Fatal()
// 	}

// 	// 确保修改后的 sid 不同
// 	assert.NotEqual(t, sid, sess.Sid())

// 	sid = sess.Sid()
// 	r = createNewSessionRequest(sid, time.Now())

// 	err = manager.Destroy(w, r)
// 	if !assert.NoError(t, err) {
// 		t.Fatal()
// 	}

// 	// Session 删除以后，重新生成
// 	sess, _, err = manager.Start(w, r)
// 	if !assert.NoError(t, err) {
// 		t.Fatal()
// 	}
// 	assert.NotEqual(t, sid, sess.Sid())
// }

// func Test_SessionStore(t *testing.T) {
// 	manager := initSessionManager(t)

// 	sid := CreateSid()
// 	r := createNewSessionRequest(sid, time.Now())
// 	w := httptest.NewRecorder()

// 	sess, _, err := manager.Start(w, r)
// 	if !assert.NoError(t, err) {
// 		t.Fatal()
// 	}

// 	sid = sess.Sid()

// 	sess.Set("uid", 110)
// 	assert.Equal(t, sess.Get("uid").MustInt(), 110)

// 	sess.Clean()
// 	assert.Equal(t, sess.Get("uid").Value(), nil)

// 	sess.Set("uid", 110)
// 	sess.Delete("uid")

// 	sess.Clean()
// 	assert.Equal(t, sess.Get("uid").Value(), nil)

// 	sess.Set("uid", 110)
// 	err = sess.Flush()
// 	assert.Equal(t, err, nil)

// 	r = createNewSessionRequest(sid, time.Now())
// 	sess, _, err = manager.Start(w, r)
// 	if !assert.NoError(t, err) {
// 		t.Fatal()
// 	}

// 	assert.Equal(t, sess.Get("uid").MustInt(), 110)

// 	c := models.SessionStore.Connector()
// 	c(func(c *mgo.Collection) {
// 		n, _ := c.Find(bson.M{
// 			sidField: sess.Sid(),
// 		}).Count()
// 		assert.Equal(t, true, n > 0)
// 	})

// 	err = sess.Destroy()
// 	assert.Equal(t, err, nil)

// 	err = sess.Flush()
// 	assert.Equal(t, err, nil)

// 	c(func(c *mgo.Collection) {
// 		n, err := c.Find(bson.M{
// 			sidField: sess.Sid(),
// 		}).Count()
// 		assert.Equal(t, err, nil)
// 		assert.Equal(t, true, n == 0)
// 	})
// }

// func Test_SessionExpire(t *testing.T) {
// 	manager := initSessionManager(t)

// 	sessionExpireSeconds := manager.provider.Config().SessionExpireSeconds()

// 	sid := CreateSid()
// 	r := createNewSessionRequest(sid, time.Now())
// 	w := httptest.NewRecorder()

// 	sess, _, err := manager.Start(w, r)
// 	if !assert.NoError(t, err) {
// 		t.Fatal()
// 	}

// 	sid = sess.Sid()

// 	sess, err = manager.provider.Read(sid)
// 	assert.Equal(t, sid, sess.Sid())

// 	c := models.SessionStore.Connector()
// 	c(func(c *mgo.Collection) {
// 		c.Update(bson.M{sidField: sid}, bson.M{updatedAtField: time.Now().Add(-sessionExpireSeconds)})
// 	})

// 	_, err = manager.provider.Read(sid)
// 	assert.Equal(t, err == ErrNotFoundSession, true)
// }

// func Test_SessionCreatedAt(t *testing.T) {
// 	manager := initSessionManager(t)

// 	sid := CreateSid()
// 	r := createNewSessionRequest(sid, time.Now())
// 	w := httptest.NewRecorder()

// 	sess, createdAt, err := manager.Start(w, r)
// 	if !assert.NoError(t, err) {
// 		t.Fatal()
// 	}

// 	r = createNewSessionRequest(sess.Sid(), createdAt)

// 	_, parsedTime, _ := manager.Start(w, r)

// 	assert.Equal(t, createdAt.UnixNano(), parsedTime.UnixNano())
// }

// func createNewSessionRequest(sid string, createdAt time.Time) *http.Request {
// 	value, _ := EncodeSecureValue(sid, TEST_SECRET_KEY, createdAt)

// 	cookie := &http.Cookie{
// 		Name:   config.CookieName,
// 		Value:  value,
// 		Path:   "/",
// 		Secure: true,
// 		MaxAge: config.SessionExpire,
// 	}
// 	r, _ := http.NewRequest("GET", "https://www.qiniu.com", nil)
// 	r.AddCookie(cookie)

// 	return r
// }
