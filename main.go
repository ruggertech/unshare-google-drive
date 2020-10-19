package main

import (
	"context"
	"fmt"
	"github.com/ruggertech/unshare-google-drive/pkg/auth"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
	"io/ioutil"
	"log"
	"strings"
)

func main() {
	fmt.Println("Reading credential.json file")
	b, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v\n", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, drive.DriveScope, drive.DriveFileScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v\n", err)
	}
	client := auth.GetClient(config)

	fmt.Println("Retrieving google drive client configuration")

	ctx := context.Background()
	srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Drive client: %v\n", err)
	}

	firstCall := true
	nextPageToken := ""
	filesGlobalCounter := 0
	pageGlobalCounter := 0
	skipGapForPrintPurposes := 400
	pageSize := int64(200)
	// TODO: Change the following email to unshare everything in google drive with the person
	emailToUnshare := "someonesEmail@gmail.com"

	fmt.Printf("Starting to read google drive files in pages of size %d, erasing all access for "+
		"the user with the email: %s\n", pageSize, emailToUnshare)

	for nextPageToken != "" || firstCall {
		firstCall = false
		r, err := srv.Files.List().PageSize(pageSize).PageToken(nextPageToken).
			//Fields("*").Do()
			Fields("nextPageToken, files(id, name, permissions)").Do()
		if err != nil {
			log.Fatalf("Unable to retrieve files: %v", err)
		}
		nextPageToken = r.NextPageToken
		pageGlobalCounter += 1
		fmt.Printf("Reading google drive page number %d:\n", pageGlobalCounter)
		if len(r.Files) == 0 {
			fmt.Println("No files found.")
		} else {
			for _, i := range r.Files {
				filesGlobalCounter += 1
				if filesGlobalCounter%skipGapForPrintPurposes == 0 {
					fmt.Printf("%s (%s) \n", i.Name, i.Id)
				}
				for _, per := range i.Permissions {
					if strings.Contains(per.EmailAddress, emailToUnshare) {
						fmt.Printf("About to unshare - fileName: \"%s\", fileId: \"%s\", permissionId: \"%s\" , role: %s, email: %s (%s)", i.Name, i.Id, per.Id, per.Role, per.EmailAddress, per.DisplayName)
						// before deleting permission, copy un-owned items
						if per.Role == "owner" {
							copiedFile := &drive.File{
								Name: i.Name,
							}

							res, err := srv.Files.Copy(i.Id, copiedFile).Do()
							if err != nil {
								fmt.Printf("\nAn error occurred when copying: %v+\n", err)
							} else {
								fmt.Printf("- file copied to my drive: %v+\n", res)
							}
						}

						delErr := srv.Permissions.Delete(i.Id, per.Id).Do()
						if delErr != nil {
							fmt.Printf("\nAn error occurred: %v+\n", delErr)
						}
						fmt.Printf("- unshared\n")
					}

					// every 20 files print to show progress
					if filesGlobalCounter%skipGapForPrintPurposes == 0 {
						fmt.Printf("Randomly printing a file name to show progress, not unsharing it - id: %s, role: %s, email: %s (%s)\n", per.Id, per.Role, per.EmailAddress, per.DisplayName)
					}
				}
			}
		}
	}

}
