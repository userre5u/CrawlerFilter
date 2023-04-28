package utils

const (
	BlacklistStr       = "Blacklist"
	WhitelistStr       = "Whitelist"
	Unknown            = "UNKNOWN"
	Fail               = "fail"
	SecretKey          = "ThisIsSecret"
	Api                = "http://ip-api.com/json/"
	Method_ok          = "[+] Method is ok"
	Method_not_allowed = "[!] Method is not allowed: %s"
	Enumeration        = "[!] Someone is doing something nasty... (path=%s)"
	ValidPath          = "[+] Path is valid: %s"
	AgentOK            = "[+] User-Agent is OK"
	AgentNotallowed    = "[!] User-Agent is not allowed: user-agent: %s"
	SessionNotok       = "[!] Invalid session key: %s"
	SessionOk          = "[+] Valid session key: %s"
	CriticalWord       = "[!] Critical word found: %s"
	CriticalNotword    = "[+] No critical words found on request"
	OPTIONS            = "OPTIONS"
	TRACE              = "TRACE"
	POST               = "POST"
	AndroidRegex       = `Mozilla\/5\.0 \(Linux; [a|A]ndroid \d{1,2}; SM-.{1,5}\) AppleWebKit\/\d{3}\.\d{2} \(KHTML, like Gecko\) Chrome\/[7|8|9]\d\.\d{1,5}\.\d{1,5}\.\d{1,5} Mobile Safari\/\d{3}\.\d{2}`
	IosRegex           = `Mozilla\/5.0 \(iPhone(\d{1,2},\d)?; (U;)?\s?CPU iPhone OS \d{2}(_\d){1,2} like Mac OS X\) AppleWebKit\/\d{3}\.\d{1,2}\.\d{1,2} \(KHTML, like Gecko\) Version\/\d{2}\.\d(\.\d)? Mobile\/[A-Za-z0-9]{6} Safari\/\d{3}\.\d{1,2}`
)
