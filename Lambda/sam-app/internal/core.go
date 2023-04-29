package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"servFunction/utils"
	"strings"

	"github.com/aws/aws-lambda-go/events"
)

type ReqInfo struct {
	DateTime      string
	Session       string
	IP            string
	Crawler       bool
	IpType        string
	UA            string
	Country       string
	SessionKey    string
	Path          string
	Method        string
	CriticalWords map[string]bool
}

type ipAPI struct {
	Status      string `json:"status"`
	CountryCode string `json:"countryCode"`
	Message     string `json:"message"`
}

func checkip(sourceIP string) string {
	// check if ip is blacklisted/whitelisted
	blacklist := [5]string{
		"1.1.1.1",
		"8.8.8.8",
		"8.8.4.4",
		"8.26.56.26",
		"9.9.9.9",
	}
	for _, ip := range blacklist {
		if ip == sourceIP {
			return utils.BlacklistStr
		}
	}
	return utils.WhitelistStr

}

func checkCountry(url string) (string, error) {
	// get country code from ip-api
	var sapi ipAPI
	resp, err := http.Get(url)
	// add context to 'ip-api' request
	if err != nil {
		return utils.Unknown, err
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return utils.Unknown, err
	}
	err = json.Unmarshal(data, &sapi)
	if err != nil {
		return utils.Unknown, err
	}
	if sapi.Status == "fail" {
		return utils.Unknown, errors.New(sapi.Message)
	}
	return sapi.CountryCode, nil

}

func (r *ReqInfo) SetSession(e events.LambdaFunctionURLRequest) {
	(*r).Session = e.RequestContext.RequestID
}

func (r *ReqInfo) SetDateTime(e events.LambdaFunctionURLRequest) {
	(*r).DateTime = e.RequestContext.Time
}

func (r *ReqInfo) GetIP(e events.LambdaFunctionURLRequest) string {
	// get ip from request and send the ip to checker function
	sourceIP := e.RequestContext.HTTP.SourceIP
	str := checkip(sourceIP)
	(*r).IpType, (*r).IP = str, sourceIP
	return fmt.Sprintf("IP is: %sed", str)

}

func (r *ReqInfo) Getcountry(e events.LambdaFunctionURLRequest) (string, error) {
	// Get country iso code and compare against blacklist countries
	blacklistCounties := [3]string{"US", "RU", "CN"}
	ipApi := utils.Api + e.RequestContext.HTTP.SourceIP
	country, err := checkCountry(ipApi)
	(*r).Country, (*r).Crawler = country, false
	if err != nil {
		return "", err
	}
	for _, ctr := range blacklistCounties {
		if ctr == country {
			(*r).Crawler = true
			break
		}
	}
	return fmt.Sprintf("Country name: %s", country), nil

}

func (r *ReqInfo) Getmethod(e events.LambdaFunctionURLRequest) (string, bool) {
	// check if incomeing requests have valid method
	msg, val := utils.Method_ok, true
	method := e.RequestContext.HTTP.Method
	(*r).Method = method
	if method == utils.POST || method == utils.TRACE || method == utils.OPTIONS {
		(*r).Crawler = true
		msg = fmt.Sprintf(utils.Method_not_allowed, method)
		val = false
	}
	return msg, val
}

func (r *ReqInfo) GetPath(e events.LambdaFunctionURLRequest) (string, bool) {
	// check if incoming request has invalid path (can log enumerations ...)
	path := e.RequestContext.HTTP.Path
	(*r).Path = path
	if path != "/" {
		(*r).Crawler = true
		return fmt.Sprintf(utils.Enumeration, path), false
	}
	return fmt.Sprintf(utils.ValidPath, path), true
}

func (r *ReqInfo) GetAgent(e events.LambdaFunctionURLRequest) (string, bool) {
	// check if request has valid/invalid user agent
	msg, val := utils.AgentOK, true
	userAgent := e.RequestContext.HTTP.UserAgent
	matchAndroid, matchIos := regexp.MustCompile(utils.AndroidRegex), regexp.MustCompile(utils.IosRegex)
	if !(matchAndroid.MatchString(userAgent) || matchIos.MatchString(userAgent)) {
		(*r).Crawler = true
		msg = fmt.Sprintf(utils.AgentNotallowed, userAgent)
		val = false
	}
	(*r).UA = userAgent
	return msg, val

}

func (r *ReqInfo) GetSessionKey(e events.LambdaFunctionURLRequest) (string, bool) {
	// check if incoming request has valid/invalid/missing session key
	key := e.Headers["SessionKey"]
	(*r).SessionKey = key
	if key != utils.SecretKey {
		(*r).Crawler = true
		return fmt.Sprintf(utils.SessionNotok, key), false
	}
	return fmt.Sprintf(utils.SessionOk, key), true

}

func (r *ReqInfo) GetBody(e events.LambdaFunctionURLRequest) (string, bool) {
	// Check if 'Body' contains suspicious words
	msg, val := utils.CriticalNotword, true
	body := e.Body
	for keyStr := range r.CriticalWords {
		boolValue := strings.Contains(strings.ToLower(body), strings.ToLower(keyStr))
		if boolValue {
			r.CriticalWords[keyStr] = true
			msg = fmt.Sprintf(utils.CriticalWord, keyStr)
			val = false
		}
	}
	return msg, val

}
