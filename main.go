package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	cookiemonster "github.com/MercuryEngineering/CookieMonster"
	"github.com/fatih/color"
	clr "github.com/gookit/color"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
)

func main() {
	printLogo()
	file := getCookieTxt()
	cookies, err := cookiemonster.ParseFile(fmt.Sprintf("cookies/%s", file))
	if err != nil {
		exitSafely(err)
	}
	ct0 := ""
	for _, val := range cookies {
		if val.Name == "ct0" {
			ct0 = val.Value
		}
	}
	bearToken, err := getBearToken()
	if err != nil {
		exitSafely(err)
	}
	headers := map[string]string{
		"Authority":     "https://twitter.com",
		"Accept":        "*/*",
		"User-agent":    "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/102.0.0.0 Safari/537.36",
		"Authorization": bearToken,
		"X-csrf-token":  ct0,
	}
	color.White("Введите ник твиттера: ")
	nickname := ""
	if _, err = fmt.Scanln(&nickname); err != nil {
		exitSafely(err)
	}
	if strings.HasPrefix(nickname, "@") {
		nickname = strings.ReplaceAll(nickname, "@", "")
	}
	id, err := getUserId(cookies, headers, nickname)
	if err != nil {
		exitSafely(err)
	}
	coord, err := firstRequest(cookies, headers, id)
	if err != nil {
		exitSafely(err)
	}
	for {
		if coord != "" && coord != "Finish!" {
			coord, err = mainRequest(cookies, headers, coord, id)
			if err != nil {
				exitSafely(err)
			}
		} else if coord == "" {
			exitSafely(errors.New("error in getting coordinates"))
		} else {
			finish()
		}
	}

}
func getCookieTxt() string {
	files, err := ioutil.ReadDir("./cookies")
	if err != nil {
		exitSafely(err)
	}
	file := ""
	for _, f := range files {
		file = f.Name()
	}
	return file
}
func getBearToken() (string, error) {
	url := "https://abs.twimg.com/responsive-web/client-web/main.3805b556.js"
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	resp, err := client.Do(req)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	split := strings.Split(string(body), "r=\"ACTION_FLUSH\"")
	split = strings.Split(split[1], ",s=\"")
	split = strings.Split(split[1], "\"")
	bear_token := fmt.Sprintf("Bearer %s", split[0])
	return bear_token, nil
}
func exitSafely(err error) {
	txt := fmt.Sprintf("<fg=red>ERROR: %v\n</>", err)
	clr.Printf(txt)
	color.Red("Press ENTER to EXIT\n")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
	os.Exit(0)
}
func finish() {
	color.Green("FINISH!\nPress ENTER to EXIT\n")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
	os.Exit(0)
}
func mainRequest(cookies []*http.Cookie, headers map[string]string, coord string, id string) (string, error) {
	jar, _ := cookiejar.New(nil)
	parsedURL, _ := url.Parse("https://twitter.com/i/api/graphql/Ezj6LyOvJyMwsQUQPIIsww/Followers")
	jar.SetCookies(parsedURL, cookies)
	client := &http.Client{Jar: jar}
	req, err := http.NewRequest("GET", "https://twitter.com/i/api/graphql/Ezj6LyOvJyMwsQUQPIIsww/Followers", nil)
	if err != nil {
		return "", err
	}
	q := req.URL.Query()
	q.Add("variables", "{\"userId\":\""+id+"\",\"count\":20,\"cursor\":\""+coord+"\",\"includePromotedContent\":false,\"withSuperFollowsUserFields\":true,\"withDownvotePerspective\":true,\"withReactionsMetadata\":false,\"withReactionsPerspective\":false,\"withSuperFollowsTweetFields\":true}")
	q.Add("features", "{\"dont_mention_me_view_api_enabled\":true,\"interactive_text_enabled\":true,\"responsive_web_uc_gql_enabled\":false,\"vibe_tweet_context_enabled\":false,\"responsive_web_edit_tweet_api_enabled\":false,\"standardized_nudges_misinfo\":false,\"responsive_web_enhance_cards_enabled\":false}")
	req.URL.RawQuery = q.Encode()
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var result js
	var cursor curs

	if err = json.Unmarshal(body, &cursor); err != nil {
		return "", err
	}
	for _, i := range cursor.Data.User.Result.Timeline.Timeline.Instructions {
		for _, j := range i.Entries {
			if j.Content.Value != "" && !(strings.HasPrefix(j.Content.Value, "-1")) {
				coord = j.Content.Value
			}
		}
	}
	if err = json.Unmarshal(body, &result); err != nil {
		return "", err
	}
	var names []string
	for _, i := range result.Data.User.Result.Timeline.Timeline.Instructions {
		for _, j := range i.Entries {
			if j.Content.ItemContent.UserResults.Result.Legacy.ScreenName != "" {
				names = append(names, j.Content.ItemContent.UserResults.Result.Legacy.ScreenName)
			}
		}
	}
	if names != nil {
		err = appendToFile("result.txt", names)
		if err != nil {
			return "", err
		}
	} else {
		return "Finish!", nil
	}

	return coord, nil
}
func firstRequest(cookies []*http.Cookie, headers map[string]string, id string) (string, error) {
	coord := ""
	var cursor curs
	var result js
	jar, _ := cookiejar.New(nil)
	parsedURL, _ := url.Parse("https://twitter.com/i/api/graphql/Ezj6LyOvJyMwsQUQPIIsww/Followers")
	jar.SetCookies(parsedURL, cookies)
	client := &http.Client{Jar: jar}
	req, err := http.NewRequest("GET", "https://twitter.com/i/api/graphql/Ezj6LyOvJyMwsQUQPIIsww/Followers", nil)
	if err != nil {
		return "", err
	}
	q := req.URL.Query()
	q.Add("variables", "{\"userId\":\""+id+"\",\"count\":20,\"includePromotedContent\":false,\"withSuperFollowsUserFields\":true,\"withDownvotePerspective\":true,\"withReactionsMetadata\":false,\"withReactionsPerspective\":false,\"withSuperFollowsTweetFields\":true}{\"userId\":\"877807935493033984\",\"count\":20,\"includePromotedContent\":false,\"withSuperFollowsUserFields\":true,\"withDownvotePerspective\":true,\"withReactionsMetadata\":false,\"withReactionsPerspective\":false,\"withSuperFollowsTweetFields\":true}")
	q.Add("features", "{\"dont_mention_me_view_api_enabled\":true,\"interactive_text_enabled\":true,\"responsive_web_uc_gql_enabled\":false,\"vibe_tweet_context_enabled\":false,\"responsive_web_edit_tweet_api_enabled\":false,\"standardized_nudges_misinfo\":false,\"responsive_web_enhance_cards_enabled\":false}")
	req.URL.RawQuery = q.Encode()
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if err = json.Unmarshal(body, &cursor); err != nil {
		return "", err
	}
	for _, i := range cursor.Data.User.Result.Timeline.Timeline.Instructions {
		for _, j := range i.Entries {
			if j.Content.Value != "" && !(strings.HasPrefix(j.Content.Value, "-1")) && j.Content.Value != "{{{{{{[]}}}}}}" {
				coord = j.Content.Value
			}
		}
	}
	if coord == "" {
		return "", nil
	}
	if err = json.Unmarshal(body, &result); err != nil {
		return "", err
	}
	var names []string
	for _, i := range result.Data.User.Result.Timeline.Timeline.Instructions {
		for _, j := range i.Entries {
			if j.Content.ItemContent.UserResults.Result.Legacy.ScreenName != "" {
				names = append(names, j.Content.ItemContent.UserResults.Result.Legacy.ScreenName)
			}
		}
	}
	err = writeFile("result.txt", names)
	if err != nil {
		exitSafely(err)
	}
	return coord, nil
}
func getUserId(cookies []*http.Cookie, headers map[string]string, name string) (string, error) {
	jar, _ := cookiejar.New(nil)
	parsedURL, _ := url.Parse("https://twitter.com/i/api/graphql/Bhlf1dYJ3bYCKmLfeEQ31A/UserByScreenName")
	jar.SetCookies(parsedURL, cookies)
	client := &http.Client{Jar: jar}
	req, err := http.NewRequest("GET", "https://twitter.com/i/api/graphql/Bhlf1dYJ3bYCKmLfeEQ31A/UserByScreenName", nil)
	if err != nil {
		return "", err
	}
	q := req.URL.Query()
	q.Add("variables", "{\"screen_name\":\""+name+"\", \"withSafetyModeUserFields\":true, \"withSuperFollowsUserFields\":true}")
	req.URL.RawQuery = q.Encode()
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var id getId
	if err = json.Unmarshal(body, &id); err != nil {
		return "", err
	}
	return id.Data.User.Result.RestID, nil
}
func writeFile(filename string, items []string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, item := range items {
		if _, err = file.WriteString(item + "\n"); err != nil {
			return err
		}
	}

	return nil
}
func appendToFile(filename string, items []string) error {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, item := range items {
		if _, err = file.WriteString(item + "\n"); err != nil {
			return err
		}
	}

	return nil
}
func printLogo() {
	logo := `
████████╗██╗    ██╗██╗████████╗████████╗███████╗██████╗       ██████╗  █████╗ ██████╗ ███████╗███████╗██████╗ 
╚══██╔══╝██║    ██║██║╚══██╔══╝╚══██╔══╝██╔════╝██╔══██╗      ██╔══██╗██╔══██╗██╔══██╗██╔════╝██╔════╝██╔══██╗
   ██║   ██║ █╗ ██║██║   ██║      ██║   █████╗  ██████╔╝█████╗██████╔╝███████║██████╔╝███████╗█████╗  ██████╔╝
   ██║   ██║███╗██║██║   ██║      ██║   ██╔══╝  ██╔══██╗╚════╝██╔═══╝ ██╔══██║██╔══██╗╚════██║██╔══╝  ██╔══██╗
   ██║   ╚███╔███╔╝██║   ██║      ██║   ███████╗██║  ██║      ██║     ██║  ██║██║  ██║███████║███████╗██║  ██║
   ╚═╝    ╚══╝╚══╝ ╚═╝   ╚═╝      ╚═╝   ╚══════╝╚═╝  ╚═╝      ╚═╝     ╚═╝  ╚═╝╚═╝  ╚═╝╚══════╝╚══════╝╚═╝  ╚═╝
                                                                                                              `

	color.Magenta(logo)

}

type js struct {
	Data struct {
		User struct {
			Result struct {
				Timeline struct {
					Timeline struct {
						Instructions []struct {
							Entries []struct {
								EntryID   string `json:"entryId"`
								SortIndex string `json:"sortIndex"`
								Content   struct {
									EntryType   string `json:"entryType"`
									ItemContent struct {
										ItemType    string `json:"itemType"`
										UserResults struct {
											Result struct {
												Legacy struct {
													ScreenName string `json:"screen_name"`
												} `json:"legacy"`
											} `json:"result"`
										} `json:"user_results"`
										UserDisplayType string `json:"userDisplayType"`
									} `json:"itemContent"`
								} `json:"content"`
							} `json:"entries,omitempty"`
						} `json:"instructions"`
					} `json:"timeline"`
				} `json:"timeline"`
			} `json:"result"`
		} `json:"user"`
	} `json:"data"`
}
type curs struct {
	Data struct {
		User struct {
			Result struct {
				Timeline struct {
					Timeline struct {
						Instructions []struct {
							Entries []struct {
								EntryID   string `json:"entryId"`
								SortIndex string `json:"sortIndex"`
								Content   struct {
									EntryType string `json:"entryType"`
									Value     string `json:"value"`
								} `json:"content"`
							} `json:"entries,omitempty"`
						} `json:"instructions"`
					} `json:"timeline"`
				} `json:"timeline"`
			} `json:"result"`
		} `json:"user"`
	} `json:"data"`
}
type getId struct {
	Data struct {
		User struct {
			Result struct {
				RestID string `json:"rest_id"`
			} `json:"result"`
		} `json:"user"`
	} `json:"data"`
}
