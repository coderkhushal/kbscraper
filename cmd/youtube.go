/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"strconv"

	"github.com/charmbracelet/huh"
	"github.com/coderkhushal/kbscraper/scrapers"
	"github.com/pterm/pterm"
	"github.com/pterm/pterm/putils"
	"github.com/spf13/cobra"
)

var youtubeCmd = &cobra.Command{
	Use:   "youtube",
	Short: "Scrape youtube data",
	Long:  `Scrape youtube data by providing the channel URL`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("\n")
		letters := pterm.DefaultBigText.WithLetters(putils.LettersFromString("KbScraper"))
		letters.Render()

		variant := ""
		form := huh.NewForm(
			huh.NewGroup(
				// Ask the user for a base burger and toppings.
				huh.NewSelect[string]().
					Title("What you want to Scrape?").
					Options(
						huh.NewOption("Scrape posts", "Posts"),
						huh.NewOption("Scrape comments", "Comments"),
						huh.NewOption("scrape commenters", "Commenters"),
					).
					Value(&variant), // store the chosen option in the "burger" variable
			),
		)
		if err := form.Run(); err != nil {
			fmt.Println(err)
			return
		}
		if variant == "" {
			return
		} else if variant == "Posts" {
			postCommand()

		} else if variant == "Comments" {
			commentsCommand()

		} else if variant == "Commenters" {
			commentersCommand()
		}
	},
}

// handles the posts input and call the scraper function
func postCommand() {
	var url string = ""
	huh.NewInput().
		Title("Enter the Channel Url ").
		Value(&url).
		Run()
	if url == "" {
		fmt.Println("url is empty ")
		return
	}
	videos, err := scrapers.ScrapeChannelVideos(url)

	if err != nil {
		fmt.Println(err)
		return
	}

	var exportType string = ""
	form := huh.NewForm(
		huh.NewGroup(

			huh.NewSelect[string]().
				Title("Select the format of file for exporting?").
				Options(
					huh.NewOption("CSV", "CSV"),
					huh.NewOption("JSON", "JSON"),
				).
				Value(&exportType),
		),
	)
	if err := form.Run(); err != nil {
		fmt.Println(err)
		return
	}
	if exportType == "" {
		return
	}
	if exportType == "CSV" {

		filename, filepath := getFileDetailsInput()
		err := scrapers.ExportVideosToCSV(videos, filename, filepath)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("Exported Successfully")
		}

	} else if exportType == "JSON" {
		filename, filepath := getFileDetailsInput()
		err := scrapers.ExportVideosToJSON(videos, filename, filepath)
		if err != nil {

			fmt.Println(err)
		} else {
			fmt.Println("Exported Successfully")
		}

	}
}

func commentersCommand() {

	var url string = ""
	huh.NewInput().
		Title("Enter the video Url ").
		Value(&url).
		Run()
	var maxcommentsstr string
	huh.NewInput().
		Title("Enter maximum number of comments you want to scrape ? eg : 10   ").
		Value(&maxcommentsstr).
		Run()
	maxcomments, err := strconv.Atoi(maxcommentsstr)
	if err != nil {
		fmt.Println(err)
		return
	}
	comments, err := scrapers.ScrapeVideoComments(url, maxcomments)

	if err != nil {
		fmt.Println(err)
		return
	}
	commenters, err := scrapers.ScrapeVideoCommentersInfo(comments)
	if err != nil {
		fmt.Println(err)
		return
	}
	var exportType string = ""
	form := huh.NewForm(
		huh.NewGroup(
			// Ask the user for a base burger and toppings.
			huh.NewSelect[string]().
				Title("Select the format of file for exporting?").
				Options(
					huh.NewOption("CSV", "CSV"),
					huh.NewOption("JSON", "JSON"),
				).
				Value(&exportType), // store the chosen option in the "burger" variable
		),
	)
	if err := form.Run(); err != nil {
		fmt.Println(err)
		return
	}
	if exportType == "" {
		return
	}
	if exportType == "CSV" {
		filename, filepath := getFileDetailsInput()
		err := scrapers.ExportCommentersToCSV(commenters, filename, filepath)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("Exported Successfully")
		}
	} else if exportType == "JSON" {
		filename, filepath := getFileDetailsInput()
		err := scrapers.ExportCommentersToJson(commenters, filename, filepath)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("Exported Successfully")
		}

	}
}

// handles the comments input and call the scraper function
func commentsCommand() {
	var url string = ""
	huh.NewInput().
		Title("Enter the video Url ").
		Value(&url).
		Run()
	var maxcommentsstr string
	huh.NewInput().
		Title("Enter maximum number of comments you want to scrape ? eg : 10   ").
		Value(&maxcommentsstr).
		Run()
	maxcomments, err := strconv.Atoi(maxcommentsstr)
	if err != nil {
		fmt.Println(err)
		return
	}
	comments, err := scrapers.ScrapeVideoComments(url, maxcomments)

	if err != nil {
		fmt.Println(err)
		return
	}
	var exportType string = ""
	form := huh.NewForm(
		huh.NewGroup(
			// Ask the user for a base burger and toppings.
			huh.NewSelect[string]().
				Title("Select the format of file for exporting?").
				Options(
					huh.NewOption("CSV", "CSV"),
					huh.NewOption("JSON", "JSON"),
				).
				Value(&exportType), // store the chosen option in the "burger" variable
		),
	)
	if err := form.Run(); err != nil {
		fmt.Println(err)
		return
	}
	if exportType == "" {
		return
	}
	if exportType == "CSV" {
		filename, filepath := getFileDetailsInput()
		err := scrapers.ExportCommentsToCSV(comments, filename, filepath)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("Exported Successfully")
		}
	} else if exportType == "JSON" {
		filename, filepath := getFileDetailsInput()
		err := scrapers.ExportCommentsToJSON(comments, filename, filepath)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("Exported Successfully")
		}

	}
}
func getFileDetailsInput() (string, string) {
	// return filename, filepath
	var (
		filename string
		filepath string
	)
	huh.NewInput().
		Title("Enter the filename ").
		Value(&filename).
		Run()

	huh.NewInput().
		Title("Enter the filepath ").
		Value(&filepath).
		Run()

	return filename, filepath
}

func init() {
	rootCmd.AddCommand(youtubeCmd)

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// youtubeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// youtubeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
