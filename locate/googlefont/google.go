package googlefont

import (
	"github.com/npillmayer/fontloading"
	"github.com/npillmayer/fontloading/locate"
	"github.com/npillmayer/schuko"
	"github.com/npillmayer/schuko/schukonf/testconfig"
	"github.com/npillmayer/schuko/tracing"
)

// tracer writes to trace with key 'tyse.font'
func tracer() tracing.Trace {
	return tracing.Select("tyse.font")
}

func Find(conf schuko.Configuration) locate.FontLocator {
	return func(descr fontloading.Descriptor) (fontloading.ScalableFont, error) {
		pattern := descr.Pattern
		style := descr.Style
		weight := descr.Weight
		return FindGoogleFont(conf, pattern, style, weight)
	}
}

func SimpleConfig(appkey string) schuko.Configuration {
	conf := make(testconfig.Conf)
	conf.Set("app-key", appkey)
	return conf
}
