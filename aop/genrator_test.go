package aop

import (
	"fmt"
	"github.com/gomelon/meta"
	"github.com/gomelon/meta-templates/aop/templates"
	"os"
	"testing"
)

func TestTemplateGen(t *testing.T) {

	workdir, _ := os.Getwd()
	path := workdir + "/testdata"
	generator, err := meta.NewTemplateGenerator(path, templates.Interface, meta.WithMetas(AllMetas()))
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	//err = generator.Print()
	err = generator.Generate()
	//err = generator.Generate()
	if err != nil {
		fmt.Println(err.Error())
	}
}
