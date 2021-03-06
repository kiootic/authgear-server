package model

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestParseUserAgent(t *testing.T) {
	Convey("ParseUserAgent", t, func() {
		Convey("should parse browser UA correctly", func() {
			ua := ParseUserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/75.0.3770.142 Safari/537.36")
			So(ua, ShouldResemble, UserAgent{
				Raw:         "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/75.0.3770.142 Safari/537.36",
				Name:        "Chrome",
				Version:     "75.0.3770",
				OS:          "Mac OS X",
				OSVersion:   "10.14.5",
				DeviceModel: "",
			})
		})
		Convey("should parse Skygear SDK UA correctly", func() {
			ua := ParseUserAgent("io.skygear.test/1.0.1 (Skygear; iPhone11,8; iOS 12.0) SKYKit/2.0.1")
			So(ua, ShouldResemble, UserAgent{
				Raw:         "io.skygear.test/1.0.1 (Skygear; iPhone11,8; iOS 12.0) SKYKit/2.0.1",
				Name:        "io.skygear.test",
				Version:     "1.0.1",
				OS:          "iOS",
				OSVersion:   "12.0",
				DeviceModel: "Apple iPhone11,8",
			})

			ua = ParseUserAgent("io.skygear.test/1.3.0 (Skygear; Samsung GT-S5830L; Android 9.0) io.skygear.skygear/2.2.0")
			So(ua, ShouldResemble, UserAgent{
				Raw:         "io.skygear.test/1.3.0 (Skygear; Samsung GT-S5830L; Android 9.0) io.skygear.skygear/2.2.0",
				Name:        "io.skygear.test",
				Version:     "1.3.0",
				OS:          "Android",
				OSVersion:   "9.0",
				DeviceModel: "Samsung GT-S5830L",
			})
		})
	})
}
