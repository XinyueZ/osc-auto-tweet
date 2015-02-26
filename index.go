

const (
	TWEET_PUT_API_URL = "https://www.oschina.net/action/openapi/tweet_pub"
)


type Error string

func (e Error) Error() string {
	return string(e)
}

func init() {
	http.HandleFunc("/login", handleLogin)
	http.HandleFunc("/tweetList", handleTweetList)
	http.HandleFunc("/myTweetList", handleMyTweetList)
	http.HandleFunc("/hotspotTweetList", handleHotspotTweetList)
	http.HandleFunc("/tweetPub", handleTweetPub)
	http.HandleFunc("/friendsList", handleFriendsList)
	http.HandleFunc("/userInformation", handlePersonal)
	http.HandleFunc("/updateRelation", handleUpdateRelation)
}


func writeMessage(msg string, pLogined *Logined, ch chan bool) {
	client := new(http.Client)
	jsonStr := fmt.Sprintf(`access_token=%s&msg=%s`, pLogined.Token.AccessToken, msg)
	if r, e := http.NewRequest("POST", TWEET_PUT_API_URL, bytes.NewBufferString(jsonStr)); e == nil {
		r.Header.Add("Content-Type", API_REQTYPE)
		if resp, e := client.Do(r); e == nil {
			printResponse(resp)
			ch <- true
		} else {
			panic(e)
		}
	} else {
		panic(e)
	}
	return
}