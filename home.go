package osc_auto_tweet

//Use library github.com/XinyueZ/osc-server
//Tweet top Hack-News automatically.
//Demo show of the library.

import (
	"appengine"
	"appengine/urlfetch"

	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/XinyueZ/osc-server/src/common"
	"github.com/XinyueZ/osc-server/src/tweet"
	"github.com/XinyueZ/osc-server/src/user"
)

const (
	API_HOST             = "https://hacker-news.firebaseio.com"
	API_VERSION          = "/v0"
	GET_TOP_STORIES      = "/topstories.json"
	GET_ITEM_DETAILS     = "/item/%d.json"
	API_GET_TOP_STORIES  = API_HOST + API_VERSION + GET_TOP_STORIES
	API_GET_ITEM_DETAILS = API_HOST + API_VERSION + GET_ITEM_DETAILS

	ST_SUSS       = 200
	ST_NO_CONTENT = 201
	ST_FAIL       = 300

	METHOD = "GET"
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
	defer func() {
		if err := recover(); err != nil {
			cxt.Errorf("handleTweet: %v", err)

			s := fmt.Sprintf(`{"status":%d, "msg": "%v"}`, ST_FAIL, err)
			w.Header().Set("Content-Type", common.API_RESTYPE)
			fmt.Fprintf(w, s)
		}
	}()
	pUser := user.NewOscUser(ACCOUNT, PASSWORD, APP_ID, APP_SEC)
	go pUser.Login(cxt, chLogin)

	//Get cookie.
	pLogined := <-chLogin
	if pLogined != nil {
		chResult := make(chan *common.Result)
		go postTweet(w, r, pLogined, chResult)
		<-chResult
	} else {
		panic("Login failed.")
	}
}

//Tweet message on oschina.
//Content is fetched from Hack-News.
//The first of TOP will be used.
func postTweet(w http.ResponseWriter, r *http.Request, pLogined *user.Logined, chResult chan *common.Result) {
	session := pLogined.Cookie.Value           //Got user session.
	uid := pLogined.Token.UID                  //User oschina.net ID.
	access_token := pLogined.Token.AccessToken //Access-toke to OpenAPI.

	cxt := appengine.NewContext(r)
	pClient := urlfetch.Client(cxt)

	pDetail := getFirstItemDetail(pClient)//Get first Hack-News.

	jsonOut := ""
	if pDetail != nil {
		if pDetail.Url == "" {
			pDetail.Url = fmt.Sprintf("https://news.ycombinator.com/item?id=%d" , pDetail.Id) //No url, then direct to comment.
		}
		content := fmt.Sprintf("1024+奇葩。 %s %s", pDetail.Title, pDetail.Url)
		tweet.TweetPub(cxt, uid, session, access_token, content, chResult)
		jsonOut = fmt.Sprintf(`{"status":%d, "content":"%s"}`, ST_SUSS, content)
	} else {
		chResult <- nil
		jsonOut = fmt.Sprintf(`{"status":%d, "content":"%s"}`, ST_NO_CONTENT, "No detail found.")
	}
	w.Header().Set("Content-Type", common.API_RESTYPE)
	fmt.Fprintf(w, jsonOut)
}

//Detail of a Hack-News
type ItemDetail struct {
	By    string  `by`
	Id    int64   `id`
	Kids  []int64 `kids`
	Score int64   `score`
	Text  string  `text`
	Time  int64   `time`
	Title string  `title`
	Type  string  `type`
	Url   string  `url`
}

//Get list of top news.
func getTopStories(pClient *http.Client) []int64 {
	if req, err := http.NewRequest(METHOD, API_GET_TOP_STORIES, nil); err == nil {
		r, err := pClient.Do(req)
		defer r.Body.Close()
		if err == nil {
			if bytes, err := ioutil.ReadAll(r.Body); err == nil {
				topStoresRes := make([]int64, 0)
				json.Unmarshal(bytes, &topStoresRes)
				return topStoresRes
			} else {
				panic(err)
			}
		} else {
			panic(err)
		}
	} else {
		panic(err)
	}
	return nil
}

//Load Hack-News detail.
func loadDetail(pClient *http.Client, itemId int64) (pDetail *ItemDetail) {
	api := fmt.Sprintf(API_GET_ITEM_DETAILS, itemId)
	pDetail = new(ItemDetail)
	if req, err := http.NewRequest(METHOD, api, nil); err == nil {
		res, err := pClient.Do(req)
		if res != nil {
			defer res.Body.Close()
		}
		if err == nil {
			if bytes, err := ioutil.ReadAll(res.Body); err == nil {
				json.Unmarshal(bytes, &pDetail)
			} else {
				panic(err)
			}
		} else {
			panic(err)
		}
	} else {
		panic(err)
	}
	return
}

//Get a detail of first Hack-News.
func getFirstItemDetail(pClient *http.Client) (pDetail *ItemDetail) {
	var ids []int64 = getTopStories(pClient)
	if ids != nil && len(ids) > 0 {
		id := ids[0]
		pDetail = loadDetail(pClient, id)
	} else {
		pDetail = nil
	}
	return
}
