package cmd

import (
        "github.com/spf13/cobra"
        "github.com/gholt/blackfridaytext"
        "os"
        "fmt"
        "net/http"
        "io/ioutil"
        "regexp"
)

const catalogBucket = "https://s3-eu-west-1.amazonaws.com/dev.stack.foundation/catalog/"

var docsImage string

var docsCmd = &cobra.Command{
        Use:   "docs",
        Short: "Show documentation about a particular official Docker image",
        Long:  `Show documentation available about a particular official Docker image.

Shows documentation available about a particular offical Docker image.`,
        Run: func(command *cobra.Command, args []string) {
                if len(docsImage) < 1 {
                        fmt.Println("Specify a valid image to print documentation")
                        command.Help()
                        return
                }

                response, err := http.Get(catalogBucket + docsImage + ".md")
                if err != nil {
                        fmt.Printf("Could not retrieve documentation for %v - is it an official Docker image?", args[0])
                        fmt.Println("Try running the \"catalog\" command to list all Docker offical images")
                        return
                }
                defer response.Body.Close()

                if response.StatusCode != 200 {
                        fmt.Printf("Could not retrieve documentation for %v - is it an official Docker image?", args[0])
                        fmt.Println("Try running the \"catalog\" command to list all Docker offical images")
                        return
                }

                markdown, _ := ioutil.ReadAll(response.Body)

                commentRegex, _ := regexp.Compile("(?s)<!--.*-->")
                logoRegex, _ := regexp.Compile("!\\[logo].*png\\)")
                linkRegex, _ := regexp.Compile("\\[(.+?)]\\((.+?)\\)")

                markdown = commentRegex.ReplaceAll(markdown, []byte(""))
                markdown = logoRegex.ReplaceAll(markdown, []byte(""))
                markdown = linkRegex.ReplaceAll(markdown, []byte("$1 ($2)"))

                opt := &blackfridaytext.Options{
                        Color: false,
                        HeaderPrefix: []byte("=="),
                        HeaderSuffix: []byte("=="),
                }
                _, output := blackfridaytext.MarkdownToText(markdown, opt)

                emphasisRegex, _ := regexp.Compile("\\*\\*([[:alnum:]].+[[:alnum:]])\\*\\*")
                output = emphasisRegex.ReplaceAll(output, []byte("$1"))

                emphasisRegex, _ = regexp.Compile("\\*([[:alnum:]].+[[:alnum:]])\\*")
                output = emphasisRegex.ReplaceAll(output, []byte("$1"))

                os.Stdout.Write(output)
        },
}

func init() {
        docsCmd.Flags().StringVarP(&docsImage, "image", "i", "", "Image to display documentation for")
        RootCmd.AddCommand(docsCmd)
}