package gjson

import (
	"strings"
	"testing"
)

var mediumJson = `{
 "person": {
    "id": "d50887ca-a6ce-4e59-b89f-14f0b5d03b03",
    "name": {
      "fullName": "Leonid Bugaev",
      "givenName": "Leonid",
      "familyName": "Bugaev"
    },
	"schools": [
		"nuaa", "mit"
	],
    "email": "leonsbox@gmail.com",
    "gender": "male",
    "working": true,
	"resigned": false,
    "location": "Saint Petersburg, Saint Petersburg, RU",
    "geo": {
      "city": "Saint Petersburg",
      "state": "Saint Petersburg",
      "country": "Russia",
      "lat": 59.9342802,
      "lng": 30.3350986
    },
    "bio": "Senior engineer at Granify.com",
    "site": "http://flickfaver.com",
    "avatar": "https://d1ts43dypk8bqh.cloudfront.net/v1/avatars/d50887ca-a6ce-4e59-b89f-14f0b5d03b03",
    "employment": {
      "name": "www.latera.ru",
      "name": "Software Engineer",
      "domain": "gmail.com"
    },
    "facebook": {
      "handle": "leonid.bugaev"
    },
    "github": {
      "handle": "buger",
      "id": 14009,
      "avatar": "https://avatars.githubusercontent.com/u/14009?v=3",
      "company": "Granify",
      "blog": "http://leonsbox.com",
      "followers": 95,
      "following": 10
    },
    "twitter": {
      "handle": "flickfaver",
      "id": 77004410,
      "bio": null,
      "followers": 2,
      "following": 1,
      "statuses": 5,
      "favorites": 0,
      "location": "",
      "site": "http://flickfaver.com",
      "avatar": null
    },
    "linkedin": {
      "handle": "in/leonidbugaev"
    },
    "googleplus": {
      "handle": null
    },
    "angellist": {
      "handle": "leonid-bugaev",
      "id": 61541,
      "bio": "Senior engineer at Granify.com",
      "blog": "http://buger.github.com",
      "site": "http://buger.github.com",
      "followers": 41,
      "avatar": "https://d1qb2nb5cznatu.cloudfront.net/users/61541-medium_jpg?1405474390"
    },
    "klout": {
      "handle": null,
      "score": null
    },
    "foursquare": {
      "handle": null
    },
    "aboutme": {
      "handle": "leonid.bugaev",
      "bio": null,
      "avatar": null
    },
    "gravatar": {
      "handle": "buger",
      "urls": [
			"https://github.com",
			"https://www.google.com"
      ],
      "avatar": "http://1.gravatar.com/avatar/f7c8edd577d13b8930d5522f28123510",
      "avatars": [
        {
          "url": "http://1.gravatar.com/avatar/f7c8edd577d13b8930d5522f28123510",
          "type": "thumbnail"
        }
      ]
    },
    "fuzzy": false
  },
  "company": null
}`

func TestParseMany(t *testing.T) {
	var root PathNode
	var tests = []struct {
		path   string
		expect interface{}
	}{
		{"person.id", "d50887ca-a6ce-4e59-b89f-14f0b5d03b03"},
	}

	for _, test := range tests {
		root.Add(strings.Split(test.path, "."), test.path)
	}
	ParseMany([]byte(mediumJson), &root, func(node *PathNode, value Value) {
		t.Logf("%s %s", node.FullPath, value.Str)
	})
}

var root PathNode
var tests = []struct {
	path   string
	expect interface{}
}{
	{"person.id", "d50887ca-a6ce-4e59-b89f-14f0b5d03b03"},
}

func init() {
	for _, test := range tests {
		root.Add(strings.Split(test.path, "."), test.path)
	}
}

func BenchmarkParseMany(b *testing.B) {

	b.ReportAllocs()

	var buf = []byte(mediumJson)
	for i := 0; i < b.N; i++ {
		ParseMany(buf, &root, func(node *PathNode, value Value) {
			//b.Logf("%s %s", node.FullPath, value.Str)
		})
	}
}
