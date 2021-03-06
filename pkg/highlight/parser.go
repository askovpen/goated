package highlight

import (
	"errors"
	"fmt"
	"regexp"

	"gopkg.in/yaml.v3"
)

// A Group represents a syntax group
type Group uint8

// Groups contains all of the groups that are defined
// You can access them in the map via their string name
var Groups map[string]Group
var numGroups Group

// String returns the group name attached to the specific group
func (g Group) String() string {
	for k, v := range Groups {
		if v == g {
			return k
		}
	}
	return ""
}

// A Def is a full syntax definition for a language
// It has a filetype, information about how to detect the filetype based
// on filename or header (the first line of the file)
// Then it has the rules which define how to highlight the file
type Def struct {
	*Header

	rules *rules
}

// Header struct
type Header struct {
	FileType string
	FtDetect [2]*regexp.Regexp
}

// File struct
type File struct {
	FileType string

	yamlSrc map[string]interface{}
}

// A Pattern is one simple syntax rule
// It has a group that the rule belongs to, as well as
// the regular expression to match the pattern
type pattern struct {
	group Group
	regex *regexp.Regexp
}

// rules defines which patterns and regions can be used to highlight
// a filetype
type rules struct {
	regions  []*region
	patterns []*pattern
}

// A region is a highlighted region (such as a multiline comment, or a string)
// It belongs to a group, and has start and end regular expressions
// A region also has rules of its own that only apply when matching inside the
// region and also rules from the above region do not match inside this region
// Note that a region may contain more regions
type region struct {
	group      Group
	limitGroup Group
	parent     *region
	start      *regexp.Regexp
	end        *regexp.Regexp
	skip       *regexp.Regexp
	rules      *rules
}

func init() {
	Groups = make(map[string]Group)
}

// ParseFtDetect func
func ParseFtDetect(file *File) (r [2]*regexp.Regexp, err error) {
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("pkg: %v", r)
			}
		}
	}()

	rules := file.yamlSrc
	loaded := 0
	for k, v := range rules {
		if k == "detect" {
			ftdetect := v.(map[string]interface{})
			if len(ftdetect) >= 1 {
				syntax, err := regexp.Compile(ftdetect["filename"].(string))
				if err != nil {
					return r, err
				}

				r[0] = syntax
			}
			if len(ftdetect) >= 2 {
				header, err := regexp.Compile(ftdetect["header"].(string))
				if err != nil {
					return r, err
				}

				r[1] = header
			}
			loaded++
		}

		if loaded >= 2 {
			break
		}
	}

	if loaded == 0 {
		return r, errors.New("no detect regexes found")
	}

	return r, err
}

// ParseFile func
func ParseFile(input []byte) (f *File, err error) {
	// This is just so if we have an error, we can exit cleanly and return the parse error to the user
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("pkg: %v", r)
			}
		}
	}()

	var rules map[string]interface{}
	if err = yaml.Unmarshal(input, &rules); err != nil {
		return nil, err
	}
	f = new(File)
	f.yamlSrc = rules
	for k, v := range rules {
		if k == "filetype" {
			filetype := v.(string)

			f.FileType = filetype
			break
		}
	}

	return f, err
}

// ParseDef parses an input syntax file into a highlight Def
func ParseDef(f *File, header *Header) (s *Def, err error) {
	// This is just so if we have an error, we can exit cleanly and return the parse error to the user
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("pkg: %v", r)
			}
		}
	}()

	rules := f.yamlSrc

	s = new(Def)
	s.Header = header

	for k, v := range rules {
		if k == "rules" {
			inputRules := v.([]interface{})

			rules, err := parseRules(inputRules, nil)
			if err != nil {
				return nil, err
			}

			s.rules = rules
		}
	}
	return s, err
}

func parseRules(input []interface{}, curRegion *region) (ru *rules, err error) {
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("pkg: %v", r)
			}
		}
	}()
	ru = new(rules)

	for _, v := range input {
		rule := v.(map[string]interface{})
		for k, val := range rule {
			group := k

			switch object := val.(type) {
			case string:
				// Pattern
				r, err := regexp.Compile(object)
				if err != nil {
					return nil, err
				}

				groupStr := group
				if _, ok := Groups[groupStr]; !ok {
					numGroups++
					Groups[groupStr] = numGroups
				}
				groupNum := Groups[groupStr]
				ru.patterns = append(ru.patterns, &pattern{groupNum, r})
			default:
				return nil, fmt.Errorf("bad type %T", object)
			}
		}
	}

	return ru, nil
}
