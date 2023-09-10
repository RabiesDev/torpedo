# torpedo
## What is this
Slither.io Bot tool written in Go  
Please support us with a star!

Preview:

https://github.com/RabiesDev/torpedo/assets/66158890/11064dd8-c7ec-453a-a2bb-cab3806829b3

## How to use
### Build
Clone this repository and run the executable built with `go build`  
Config can be specified via argument (default is Config in the same directory)

### Sync
First of all, please activate `sync_api`.  
Next, install [Frontier Slither](https://chrome.google.com/webstore/detail/frontier-slither/jkfiikecahagonfbnjfhjphocjlaacmc) for coordinate acquisition.  
Open the console by pressing `ctrl+shift+p` on the slither.io page,  
Drag and drop `browser_script.js` from `assets` into it.

### Config
`server_address` should be the address of the server as it is.  
`proxies` should be the path to the text file containing the proxies.  
`point_offset` can be an offset from the head

## Notes
Please run this tool at your own risk!
