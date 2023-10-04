# torpedo
## What is this
Slither.io Bot tool written in Go  
Please support us with a star!

Preview:

https://github.com/RabiesDev/torpedo/assets/66158890/11064dd8-c7ec-453a-a2bb-cab3806829b3

## How to use
### Step-by-step guide
1. Clone this repository
2. issue npm install in sync_api
3. issue `go build` one level higher, which produces an executable
4. You need to use chromium or chrome, you might need to allow insecure content from slither.io in order to make it load (or otherwise make sure your browser doesn't redirect slither.io to https)
5. Next, install [Frontier Slither](https://chrome.google.com/webstore/detail/frontier-slither/jkfiikecahagonfbnjfhjphocjlaacmc) for coordinate acquisition.
6. edit config.json
7. edit proxies.txt if needed
8. run the executable produced in step 3 first with the param --help, which would provide an explanation how to run it.
9. cd into sync_api and run start.bat
10. Open the console in your browser while slither.io game is active by pressing `ctrl+shift+p`
11. Drag and drop `browser_script.js` from `assets` into it.

### Config
 * you can find an example in the workspace directory
 * `server_address` should be the address of the server as it is.
 * `proxies` should be the path to the text file containing the proxies.
 * `point_offset` can be an offset from the head
 * proxies.txt has a format of proto://host:port/ one per line

## Notes
Please run this tool at your own risk!
