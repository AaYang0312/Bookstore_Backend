package cache

import "testing"

func TestCollectionCacheKeysIncludeVersion(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{name: "hot books", got: hotBooksCacheKey(12, 10), want: "book:hot:v12:10"},
		{name: "new books", got: newBooksCacheKey(12, 8), want: "book:new:v12:8"},
		{name: "book list", got: bookListCacheKey(12, 2, 20), want: "book:list:v12:2:20"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.got != test.want {
				t.Fatalf("缓存 Key 不正确：得到 %q，期望 %q", test.got, test.want)
			}
		})
	}
}
