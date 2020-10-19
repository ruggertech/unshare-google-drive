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
	b, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, drive.DriveScope, drive.DriveFileScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := auth.GetClient(config)

	ctx := context.Background()
	srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Drive client: %v", err)
	}

	firstCall := true
	nextPageToken := ""
	filesGlobalCounter := 0
	pageGlobalCounter := 0
	skipGapForPrintPurposes := 400
	pageSize := int64(200)
	// TODO: Change the following email to unshare everything in google drive with the person
	emailToUnshare := "someonesEmail@gmail.com"

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
		fmt.Printf("Page %d:\n", pageGlobalCounter)
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
						fmt.Printf("ALERT - fileName: \"%s\", fileId: \"%s\", permissionId: \"%s\" , role: %s, email: %s (%s)\n", i.Name, i.Id, per.Id, per.Role, per.EmailAddress, per.DisplayName)
						// before deleting permission, copy un-owned items
						if per.Role == "owner" {
							copiedFile := &drive.File{
								Name: i.Name,
							}

							res, err := srv.Files.Copy(i.Id, copiedFile).Do()
							if err != nil {
								fmt.Printf("An error occurred when copying: %v+\n", err)
							} else {
								fmt.Printf("copied file: %v+\n", res)
							}
						}

						fmt.Printf("Deleting the permission\n")
						delErr := srv.Permissions.Delete(i.Id, per.Id).Do()
						if delErr != nil {
							fmt.Printf("An error occurred: %v+\n", delErr)
						}
					}

					// every 20 files print to show progress
					if filesGlobalCounter%skipGapForPrintPurposes == 0 {
						fmt.Printf("Normal - id: %s, role: %s, email: %s (%s)\n", per.Id, per.Role, per.EmailAddress, per.DisplayName)
					}
				}
			}
		}
	}

}
