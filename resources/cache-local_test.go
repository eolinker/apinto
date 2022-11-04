package resources

import (
	"fmt"
	"github.com/coocood/freecache"
)

func ExampleFreeCache() {
	key := []byte("test")
	value := []byte("value")

	client := freecache.NewCache(0)
	set, err := client.GetOrSet(key, value, 100)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("value1=%s\n", string(set))
	set, err = client.GetOrSet(key, value, 100)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("value=%s\n", string(set))
	// output:
	//value1=
	//value=value

}
