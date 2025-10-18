package datetime

import (
    "encoding/json"
    "fmt"
    "strconv"
    "strings"
    "time"
)

// UnixOrTime unmarshals either a numeric Unix timestamp (seconds or milliseconds)
// or a datetime string (RFC3339 or "2006-01-02 15:04:05") into seconds since epoch.
type UnixOrTime struct {
    sec int64
}

func (u *UnixOrTime) AsUnix() int64 { return u.sec }

func (u *UnixOrTime) UnmarshalJSON(b []byte) error {
    s := strings.TrimSpace(string(b))
    if len(s) == 0 || s == "null" {
        u.sec = 0
        return nil
    }

    // Quoted string => parse as datetime
    if s[0] == '"' && s[len(s)-1] == '"' {
        var str string
        if err := json.Unmarshal(b, &str); err != nil {
            return err
        }
        str = strings.TrimSpace(str)
        if str == "" {
            u.sec = 0
            return nil
        }
        layouts := []string{
            time.RFC3339,
            "2006-01-02 15:04:05",
            "2006-01-02T15:04:05",
            "2006/01/02 15:04:05",
        }
        for _, layout := range layouts {
            if tm, err := time.ParseInLocation(layout, str, time.Local); err == nil {
                u.sec = tm.Unix()
                return nil
            }
        }
        return fmt.Errorf("invalid datetime format: %s", str)
    }

    // Numeric path
    s = strings.Trim(s, "\"")
    if strings.ContainsRune(s, '.') {
        f, err := strconv.ParseFloat(s, 64)
        if err != nil {
            return fmt.Errorf("invalid numeric timestamp: %s", s)
        }
        if f >= 1e12 { // ms
            u.sec = int64(f / 1000)
        } else {
            u.sec = int64(f)
        }
        return nil
    }
    iv, err := strconv.ParseInt(s, 10, 64)
    if err != nil {
        return fmt.Errorf("invalid integer timestamp: %s", s)
    }
    if iv >= 1e12 { // ms
        iv /= 1000
    }
    u.sec = iv
    return nil
}

