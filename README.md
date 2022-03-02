# LivePhish-Downloader
LivePhish downloader written in Go..
![](https://i.imgur.com/FKfmvWg.png)
[Windows, Linux and macOS binaries](https://github.com/Sorrow446/LivePhish-Downloader/releases)

# Setup
Input credentials into config file.
Configure any other options if needed.
|Option|Info|
| --- | --- |
|email|Email address.
|password|Password.
|format|Download quality. 1 = AAC 150, 2 = 16-bit / 44.1 kHz ALAC, 3 = 16-bit / 44.1 kHz FLAC.
|outPath|Where to download to. Path will be made if it doesn't already exist.

# Usage
Args take priority over the config file.

Download two albums:   
`lp_dl_x64.exe https://plus.livephish.com/#/catalog/recording/1747 https://www.livephish.com/browse/music/0,1748`

Download a single album and from two text files:   
`lp_dl_x64.exe https://plus.livephish.com/#/catalog/recording/1747 G:\1.txt G:\2.txt`

```
 __    _         _____ _   _     _      ____                _           _
|  |  |_|_ _ ___|  _  | |_|_|___| |_   |    \ ___ _ _ _ ___| |___ ___ _| |___ ___
|  |__| | | | -_|   __|   | |_ -|   |  |  |  | . | | | |   | | . | .'| . | -_|  _|
|_____|_|\_/|___|__|  |_|_|_|___|_|_|  |____/|___|_____|_|_|_|___|__,|___|___|_|

Usage: main.exe [--format FORMAT] [--outpath OUTPATH] URLS [URLS ...]

Positional arguments:
  URLS

Options:
  --format FORMAT, -f FORMAT
                         Download quality. 1 = AAC 150, 2 = 16-bit / 44.1 kHz ALAC, 3 = 16-bit / 44.1 kHz FLAC. [default: -1]
  --outpath OUTPATH, -o OUTPATH
                         Where to download to. Path will be made if it doesn't already exist.
  --help, -h             display this help and exit
  ```
 
# Disclaimer
- I will not be responsible for how you use LivePhish Downloader.    
- Nugs and LivePhish brand and names are the registered trademarks of their respective owners.    
- LivePhish Downloader has no partnership, sponsorship or endorsement with Nugs, LivePhish or Phish.
