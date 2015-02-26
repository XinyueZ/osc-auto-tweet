package osc_auto_tweet

import (
	"appengine"

	"net/http"
 
	"github.com/XinyueZ/osc-server/src/common"
	"github.com/XinyueZ/osc-server/src/tweet"
	"github.com/XinyueZ/osc-server/src/user"
)

type Error string

func (e Error) Error() string {
	return string(e)
}

func init() {
	http.HandleFunc("/tweet", handleTweet)
}

func handleTweet(w http.ResponseWriter, r *http.Request) {
	cxt := appengine.NewContext(r)
	chLogin := make(chan *user.Logined)
	chResult := make(chan *common.Result)
	
	defer func() {
		if err := recover(); err != nil {
			cxt.Errorf("handleTweet: %v", err)
		}
	}()

	pUser := user.NewOscUser(ACCOUNT, PASSWORD, APP_ID, APP_SEC)
	go pUser.Login(cxt, chLogin)

	//Get cookie.
	pLogined := <-chLogin
	session := pLogined.Cookie.Value    //Got user session. 
	uid := pLogined.Token.UID
	access_token := pLogined.Token.AccessToken

	tweet.TweetPub(cxt, uid, session, access_token, "hello world is funny.", chResult)
	<-chResult
}
