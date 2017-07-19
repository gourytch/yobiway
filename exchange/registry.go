package exchange

import "fmt"

var Registry map[string]Exchange = make(map[string]Exchange)

func Register(xcg Exchange) error {
	name := xcg.GetName()
	if _, ok := Registry[name]; ok {
		return fmt.Errorf("exchange already registered: `%s`", name)
	}
	Registry[name] = xcg
	return nil
}


