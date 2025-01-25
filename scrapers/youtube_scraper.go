package scrapers

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/coderkhushal/kbscraper/outputs"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/gocolly/colly"
)

// initialize a data structure to keep the scraped data

func ScrapeVideoComments(videoLink string, maxComments int) ([]outputs.YoutubeComment, error) {
	// check if the video link is valid or not
	if videoLink == "" {

		return nil, fmt.Errorf("Video link is empty")
	} else if strings.Contains(videoLink, "youtube.com/@") {

		return nil, fmt.Errorf("Invalid video link")
	}

	browser := rod.New().ControlURL(launcher.New().Headless(true).MustLaunch()).MustConnect()
	defer browser.MustClose()

	page := browser.MustPage(videoLink)
	page.MustWaitLoad()
	fmt.Println("Scraping comments...")
	time.Sleep(5 * time.Second)

	if _, err := page.Eval(`window.scrollBy(0, 1000)`); err != nil {

	}
	var comments []*rod.Element
	comments = page.MustElements(`ytd-comment-thread-renderer`)
	if len(comments) == 0 {
		fmt.Println("comments not found, retrying... ")
		for i := 0; i < 3; i++ {
			time.Sleep(2 * time.Second)
			comments = page.MustElements(`ytd-comment-thread-renderer`)
			if len(comments) > 0 {
				break
			}

		}
	}
	if len(comments) == 0 {

		return nil, fmt.Errorf("No comments found")
	}

	// fetch more comments
	for len(comments) < maxComments {
		if _, err := page.Eval(`window.scrollBy(0, 1000)`); err != nil {
			return nil, fmt.Errorf("something went wrong. please check link or your network connection")
		}
		time.Sleep(2 * time.Second)
		comments = page.MustElements(`ytd-comment-thread-renderer`)
	}

	fmt.Printf("%d comments found\n", len(comments))
	commentsData := []outputs.YoutubeComment{}
	for _, comment := range comments {
		author := comment.MustElement(`a#author-text`).MustText()
		comment := comment.MustElement(`yt-attributed-string`).MustText()
		commentsData = append(commentsData, outputs.YoutubeComment{
			Author:  author,
			Comment: comment,
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
		file, err := os.Create("youtube.json")
		if err != nil {

			return nil, fmt.Errorf("Error creating file %s", err)
		}
		defer file.Close()
		enc := json.NewEncoder(file)
		enc.SetIndent("", "  ")
		enc.Encode(videos)
		// fmt.Printf("Extracted youtube data %+v \n", contents)

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
	headers := []string{"Author", "Comment"}
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
