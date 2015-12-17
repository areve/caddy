package setup

import (
	//~ "io/ioutil"
	"fmt"
	//~ "os"
	//~ "path/filepath"
	//~ "strconv"
	"testing"
	//~ "time"

	//~ "github.com/mholt/caddy/server"
)


func TestHide(t *testing.T) {

	//~ tempDirPath, err := getTempDirPath()
	//~ if err != nil {
		//~ t.Fatalf("BeforeTest: Failed to find an existing directory for testing! Error was: %v", err)
	//~ }
	//~ nonExistantDirPath := filepath.Join(tempDirPath, strconv.Itoa(int(time.Now().UnixNano())))

	//~ tempTemplate, err := ioutil.TempFile(".", "tempTemplate")
	//~ if err != nil {
		//~ t.Fatalf("BeforeTest: Failed to create a temporary file in the working directory! Error was: %v", err)
	//~ }
	//~ defer os.Remove(tempTemplate.Name())

	//~ tempTemplatePath := filepath.Join(".", tempTemplate.Name())

	for i, test := range []struct {
		input             string
		expectedHideConfig string
		shouldErr         bool
	}{
		// test case #0 missing parameters
		{"hide", "", true},

		// test case #1 multiple names
		{"hide {\n name 1 2 3 \n}", "{<nil> [{false [] [] [1 2 3] []}]}", false},

		// test case #2 multiple directives, name and suffix
		{"hide {\n prefix 1 2 3 \n}\nhide {\n suffix 4  \n}\n", "{<nil> [{false [1 2 3] [] [] []} {false [] [4] [] []}]}", false},

		// test case #3 one match case name
		{"hide matchcase { name 1\n}", "{<nil> [{true [] [] [1] []}]}", false},

		// test case #4 paths
		{"hide {\n path / /foo \"/foo/bar baz\" \n}", "{<nil> [{false [] [] [] [/ /foo /foo/bar baz]}]}", false},

		// test case #5 multiple mixed match case directives, name and suffix
		{"hide {\n prefix 1 2 3 \n}\nhide matchcase {\n suffix 4  \n}", "{<nil> [{false [1 2 3] [] [] []} {true [] [4] [] []}]}", false},

		// test case #6 empty block
		{"hide {}", "", true},

		// test case #7 matchcase empty block
		{"hide matchcase {}", "", true},

		// test case #8 name no with no names
		{"hide {\n name\n }", "", true},

		// test case #9 wrong options
		{"hide foobar {\n name\n }", "", true},

		// test case #10 wrong options
		{"hide {\n foobar 1\n }", "", true},

	} {
		c := NewTestController(test.input)
		_, err := Hide(c)
		if err != nil && !test.shouldErr {
			t.Errorf("Test case #%d recieved an error of %v", i, err)
		} else if test.shouldErr {
			continue
		}

		result := fmt.Sprint(c.Hide)
		if (test.expectedHideConfig != result) {
				t.Errorf("Test case #%d expected a HideConfig of %v, but got %v", i, test.expectedHideConfig, result)
		}
	}
}
