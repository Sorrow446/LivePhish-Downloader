package main

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/alexflint/go-arg"
	"github.com/dustin/go-humanize"
)

const (
	clientId        = "Fujeij8d764ydxcnh4676scsr7f4"
	developerKey    = "njeurd876frhdjxy6sxxe721"
	sigKey          = "jdfirj8475jf_"
	timestampLayout = "01/02/2006 15:04:05"
	userAgent       = "LivePhish/3.4.5.357 (Android; 7.1.2; Asus; ASUS_Z01QD)"
	apiBase         = "https://www.livephish.com/"
	apiBase2        = "https://id.livephish.com/connect/"
	plusUrl         = "https://plus.livephish.com/"
)

var (
	jar, _       = cookiejar.New(nil)
	client       = &http.Client{Jar: jar}
	regexStrings = [2]string{
		`^https://plus.livephish.com/#/catalog/recording/(\d+)$`,
		`^https://www.livephish.com/browse/music/0,(\d+)/[\w-]+$`,
	}
)

func (wc *WriteCounter) Write(p []byte) (int, error) {
	var speed int64 = 0
	n := len(p)
	wc.Downloaded += int64(n)
	percentage := float64(wc.Downloaded) / float64(wc.Total) * float64(100)
	wc.Percentage = int(percentage)
	toDivideBy := time.Now().UnixMilli() - wc.StartTime
	if toDivideBy != 0 {
		speed = int64(wc.Downloaded) / toDivideBy * 1000
	}
	fmt.Printf("\r%d%% @ %s/s, %s/%s ", wc.Percentage, humanize.Bytes(uint64(speed)),
		humanize.Bytes(uint64(wc.Downloaded)), wc.TotalStr)
	return n, nil
}

func handleErr(errText string, err error, _panic bool) {
	errString := fmt.Sprintf("%s\n%s", errText, err)
	if _panic {
		panic(errString)
	}
	fmt.Println(errString)
}

func wasRunFromSrc() bool {
	buildPath := filepath.Join(os.TempDir(), "go-build")
	return strings.HasPrefix(os.Args[0], buildPath)
}

func getScriptDir() (string, error) {
	var (
		ok    bool
		err   error
		fname string
	)
	runFromSrc := wasRunFromSrc()
	if runFromSrc {
		_, fname, _, ok = runtime.Caller(0)
		if !ok {
			return "", errors.New("Failed to get script filename.")
		}
	} else {
		fname, err = os.Executable()
		if err != nil {
			return "", err
		}
	}
	return filepath.Dir(fname), nil
}

func readTxtFile(path string) ([]string, error) {
	var lines []string
	f, err := os.OpenFile(path, os.O_RDONLY, 0755)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			lines = append(lines, line)
		}
	}
	if scanner.Err() != nil {
		return nil, scanner.Err()
	}
	return lines, nil
}

func contains(lines []string, value string) bool {
	for _, line := range lines {
		if strings.EqualFold(line, value) {
			return true
		}
	}
	return false
}

func processUrls(urls []string) ([]string, error) {
	var (
		processed []string
		txtPaths  []string
	)
	for _, _url := range urls {
		if strings.HasSuffix(_url, ".txt") && !contains(txtPaths, _url) {
			txtLines, err := readTxtFile(_url)
			if err != nil {
				return nil, err
			}
			for _, txtLine := range txtLines {
				if !contains(processed, txtLine) {
					processed = append(processed, txtLine)
				}
			}
			txtPaths = append(txtPaths, _url)
		} else {
			if !contains(processed, _url) {
				processed = append(processed, _url)
			}
		}
	}
	return processed, nil
}

func parseCfg() (*Config, error) {
	resolveFormat := map[int]int{
		1: 4,
		2: 2,
		3: 3,
	}
	cfg, err := readConfig()
	if err != nil {
		return nil, err
	}
	args := parseArgs()
	if args.Format != -1 {
		cfg.Format = args.Format
	}
	if !(cfg.Format >= 1 && cfg.Format <= 3) {
		return nil, errors.New("Format must be between 1 and 3.")
	}
	if !(cfg.EpochCompensation >= 0 && cfg.EpochCompensation <= 100) {
		return nil, errors.New("Epoch compensation must be between 0 and 100.")
	}
	cfg.Format = resolveFormat[cfg.Format]
	if args.OutPath != "" {
		cfg.OutPath = args.OutPath
	}
	if cfg.OutPath == "" {
		cfg.OutPath = "LivePhish downloads"
	}
	cfg.Urls, err = processUrls(args.Urls)
	if err != nil {
		errString := fmt.Sprintf("Failed to process URLs.\n%s", err)
		return nil, errors.New(errString)
	}
	return cfg, nil
}

func readConfig() (*Config, error) {
	data, err := ioutil.ReadFile("config.json")
	if err != nil {
		return nil, err
	}
	var obj Config
	err = json.Unmarshal(data, &obj)
	if err != nil {
		return nil, err
	}
	return &obj, nil
}

func parseArgs() *Args {
	var args Args
	arg.MustParse(&args)
	return &args
}

func makeDirs(path string) error {
	return os.MkdirAll(path, 0755)
}

func fileExists(path string) (bool, error) {
	f, err := os.Stat(path)
	if err == nil {
		return !f.IsDir(), nil
	} else if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func sanitize(filename string) string {
	regex := regexp.MustCompile(`[\/:*?"><|]`)
	sanitized := regex.ReplaceAllString(filename, "_")
	return sanitized
}

func auth(email, pwd string) (string, error) {
	data := url.Values{}
	data.Set("client_id", clientId)
	data.Set("grant_type", "password")
	data.Set("scope", "offline_access nugsnet:api nugsnet:legacyapi")
	data.Set("username", email)
	data.Set("password", pwd)
	req, err := http.NewRequest(http.MethodPost, apiBase2+"token", strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Add("User-Agent", userAgent)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	do, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer do.Body.Close()
	if do.StatusCode != http.StatusOK {
		return "", errors.New(do.Status)
	}
	var obj Auth
	err = json.NewDecoder(do.Body).Decode(&obj)
	if err != nil {
		return "", err
	}
	return obj.AccessToken, nil
}

func getUsertoken(email, pwd string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, apiBase+"secureApi.aspx", nil)
	if err != nil {
		return "", err
	}
	query := url.Values{}
	query.Set("method", "session.getUserToken")
	query.Set("clientID", clientId)
	query.Set("developerKey", developerKey)
	query.Set("user", email)
	query.Set("pw", pwd)
	req.URL.RawQuery = query.Encode()
	req.Header.Add("User-Agent", userAgent)
	do, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer do.Body.Close()
	if do.StatusCode != http.StatusOK {
		return "", errors.New(do.Status)
	}
	var obj UserToken
	err = json.NewDecoder(do.Body).Decode(&obj)
	if err != nil {
		return "", err
	}
	return obj.Response.TokenValue, nil
}

func getUserInfo(token string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, apiBase2+"userinfo", nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("User-Agent", userAgent)
	do, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer do.Body.Close()
	if do.StatusCode != http.StatusOK {
		return "", errors.New(do.Status)
	}
	var obj UserInfo
	err = json.NewDecoder(do.Body).Decode(&obj)
	if err != nil {
		return "", err
	}
	return obj.Sub, nil
}

func getSubInfo(email, token, sessToken string) (*SubInfo, error) {
	req, err := http.NewRequest(http.MethodGet, apiBase+"secureApi.aspx", nil)
	if err != nil {
		return nil, err
	}
	query := url.Values{}
	query.Set("method", "user.getSubscriberInfo")
	query.Set("developerKey", developerKey)
	query.Set("user", email)
	query.Set("token", sessToken)
	req.URL.RawQuery = query.Encode()
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("User-Agent", userAgent)
	do, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer do.Body.Close()
	if do.StatusCode != http.StatusOK {
		return nil, errors.New(do.Status)
	}
	var obj SubInfo
	err = json.NewDecoder(do.Body).Decode(&obj)
	if err != nil {
		return nil, err
	}
	return &obj, nil
}

func parseTimestamps(start, end string) (string, string) {
	startTime, _ := time.Parse(timestampLayout, start)
	endTime, _ := time.Parse(timestampLayout, end)
	parsedStart := strconv.FormatInt(startTime.Unix(), 10)
	parsedEnd := strconv.FormatInt(endTime.Unix(), 10)
	return parsedStart, parsedEnd
}

func parseStreamParams(subInfo *SubInfo) *StreamParams {
	streamParams := &StreamParams{
		SubscriptionID:          subInfo.Response.SubscriptionInfo.SubscriptionID,
		SubCostplanIDAccessList: subInfo.Response.SubscriptionInfo.SubCostplanIDAccessList,
		UserID:                  strconv.Itoa(subInfo.Response.SubscriptionInfo.UserID),
		StartStamp:              strconv.Itoa(subInfo.Response.SubscriptionInfo.StartDateStamp),
		EndStamp:                strconv.Itoa(subInfo.Response.SubscriptionInfo.EndDateStamp),
	}
	return streamParams
}

func checkUrl(url string) string {
	for _, regexString := range regexStrings {
		regex := regexp.MustCompile(regexString)
		match := regex.FindStringSubmatch(url)
		if match != nil {
			return match[1]
		}
	}
	return ""
}

func getAlbumMeta(albumId string) (*AlbumMeta, error) {
	req, err := http.NewRequest(http.MethodGet, apiBase+"api.aspx", nil)
	if err != nil {
		return nil, err
	}
	query := url.Values{}
	query.Set("method", "catalog.container")
	query.Set("containerID", albumId)
	query.Set("vdisp", "1")
	req.URL.RawQuery = query.Encode()
	req.Header.Add("User-Agent", userAgent)
	do, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer do.Body.Close()
	if do.StatusCode != http.StatusOK {
		return nil, errors.New(do.Status)
	}
	var obj AlbumMeta
	err = json.NewDecoder(do.Body).Decode(&obj)
	if err != nil {
		return nil, err
	}
	return &obj, nil
}

func generateSig(epochComp int64) (string, string, error) {
	timestamp := time.Now().Unix() + epochComp
	timestampStr := strconv.FormatInt(timestamp, 10)
	h := md5.New()
	_, err := h.Write([]byte(sigKey + timestampStr))
	if err != nil {
		return "", "", err
	}
	return hex.EncodeToString(h.Sum(nil)), timestampStr, nil
}

func getStreamMeta(trackId, format int, streamParams *StreamParams, epochComp int64) (string, error) {
	req, err := http.NewRequest(http.MethodGet, apiBase+"bigriver/subPlayer.aspx", nil)
	if err != nil {
		return "", err
	}
	query := url.Values{}
	sig, timestamp, err := generateSig(epochComp)
	if err != nil {
		return "", err
	}
	query.Set("trackID", strconv.Itoa(trackId))
	query.Set("app", "1")
	query.Set("platformID", strconv.Itoa(format))
	query.Set("subscriptionID", streamParams.SubscriptionID)
	query.Set("subCostplanIDAccessList", streamParams.SubCostplanIDAccessList)
	query.Set("nn_userID", streamParams.UserID)
	query.Set("startDateStamp", streamParams.StartStamp)
	query.Set("endDateStamp", streamParams.EndStamp)
	query.Set("tk", sig)
	query.Set("lxp", timestamp)
	req.URL.RawQuery = query.Encode()
	req.Header.Add("User-Agent", "LivePhishAndroid")
	do, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer do.Body.Close()
	if do.StatusCode != http.StatusOK {
		return "", errors.New(do.Status)
	}
	var obj StreamMeta
	err = json.NewDecoder(do.Body).Decode(&obj)
	if err != nil {
		return "", err
	}
	return obj.StreamLink, nil
}

func queryQuality(streamUrl string) *Quality {
	qualityMap := map[string]Quality{
		".alac16/": {Specs: "16-bit / 44.1 kHz ALAC", Extension: ".m4a"},
		".flac16/": {Specs: "16-bit / 44.1 kHz FLAC", Extension: ".flac"},
		".aac150/": {Specs: "AAC 150", Extension: ".m4a"},
	}
	for k, v := range qualityMap {
		if strings.Contains(streamUrl, k) {
			return &v
		}
	}
	return nil
}

func downloadTrack(trackPath, url string) error {
	f, err := os.OpenFile(trackPath, os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		return err
	}
	defer f.Close()
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Referer", plusUrl)
	req.Header.Add("User-Agent", userAgent)
	req.Header.Add("Range", "bytes=0-")
	do, err := client.Do(req)
	if err != nil {
		return err
	}
	defer do.Body.Close()
	if do.StatusCode != http.StatusOK && do.StatusCode != http.StatusPartialContent {
		return errors.New(do.Status)
	}
	totalBytes := do.ContentLength
	counter := &WriteCounter{
		Total:     totalBytes,
		TotalStr:  humanize.Bytes(uint64(totalBytes)),
		StartTime: time.Now().UnixMilli(),
	}
	_, err = io.Copy(f, io.TeeReader(do.Body, counter))
	fmt.Println("")
	return err
}

func init() {
	fmt.Println(`
 __    _         _____ _   _     _      ____                _           _         
|  |  |_|_ _ ___|  _  | |_|_|___| |_   |    \ ___ _ _ _ ___| |___ ___ _| |___ ___ 
|  |__| | | | -_|   __|   | |_ -|   |  |  |  | . | | | |   | | . | .'| . | -_|  _|
|_____|_|\_/|___|__|  |_|_|_|___|_|_|  |____/|___|_____|_|_|_|___|__,|___|___|_|  
`)
}

func main() {
	scriptDir, err := getScriptDir()
	if err != nil {
		panic(err)
	}
	err = os.Chdir(scriptDir)
	if err != nil {
		panic(err)
	}
	cfg, err := parseCfg()
	if err != nil {
		handleErr("Failed to parse config file.", err, true)
	}
	err = makeDirs(cfg.OutPath)
	if err != nil {
		handleErr("Failed to make output folder.", err, true)
	}
	token, err := auth(cfg.Email, cfg.Password)
	if err != nil {
		handleErr("Failed to auth.", err, true)
	}
	userToken, err := getUsertoken(cfg.Email, cfg.Password)
	if err != nil {
		handleErr("Failed to get session token.", err, true)
	}
	subInfo, err := getSubInfo(cfg.Email, token, userToken)
	if err != nil {
		handleErr("Failed to get subcription info.", err, true)
	}
	if !subInfo.Response.SubscriptionInfo.CanStreamSubContent {
		panic("Account subscription required.")
	}
	fmt.Println(
		"Signed in successfully - " + subInfo.Response.SubscriptionInfo.PlanName + "\n",
	)
	streamParams := parseStreamParams(subInfo)
	albumTotal := len(cfg.Urls)
	for albumNum, url := range cfg.Urls {
		fmt.Printf("Album %d of %d:\n", albumNum+1, albumTotal)
		albumId := checkUrl(url)
		if albumId == "" {
			fmt.Println("Invalid URL:", url)
			continue
		}
		meta, err := getAlbumMeta(albumId)
		if err != nil {
			handleErr("Failed to get album metadata.", err, false)
			continue
		}
		albumFolder := meta.Response.ArtistName + " - " + strings.TrimRight(meta.Response.ContainerInfo, " ")
		fmt.Println(albumFolder)
		if len(albumFolder) > 120 {
			albumFolder = albumFolder[:120]
			fmt.Println("Album folder name was chopped as it exceeds 120 characters.")
		}
		albumPath := filepath.Join(cfg.OutPath, sanitize(albumFolder))
		err = makeDirs(albumPath)
		if err != nil {
			handleErr("Failed to make album folder.", err, false)
			continue
		}
		trackTotal := len(meta.Response.Tracks)
		for trackNum, track := range meta.Response.Tracks {
			trackNum++
			streamUrl, err := getStreamMeta(track.TrackID, cfg.Format, streamParams, cfg.EpochCompensation)
			if err != nil {
				handleErr("Failed to get track stream metadata.", err, false)
				continue
			}
			if streamUrl == "" {
				fmt.Println(
					"The API didn't return a track stream URL. " +
						"Try increasing or decreasing the epoch compensation option.",
				)
				continue
			}
			quality := queryQuality(streamUrl)
			if quality == nil {
				fmt.Println("The API returned an unsupported format.")
				continue
			}
			trackFname := fmt.Sprintf(
				"%02d. %s%s", trackNum, sanitize(track.SongTitle), quality.Extension,
			)
			trackPath := filepath.Join(albumPath, trackFname)
			exists, err := fileExists(trackPath)
			if err != nil {
				handleErr("Failed to check if track already exists locally.", err, false)
				continue
			}
			if exists {
				fmt.Println("Track already exists locally.")
				continue
			}
			fmt.Printf(
				"Downloading track %d of %d: %s - %s\n", trackNum, trackTotal, track.SongTitle,
				quality.Specs,
			)
			err = downloadTrack(trackPath, streamUrl)
			if err != nil {
				handleErr("Failed to download track.", err, false)
			}
		}
	}
}
