# redditpic #
A fairly simple program to import images from specified subreddit and show in desired image viewer(default: sxiv) written in go. 

## Usage ##
```
Usage: redditpic [options]
  -a string
    	Sorts posts based on new,hot,top
  -p string
    	Specify the time range of posts, depends on -a being top
  -s string
    	Specify subreddit to import images from
  -v string
    	Program to open images (default "sxiv")
```

### Example ###
`./redditpic -s MinimalWallpaper -p month -a top -v sxiv`
