package scrapers

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/coderkhushal/kbscraper/outputs"
	"github.com/gocolly/colly"
	"github.com/tebeka/selenium"
)

var (
	chromeDriverPath = "/usr/bin/chromedriver"
	port             = 8080
)

func ScrapeVideoCommentersInfo(comments []outputs.YoutubeComment) ([]outputs.YoutubeCommenter, error) {
	// func ScrapeVideoCommentersInfo() {
	commenters := []outputs.YoutubeCommenter{}
	// time.Sleep(2 * time.Second)
	// // create selenium driver
	// opts := []selenium.ServiceOption{}
	// service, err := selenium.NewChromeDriverService(chromeDriverPath, port, opts...)
	// if err != nil {
	// 	log.Printf("Failed to start ChromeDriver service: %v", err)
	// 	return commenters, nil
	// }
	// defer service.Stop()

	// caps := selenium.Capabilities{"browserName": "chrome"}
	// // caps["goog:chromeOptions"] = map[string]interface{}{"args": []string{"--headless", "--disable-gpu"}}

	// wd, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", port))
	// if err != nil {
	// 	log.Printf("Failed to connect to WebDriver: %v", err)
	// 	return nil, err
	// }
	// defer wd.Quit()

	// for _, comment := range comments {
	// 	if comment.Author == "" {
	// 		fmt.Println("empty author")
	// 		continue
	// 	}
	// 	url := fmt.Sprintf("https://www.youtube.com/%s", comment.Author)
	// 	fmt.Println(url)
	// 	if err := wd.Get(url); err != nil {
	// 		return nil, err
	// 	}
	// 	time.Sleep(2 * time.Second)
	// 	nameElem, err := wd.FindElement(selenium.ByCSSSelector, `.page-header-view-model-wiz__page-header-title.page-header-view-model-wiz__page-header-title--page-header-title-large`)
	// 	if err != nil {
	// 		fmt.Println("error finding name element")
	// 		continue
	// 	}
	// 	name, _ := nameElem.Text()

	// 	subscribersElem, err := wd.FindElement(selenium.ByCSSSelector, `.yt-core-attributed-string.yt-content-metadata-view-model-wiz__metadata-text`)
	// 	if err != nil {
	// 		fmt.Println("error finding subscribers element")
	// 		continue
	// 	}
	// 	subscribers, _ := subscribersElem.Text()
	// 	subscribernumber := strings.Split(subscribers, " ")[0]
	// 	commenters = append(commenters, outputs.YoutubeCommenter{
	// 		Name:        name,
	// 		Subscribers: subscribernumber,
	// 	})
	// }
	wg := sync.WaitGroup{}
	wg.Add(len(comments))
	for _, comment := range comments {
		go func() {
			defer wg.Done()
			if comment.Author == "" {
				return
			}
			var ytInitialData string
			coll := colly.NewCollector(
				colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3"),
			)
			coll.OnResponse(func(r *colly.Response) {
				re := regexp.MustCompile(`var ytInitialData = (.*?);</script>`)
				matches := re.FindSubmatch(r.Body)
				if len(matches) > 1 {
					ytInitialData = string(matches[1])
				}
			})

			// Handle errors
			coll.OnError(func(_ *colly.Response, err error) {
				log.Println("Error:", err)
			})
			url := fmt.Sprintf("https://www.youtube.com/%s", comment.Author)
			coll.Visit(url)

			time.Sleep(1 * time.Second)
			a := new(map[string]interface{})
			if ytInitialData != "" {
				err := json.Unmarshal([]byte(ytInitialData), &a)
				if err != nil {
					log.Println("Error parsing JSON", err)
				}
				commenter := outputs.YoutubeCommenter{
					Name:        getName(a),
					UserName:    comment.Author,
					Subscribers: getSubscribers(a),
					Image:       getImage(a),
					ChannelLink: fmt.Sprintf("www.youtube.com/%s", comment.Author),
				}
				commenters = append(commenters, commenter)

			} else {
				fmt.Println("Error extracting youtube data.")
			}
		}()
	}
	wg.Wait()
	return commenters, nil

}
func ScrapeVideoComments(videoLink string, maxComments int) ([]outputs.YoutubeComment, error) {
	if videoLink == "" {
		return nil, fmt.Errorf("Video link is empty")
	} else if strings.Contains(videoLink, "youtube.com/@") {
		return nil, fmt.Errorf("Invalid video link")
	}

	opts := []selenium.ServiceOption{}
	service, err := selenium.NewChromeDriverService(chromeDriverPath, port, opts...)
	if err != nil {
		log.Printf("Failed to start ChromeDriver service: %v", err)
		return nil, err
	}
	defer service.Stop()

	caps := selenium.Capabilities{"browserName": "chrome"}
	caps["goog:chromeOptions"] = map[string]interface{}{"args": []string{"--headless", "--disable-gpu"}}

	wd, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", port))
	if err != nil {
		log.Printf("Failed to connect to WebDriver: %v", err)
		return nil, err
	}
	defer wd.Quit()

	if err := wd.Get(videoLink); err != nil {
		return nil, fmt.Errorf("Failed to load video link: %v", err)
	}

	fmt.Println("Scraping comments...")
	time.Sleep(10 * time.Second)

	for i := 0; i < 3; i++ {
		_, err := wd.ExecuteScript("window.scrollBy(0, 1000);", nil)
		if err != nil {
			return nil, fmt.Errorf("Failed to scroll the page: %v", err)
		}
		time.Sleep(2 * time.Second)
	}

	commentElems, err := wd.FindElements(selenium.ByCSSSelector, `ytd-comment-thread-renderer`)
	if err != nil || len(commentElems) == 0 {
		return nil, fmt.Errorf("No comments found")
	}

	for len(commentElems) < maxComments {
		_, err := wd.ExecuteScript("window.scrollBy(0, 1000);", nil)
		if err != nil {
			return nil, fmt.Errorf("Failed to fetch more comments: %v", err)
		}
		time.Sleep(2 * time.Second)

		commentElems, _ = wd.FindElements(selenium.ByCSSSelector, `ytd-comment-thread-renderer`)
	}

	fmt.Printf("%d comments found\n", len(commentElems))

	var commentsData []outputs.YoutubeComment
	for _, elem := range commentElems {
		authorElem, err := elem.FindElement(selenium.ByCSSSelector, `a#author-text`)
		if err != nil {
			fmt.Println("Error finding author element")
			continue
		}
		author, _ := authorElem.Text()

		timeElem, err := elem.FindElement(selenium.ByCSSSelector, `span#published-time-text`)
		if err != nil {
			fmt.Println("Error finding time element")
			continue
		}
		publishedTime, _ := timeElem.Text()
		likeElem, err := elem.FindElement(selenium.ByCSSSelector, `span#vote-count-middle`)
		if err != nil {
			fmt.Println("Error finding time element")
			continue
		}
		likes, _ := likeElem.Text()

		commentElem, err := elem.FindElement(selenium.ByCSSSelector, `yt-attributed-string`)
		if err != nil {
			fmt.Println("Error finding comment element")
			continue
		}
		commentText, _ := commentElem.Text()

		commentsData = append(commentsData, outputs.YoutubeComment{
			Author:  author,
			Comment: commentText,
			Time:    publishedTime,
			Likes:   likes,
		})

	}

	return commentsData, nil
}

func ScrapeChannelVideos(channelLink string) ([]outputs.YoutubeVideo, error) {

	var ytInitialData string
	videos := []outputs.YoutubeVideo{}
	C.OnResponse(func(r *colly.Response) {
		re := regexp.MustCompile(`var ytInitialData = (.*?);</script>`)
		matches := re.FindSubmatch(r.Body)
		if len(matches) > 1 {
			ytInitialData = string(matches[1])
		}
	})

	// Handle errors
	C.OnError(func(_ *colly.Response, err error) {
		log.Println("Error:", err)
	})

	// Visit the channel videos URL
	C.Visit(channelLink + "/videos")
	time.Sleep(1 * time.Second)

	a := new(map[string]interface{})
	if ytInitialData != "" {
		fmt.Println("Extracted ytInitialData")
		err := json.Unmarshal([]byte(ytInitialData), &a)
		if err != nil {
			log.Println("Error parsing JSON", err)
		}
		// check if file is already present
		_, err = os.Stat("youtube.json")
		if err == nil {
			os.Remove("youtube.json")
		}
		contents, ok := (*a)["contents"].(map[string]interface{})["twoColumnBrowseResultsRenderer"].(map[string]interface{})["tabs"].([]interface{})[1].(map[string]interface{})["tabRenderer"].(map[string]interface{})["content"].(map[string]interface{})["richGridRenderer"].(map[string]interface{})["contents"].([]interface{})
		if !ok {

			return nil, fmt.Errorf("Error extracting tabs")
		}
		if len(contents) == 0 {
			fmt.Println("No videos found.")
			return nil, fmt.Errorf("No videos found")
		}
		for _, content := range contents {
			videos = append(videos, outputs.YoutubeVideo{
				Title:     getTitle(content),
				Url:       getUrl(content),
				Thumbnail: getThumbnail(content),
				Duration:  getDuration(content),
			})
		}

	} else {
		fmt.Println("Error extracting youtube data.")
	}
	return videos, nil
}
func getUrl(SingleContent interface{}) string {
	url := ""
	if SingleContent != nil {
		content, ok := SingleContent.(map[string]interface{})

		if ok {
			richItemRenderer, ok := content["richItemRenderer"].(map[string]interface{})
			if ok {
				videoRenderer, ok := richItemRenderer["content"].(map[string]interface{})["videoRenderer"].(map[string]interface{})
				if ok {
					navigationEndpoint, ok := videoRenderer["navigationEndpoint"].(map[string]interface{})
					if ok {
						commandMetadata, ok := navigationEndpoint["commandMetadata"].(map[string]interface{})
						if ok {
							webCommandMetadata, ok := commandMetadata["webCommandMetadata"].(map[string]interface{})
							if ok {

								u, ok := webCommandMetadata["url"].(string)
								if ok {
									url = fmt.Sprintf("www.youtube.com/%s", u)
								}
							}
						}
					}
				}
			}
		}
	}
	return url
}
func getDuration(SingleContent interface{}) string {
	duration := ""
	if SingleContent != nil {
		content, ok := SingleContent.(map[string]interface{})
		if ok {
			richItemRenderer, ok := content["richItemRenderer"].(map[string]interface{})
			if ok {
				videoRenderer, ok := richItemRenderer["content"].(map[string]interface{})["videoRenderer"].(map[string]interface{})
				if ok {
					lengthText, ok := videoRenderer["lengthText"].(map[string]interface{})
					if ok {
						simpleText, ok := lengthText["simpleText"].(string)
						if ok {
							duration = simpleText
						}
					}
				}
			}
		}
	}
	return duration
}
func getTitle(SingleContent interface{}) string {
	title := ""
	if SingleContent != nil {
		content, ok := SingleContent.(map[string]interface{})
		if ok {
			richItemRenderer, ok := content["richItemRenderer"].(map[string]interface{})
			if ok {
				videoRenderer, ok := richItemRenderer["content"].(map[string]interface{})["videoRenderer"].(map[string]interface{})
				if ok {
					title, ok = videoRenderer["title"].(map[string]interface{})["simpleText"].(string)
					if !ok {
						title = videoRenderer["title"].(map[string]interface{})["runs"].([]interface{})[0].(map[string]interface{})["text"].(string)
					}
				}
			}
		}
	}
	return title
}
func getThumbnail(SingleContent interface{}) string {
	thumbnail := ""
	if SingleContent != nil {
		content, ok := SingleContent.(map[string]interface{})
		if ok {
			richItemRenderer, ok := content["richItemRenderer"].(map[string]interface{})
			if ok {
				videoRenderer, ok := richItemRenderer["content"].(map[string]interface{})["videoRenderer"].(map[string]interface{})
				if ok {
					thumbnail, ok = videoRenderer["thumbnail"].(map[string]interface{})["thumbnails"].([]interface{})[0].(map[string]interface{})["url"].(string)
				}
			}
		}
	}
	return thumbnail
}

func ExportCommentsToJSON(comments []outputs.YoutubeComment, filename string, filepath string) error {
	commentBytes, err := json.Marshal(comments)
	if err != nil {
		return fmt.Errorf("error converting comments to writeable form %s \n", err)

	}
	err = outputs.WriteJson(commentBytes, filename, filepath)
	if err != nil {
		return err
	}
	return nil
}
func ExportVideosToJSON(videos []outputs.YoutubeVideo, filename string, filepath string) error {
	videoBytes, err := json.Marshal(videos)
	if err != nil {
		return fmt.Errorf("error converting videos to writeable form %s \n", err)

	}
	err = outputs.WriteJson(videoBytes, filename, filepath)
	if err != nil {
		return err
	}
	return nil
}
func ExportCommentsToCSV(comments []outputs.YoutubeComment, filename string, filepath string) error {
	commentBytes, err := json.Marshal(comments)
	if err != nil {
		return fmt.Errorf("error converting comments to writeable form %s \n", err)

	}
	headers := []string{"Author", "Comment", "Time", "Likes"}
	err = outputs.WriteCsv("COMMENT", headers, commentBytes, filename, filepath)
	if err != nil {
		return err
	}
	return nil
}
func ExportVideosToCSV(videos []outputs.YoutubeVideo, filename string, filepath string) error {
	videoBytes, err := json.Marshal(videos)
	if err != nil {
		return fmt.Errorf("error converting videos to writeable form %s \n", err)

	}
	headers := []string{"Title", "Url", "Thumbnail"}
	err = outputs.WriteCsv("VIDEO", headers, videoBytes, filename, filepath)
	if err != nil {
		return err
	}
	return nil
}
func ExportCommentersToCSV(commenters []outputs.YoutubeCommenter, filename string, filepath string) error {
	commenterBytes, err := json.Marshal(commenters)
	if err != nil {
		return fmt.Errorf("error converting commenters to writeable form %s \n", err)

	}
	headers := []string{"Name", "Username", "Image", "Subscribers", "ChannelLink"}
	err = outputs.WriteCsv("COMMENTER", headers, commenterBytes, filename, filepath)
	if err != nil {
		return err
	}
	return nil
}
func ExportCommentersToJson(commenters []outputs.YoutubeCommenter, filename string, filepath string) error {
	commenterBytes, err := json.Marshal(commenters)
	if err != nil {
		return fmt.Errorf("error converting commenters to writeable form %s \n", err)

	}
	err = outputs.WriteJson(commenterBytes, filename, filepath)
	if err != nil {
		return err
	}
	return nil
}
func createSeleniumDriver() (*selenium.WebDriver, error) {
	opts := []selenium.ServiceOption{}
	service, err := selenium.NewChromeDriverService(chromeDriverPath, port, opts...)
	if err != nil {
		log.Printf("Failed to start ChromeDriver service: %v", err)
		return nil, err
	}
	defer service.Stop()

	caps := selenium.Capabilities{"browserName": "chrome"}
	// caps["goog:chromeOptions"] = map[string]interface{}{"args": []string{"--headless", "--disable-gpu"}}

	wd, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", port))
	if err != nil {
		log.Printf("Failed to connect to WebDriver: %v", err)
		return nil, err
	}
	defer wd.Quit()
	return &wd, nil
}
func getImage(a *map[string]interface{}) string {
	metadata, ok := (*a)["metadata"].(map[string]interface{})
	if ok {
		channelMetadataRenderer, ok := metadata["channelMetadataRenderer"].(map[string]interface{})
		if ok {
			avatar, ok := channelMetadataRenderer["avatar"].(map[string]interface{})

			if ok {
				images, ok := avatar["thumbnails"].([]interface{})
				if ok {
					image := images[0].(map[string]interface{})["url"].(string)
					return image
				}
			}

		}
	}

	return ""
}
func getSubscribers(a *map[string]interface{}) string {
	header, ok := (*a)["header"].(map[string]interface{})
	if ok {
		pageHeaderRenderer, ok := header["pageHeaderRenderer"].(map[string]interface{})
		if ok {
			content, ok := pageHeaderRenderer["content"].(map[string]interface{})
			if ok {
				pageHeaderViewModel, ok := content["pageHeaderViewModel"].(map[string]interface{})
				if ok {
					metadata, ok := pageHeaderViewModel["metadata"].(map[string]interface{})
					if ok {
						contentMetadataViewModel, ok := metadata["contentMetadataViewModel"].(map[string]interface{})
						if ok {

							metadataRows, ok := contentMetadataViewModel["metadataRows"].([]interface{})
							if ok {

								if len(metadataRows) > 1 {

									metadataParts, ok := metadataRows[1].(map[string]interface{})["metadataParts"].([]interface{})
									if ok {

										text, ok := metadataParts[0].(map[string]interface{})["text"].(map[string]interface{})
										if ok {

											return strings.Split(text["content"].(string), " ")[0]
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}

	return ""
}
func getName(a *map[string]interface{}) string {
	header, ok := (*a)["header"].(map[string]interface{})
	if ok {
		pageHeaderRenderer, ok := header["pageHeaderRenderer"].(map[string]interface{})
		if ok {
			pageTitle, ok := pageHeaderRenderer["pageTitle"].(string)
			if ok {
				return pageTitle
			}
		}

	}

	return ""
}
func getVideosLength(a *map[string]interface{}) string {
	contents, ok := (*a)["contents"].(map[string]interface{})["twoColumnBrowseResultsRenderer"].(map[string]interface{})["tabs"].([]interface{})[1].(map[string]interface{})["tabRenderer"].(map[string]interface{})["content"].(map[string]interface{})["richGridRenderer"].(map[string]interface{})["contents"].([]interface{})
	if !ok {
		return ""
	}
	return fmt.Sprintf("%d", len(contents))
}
