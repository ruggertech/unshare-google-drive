# unshare-google-drive

Scans your google drive and un-shares any file shared with the person whose email you provided to the script

1. Inside main.go change the email address to the email of the person you wish to remove from 
all the files in your google drive.
2. Click  `enable drive api` in https://developers.google.com/drive/api/v3/quickstart/go
and download the file `credentials.json`.  Put it in the application directory.
3. In the command line type:
`make run` and follow the instructions

The code uses google drive apis v3:  https://developers.google.com/drive/api/v3/reference
 