package outputs

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
)

type YoutubeCommenter struct {
	Name        string
	UserName    string
	Image       string
	Subscribers string
	ChannelLink string
}
type YoutubeVideo struct {
	Title     string
	Url       string
	Thumbnail string
	Duration  string
}
type YoutubeComment struct {
	Author  string
	Comment string
	Time    string
	Likes   string
}

func WriteCsv(contentType string, header []string, content []byte, filename string, filepath string) error {
	if content == nil || filename == "" || filepath == "" {
		return fmt.Errorf("content, filename or filepath is empty")
	}
	fmt.Println("Writing to CSV file")

	file, err := os.Create(filepath + filename)
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	switch contentType {
	case "VIDEO":
		var data []YoutubeVideo
		err = json.Unmarshal(content, &data)
		if err != nil {
			return fmt.Errorf("error unmarshalling data %s \n", err)
		}
		err = writer.Write(header)
		if err != nil {
			return fmt.Errorf("error writing header %s \n", err)
		}
		for _, video := range data {
			err = writer.Write([]string{video.Title, video.Url, video.Thumbnail})
			if err != nil {
				return fmt.Errorf("error writing data %s \n", err)
			}
		}

	case "COMMENT":
		var data []YoutubeComment
		err = json.Unmarshal(content, &data)
		if err != nil {
			return fmt.Errorf("error unmarshalling data %s \n", err)
		}
		err = writer.Write(header)
		if err != nil {
			return fmt.Errorf("error writing header %s \n", err)
		}
		for _, comment := range data {

			err = writer.Write([]string{comment.Author, comment.Comment, comment.Time, comment.Likes})
			if err != nil {
				return fmt.Errorf("error writing data %s \n", err)
			}
		}

	case "COMMENTER":
		var data []YoutubeCommenter
		err = json.Unmarshal(content, &data)
		if err != nil {
			return fmt.Errorf("error unmarshalling data %s \n", err)
		}
		err = writer.Write(header)
		if err != nil {
			return fmt.Errorf("error writing header %s \n", err)
		}
		for _, commenter := range data {
			err = writer.Write([]string{commenter.Name, commenter.UserName, commenter.Image, commenter.Subscribers, commenter.ChannelLink})
			if err != nil {
				return fmt.Errorf("error writing data %s \n", err)
			}
		}

	}

	return nil
}
